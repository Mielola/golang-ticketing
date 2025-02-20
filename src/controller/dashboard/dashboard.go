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
	OpenTickets     int `json:"open_tickets"`
	PendingTickets  int `json:"pending_tickets"`
	ResolvedTickets int `json:"resolved_tickets"`
	TotalTickets    int `json:"total_tickets"`
}
type DashboardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    DataContent `json:"data"`
}

type DataContent struct {
	Summary       Tickets                  `json:"summary"`
	RecentTickets []map[string]interface{} `json:"recent_tickets"`
	UserLogs      []UserLogResponse        `json:"user_logs"`
}

type UserLogResponse struct {
	types.UserResponseWithoutToken
	LoginDate string `json:"login_date"`
	LoginTime string `json:"login_time"`
}

func GetDashboard(c *gin.Context) {
	var tickets Tickets
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

	var recentTickets []map[string]interface{}
	if err := DB.Table("tickets").Select(`*`).Order("created_at DESC").Limit(5).Scan(&recentTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get recent tickets", "error": err.Error()})
		return
	}

	var rawLogs []struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
	}

	var users []types.UserResponseWithoutToken
	if err := DB.Table("user_logs").Select("*").Scan(&rawLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := DB.Table("users").
		Select("users.*").
		Joins("JOIN user_logs ON user_logs.user_email = users.email").
		Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	userMap := make(map[string]types.UserResponseWithoutToken)
	for _, user := range users {
		userMap[user.Email] = user
	}

	var response []UserLogResponse
	for _, log := range rawLogs {
		userData, exists := userMap[log.UserEmail]
		if !exists {
			continue
		}

		response = append(response, UserLogResponse{
			UserResponseWithoutToken: userData,
			LoginDate:                log.LoginTime.Format("2006-01-02"),
			LoginTime:                log.LoginTime.Format("15:04:05"),
		})
	}

	// Gunakan struct agar urutan tidak berubah
	dashboardData := DashboardResponse{
		Success: true,
		Message: "Dashboard data retrieved successfully",
		Data: DataContent{
			Summary:       tickets,
			RecentTickets: recentTickets,
			UserLogs:      response,
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

func SetDB(db *gorm.DB) {
	DB = db
}
