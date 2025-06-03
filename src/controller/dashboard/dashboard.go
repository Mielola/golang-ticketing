package dashboard

import (
	"fmt"
	"my-gin-project/src/database"
	"my-gin-project/src/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboard(c *gin.Context) {
	DB := database.GetDB()

	// --------------------------------------------
	// @ GET Tickets Summary
	// --------------------------------------------
	var tickets types.TicketsResponse
	if err := DB.Table("tickets").
		Select(`
			COUNT(CASE WHEN status = 'New' THEN 1 END) as open_tickets,
			COUNT(CASE WHEN status = 'On Progress' THEN 1 END) as pending_tickets,
			COUNT(CASE WHEN status = 'Resolved' THEN 1 END) as resolved_tickets,
			COUNT(CASE WHEN priority = 'Critical' THEN 1 END) as critical_tickets,
			COUNT("*") as total_tickets
		`).
		Scan(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get dashboard data", "error": err.Error()})
		return
	}

	// --------------------------------------------
	// @ GET Tickets Charts By Products
	// --------------------------------------------
	var ticketsCharts []struct {
		Name          string `json:"name"`
		TotalTickets  string `json:"total_tickets"`
		TotalCategory string `json:"total_category"`
	}

	if err := DB.Table("products").
		Select(`
        products.name,
        COUNT(DISTINCT tickets.id) as total_tickets,
        COUNT(DISTINCT category.id) as total_category
    `).
		Joins("LEFT JOIN tickets ON products.name = tickets.products_name").
		Joins("LEFT JOIN category ON products.id = category.products_id").
		Group("products.name").
		Scan(&ticketsCharts).Error; err != nil {

		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Chart :" + err.Error(),
		})
		return
	}

	// --------------------------------------------
	// @ GET Recent Tickets
	// --------------------------------------------
	type RecentTicket struct {
		TrackingID    string     `json:"tracking_id"`
		Category      string     `json:"category_name"`
		CreatedAt     time.Time  `json:"created_at"`
		DetailKendala string     `json:"detail_kendala"`
		HariMasuk     *time.Time `json:"hari_masuk"`
		WaktuMasuk    string     `json:"waktu_masuk"`
		Subject       string     `json:"subject"`
		UserEmail     string     `json:"user_email"`
	}

	var recentTickets []RecentTicket
	if err := DB.Table("tickets").
		Select("tracking_id, created_at, detail_kendala, hari_masuk, waktu_masuk, subject, user_email").
		Order("created_at DESC").
		Limit(5).
		Scan(&recentTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get recent tickets", "error": err.Error()})
		return
	}

	// Format response
	type FormattedRecentTicket struct {
		TrackingID    string  `json:"tracking_id"`
		CreatedAt     string  `json:"created_at"`
		DetailKendala string  `json:"detail_kendala"`
		HariMasuk     *string `json:"hari_masuk,omitempty"`
		WaktuMasuk    string  `json:"waktu_masuk"`
		Subject       string  `json:"subject"`
		UserEmail     string  `json:"user_email"`
	}

	var formattedTickets []FormattedRecentTicket
	for _, ticket := range recentTickets {
		var formattedHariMasuk *string

		if ticket.HariMasuk != nil {
			formatted := ticket.HariMasuk.Format("2006-01-02")
			formattedHariMasuk = &formatted
		}

		formattedTickets = append(formattedTickets, FormattedRecentTicket{
			TrackingID:    ticket.TrackingID,
			CreatedAt:     ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			DetailKendala: ticket.DetailKendala,
			HariMasuk:     formattedHariMasuk,
			WaktuMasuk:    ticket.WaktuMasuk,
			Subject:       ticket.Subject,
			UserEmail:     ticket.UserEmail,
		})
	}

	// --------------------------------------------
	// @ GET User Logs
	// --------------------------------------------
	var userLogs []types.UserLogResponse

	if err := DB.Table("user_logs").
		Select("user_logs.login_time, users.avatar, users.email, users.name, users.role, users.avatar, users.status, user_logs.shift_name").
		Joins("JOIN users ON user_logs.user_email = users.email").
		Order("user_logs.login_time DESC").
		Find(&userLogs).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data User Logs",
		})
		return
	}

	formattedUserLogs := make([]map[string]interface{}, 0)
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	for _, user := range userLogs {
		formattedUserLogs = append(formattedUserLogs, map[string]interface{}{
			"email":      user.Email,
			"name":       user.Name,
			"role":       user.Role,
			"avatar":     baseURL + *user.Avatar,
			"shift_name": user.ShiftName,
			"status":     user.Status,
			"login_date": user.LoginTime.Format("2006-01-02"),
			"login_time": user.LoginTime.Format("15:04:05"),
		})
	}

	// --------------------------------------------
	// @ Build Final Response
	// --------------------------------------------
	var recentTicketsMap []map[string]interface{}
	for _, ticket := range formattedTickets {
		ticketMap := map[string]interface{}{
			"tracking_id":    ticket.TrackingID,
			"created_at":     ticket.CreatedAt,
			"detail_kendala": ticket.DetailKendala,
			"hari_masuk":     ticket.HariMasuk,
			"waktu_masuk":    ticket.WaktuMasuk,
			"subject":        ticket.Subject,
			"user_email":     ticket.UserEmail,
		}
		recentTicketsMap = append(recentTicketsMap, ticketMap)
	}

	dashboardData := types.DashboardResponse{
		Success: true,
		Message: "Dashboard data retrieved successfully",
		Data: types.DataContent{
			Summary:              tickets,
			RecentTickets:        recentTicketsMap,
			UserLogs:             formattedUserLogs,
			ChartTicketsProudcts: ticketsCharts,
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

// @ GET
func GetForm(c *gin.Context) {
	DB := database.GetDB()
	var Product struct {
		Name string `json:"name"`
	}

	var categories []struct {
		ID           uint32 `json:"id"`
		CategoryName string `json:"category_name"`
	}

	if err := c.ShouldBindJSON(&Product); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Body must be JSON",
			Data:    nil,
		})
		return
	}

	if Product.Name == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Product name is required",
			Data:    nil,
		})
		return
	}

	if err := DB.Table("category").
		Select("id, category_name").
		Where("products_id = (SELECT id FROM products WHERE name = ?)", Product.Name).
		Pluck("category_name", &categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Data retrieved successfully",
		Data:    categories,
	})
}
