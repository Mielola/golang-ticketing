package ticket

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

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
	var tickets []models.Ticket

	if err := DB.Preload("Category").
		Preload("User").
		Preload("Place").
		Order("created_at DESC").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	// Map user email ke User struct untuk akses avatar dll
	userMap := make(map[string]models.User)
	for _, ticket := range tickets {
		userMap[ticket.UserEmail] = ticket.User
	}

	// Ambil semua tracking_id untuk query last reply
	var trackingIDs []string
	for _, t := range tickets {
		trackingIDs = append(trackingIDs, t.TrackingID)
	}

	// Ambil last reply per ticket dari user_tickets
	var lastRepliesRaw []struct {
		TicketsID string
		UserEmail string
		NewStatus string
		UpdateAt  time.Time
	}
	DB.Raw(`
        SELECT ut.*
        FROM user_tickets ut
        INNER JOIN (
            SELECT tickets_id, MAX(update_at) AS max_update
            FROM user_tickets
            WHERE tickets_id IN ?
            GROUP BY tickets_id
        ) latest ON ut.tickets_id = latest.tickets_id AND ut.update_at = latest.max_update
    `, trackingIDs).Scan(&lastRepliesRaw)

	// Map last reply per ticket
	lastReplyMap := make(map[string]struct {
		UserEmail string
		NewStatus string
		UpdateAt  time.Time
	})
	lastReplyEmails := make(map[string]bool)
	for _, r := range lastRepliesRaw {
		lastReplyMap[r.TicketsID] = struct {
			UserEmail string
			NewStatus string
			UpdateAt  time.Time
		}{r.UserEmail, r.NewStatus, r.UpdateAt}
		if r.UserEmail != "" {
			lastReplyEmails[r.UserEmail] = true
		}
	}

	// Ambil user last replier
	var lastRepliers []models.User
	if len(lastReplyEmails) > 0 {
		var emails []string
		for e := range lastReplyEmails {
			emails = append(emails, e)
		}
		DB.Where("email IN ?", emails).Find(&lastRepliers)
	}

	lastReplierMap := make(map[string]models.User)
	for _, user := range lastRepliers {
		if user.Avatar != "" {
			user.Avatar = baseURL + user.Avatar
		}
		lastReplierMap[user.Email] = user
	}

	// Susun hasil akhir
	var formattedTickets []map[string]interface{}
	for _, ticket := range tickets {
		ticketCreator := userMap[ticket.UserEmail]
		if ticketCreator.Avatar != "" {
			ticketCreator.Avatar = baseURL + ticketCreator.Avatar
		}

		var lastReplier *models.User
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
			"category":       ticket.Category.CategoryName,
			"priority":       ticket.Priority,
			"places_id":      ticket.PlacesID,
			"places_name":    "Not Found",
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"no_whatsapp":    ticket.NoWhatsapp,
			"detail_kendala": ticket.DetailKendala,
			"pic":            ticket.PIC,
			"created_date":   ticket.CreatedAt.Format("2006-01-02"),
			"created_time":   ticket.CreatedAt.Format("15:04:05"),
			"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		})
		if ticket.Place != nil {
			formattedTickets[len(formattedTickets)-1]["places_name"] = ticket.Place.Name
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tickets retrieved successfully",
		"data":    formattedTickets,
	})
}

