package dashboard

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "root:@tcp(db:3306)/commandcenter?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	fmt.Println("Dashboard: Connected to MySQL")
}

func init() {
	InitDB()
}

type Tickets struct {
	OpenTickets    int `json:"open_tickets"`
	ClosedTickets  int `json:"closed_tickets"`
	PendingTickets int `json:"pending_tickets"`
	TotalTickets   int `json:"total_tickets"`
}

func GetDashboard(c *gin.Context) {
	var tickets Tickets

	// @Get Tickets
	if err := DB.Table("tickets").
		Select(`
			COUNT(CASE WHEN status = 'New' THEN 1 END) as open_tickets,
			COUNT(CASE WHEN status = 'On Progress' THEN 1 END) as closed_tickets,
			COUNT(CASE WHEN status = 'Resolved' THEN 1 END) as pending_tickets,
			COUNT("*") as total_tickets
		`).
		Scan(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get dashboard data",
			"error":   err.Error(),
		})
		return
	}

	// Menggunakan slice untuk menampung beberapa tiket
	var recentTickets []map[string]interface{}
	if err := DB.Table("tickets").
		Select(`*`).Order("created_at DESC").Limit(5).Scan(&recentTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get recent tickets",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dashboard data retrieved successfully",
		"data": gin.H{
			"summary":        tickets,
			"recent_tickets": recentTickets,
		},
	})
}

func SetDB(db *gorm.DB) {
	DB = db
}
