package ticket

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func toSolvedTime(duration time.Duration) *string {
	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60
	solved := fmt.Sprintf("%d hours %d minutes %d seconds", h, m, s)
	return &solved
}

func CheckTicketsDeadline(c *gin.Context) {
	DB := database.GetDB()
	var tickets []types.TicketsResponseAll

	if err := DB.Table("tickets").Select("*").Order("priority DESC").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	for _, ticket := range tickets {
		var ticketCreator types.TicketsCreator

		if err := DB.Table("users").
			Select("email, name, avatar").
			Where("email = ?", ticket.UserEmail).
			Scan(&ticketCreator).Error; err != nil {
			return
		}

		var lastReply struct {
			UserEmail string    `json:"user_email"`
			NewStatus string    `json:"new_status"`
			UpdatedAt time.Time `json:"update_at"`
		}

		if err := DB.Table("user_tickets").
			Select("user_email, new_status, update_at").
			Where("tickets_id = ?", ticket.TrackingID).
			Order("update_at DESC").
			Limit(1).
			Scan(&lastReply).Error; err != nil {
			lastReply = struct {
				UserEmail string    `json:"user_email"`
				NewStatus string    `json:"new_status"`
				UpdatedAt time.Time `json:"update_at"`
			}{}
		}

		if lastReply.UserEmail != "" {
			var replierInfo struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			}

			if err := DB.Table("users").
				Select("email, name, avatar").
				Where("email = ?", lastReply.UserEmail).
				Scan(&replierInfo).Error; err == nil {
			}
		}

		if ticket.Status != "Resolved" && time.Since(ticket.CreatedAt) > 24*time.Hour && ticket.Priority != "Critical" {

			// Update the ticket's priority to "High"
			if err := DB.Table("tickets").
				Where("tracking_id = ?", ticket.TrackingID).
				Update("priority", "Critical").Error; err != nil {
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{
					Success: false,
					Message: err.Error(),
					Data:    nil,
				})
				return
			}

			// Save to History
			saveHistory := struct {
				UserEmail string `json:"user_email"`
				NewStatus string `json:"new_status"`
				TicketsID string `json:"ticket_id"`
				Priority  string `json:"priority"`
				Details   string `json:"details"`
			}{
				UserEmail: ticketCreator.Email,
				NewStatus: ticket.Status,
				TicketsID: ticket.TrackingID,
				Priority:  "Critical",
				Details:   "Sistem Otmotatis Mengupdate Prioritas Ticket",
			}

			if err := DB.Table("user_tickets").Create(&saveHistory).Error; err != nil {
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{
					Success: false,
					Message: err.Error(),
				})
				return
			}

		}
	}
}