// @GET
func GetTicketsLogs(c *gin.Context) {

	type TicketsCreator struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}

	type TicketsLogs struct {
		ID             uint64         `json:"id"`
		TicketsId      string         `json:"tickets_id"`
		NewStatus      string         `json:"new_status"`
		Priority       string         `json:"priority"`
		Details        string         `json:"details"`
		UpdateAt       time.Time      `json:"update_at"`
		UpdateAtString string         `json:"update_at_string"`
		User           TicketsCreator `json:"user"`
	}

	type ResponseFormat struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	DB := database.GetDB()

	var userTickets []models.UserTicket
	err := DB.Preload("User").Order("update_at DESC").Find(&userTickets).Error
	if err != nil {
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

	formattedLogs := make([]TicketsLogs, 0, len(userTickets))
	for _, ut := range userTickets {
		avatar := ut.User.Avatar
		if avatar != "" && !strings.HasPrefix(avatar, "http") {
			avatar = baseURL + avatar
		}

		updateAtStr := ""
		if !ut.UpdateAt.IsZero() {
			updateAtStr = ut.UpdateAt.Format("2006-01-02 15:04:05")
		}

		formattedLogs = append(formattedLogs, TicketsLogs{
			ID:             ut.ID,
			TicketsId:      ut.TicketsID,
			NewStatus:      ut.NewStatus,
			Priority:       ut.Priority,
			Details:        ut.Details,
			UpdateAt:       ut.UpdateAt,
			UpdateAtString: updateAtStr,
			User: TicketsCreator{
				Email:  ut.User.Email,
				Name:   ut.User.Name,
				Avatar: avatar,
			},
		})
	}

	c.JSON(http.StatusOK, ResponseFormat{
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

	// Query tickets dengan JOIN langsung ke users untuk mendapatkan creator info
	type TicketWithCreator struct {
		types.TicketsResponseAll
		CreatorEmail  string `gorm:"column:creator_email"`
		CreatorName   string `gorm:"column:creator_name"`
		CreatorAvatar string `gorm:"column:creator_avatar"`
	}

	tickets := make([]TicketWithCreator, 0)
	if err := DB.Table("tickets").
		Select(`tickets.*, 
			category.category_name, 
			places.name AS places_name,
			users.email AS creator_email,
			users.name AS creator_name,
			users.avatar AS creator_avatar`).
		Where("DATE(tickets.created_at) BETWEEN ? AND ?", startDate.String(), endDate.String()).
		Order("tickets.created_at DESC").
		Joins("LEFT JOIN category ON tickets.category_id = category.id").
		Joins("LEFT JOIN places ON tickets.places_id = places.id").
		Joins("LEFT JOIN users ON tickets.user_email = users.email").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	if len(tickets) == 0 {
		c.JSON(http.StatusOK, types.ResponseFormat{
			Success: true,
			Message: startDateStr + " to " + endDateStr,
			Data:    []map[string]interface{}{},
		})
		return
	}

	// Kumpulkan semua tracking IDs untuk query last replies
	trackingIDs := make([]string, len(tickets))
	for i, ticket := range tickets {
		trackingIDs[i] = ticket.TrackingID
	}

	// Query semua last replies sekaligus
	type LastReplyWithUser struct {
		TicketsID     string    `gorm:"column:tickets_id"`
		UserEmail     string    `gorm:"column:user_email"`
		NewStatus     string    `gorm:"column:new_status"`
		UpdatedAt     time.Time `gorm:"column:update_at"`
		ReplierName   string    `gorm:"column:replier_name"`
		ReplierAvatar string    `gorm:"column:replier_avatar"`
	}

	lastReplies := make([]LastReplyWithUser, 0)
	if err := DB.Table("user_tickets").
		Select(`user_tickets.tickets_id,
			user_tickets.user_email,
			user_tickets.new_status,
			user_tickets.update_at,
			users.name AS replier_name,
			users.avatar AS replier_avatar`).
		Where("user_tickets.tickets_id IN (?)", trackingIDs).
		Joins("LEFT JOIN users ON user_tickets.user_email = users.email").
		Where(`user_tickets.update_at IN (
			SELECT MAX(ut2.update_at) 
			FROM user_tickets ut2 
			WHERE ut2.tickets_id = user_tickets.tickets_id
		)`).
		Find(&lastReplies).Error; err != nil {
		// Jika error, lanjutkan tanpa last reply info
		lastReplies = []LastReplyWithUser{}
	}

	// Buat map untuk akses cepat last replies berdasarkan tracking_id
	lastReplyMap := make(map[string]LastReplyWithUser)
	for _, reply := range lastReplies {
		lastReplyMap[reply.TicketsID] = reply
	}

	// Format response
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	formattedTickets := make([]map[string]interface{}, 0, len(tickets))
	for _, ticket := range tickets {
		// Format creator info
		ticketCreator := types.TicketsCreator{
			Email:  ticket.CreatorEmail,
			Name:   ticket.CreatorName,
			Avatar: ticket.CreatorAvatar,
		}
		if ticketCreator.Avatar != "" {
			ticketCreator.Avatar = baseURL + ticketCreator.Avatar
		}

		// Format last replier info
		var lastReplier *struct {
			Email  string `json:"email"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}

		if lastReply, exists := lastReplyMap[ticket.TrackingID]; exists && lastReply.UserEmail != "" {
			avatar := lastReply.ReplierAvatar
			if avatar != "" {
				avatar = baseURL + avatar
			}
			lastReplier = &struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			}{
				Email:  lastReply.UserEmail,
				Name:   lastReply.ReplierName,
				Avatar: avatar,
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
			"places_name":    ticket.PlacesName,
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
		ProductsName string  `json:"products_name" binding:"required"`
		PlacesID     *string `json:"places_id"`
		CategoryID   string  `json:"category_id"`
		StartDate    string  `json:"start_date" binding:"required"`
		EndDate      string  `json:"end_date" binding:"required"`
		Status       string  `json:"status" binding:"required"`
		StartTime    string  `json:"start_time" binding:"required"`
		EndTime      string  `json:"end_time" binding:"required"`
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

	var tickets []struct {
		TrackingID      string    `json:"tracking_id"`
		CreatedAt       time.Time `json:"created_at"`
		Subject         string    `json:"subject"`
		HariMasuk       time.Time `json:"hari_masuk"`
		WaktuMasuk      string    `json:"waktu_masuk"`
		CategoryName    string    `json:"category_name"`
		PlacesName      *string   `json:"places_name"`
		ResponDiberikan string    `json:"respon_diberikan"`
	}

	var chartPriority struct {
		Low    int `json:"low"`
		Medium int `json:"medium"`
		High   int `json:"high"`
	}

	var chartPlace []struct {
		Name         string `json:"name"`
		TotalTickets int    `json:"total_tickets"`
	}

	type PriorityItem struct {
		Label string `json:"label"`
		Value int    `json:"value"`
	}

	var chartCategory []struct {
		CategoryName string `json:"category_name"`
		TotalTickets int    `json:"total_tickets"`
	}

	// Query chart priority
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

	// Query chart Places
	if err := DB.Table("places").
		Select("places.name, COUNT(tickets.id) AS total_tickets").
		Joins("LEFT JOIN tickets ON places.id = tickets.places_id AND tickets.products_name = ?", input.ProductsName).
		Group("places.name").
		Having("COUNT(tickets.id) > 0").
		Find(&chartPlace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Query chart category
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

	// Dynamic Query
	filters := DB.Table("tickets").
		Select("tickets.tracking_id, tickets.created_at, tickets.subject, tickets.hari_masuk, tickets.waktu_masuk, category.category_name, tickets.respon_diberikan, places.name AS places_name").
		Joins("LEFT JOIN category ON tickets.category_id = category.id").
		Joins("LEFT JOIN places ON tickets.places_id = places.id").
		Where("tickets.created_at BETWEEN ? AND ? AND products_name = ?", startDateTime, endDateTime, input.ProductsName)

	// Jika status tidak "all"
	if strings.ToLower(input.Status) != "all" {
		filters = filters.Where("tickets.status = ?", input.Status)
	}

	// Jika places_id tidak "all" dan tidak kosong
	if input.PlacesID != nil && strings.ToLower(*input.PlacesID) != "all" {
		filters = filters.Where("tickets.places_id = ?", input.PlacesID)
	}

	// Jika category_id tidak "all" dan tidak kosong
	if strings.ToLower(input.CategoryID) != "all" && input.CategoryID != "" {
		filters = filters.Where("tickets.category_id = ?", input.CategoryID)
	}

	// Execute query
	if err := filters.Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
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

	placesItems := make([]map[string]interface{}, 0)
	for _, places := range chartPlace {
		placesItems = append(placesItems, map[string]interface{}{
			"places_name":   places.Name,
			"total_tickets": places.TotalTickets,
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
			"places_name":   ticket.PlacesName,
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
			"ChartPlaces":   placesItems,
			"ChartPriority": priorityItems,
			"ChartCategory": categoryItems,
		},
	})
}

// @POST
func AddTicket(c *gin.Context) {
	DB := database.GetDB()

	// Struktur input dengan validasi
	var inputJSON struct {
		HariMasuk       string  `json:"hari_masuk" binding:"required"`
		HariRespon      string  `json:"hari_respon" binding:"required"`
		WaktuMasuk      string  `json:"waktu_masuk" binding:"required"`
		WaktuRespon     string  `json:"waktu_respon" binding:"required"`
		CategoryId      uint64  `json:"category_id" binding:"required"`
		PlacesID        *uint64 `json:"places_id"`
		Subject         string  `json:"subject" binding:"required"`
		PIC             string  `json:"PIC"`
		DetailKendala   string  `json:"detail_kendala" binding:"required"`
		ResponDiberikan string  `json:"respon_diberikan" binding:"required"`
		NoWhatsapp      string  `json:"no_whatsapp" binding:"required"`
		Priority        string  `json:"priority" binding:"required"`
		ProductsName    string  `json:"products_name" binding:"required"`
	}

	// Bind input JSON
	if err := c.ShouldBindJSON(&inputJSON); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Ambil token dari header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	// Cari user berdasarkan token
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := DB.Table("users").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, types.ResponseFormat{
			Success: false,
			Message: "User not found or invalid token",
		})
		return
	}

	// Parsing tanggal dengan validasi
	hariMasuk, err := time.Parse("2006-01-02", inputJSON.HariMasuk)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hari_masuk format. Expected YYYY-MM-DD"})
		return
	}

	hariRespon, err := time.Parse("2006-01-02", inputJSON.HariRespon)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hari_respon format. Expected YYYY-MM-DD"})
		return
	}

	// Inisialisasi struct ticket
	ticket := models.Ticket{
		HariMasuk:       hariMasuk,
		WaktuMasuk:      inputJSON.WaktuMasuk,
		HariRespon:      &hariRespon,
		WaktuRespon:     inputJSON.WaktuRespon,
		UserName:        user.Name,
		UserEmail:       user.Email,
		CategoryId:      inputJSON.CategoryId,
		Priority:        inputJSON.Priority,
		Subject:         inputJSON.Subject,
		DetailKendala:   inputJSON.DetailKendala,
		PIC:             inputJSON.PIC,
		ResponDiberikan: inputJSON.ResponDiberikan,
		NoWhatsapp:      inputJSON.NoWhatsapp,
		ProductsName:    inputJSON.ProductsName,
		TrackingID:      generateTrackingID(inputJSON.ProductsName),
	}

	// Handle PlacesID yang bisa nil
	if inputJSON.PlacesID != nil {
		ticket.PlacesID = *&inputJSON.PlacesID
	}

	// Buat history
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

	// Transaction simpan ticket dan history sekaligus
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("tickets").Create(&ticket).Error; err != nil {
			return err
		}

		if err := tx.Table("user_tickets").Create(&history).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "Ticket added successfully",
		Data:    ticket,
	})
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// @POST
func ImportTicketsArray(c *gin.Context) {
	DB := database.GetDB()

	// Struktur untuk terima array
	var inputJSON []struct {
		HariMasuk       string  `json:"hari_masuk" binding:"required"`
		HariRespon      string  `json:"hari_respon" binding:"required"`
		WaktuMasuk      string  `json:"waktu_masuk" binding:"required"`
		WaktuRespon     string  `json:"waktu_respon" binding:"required"`
		CategoryId      uint64  `json:"category_id" binding:"required"`
		Subject         string  `json:"subject" binding:"required"`
		Status          string  `json:"status" binding:"required"`
		PIC             *string `json:"PIC"`
		PlacesID        *uint64 `json:"places_id"`
		DetailKendala   string  `json:"detail_kendala" binding:"required"`
		ResponDiberikan string  `json:"respon_diberikan" binding:"required"`
		NoWhatsapp      *string `json:"no_whatsapp"`
		Priority        string  `json:"priority" binding:"required"`
		ProductsName    string  `json:"products_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&inputJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil token user
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	// Get user
	var user struct {
		Name  string
		Email string
	}
	if err := DB.Table("users").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var importedTickets []models.Ticket
	var ticketHistories []map[string]interface{}

	for _, item := range inputJSON {
		hariMasuk, _ := time.Parse("2006-01-02", item.HariMasuk)
		hariRespon, _ := time.Parse("2006-01-02", item.HariRespon)
		timeSolved := "On Progress"

		if item.Status == "Resolved" {
			timeSolved = "Generate By Excel"
		}

		ticket := models.Ticket{
			HariMasuk:       hariMasuk,
			HariRespon:      &hariRespon,
			WaktuMasuk:      item.WaktuMasuk,
			WaktuRespon:     item.WaktuRespon,
			CategoryId:      item.CategoryId,
			Subject:         item.Subject,
			PIC:             getStringValue(item.PIC),
			PlacesID:        item.PlacesID,
			DetailKendala:   item.DetailKendala,
			ResponDiberikan: item.ResponDiberikan,
			Status:          item.Status,
			SolvedTime:      timeSolved,
			NoWhatsapp:      getStringValue(item.NoWhatsapp),
			Priority:        item.Priority,
			ProductsName:    item.ProductsName,
			UserName:        user.Name,
			UserEmail:       user.Email,
			TrackingID:      generateTrackingID(item.ProductsName),
		}

		importedTickets = append(importedTickets, ticket)

		history := map[string]interface{}{
			"user_email": user.Email,
			"new_status": item.Status,
			"tickets_id": ticket.TrackingID,
			"priority":   item.Priority,
			"details":    "Membuat Tiket Baru via Excel",
		}

		ticketHistories = append(ticketHistories, history)
	}

	tx := DB.Begin()

	if err := tx.Table("tickets").Create(&importedTickets).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to import tickets: " + err.Error(),
		})
		return
	}

	if err := tx.Table("user_tickets").Create(&ticketHistories).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create ticket histories: " + err.Error(),
		})
		return
	}
	tx.Commit()

	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "Success Import Ticket",
		Data:    importedTickets,
	})
}

