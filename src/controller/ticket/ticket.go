package ticket

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"my-gin-project/src/types"

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

// @GET
func GetAllTickets(c *gin.Context) {
	var tickets []types.Tickets
	var User struct {
		Email  string `json:"email"`
		Name   string `json:""name`
		Avatar string `json:"avatar"`
	}

	if err := DB.Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to get tickets",
			Data:    nil,
		})
		return
	}

	if err := DB.Table("users").Select("*").Joins("JOIN tickets ON tickets.user_email = users.email").Scan(&User).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to get user",
			Data:    nil,
		})
		return
	}

	baseURL := "http://localhost:8080/storage/images/"
	if User.Avatar != "" {
		User.Avatar = baseURL + User.Avatar
	}

	var formattedTickets []map[string]interface{}
	for _, ticket := range tickets {
		formattedTickets = append(formattedTickets, map[string]interface{}{
			"id":             ticket.ID,
			"tracking_id":    ticket.TrackingID,
			"hari_masuk":     ticket.HariMasuk.Format("2006-01-02"), // Format YYYY-MM-DD
			"waktu_masuk":    ticket.WaktuMasuk,
			"user":           User,
			"category":       ticket.Category,
			"priority":       ticket.Priority,
			"status":         ticket.Status,
			"subject":        ticket.Subject,
			"detail_kendala": ticket.DetailKendala,
			"owner":          ticket.Owner,
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

	var tickets []types.Tickets
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

// @GET
func GetTicketsByCategory(c *gin.Context) {
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

// @POST
func AddTicket(c *gin.Context) {
	var input types.Tickets
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

	saveHistory := struct {
		UserEmail string `json:"user_email"`
		Status    string `json:"status"`
		TicketsID string `json:"ticket_id"`
	}{
		UserEmail: input.UserEmail,
		Status:    input.Status,
		TicketsID: input.TrackingID,
	}

	if err := DB.Table("user_tickets").Create(&saveHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ticket added successfully", "ticket": input})
}

// @POST
func UpdateStatus(c *gin.Context) {
	// User Input
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	var user struct {
		Email string `json:"email"`
	}

	// @GET Token from Header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP is required"})
		return
	}

	// @GET User Email from Token
	if err := DB.Table("users").Select("email").Where("OTP = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// @Bind JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// @GET Ticket from Database
	var ticket types.Tickets
	if err := DB.Where("tracking_id = ?", c.Param("tracking_id")).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Ticket not found"})
		return
	}

	ticket.Status = input.Status

	saveHistory := struct {
		UserEmail string `json:"user_email"`
		Status    string `json:"status"`
		TicketsID string `json:"ticket_id"`
	}{
		UserEmail: user.Email,
		Status:    input.Status,
		TicketsID: c.Param("tracking_id"),
	}

	if err := DB.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Table("user_tickets").Create(&saveHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "ticket": ticket})
}

// @DELETE
func DeleteTicket(c *gin.Context) {
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

func init() {
	InitDB()
}
