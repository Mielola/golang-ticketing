package dashboard

import (
	"fmt"
	"log"
	"my-gin-project/src/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "root:@tcp(localhost:3306)/commandcenter?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	fmt.Println("Dashboard: Connected to MySQL")
}

func init() {
	InitDB()
}

func GetDashboard(c *gin.Context) {
	// --------------------------------------------
	// @ GET Tickets Summary
	// --------------------------------------------
	var tickets types.TicketsResponse
	if err := DB.Table("tickets").
		Select(`
			COUNT(CASE WHEN status = 'New' THEN 1 END) as open_tickets,
			COUNT(CASE WHEN status = 'On Progress' THEN 1 END) as pending_tickets,
			COUNT(CASE WHEN status = 'Resolved' THEN 1 END) as resolved_tickets,
			COUNT("*") as total_tickets
		`).
		Scan(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get dashboard data", "error": err.Error()})
		return
	}

	// --------------------------------------------
	// @ GET Recent Tickets
	// --------------------------------------------
	type RecentTicket struct {
		Category      int        `json:"category"`
		CreatedAt     time.Time  `json:"created_at"`
		DetailKendala string     `json:"detail_kendala"`
		DueDate       *time.Time `json:"due_date"`
		HariMasuk     *time.Time `json:"hari_masuk"`
		WaktuMasuk    string     `json:"waktu_masuk"`
		Subject       string     `json:"subject"`
		UserEmail     string     `json:"user_email"`
	}

	var recentTickets []RecentTicket
	if err := DB.Table("tickets").
		Select("category, created_at, detail_kendala, due_date, hari_masuk, waktu_masuk, subject, user_email").
		Order("created_at DESC").
		Limit(10).
		Scan(&recentTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get recent tickets", "error": err.Error()})
		return
	}

	// Format `hari_masuk` sebelum dikirim ke response
	type FormattedRecentTicket struct {
		Category      int     `json:"category"`
		CreatedAt     string  `json:"created_at"`
		DetailKendala string  `json:"detail_kendala"`
		DueDate       *string `json:"due_date,omitempty"`
		HariMasuk     *string `json:"hari_masuk,omitempty"`
		WaktuMasuk    string  `json:"waktu_masuk"`
		Subject       string  `json:"subject"`
		UserEmail     string  `json:"user_email"`
	}

	var formattedTickets []FormattedRecentTicket
	for _, ticket := range recentTickets {
		var formattedHariMasuk, formattedDueDate *string

		if ticket.HariMasuk != nil {
			formatted := ticket.HariMasuk.Format("2006-01-02")
			formattedHariMasuk = &formatted
		}

		if ticket.DueDate != nil {
			formatted := ticket.DueDate.Format("2006-01-02")
			formattedDueDate = &formatted
		}

		formattedTickets = append(formattedTickets, FormattedRecentTicket{
			Category:      ticket.Category,
			CreatedAt:     ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			DetailKendala: ticket.DetailKendala,
			DueDate:       formattedDueDate,
			HariMasuk:     formattedHariMasuk,
			WaktuMasuk:    ticket.WaktuMasuk,
			Subject:       ticket.Subject,
			UserEmail:     ticket.UserEmail,
		})
	}

	// --------------------------------------------
	// @ GET User Logs
	// --------------------------------------------
	var rawLogs []struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
	}

	var users []types.UserResponseWithoutToken
	if err := DB.Table("user_logs").Select("*").Order("login_time DESC").Scan(&rawLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := DB.Table("users").
		Select("users.*").
		Joins("JOIN user_logs ON user_logs.user_email = users.email").
		Order("user_logs.login_time DESC").
		Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// --------------------------------------------
	// @ Format User Data
	// --------------------------------------------
	baseURL := "http://localhost:8080/storage/images/"
	userMap := make(map[string]types.UserResponseWithoutToken)
	for _, user := range users {
		if user.Avatar != nil {
			avatarURL := baseURL + *user.Avatar
			user.Avatar = &avatarURL
		}
		userMap[user.Email] = user
	}

	var userResponse []types.UserLogResponse
	for _, log := range rawLogs {
		userData, exists := userMap[log.UserEmail]
		if !exists {
			continue
		}

		userResponse = append(userResponse, types.UserLogResponse{
			UserResponseWithoutToken: userData,
			LoginDate:                log.LoginTime.Format("2006-01-02"),
			LoginTime:                log.LoginTime.Format("15:04:05"),
		})
	}

	// --------------------------------------------
	// @ Build Final Response
	// --------------------------------------------
	var recentTicketsMap []map[string]interface{}
	for _, ticket := range formattedTickets {
		ticketMap := map[string]interface{}{
			"category":       ticket.Category,
			"created_at":     ticket.CreatedAt,
			"detail_kendala": ticket.DetailKendala,
			"due_date":       ticket.DueDate,
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
			Summary:       tickets,
			RecentTickets: recentTicketsMap,
			UserLogs:      userResponse,
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

func SetDB(db *gorm.DB) {
	DB = db
}