func generateTrackingID(productName string) string {
	// Membuat prefix dari inisial nama produk
	words := strings.Fields(productName)
	var prefix string
	for _, word := range words {
		prefix += strings.ToUpper(string(word[0]))
	}

	// Format tanggal: YYMMDD
	tanggal := time.Now().Format("060102")

	// Ambil 6 karakter pertama dari UUID (cukup unik)
	uuidPart := strings.ToUpper(uuid.New().String()[:6])

	// Format akhir
	trackingID := fmt.Sprintf("%s-%s-%s", prefix, tanggal, uuidPart)
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
		"place_id":       ticket.PlacesID,
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
func ResolvedTicket(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Status             string `json:"status" binding:"required"`
		CategoryResolvedId uint64 `json:"category_resolved_id" binding:"required"`
		NoteResolved       string `json:"note_resolved" binding:"required"`
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
		ID                 uint      `gorm:"primaryKey" json:"id"`
		TrackingID         string    `json:"tracking_id"`
		ProductsName       string    `json:"products_name"`
		HariMasuk          time.Time `json:"hari_masuk"`
		WaktuMasuk         string    `json:"waktu_masuk"`
		HariRespon         time.Time `json:"hari_respon,omitempty"`
		WaktuRespon        string    `json:"waktu_respon,omitempty"`
		UserName           string    `json:"user_name,omitempty"`
		UserEmail          string    `json:"user_email"`
		NoWhatsapp         string    `json:"no_whatsapp"`
		Priority           string    `json:"priority"`
		CategoryResolvedId uint64    `json:"category_resolved_id"`
		NoteResolved       string    `json:"note_resolved"`
		Status             string    `json:"status"`
		Subject            string    `json:"subject"`
		DetailKendala      string    `json:"detail_kendala"`
		PIC                string    `json:"PIC"`
		ResponDiberikan    string    `json:"respon_diberikan,omitempty"`
		CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
		SolvedTime         *string   `json:"solved_time,omitempty"`
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
	ticket.CategoryResolvedId = input.CategoryResolvedId
	ticket.NoteResolved = input.NoteResolved

	saveHistory := types.UserTicketHistory{
		UserEmail: user.Email,
		NewStatus: input.Status,
		TicketsID: c.Param("tracking_id"),
		Priority:  ticket.Priority,
		Details:   input.NoteResolved,
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
		PlacesID        uint32    `json:"places_id"`
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
		PlacesID        *uint64   `gorm:"index"`
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
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{Success: false, Message: "Token is required"})
		return
	}

	var user struct {
		Name  string
		Email string
	}
	if err := DB.Table("users").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{Success: false, Message: "Invalid token or user not found"})
		return
	}

	var ticket models.Ticket
	if err := DB.Table("tickets").Where("tracking_id = ?", c.Param("tracking_id")).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{Success: false, Message: "Ticket not found"})
		return
	}

	waktuMasuk, err := combineDateTime(ticket.HariMasuk, ticket.WaktuMasuk)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{Success: false, Message: "Invalid waktu_masuk"})
		return
	}
	var waktuRespon *time.Time
	if ticket.WaktuRespon != "" && ticket.HariRespon != nil {
		combined, err := combineDateTime(*ticket.HariRespon, ticket.WaktuRespon)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ResponseFormat{Success: false, Message: "Invalid waktu_respon"})
			return
		}
		waktuRespon = &combined
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tempTicket := models.TempTickets{
		TrackingID:      ticket.TrackingID,
		HariMasuk:       ticket.HariMasuk,
		WaktuMasuk:      waktuMasuk.Format("15:04:05"),
		HariRespon:      ticket.HariRespon,
		WaktuRespon:     waktuRespon.Format("15:04:05"),
		SolvedTime:      ticket.SolvedTime,
		UserEmail:       ticket.UserEmail,
		DeletedBy:       user.Email,
		NoWhatsapp:      ticket.NoWhatsapp,
		CategoryId:      ticket.CategoryId,
		ProductsName:    ticket.ProductsName,
		Priority:        ticket.Priority,
		Status:          ticket.Status,
		Subject:         ticket.Subject,
		DetailKendala:   ticket.DetailKendala,
		PIC:             ticket.PIC,
		ResponDiberikan: ticket.ResponDiberikan,
		CreatedAt:       ticket.CreatedAt,
		UpdatedAt:       ticket.UpdatedAt,
		DeletedAt:       timePtr(time.Now()),
	}
	if err := tx.Table("temp_tickets").Save(&tempTicket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{Success: false, Message: "Failed to save temporary ticket"})
		return
	}

	var userTickets []models.UserTicket
	if err := tx.Table("user_tickets").Where("tickets_id = ?", ticket.TrackingID).Find(&userTickets).Error; err == nil {
		if len(userTickets) > 0 {
			// batch insert
			var tempUserTickets []models.TempUserTickets
			for _, ut := range userTickets {
				tempUserTickets = append(tempUserTickets, models.TempUserTickets{
					TicketsID: ut.TicketsID,
					UserEmail: ut.UserEmail,
					NewStatus: ut.NewStatus,
					UpdateAt:  ut.UpdateAt,
					Priority:  ut.Priority,
					Details:   ut.Details,
				})
			}
			if err := tx.Table("temp_user_tickets").Create(&tempUserTickets).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{Success: false, Message: "Failed to save temporary user tickets"})
				return
			}
		}
	}

	if err := tx.Table("user_tickets").Where("tickets_id = ?", ticket.TrackingID).Delete(&models.UserTicket{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{Success: false, Message: "Failed to delete user tickets"})
		return
	}

	if err := tx.Table("tickets").Where("tracking_id = ?", ticket.TrackingID).Delete(&models.Ticket{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{Success: false, Message: "Failed to delete ticket"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{Success: false, Message: "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{Success: true, Message: "Ticket deleted and moved to temporary tickets"})
}

// @POST
func RestoreTicket(c *gin.Context) {
	DB := database.GetDB()
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	var user struct {
		Name  string
		Email string
	}
	if err := DB.Table("users").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid token or user not found",
		})
		return
	}

	var tempTicket models.TempTickets
	if err := DB.Table("temp_tickets").Where("tracking_id = ?", c.Param("tracking_id")).First(&tempTicket).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Temporary ticket not found",
		})
		return
	}

	waktuMasuk, err := combineDateTime(tempTicket.HariMasuk, tempTicket.WaktuMasuk)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid waktu_masuk format",
		})
		return
	}

	var waktuRespon string
	if tempTicket.WaktuRespon != "" && tempTicket.HariRespon != nil {
		waktuRespon = tempTicket.WaktuRespon
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ticket := models.Ticket{
		TrackingID:      tempTicket.TrackingID,
		HariMasuk:       tempTicket.HariMasuk,
		WaktuMasuk:      waktuMasuk.Format("15:04:05"),
		HariRespon:      tempTicket.HariRespon,
		WaktuRespon:     waktuRespon,
		SolvedTime:      tempTicket.SolvedTime,
		UserEmail:       tempTicket.UserEmail,
		NoWhatsapp:      tempTicket.NoWhatsapp,
		CategoryId:      tempTicket.CategoryId,
		ProductsName:    tempTicket.ProductsName,
		Priority:        tempTicket.Priority,
		Status:          tempTicket.Status,
		Subject:         tempTicket.Subject,
		DetailKendala:   tempTicket.DetailKendala,
		PIC:             tempTicket.PIC,
		ResponDiberikan: tempTicket.ResponDiberikan,
		CreatedAt:       tempTicket.CreatedAt,
		UpdatedAt:       tempTicket.UpdatedAt,
	}

	if err := tx.Table("tickets").Save(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to restore ticket",
		})
		return
	}

	var tempUserTickets []models.TempUserTickets
	if err := tx.Table("temp_user_tickets").Where("tickets_id = ?", tempTicket.TrackingID).Find(&tempUserTickets).Error; err == nil {
		if len(tempUserTickets) > 0 {
			var userTickets []models.UserTicket
			for _, tut := range tempUserTickets {
				userTickets = append(userTickets, models.UserTicket{
					TicketsID: tut.TicketsID,
					UserEmail: tut.UserEmail,
					NewStatus: tut.NewStatus,
					UpdateAt:  tut.UpdateAt,
					Priority:  tut.Priority,
					Details:   tut.Details,
				})
			}
			if err := tx.Table("user_tickets").Create(&userTickets).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{
					Success: false,
					Message: "Failed to restore user tickets",
				})
				return
			}
		}
	}

	// Hapus data sementara user tickets & tiket dalam 1 delete per tabel
	if err := tx.Table("temp_user_tickets").Where("tickets_id = ?", tempTicket.TrackingID).Delete(&models.TempUserTickets{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete temporary user tickets",
		})
		return
	}

	if err := tx.Table("temp_tickets").Where("tracking_id = ?", tempTicket.TrackingID).Delete(&models.TempTickets{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete temporary ticket",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Ticket restored successfully",
	})
}