// @GET
func GetAllTickets(c *gin.Context) {
	DB := database.GetDB()
	var tickets []types.TicketsResponseAll
	if err := DB.Table("tickets").
		Select("tickets.*, category.category_name").
		Joins("LEFT JOIN category ON tickets.category_id = category.id").
		Order("created_at ASC").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	emailSet := make(map[string]bool)

	for _, ticket := range tickets {
		emailSet[ticket.UserEmail] = true
	}

	var emails []string
	for email := range emailSet {
		emails = append(emails, email)
	}

	var users []types.TicketsCreator
	if err := DB.Table("users").
		Select("email, name, avatar").
		Where("email IN ?", emails).
		Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to fetch users",
			Data:    nil,
		})
		return
	}

	userMap := make(map[string]types.TicketsCreator)
	for _, user := range users {
		if user.Avatar != "" {
			user.Avatar = baseURL + user.Avatar
		}
		userMap[user.Email] = user
	}

	// 1. Ambil semua tracking_id dari tickets
	var trackingIDs []string
	for _, t := range tickets {
		trackingIDs = append(trackingIDs, t.TrackingID)
	}

	// 2. Ambil last reply per tiket (latest update_at)
	var lastRepliesRaw []struct {
		TicketsID string    `json:"tickets_id"`
		UserEmail string    `json:"user_email"`
		NewStatus string    `json:"new_status"`
		UpdatedAt time.Time `json:"update_at"`
	}
	DB.Raw(`
	SELECT ut.*
	FROM user_tickets ut
	INNER JOIN (
		SELECT tickets_id, MAX(update_at) AS max_update
		FROM user_tickets
		WHERE tickets_id IN ?
		GROUP BY tickets_id
	) latest ON ut.tickets_id = latest.tickets_id AND ut.update_at = latest.max_update`, trackingIDs).Scan(&lastRepliesRaw)

	// 3. Buat map tickets_id -> reply
	lastReplyMap := make(map[string]struct {
		UserEmail string
		NewStatus string
		UpdatedAt time.Time
	})
	lastReplyEmails := make(map[string]bool)
	for _, r := range lastRepliesRaw {
		lastReplyMap[r.TicketsID] = struct {
			UserEmail string
			NewStatus string
			UpdatedAt time.Time
		}{r.UserEmail, r.NewStatus, r.UpdatedAt}
		if r.UserEmail != "" {
			lastReplyEmails[r.UserEmail] = true
		}
	}

	// 4. Ambil data user yang jadi last replier
	var lastRepliers []types.TicketsCreator
	if len(lastReplyEmails) > 0 {
		var emails []string
		for e := range lastReplyEmails {
			emails = append(emails, e)
		}

		DB.Table("users").
			Select("email, name, avatar").
			Where("email IN ?", emails).
			Scan(&lastRepliers)
	}

	// 5. Buat map email -> user (last replier)
	lastReplierMap := make(map[string]types.TicketsCreator)
	for _, user := range lastRepliers {
		if user.Avatar != "" {
			user.Avatar = baseURL + user.Avatar
		}
		lastReplierMap[user.Email] = user
	}

	// 6. Susun hasil akhir
	var formattedTickets []map[string]interface{}
	for _, ticket := range tickets {
		ticketCreator := userMap[ticket.UserEmail]

		var lastReplier *types.TicketsCreator
		if lr, ok := lastReplyMap[ticket.TrackingID]; ok {
			if user, ok := lastReplierMap[lr.UserEmail]; ok {
				lastReplier = &user
			}
		}

		formattedTickets = append(formattedTickets, map[string]interface{}{
			"id":             ticket.ID,
			"tracking_id":    ticket.TrackingID,
			"products_name":  ticket.ProductsName,
			"hari_masuk":     ticket.HariMasuk.Format("2006-01-02"),
			"waktu_masuk":    ticket.WaktuMasuk,
			"solved_time":    ticket.SolvedTime,
			"user":           ticketCreator,
			"last_replier":   lastReplier,
			"category":       ticket.CategoryName,
			"priority":       ticket.Priority,
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"no_whatsapp":    ticket.NoWhatsapp,
			"detail_kendala": ticket.DetailKendala,
			"pic":            ticket.PIC,
			"created_date":   ticket.CreatedAt.Format("2006-01-02"),
			"created_time":   ticket.CreatedAt.Format("15:04:05"),
			"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data:    formattedTickets,
	})
}

// @GET
func GetTicketsLogs(c *gin.Context) {
	var ticketLogs []types.TicketsLogsRaw
	DB := database.GetDB()

	if err := DB.Table("user_tickets").
		Select(`
			user_tickets.*, 
			users.email as user_email, 
			users.name as user_name, 
			users.avatar as user_avatar
		`).
		Joins("LEFT JOIN users ON user_tickets.user_email = users.email").
		Order("user_tickets.update_at DESC").
		Find(&ticketLogs).Error; err != nil {

		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Ubah ke format yang diinginkan
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	formattedLogs := make([]types.TicketsLogs, 0, len(ticketLogs))
	for _, log := range ticketLogs {

		if log.UserAvatar != "" {
			log.UserAvatar = baseURL + log.UserAvatar
		}

		formattedLogs = append(formattedLogs, types.TicketsLogs{
			ID:        log.ID,
			TicketsId: log.TicketsId,
			NewStatus: log.NewStatus,
			Priority:  log.Priority,
			Details:   log.Details,
			UpdateAt:  log.UpdateAt,
			UpdateAtString: func() string {
				if log.UpdateAt != nil {
					return log.UpdateAt.Format("2006-01-02 15:04:05")
				}
				return ""
			}(),
			User: types.TicketsCreator{
				Email:  log.UserEmail,
				Name:   log.UserName,
				Avatar: log.UserAvatar,
			},
		})
	}

	// Kirim response
	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets logs retrieved successfully",
		Data:    formattedLogs,
	})
}

