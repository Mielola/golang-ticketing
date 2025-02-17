package ticket

import (
	"fmt"
	"log"
	"net/http"

	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Ticket struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TrackingID string `json:"migration"`
	HariMasuk  string `json:"batch"`
}

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

func GetAllTickets(c *gin.Context) {
	var migrations []types.Tickets
	if err := DB.Find(&migrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All tickets retrieved successfully", "data": migrations})
}

func init() {
	InitDB()
}