// @GET
func GetDeletedTickets(c *gin.Context) {
	DB := database.GetDB()

	// Wajib: inisialisasi slice kosong
	deletedTickets := make([]struct {
		TrackingID   string    `json:"tracking_id"`
		CategoryName string    `json:"category"`
		ProductName  string    `json:"product"`
		DeletedBy    string    `json:"deleted_by"`
		DeletedAt    time.Time `json:"deleted_at"`
	}, 0)

	err := DB.Table("temp_tickets").
		Select("temp_tickets.tracking_id, category.category_name as category_name, temp_tickets.products_name as product_name, temp_tickets.deleted_by, temp_tickets.deleted_at").
		Joins("LEFT JOIN category ON temp_tickets.category_id = category.id").
		Order("temp_tickets.deleted_at DESC").
		Scan(&deletedTickets).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to retrieve deleted tickets",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Deleted tickets fetched successfully",
		Data:    deletedTickets,
	})
}

func DeleteTempTickets(c *gin.Context) {
	DB := database.GetDB()

	// Ambil tracking_id dari parameter
	trackingID := c.Param("tracking_id")

	// Cek apakah tiket ada di temp_tickets
	var tempTicket models.TempTickets
	if err := DB.Table("temp_tickets").Where("tracking_id = ?", trackingID).First(&tempTicket).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Temporary ticket not found",
		})
		return
	}

	// Hapus semua relasi user_tickets yang ada di temp_user_tickets
	if err := DB.Table("temp_user_tickets").Where("tickets_id = ?", trackingID).Delete(&models.TempUserTickets{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete related temporary user tickets",
		})
		return
	}

	// Hapus tiket dari temp_tickets
	if err := DB.Table("temp_tickets").Where("tracking_id = ?", trackingID).Delete(&models.TempTickets{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete temporary ticket",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Temporary ticket deleted successfully",
	})
}

// Helper untuk pointer waktu
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper gabungkan date + time string jadi DATETIME Go
func combineDateTime(date time.Time, timeStr string) (time.Time, error) {
	dateStr := date.Format("2006-01-02")
	fullDateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)
	layout := "2006-01-02 15:04:05"
	return time.Parse(layout, fullDateTimeStr)
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

	-- Join hanya ambil shift terakhir per user
	LEFT JOIN (
		SELECT user_email, MAX(shift_id) AS shift_id
		FROM employee_shifts
		GROUP BY user_email
	) AS employee_shifts ON users.email = employee_shifts.user_email

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
		tickets.created_at DESC;
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