// @GET
func GetTicketsByDateRange(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	DB := database.GetDB()

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Start date and end date are required",
			Data:    nil,
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid StartDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid EndDate format, use YYYY-MM-DD"})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "StartDate cannot be after EndDate"})
		return
	}

	tickets := make([]types.TicketsResponseAll, 0)
	if err := DB.Table("tickets").
		Select("tickets.*, category.category_name").
		Where("DATE(created_at) BETWEEN ? AND ?", startDate.String(), endDate.String()).
		Order("created_at DESC").
		Joins("LEFT JOIN category ON tickets.category_id = category.id").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	formattedTickets := make([]map[string]interface{}, 0)
	for _, ticket := range tickets {
		var ticketCreator types.TicketsCreator

		if err := DB.Table("users").
			Select("email, name, avatar").
			Where("email = ?", ticket.UserEmail).
			Scan(&ticketCreator).Error; err != nil {
			return
		}

		if ticketCreator.Avatar != "" {
			ticketCreator.Avatar = baseURL + ticketCreator.Avatar
		}

		var lastReply struct {
			UserEmail string    `json:"user_email"`
			NewStatus string    `json:"new_status"`
			UpdatedAt time.Time `json:"update_at"`
		}

		if err := DB.Table("user_tickets").
			Select("user_email, new_status, update_at").
			Where("tickets_id = ?", ticket.TrackingID).
			Order("update_at DESC").
			Limit(1).
			Scan(&lastReply).Error; err != nil {
			lastReply = struct {
				UserEmail string    `json:"user_email"`
				NewStatus string    `json:"new_status"`
				UpdatedAt time.Time `json:"update_at"`
			}{}
		}

		// Get last replier's information
		var lastReplier *struct {
			Email  string `json:"email"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}

		if lastReply.UserEmail != "" {
			var replierInfo struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			}

			if err := DB.Table("users").
				Select("email, name, avatar").
				Where("email = ?", lastReply.UserEmail).
				Scan(&replierInfo).Error; err == nil {

				if replierInfo.Avatar != "" {
					replierInfo.Avatar = baseURL + replierInfo.Avatar
				}
				lastReplier = &replierInfo
			}
		}

		formattedTickets = append(formattedTickets, map[string]interface{}{
			"id":             ticket.ID,
			"tracking_id":    ticket.TrackingID,
			"products_name":  ticket.ProductsName,
			"hari_masuk":     ticket.HariMasuk.Format("2006-01-02"),
			"waktu_masuk":    ticket.WaktuMasuk,
			"solved_time":    ticket.SolvedTime,
			"user":           ticketCreator,
			"last_replier":   lastReplier,
			"category":       ticket.CategoryName,
			"priority":       ticket.Priority,
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"no_whatsapp":    ticket.NoWhatsapp,
			"detail_kendala": ticket.DetailKendala,
			"pic":            ticket.PIC,
			"created_date":   ticket.CreatedAt.Format("2006-01-02"),
			"created_time":   ticket.CreatedAt.Format("15:04:05"),
			"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		})
	}

	// Kirim respons
	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: startDateStr + " to " + endDateStr,
		Data:    formattedTickets,
	})
}

// @GET
func GetTicketsByCategory(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		Category string `json:"category"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if input.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	var tickets []types.Tickets
	if err := DB.Select("*").Joins("JOIN category ON tickets.category = category.id").Where("category.category_name = ?", input.Category).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tickets retrieved successfully", "data": tickets})
}

// @GET
func GetTicketsByStatus(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if input.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	var tickets []types.Tickets
	if err := DB.Table("tickets").Where("status = ?", input.Status).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tickets retrieved successfully", "data": tickets})
}

// @GET
func GetTicketsByPriority(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		Priority string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if input.Priority == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Priority is required"})
		return
	}

	var tickets []types.Tickets
	if err := DB.Table("tickets").Where("priority = ?", input.Priority).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tickets retrieved successfully", "data": tickets})
}

