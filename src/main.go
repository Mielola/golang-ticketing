package main

import (
	"fmt"
	"math/rand"
	"my-gin-project/src/database"
	"my-gin-project/src/routes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Ticket struct untuk tabel ticket
type Ticket struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TrackingID      string     `json:"tracking_id"`
	HariMasuk       string     `json:"hari_masuk"`
	WaktuMasuk      string     `json:"waktu_masuk"`
	HariRespon      string     `json:"hari_respon,omitempty"`
	WaktuRespon     string     `json:"waktu_respon,omitempty"`
	NamaAdmin       string     `json:"nama_admin,omitempty"`
	Email           string     `json:"email"`
	Category        string     `json:"category"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	Subject         string     `json:"subject"`
	DetailKendala   string     `json:"detail_kendala"`
	Owner           string     `json:"owner"`
	TimeWorked      *int       `json:"time_worked,omitempty"` // Menggunakan pointer untuk mendukung null
	DueDate         *time.Time `json:"due_date,omitempty"`    // Menggunakan pointer untuk mendukung null
	KategoriMasalah string     `json:"kategori_masalah,omitempty"`
	ResponDiberikan string     `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func AddTicket(c *gin.Context) {
	var input Ticket
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tanggal := time.Now().Format("060102")
	abjad := string('A' + byte(rand.Intn(26)))
	nomorAcak := fmt.Sprintf("%03d", rand.Intn(1000))
	trackingID := fmt.Sprintf("%s%s%s", tanggal, abjad, nomorAcak)[:9]
	trackingID = fmt.Sprintf("%s-%s-%s", trackingID[:3], trackingID[3:6], trackingID[6:9])

	input.TrackingID = trackingID

	if err := DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ticket added successfully", "ticket": input})
}

func UpdateStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ticket Ticket
	if err := DB.Where("tracking_id = ?", c.Param("tracking_id")).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Ticket not found"})
		return
	}

	ticket.Status = input.Status

	if err := DB.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "ticket": ticket})
}

func DeleteTicket(c *gin.Context) {
	var ticket Ticket
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

func GetTicketsByDateRange(c *gin.Context) {
	var input struct {
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}

	// Validasi JSON input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Validasi rentang tanggal
	if input.StartDate.After(input.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "StartDate cannot be after EndDate"})
		return
	}

	var tickets []Ticket
	// Query ke database
	if err := DB.Where("hari_masuk BETWEEN ? AND ?", input.StartDate, input.EndDate).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets: " + err.Error()})
		return
	}

	// Respons
	c.JSON(http.StatusOK, gin.H{
		"message":    "Tickets retrieved successfully",
		"data":       tickets,
		"total":      len(tickets),
		"date_range": gin.H{"start_date": input.StartDate, "end_date": input.EndDate},
	})
}

func main() {

	database.InitDB()

	r := gin.Default()
	routes.SetupRoutes(r)

	// Menjalankan server
	r.Run(":8080")
}
