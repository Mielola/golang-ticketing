package statistik

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
	fmt.Println("Connected to MySQL")
}

func init() {
	InitDB()
}

func GetStatistik(c *gin.Context) {
	var input struct {
		ProductsName string `json:"products_name" binding:"required"`
		StartDate    string `json:"start_date" binding:"required"`
		EndDate      string `json:"end_date" binding:"required"`
	}

	var userTickets []struct {
		Name             string `json:"name"`
		Role             string `json:"role"`
		TotalUserRole    string `json:"total_user_role"`
		TotalUserTickets string `json:"total_user_tickets"`
	}

	var chartCategory []struct {
		CategoryName string `json:"category_name"`
		TotalTickets int    `json:"total_tickets"`
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

	// Chart Tickets Berdasarkan Prioritas
	if err := DB.Table("tickets").
		Select("COUNT(CASE WHEN priority = 'Low' THEN 1 END) AS low, COUNT(CASE WHEN priority = 'Medium' THEN 1 END) AS medium, COUNT(CASE WHEN priority = 'High' THEN 1 END) AS high").
		Where("tickets.created_at BETWEEN ? AND ? AND products_name = ?", input.StartDate, input.EndDate, input.ProductsName).
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
		Joins("LEFT JOIN tickets ON tickets.category_name = category.category_name AND tickets.created_at BETWEEN ? AND ? AND tickets.products_name = ?", input.StartDate, input.EndDate, input.ProductsName).
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
		Select("users.name, users.role, COUNT(users.role) AS total_user_role, COUNT(tickets.id) AS total_user_tickets").
		Joins("LEFT JOIN tickets ON users.email = tickets.user_email AND tickets.products_name = ? AND tickets.created_at BETWEEN ? AND ?", input.ProductsName, input.StartDate, input.EndDate).
		Group("users.name, users.role").
		Find(&userTickets).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Total Tickets By User",
		})
		return
	}

	priorityItems := []PriorityItem{
		{Label: "Low", Value: chartPriority.Low},
		{Label: "Medium", Value: chartPriority.Medium},
		{Label: "High", Value: chartPriority.High},
	}

	chartUsersRole := make([]map[string]interface{}, 0)
	for _, users := range userTickets {
		chartUsersRole = append(chartUsersRole, map[string]interface{}{
			"role":            users.Role,
			"total_user_role": users.TotalUserRole,
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
			"ChartUserTickets": userTickets,
			"ChartPriority":    priorityItems,
			"ChartCategory":    categoryItems,
			"ChartUserRole":    chartUsersRole,
		},
	})
}
