package statistik

import (
	"fmt"
	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetStatistik(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		ProductsName string `json:"products_name" binding:"required"`
		StartDate    string `json:"start_date" binding:"required"`
		EndDate      string `json:"end_date" binding:"required"`
	}

	var userTickets []struct {
		Name             string `json:"name"`
		TotalUserTickets int    `json:"total_user_tickets"`
	}

	var chartCategory []struct {
		CategoryName string `json:"category_name"`
		TotalTickets int    `json:"total_tickets"`
	}

	var chartPriority struct {
		Low      int `json:"low"`
		Medium   int `json:"medium"`
		High     int `json:"high"`
		Critical int `json:"critical"`
	}

	var chartReqCategory []struct {
		Name         string `json:"name"`
		TotalTickets int    `json:"total_tickets"`
	}

	var charUserRole []struct {
		Role          *string `json:"role"`
		TotalUserRole *string `json:"total_user_role"`
	}

	var chartPlace []struct {
		Name         string `json:"name"`
		TotalTickets int    `json:"total_tickets"`
	}

	var chartTicketPeriode []struct {
		CreatedAt    *time.Time `json:"created_at"`
		TotalTickets *string    `json:"total_tickets"`
	}

	var chartUserResolved []struct {
		Name          string `json:"name"`
		ResolvedCount int    `json:"resolved_count"`
	}

	var chartTicketProducts []struct {
		Name         string `json:"name"`
		TotalTickets int    `json:"total_tickets"`
	}

	type PriorityItem struct {
		Label string `json:"label"`
		Value int    `json:"value"`
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

	// Chart Tickets Berdasarkan Products
	if err := DB.Table("products").
		Select("products.name, COUNT(tickets.id) AS total_tickets").
		Joins("LEFT JOIN tickets ON products.name = tickets.products_name").
		Group("products.name").
		Scan(&chartTicketProducts).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Chart Tickets Berdasarkan Prioritas
	if err := DB.Table("tickets").
		Select("COUNT(CASE WHEN priority = 'Low' THEN 1 END) AS low, COUNT(CASE WHEN priority = 'Medium' THEN 1 END) AS medium, COUNT(CASE WHEN priority = 'High' THEN 1 END) AS high, COUNT(CASE WHEN priority = 'Critical' THEN 1 END) AS critical").
		Where("Date(tickets.created_at) BETWEEN ? AND ? AND products_name = ?", input.StartDate, input.EndDate, input.ProductsName).
		Find(&chartPriority).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Chart Tickets Berdasarkan Category
	if err := DB.Table("category").
		Select("category.category_name, COUNT(tickets.id) AS total_tickets").
		Joins("LEFT JOIN tickets ON tickets.category_id = category.id AND DATE(tickets.created_at) BETWEEN ? AND ? AND tickets.products_name = ?", input.StartDate, input.EndDate, input.ProductsName).
		Where("category.products_id = (SELECT id FROM products WHERE name = ?)", input.ProductsName).
		Group("category.category_name").
		Find(&chartCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Chart User total tickets
	if err := DB.Table("users").
		Select("users.name, COUNT(tickets.id) AS total_user_tickets").
		Joins("LEFT JOIN tickets ON users.email = tickets.user_email AND tickets.products_name = ? AND DATE(tickets.created_at) BETWEEN ? AND ?", input.ProductsName, input.StartDate, input.EndDate).
		Group("users.name").
		Find(&userTickets).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Total Tickets By User",
		})
		return
	}

	// Chart Total Ticket Berdasarkan Periode
	if err := DB.Table("tickets").
		Select("DATE(tickets.created_at) AS created_at, COUNT(tickets.id) AS total_tickets").
		Where("DATE(tickets.created_at) BETWEEN ? AND ? AND tickets.products_name = ?", input.StartDate, input.EndDate, input.ProductsName).
		Group("DATE(tickets.created_at)").
		Find(&chartTicketPeriode).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
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

	// Chart Total Role
	if err := DB.Table("users").
		Select("users.role, COUNT(users.role) AS total_user_role").
		Group("users.role").
		Find(&charUserRole).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Chart User Total Resolved
	if err := DB.Table("users").
		Select("users.name,users.email,COUNT(CASE WHEN user_tickets.new_status = 'Resolved' THEN user_tickets.tickets_id END) AS resolved_count").
		Joins("LEFT JOIN user_tickets ON users.email = user_tickets.user_email").
		Group("users.name, users.email").
		Order("resolved_count DESC").
		Scan(&chartUserResolved).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := DB.Table("category").
		Select("category.category_name AS name, COUNT(tickets.id) AS total_tickets").
		Joins("LEFT JOIN tickets ON tickets.category_id = category.id").
		Where("category.category_name IN ?", []string{"Refund", "Relasi", "Visit", "Remote"}).
		Group("category.category_name").
		Scan(&chartReqCategory).Error; err != nil {

		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	priorityItems := []PriorityItem{
		{Label: "Low", Value: chartPriority.Low},
		{Label: "Medium", Value: chartPriority.Medium},
		{Label: "High", Value: chartPriority.High},
		{Label: "Critical", Value: chartPriority.Critical},
	}

	ChartUserTicketsFormatted := make([]map[string]interface{}, 0)
	for _, ticketsItem := range chartTicketPeriode {
		ChartUserTicketsFormatted = append(ChartUserTicketsFormatted, map[string]interface{}{
			"created_at":    ticketsItem.CreatedAt.Format("2006-01-02"),
			"total_tickets": ticketsItem.TotalTickets,
		})
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

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Report generated successfully",
		Data: gin.H{
			"ChartUserTickets":    userTickets,
			"ChartTicketPeriode":  ChartUserTicketsFormatted,
			"ChartPriority":       priorityItems,
			"ChartCategory":       categoryItems,
			"CharUserRole":        charUserRole,
			"ChartUserResolved":   chartUserResolved,
			"ChartTicketProducts": chartTicketProducts,
			"ChartPlaces":         placesItems,
			"ChartReqCategory":    chartReqCategory,
		},
	})
}

