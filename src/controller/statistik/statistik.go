package statistik

import (
	"my-gin-project/src/database"
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

	var charUserRole []struct {
		Role          *string `json:"role"`
		TotalUserRole *string `json:"total_user_role"`
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
		},
	})
}