// @GET
func GenerateReport(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		ProductsName string `json:"products_name" binding:"required"`
		StartDate    string `json:"start_date" binding:"required"`
		EndDate      string `json:"end_date" binding:"required"`
		Status       string `json:"status" binding:"required"`
		StartTime    string `json:"start_time" binding:"required"`
		EndTime      string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Request must be in JSON format",
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid StartDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid EndDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if input.StartDate == "" || input.EndDate == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Start date, end date, status are required",
			Data:    nil,
		})
		return
	}

	var tickets []struct {
		TrackingID      string    `json:"tracking_id"`
		CreatedAt       time.Time `json:"created_at"`
		Subject         string    `json:"subject"`
		HariMasuk       time.Time `json:"hari_masuk"`
		WaktuMasuk      string    `json:"waktu_masuk"`
		CategoryName    string    `json:"category_name"`
		ResponDiberikan string    `json:"respon_diberikan"`
	}

	var chartPriority struct {
		Low    int `json:"low"`
		Medium int `json:"medium"`
		High   int `json:"high"`
	}

	type PriorityItem struct {
		Label string `json:"label"`
		Value int    `json:"value"`
	}

	var chartCategory []struct {
		CategoryName string `json:"category_name"`
		TotalTickets int    `json:"total_tickets"`
	}

	if err := DB.Table("tickets").
		Select("COUNT(CASE WHEN priority = 'Low' THEN 1 END) AS low, COUNT(CASE WHEN priority = 'Medium' THEN 1 END) AS medium, COUNT(CASE WHEN priority = 'High' THEN 1 END) AS high").
		Where("products_name = ?", input.ProductsName).
		Find(&chartPriority).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	if err := DB.Table("category").Select("category.category_name, COUNT(*) AS total_tickets").
		Joins("LEFT JOIN tickets ON category.id = tickets.category_id").
		Where("products_name = ?", input.ProductsName).Group("category_name").
		Find(&chartCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	var startDateTime = input.StartDate + " " + input.StartTime
	var endDateTime = input.EndDate + " " + input.EndTime

	if input.Status == "all" || input.Status == "All" {
		if err := DB.Table("tickets").
			Select("tickets.tracking_id, tickets.created_at, tickets.subject, tickets.hari_masuk, tickets.waktu_masuk, category.category_name, tickets.respon_diberikan").
			Joins("LEFT JOIN category ON tickets.category_id = category.id").
			Where("tickets.created_at BETWEEN ? AND ? AND products_name = ?", startDateTime, endDateTime, input.ProductsName).
			Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
	} else {
		if err := DB.Table("tickets").
			Select("tickets.tracking_id, tickets.created_at, tickets.subject, tickets.hari_masuk, tickets.waktu_masuk, category.category_name, tickets.respon_diberikan").
			Joins("LEFT JOIN category ON tickets.category_id = category.id").
			Where("tickets.created_at BETWEEN ? AND ? AND status = ? AND products_name = ?", startDateTime, endDateTime, input.Status, input.ProductsName).
			Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
	}

	priorityItems := []PriorityItem{
		{Label: "Low", Value: chartPriority.Low},
		{Label: "Medium", Value: chartPriority.Medium},
		{Label: "High", Value: chartPriority.High},
	}

	categoryItems := make([]map[string]interface{}, 0)
	for _, category := range chartCategory {
		categoryItems = append(categoryItems, map[string]interface{}{
			"category_name": category.CategoryName,
			"total_tickets": category.TotalTickets,
		})
	}

	formattedTickets := make([]map[string]interface{}, 0)
	for _, ticket := range tickets {
		formattedTickets = append(formattedTickets, map[string]interface{}{
			"tracking_id":   ticket.TrackingID,
			"created_at":    ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			"subject":       ticket.Subject,
			"respon_admin":  ticket.ResponDiberikan,
			"hari_masuk":    ticket.HariMasuk.Format("2006-01-02"),
			"waktu_masuk":   ticket.WaktuMasuk,
			"category_name": ticket.CategoryName,
		})
	}

	type ResponseFormats struct {
		Success  bool                     `json:"success"`
		Message  string                   `json:"message"`
		Products string                   `json:"products_name"`
		Data     []map[string]interface{} `json:"data"`
		Chart    interface{}              `json:"chart"`
	}

	c.JSON(http.StatusOK, ResponseFormats{
		Success:  true,
		Message:  "Report generated successfully",
		Products: input.ProductsName,
		Data:     formattedTickets,
		Chart: gin.H{
			"ChartPriority": priorityItems,
			"ChartCategory": categoryItems,
		},
	})

}