func GetStatistikByPlace(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		PlacesName string `json:"places_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON",
			"data":    nil,
		})
		return
	}

	var place models.Place
	if err := DB.Where("name = ?", input.PlacesName).First(&place).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Place not found",
			"data":    nil,
		})
		return
	}

	var tickets []models.Ticket
	if err := DB.Preload("Category").
		Preload("User").
		Preload("Place").
		Order("created_at DESC").
		Where("tickets.places_id = ?", place.ID).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// baseURL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	// userMap
	userMap := make(map[string]models.User)
	for _, t := range tickets {
		userMap[t.UserEmail] = t.User
	}

	// collect tracking ids
	var trackingIDs []string
	for _, t := range tickets {
		trackingIDs = append(trackingIDs, t.TrackingID)
	}

	// last replies
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

	var ticketStatus types.TicketsResponse
	if err := DB.Table("tickets").
		Select(`
			COUNT(CASE WHEN status = 'New' THEN 1 END) as open_tickets,
			COUNT(CASE WHEN status = 'Hold' THEN 1 END) as hold_tickets,
			COUNT(CASE WHEN status = 'On Progress' THEN 1 END) as pending_tickets,
			COUNT(CASE WHEN status = 'Resolved' THEN 1 END) as resolved_tickets,
			COUNT(CASE WHEN priority = 'Critical' THEN 1 END) as critical_tickets,
			COUNT("*") as total_tickets
		`).
		Where("places_id = ?", place.ID).
		Scan(&ticketStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get dashboard data", "error": err.Error()})
		return
	}

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

	// Format response
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

		formatted := map[string]interface{}{
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
			"places_name":    ticket.Place.Name,
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"no_whatsapp":    ticket.NoWhatsapp,
			"detail_kendala": ticket.DetailKendala,
			"pic":            ticket.PIC,
			"created_date":   ticket.CreatedAt.Format("2006-01-02"),
			"created_time":   ticket.CreatedAt.Format("15:04:05"),
			"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		}
		formattedTickets = append(formattedTickets, formatted)
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data: gin.H{
			"tickets": formattedTickets,
			"sumary":  ticketStatus,
		},
	})
}

func GetStatistikByCategory(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		CategoryName string `json:"category_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON",
			"data":    nil,
		})
		return
	}

	var category models.Category
	if err := DB.Where("category_name = ?", input.CategoryName).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Category not found",
			"data":    nil,
		})
		return
	}

	var tickets []models.Ticket
	if err := DB.Preload("Category").
		Preload("User").
		Preload("Place").
		Order("created_at DESC").
		Where("tickets.category_id = ?", category.ID).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// baseURL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	// userMap
	userMap := make(map[string]models.User)
	for _, t := range tickets {
		userMap[t.UserEmail] = t.User
	}

	// collect tracking ids
	var trackingIDs []string
	for _, t := range tickets {
		trackingIDs = append(trackingIDs, t.TrackingID)
	}

	// last replies
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

	var ticketStatus types.TicketsResponse
	if err := DB.Table("tickets").
		Select(`
			COUNT(CASE WHEN status = 'New' THEN 1 END) as open_tickets,
			COUNT(CASE WHEN status = 'Hold' THEN 1 END) as hold_tickets,
			COUNT(CASE WHEN status = 'On Progress' THEN 1 END) as pending_tickets,
			COUNT(CASE WHEN status = 'Resolved' THEN 1 END) as resolved_tickets,
			COUNT(CASE WHEN priority = 'Critical' THEN 1 END) as critical_tickets,
			COUNT("*") as total_tickets
		`).
		Where("category_id = ?", category.ID).
		Scan(&ticketStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get dashboard data", "error": err.Error()})
		return
	}

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

	// Format response
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

		formatted := map[string]interface{}{
			"id":            ticket.ID,
			"tracking_id":   ticket.TrackingID,
			"products_name": ticket.ProductsName,
			"hari_masuk":    ticket.HariMasuk.Format("2006-01-02"),
			"waktu_masuk":   ticket.WaktuMasuk,
			"solved_time":   ticket.SolvedTime,
			"user":          ticketCreator,
			"last_replier":  lastReplier,
			"category":      ticket.Category.CategoryName,
			"priority":      ticket.Priority,
			"places_id":     ticket.PlacesID,
			"places_name": func() string {
				if ticket.Place != nil {
					return ticket.Place.Name
				}
				return "Not Found"
			}(),
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"no_whatsapp":    ticket.NoWhatsapp,
			"detail_kendala": ticket.DetailKendala,
			"pic":            ticket.PIC,
			"created_date":   ticket.CreatedAt.Format("2006-01-02"),
			"created_time":   ticket.CreatedAt.Format("15:04:05"),
			"updated_at":     ticket.UpdatedAt.Format("2006-01-02"),
		}
		formattedTickets = append(formattedTickets, formatted)
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data: gin.H{
			"tickets": formattedTickets,
			"sumary":  ticketStatus,
		},
	})
}