// @POST
func AddTicket(c *gin.Context) {
	DB := database.GetDB()
	// Input structure with proper validation tags
	var inputJSON struct {
		HariMasuk       string `json:"hari_masuk" binding:"required"`
		HariRespon      string `json:"hari_respon" binding:"required"`
		WaktuMasuk      string `json:"waktu_masuk" binding:"required"`
		WaktuRespon     string `json:"waktu_respon" binding:"required"`
		CategoryId      uint64 `json:"category_id" binding:"required"`
		Subject         string `json:"subject" binding:"required"`
		PIC             string `json:"PIC"`
		DetailKendala   string `json:"detail_kendala" binding:"required"`
		ResponDiberikan string `json:"respon_diberikan" binding:"required"`
		NoWhatsapp      string `json:"no_whatsapp" binding:"required"`
		Priority        string `json:"priority" binding:"required"`
		ProductsName    string `json:"products_name" binding:"required"`
	}

	// Bind JSON to struct
	if err := c.ShouldBindJSON(&inputJSON); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Token is required",
			Data:    nil,
		})
		return
	}

	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := DB.Table("users").Where("users.token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Initialize ticket
	var ticket models.Ticket

	// Parse date fields
	hariMasuk, err := time.Parse("2006-01-02", inputJSON.HariMasuk)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hari_masuk format. Expected YYYY-MM-DD"})
		return
	}
	ticket.HariMasuk = hariMasuk

	// Assign remaining fields directly from the struct
	ticket.WaktuMasuk = inputJSON.WaktuMasuk
	hariRespon, err := time.Parse("2006-01-02", inputJSON.HariRespon)
	ticket.HariRespon = &hariRespon
	ticket.WaktuRespon = inputJSON.WaktuRespon
	ticket.UserName = user.Name
	ticket.UserEmail = user.Email
	ticket.CategoryId = inputJSON.CategoryId
	ticket.Priority = inputJSON.Priority
	ticket.Subject = inputJSON.Subject
	ticket.DetailKendala = inputJSON.DetailKendala
	ticket.PIC = inputJSON.PIC
	ticket.ResponDiberikan = inputJSON.ResponDiberikan
	ticket.NoWhatsapp = inputJSON.NoWhatsapp
	ticket.ProductsName = inputJSON.ProductsName

	// Generate tracking ID
	ticket.TrackingID = generateTrackingID(inputJSON.ProductsName)

	// Save ticket to database
	if err := DB.Table("tickets").Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Save history
	history := struct {
		UserEmail string `json:"user_email"`
		NewStatus string `json:"new_status"`
		TicketsID string `json:"ticket_id"`
		Priority  string `json:"priority"`
		Details   string `json:"details"`
	}{
		UserEmail: ticket.UserEmail,
		NewStatus: "New",
		TicketsID: ticket.TrackingID,
		Priority:  ticket.Priority,
		Details:   "Membuat Tiket Baru",
	}

	if err := DB.Table("user_tickets").Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ticket added successfully",
		"ticket":  ticket,
	})
}

func generateTrackingID(productName string) string {
	words := strings.Fields(productName)
	var prefix string
	for _, word := range words {
		prefix += strings.ToUpper(string(word[0]))
	}

	tanggal := time.Now().Format("060102")

	abjad := string('A' + byte(rand.Intn(26)))

	nomorAcak := fmt.Sprintf("%03d", rand.Intn(1000))
	trackingID := fmt.Sprintf("%s-%s%s-%s", prefix, tanggal[:3], abjad, nomorAcak)
	return trackingID
}

// @GET
func GetTicketByID(c *gin.Context) {
	DB := database.GetDB()
	trackingID := c.Param("tracking_id")

	// Ambil ticket beserta relasi user dan product
	var ticket models.Ticket
	if err := DB.Preload("User").Preload("Product").Preload("Category").
		Where("tracking_id = ?", trackingID).
		First(&ticket).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, types.ResponseFormat{
				Success: false,
				Message: "Ticket Not Found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Ambil log riwayat tiket
	var historyTickets []types.TicketsLogsRaw
	if err := DB.Table("user_tickets").
		Select(`
			user_tickets.*, 
			users.email as user_email, 
			users.name as user_name, 
			users.avatar as user_avatar
		`).
		Joins("LEFT JOIN users ON user_tickets.user_email = users.email").
		Where("user_tickets.tickets_id = ?", trackingID).
		Order("user_tickets.update_at DESC").
		Find(&historyTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Siapkan base URL untuk avatar
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	// Format logs
	var formattedLogs []types.TicketsLogs
	for _, log := range historyTickets {
		if log.UserAvatar != "" {
			log.UserAvatar = baseURL + log.UserAvatar
		}
		formattedLogs = append(formattedLogs, types.TicketsLogs{
			ID:        log.ID,
			TicketsId: log.TicketsId,
			NewStatus: log.NewStatus,
			Priority:  log.Priority,
			Details:   log.Details,
			UpdateAt:  log.UpdateAt,
			UpdateAtString: func() string {
				if log.UpdateAt != nil {
					return log.UpdateAt.Format("2006-01-02 15:04:05")
				}
				return ""
			}(),
			User: types.TicketsCreator{
				Email:  log.UserEmail,
				Name:   log.UserName,
				Avatar: log.UserAvatar,
			},
		})
	}

	// Ambil last replier dari user_tickets
	var lastReply struct {
		UserEmail string    `json:"user_email"`
		NewStatus string    `json:"new_status"`
		UpdatedAt time.Time `json:"update_at"`
	}
	DB.Table("user_tickets").
		Select("user_email, new_status, update_at").
		Where("tickets_id = ?", trackingID).
		Order("update_at DESC").
		Limit(1).
		Scan(&lastReply)

	var lastReplier *struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	if lastReply.UserEmail != "" {
		var replierInfo struct {
			Email  string `json:"email"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}
		if err := DB.Table("users").
			Select("email, name, avatar").
			Where("email = ?", lastReply.UserEmail).
			Scan(&replierInfo).Error; err == nil {

			if replierInfo.Avatar != "" {
				replierInfo.Avatar = baseURL + replierInfo.Avatar
			}
			lastReplier = &replierInfo
		}
	}

	// Format avatar ticket creator
	if ticket.User.Avatar != "" {
		ticket.User.Avatar = baseURL + ticket.User.Avatar
	}

	// Final response
	formattedTicket := map[string]interface{}{
		"id":            ticket.ID,
		"tracking_id":   ticket.TrackingID,
		"products_name": ticket.ProductsName,
		"hari_masuk":    ticket.HariMasuk.Format("2006-01-02"),
		"waktu_masuk":   ticket.WaktuMasuk,
		"solved_time":   ticket.SolvedTime,
		"user": types.TicketsCreator{
			Email:  ticket.User.Email,
			Name:   ticket.User.Name,
			Avatar: ticket.User.Avatar,
		},
		"last_replier":   lastReplier,
		"category_id":    ticket.CategoryId,
		"category":       ticket.Category.CategoryName,
		"priority":       ticket.Priority,
		"status":         ticket.Status,
		"subject":        ticket.Subject,
		"no_whatsapp":    ticket.NoWhatsapp,
		"detail_kendala": ticket.DetailKendala,
		"pic":            ticket.PIC,
		"created_date":   ticket.CreatedAt.Format("2006-01-02"),
		"created_time":   ticket.CreatedAt.Format("15:04:05"),
		"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		"history":        formattedLogs,
		"respon_admin":   ticket.ResponDiberikan,
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data:    formattedTicket,
	})
}

// @POST
func UpdateStatus(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Status string `json:"status"`
	}

	// @Bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user struct {
		Email string `json:"email"`
	}

	// @GET Token from Header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Token is Required",
		})
		return
	}

	// @GET User Email from Token
	if err := DB.Table("users").Select("email").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "User Not Found",
		})
		return
	}

	// @GET Ticket from Database
	var ticket struct {
		ID              uint      `gorm:"primaryKey" json:"id"`
		TrackingID      string    `json:"tracking_id"`
		ProductsName    string    `json:"products_name"`
		HariMasuk       time.Time `json:"hari_masuk"`
		WaktuMasuk      string    `json:"waktu_masuk"`
		HariRespon      time.Time `json:"hari_respon,omitempty"`
		WaktuRespon     string    `json:"waktu_respon,omitempty"`
		UserName        string    `json:"user_name,omitempty"`
		UserEmail       string    `json:"user_email"`
		NoWhatsapp      string    `json:"no_whatsapp"`
		Priority        string    `json:"priority"`
		Status          string    `json:"status"`
		Subject         string    `json:"subject"`
		DetailKendala   string    `json:"detail_kendala"`
		PIC             string    `json:"PIC"`
		ResponDiberikan string    `json:"respon_diberikan,omitempty"`
		CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
		SolvedTime      *string   `json:"solved_time,omitempty"`
	}
	if err := DB.Table("tickets").Where("tracking_id = ?", c.Param("tracking_id")).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if input.Status == "Resolved" {
		startTime := ticket.CreatedAt
		endTime := time.Now()

		// Hitung selisih waktu (durasi penyelesaian)
		duration := endTime.Sub(startTime)

		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		seconds := int(duration.Seconds()) % 60

		var time struct {
			SolvedTime string `json:"solved_time"`
		}
		time.SolvedTime = fmt.Sprintf("%d hours %d minutes %d seconds", hours, minutes, seconds)
		ticket.SolvedTime = &time.SolvedTime
	} else {
		ticket.SolvedTime = nil
	}

	ticket.Status = input.Status

	saveHistory := types.UserTicketHistory{
		UserEmail: user.Email,
		NewStatus: input.Status,
		TicketsID: c.Param("tracking_id"),
		Priority:  ticket.Priority,
		Details:   "Mengubah Status Tickets",
	}

	if err := DB.Table("tickets").Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := DB.Table("user_tickets").Create(&saveHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Status updated successfully",
		Data:    ticket,
	})
}

// @POST
func UpdateTicket(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		ProductsName    string    `json:"products_name"`
		CategoryId      int       `json:"category_id"`
		NoWhatsapp      string    `json:"no_whatsapp"`
		PIC             string    `json:"PIC"`
		DetailKendala   string    `json:"detail_kendala"`
		ResponDiberikan string    `json:"respon_diberikan"`
		Priority        string    `json:"priority"`
		Status          string    `json:"status"`
		HariMasuk       time.Time `json:"hari_masuk"`
		WaktuMasuk      string    `json:"waktu_masuk"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Get user email from token
	var user struct {
		Email string
	}
	if err := DB.Table("users").Select("email").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Update ticket data
	if err := DB.Table("tickets").Where("tracking_id = ?", c.Param("tracking_id")).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Update Tickets : " + err.Error(),
		})
		return
	}

	// Get updated ticket
	var ticket struct {
		TrackingID      string    `json:"tracking_id"`
		ProductsName    string    `json:"products_name"`
		HariMasuk       time.Time `json:"hari_masuk"`
		WaktuMasuk      string    `json:"waktu_masuk"`
		HariRespon      time.Time `json:"hari_respon,omitempty"`
		WaktuRespon     string    `json:"waktu_respon,omitempty"`
		UserName        string    `json:"user_name,omitempty"`
		UserEmail       string    `json:"user_email"`
		NoWhatsapp      string    `json:"no_whatsapp"`
		CategoryId      uint64    `json:"category_id"`
		Priority        string    `json:"priority"`
		Status          string    `json:"status"`
		Subject         string    `json:"subject"`
		DetailKendala   string    `json:"detail_kendala"`
		PIC             string    `json:"PIC"`
		ResponDiberikan string    `json:"respon_diberikan,omitempty"`
		CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
		SolvedTime      *string   `json:"solved_time,omitempty"`
	}
	if err := DB.Table("tickets").
		Select("*, category.category_name").
		Joins("LEFT JOIN category ON tickets.category_id = category.id").
		Where("tracking_id = ?", c.Param("tracking_id")).
		First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data Ticket : " + err.Error(),
		})
		return
	}

	// Calculate resolved time if status is Resolved
	if ticket.Status == "Resolved" {
		duration := time.Since(ticket.CreatedAt)
		ticket.SolvedTime = toSolvedTime(duration)
	} else {
		ticket.SolvedTime = nil
	}

	ticket.DetailKendala = input.DetailKendala
	ticket.PIC = input.PIC

	// Save update to ticket
	if err := DB.Table("tickets").Where("tracking_id = ?", ticket.TrackingID).Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Save history
	history := types.UserTicketHistory{
		UserEmail: user.Email,
		NewStatus: ticket.Status,
		TicketsID: ticket.TrackingID,
		Priority:  ticket.Priority,
		Details:   "Mengubah Data Tiket",
	}

	if err := DB.Table("user_tickets").Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Error Save History : " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Ticket updated successfully",
		Data:    ticket,
	})
}

// @DELETE
func DeleteTicket(c *gin.Context) {
	DB := database.GetDB()
	var ticket types.Tickets
	if err := DB.Where("tracking_id = ?", c.Param("tracking_id")).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Ticket not found"})
		return
	}

	if err := DB.Delete(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}

func HandOverTicket(c *gin.Context) {
	DB := database.GetDB()
	query := `
	SELECT 
		tickets.tracking_id, 
		users.email, 
		COALESCE(shifts.shift_name, shifts_for_admin.shift_name, '-') AS shift_name,
		tickets.status,
		tickets.created_at,
		category.category_name,
		tickets.user_name,
		tickets.products_name,
		tickets.subject,
		tickets.PIC,
		tickets.no_whatsapp,
		tickets.priority,
		users.avatar,
		COALESCE(shifts.id, shifts_for_admin.id, 0) AS shifts_id
	FROM tickets
	JOIN users ON tickets.user_email = users.email
	LEFT JOIN category ON tickets.category_id = category.id
	LEFT JOIN employee_shifts ON users.email = employee_shifts.user_email
	LEFT JOIN shifts ON employee_shifts.shift_id = shifts.id
	LEFT JOIN (
		SELECT id, shift_name, start_time, end_time
		FROM shifts
	) AS shifts_for_admin ON users.role = 'admin' 
		AND employee_shifts.shift_id IS NULL
		AND (
			(shifts_for_admin.start_time < shifts_for_admin.end_time AND TIME(tickets.created_at) BETWEEN shifts_for_admin.start_time AND shifts_for_admin.end_time)
			OR
			(shifts_for_admin.start_time > shifts_for_admin.end_time AND (TIME(tickets.created_at) >= shifts_for_admin.start_time OR TIME(tickets.created_at) <= shifts_for_admin.end_time))
		)
	WHERE tickets.status != 'Resolved'
	AND (
		users.role = 'admin'
		OR employee_shifts.shift_id IS NOT NULL
	)
	ORDER BY 
		CASE tickets.priority
			WHEN 'High' THEN 1
			WHEN 'Medium' THEN 2
			WHEN 'Low' THEN 3
			ELSE 4
		END,
		tickets.created_at ASC;

	`

	// Struct untuk raw query
	var rawTickets []struct {
		TrackingID   string
		CreatedAt    time.Time
		Status       string
		UserName     string
		Avatar       string
		Subject      string
		PIC          string
		NoWhatsapp   string
		Priority     string
		CategoryName string
		ProductsName string
		Email        string
		ShiftName    string
		ShiftsId     string
	}

	if err := DB.Raw(query).Scan(&rawTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Struct untuk response final dengan CreatedAt sudah di-format
	type TicketResponse struct {
		TrackingID   string `json:"tracking_id"`
		CreatedAt    string `json:"created_at"`
		Status       string `json:"status"`
		UserName     string `json:"user_name"`
		Avatar       string `json:"avatar"`
		Subject      string `json:"subject"`
		PIC          string `json:"PIC"`
		NoWhatsapp   string `json:"no_whatsapp"`
		Priority     string `json:"priority"`
		CategoryName string `json:"category_name"`
		ProductsName string `json:"products_name"`
		Email        string `json:"email"`
		ShiftName    string `json:"shift_name"`
		ShiftsId     string `json:"shifts_id"`
	}

	layout := "02-01-2006 15:04"
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	var tickets []TicketResponse

	for _, t := range rawTickets {
		avatar := t.Avatar
		if avatar != "" && !strings.HasPrefix(avatar, "http") {
			avatar = baseURL + avatar
		}

		tickets = append(tickets, TicketResponse{
			TrackingID:   t.TrackingID,
			CreatedAt:    t.CreatedAt.Format(layout),
			Status:       t.Status,
			UserName:     t.UserName,
			Avatar:       avatar,
			Subject:      t.Subject,
			PIC:          t.PIC,
			NoWhatsapp:   t.NoWhatsapp,
			Priority:     t.Priority,
			CategoryName: t.CategoryName,
			ProductsName: t.ProductsName,
			Email:        t.Email,
			ShiftName:    t.ShiftName,
			ShiftsId:     t.ShiftsId,
		})
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get Ticket",
		Data:    tickets,
	})
}
