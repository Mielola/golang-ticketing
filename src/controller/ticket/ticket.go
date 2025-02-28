package ticket

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"my-gin-project/src/types"

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
	fmt.Println("Connected to MySQL")
}

// @GET
func GetAllTickets(c *gin.Context) {
	var tickets []types.TicketsResponseAll

	if err := DB.Table("tickets").Select("*").Order("tickets.created_at DESC").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	baseURL := "http://localhost:8080/storage/images/"
	var formattedTickets []map[string]interface{}
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
	var formattedLogs []types.TicketsLogs
	for _, log := range ticketLogs {

		baseURL := "http://localhost:8080/storage/images/"
		if log.UserAvatar != "" {
			log.UserAvatar = baseURL + log.UserAvatar
		}

		formattedLogs = append(formattedLogs, types.TicketsLogs{
			ID:            log.ID,
			TicketsId:     log.TicketsId,
			NewStatus:     log.NewStatus,
			CurrentStatus: log.CurrentStatus,
			UpdateAt:      log.UpdateAt,
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
	// Input structure with proper validation tags
	var inputJSON struct {
		HariMasuk       string `json:"hari_masuk" binding:"required"`
		HariRespon      string `json:"hari_respon" binding:"required"`
		WaktuMasuk      string `json:"waktu_masuk" binding:"required"`
		WaktuRespon     string `json:"waktu_respon" binding:"required"`
		CategoryName    string `json:"category_name" binding:"required"`
		Subject         string `json:"subject" binding:"required"`
		PIC             string `json:"PIC" binding:"required"`
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
	var ticket types.TicketsInput

	// Parse date fields
	hariMasuk, err := time.Parse("2006-01-02", inputJSON.HariMasuk)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hari_masuk format. Expected YYYY-MM-DD"})
		return
	}
	ticket.HariMasuk = hariMasuk

	// Assign remaining fields directly from the struct
	ticket.WaktuMasuk = inputJSON.WaktuMasuk
	ticket.HariRespon = inputJSON.HariRespon
	ticket.WaktuRespon = inputJSON.WaktuRespon
	ticket.UserName = user.Name
	ticket.UserEmail = user.Email
	ticket.CategoryName = inputJSON.CategoryName
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
		UserEmail     string `json:"user_email"`
		NewStatus     string `json:"new_status"`
		CurrentStatus string `json:"current_status"`
		TicketsID     string `json:"ticket_id"`
	}{
		UserEmail:     ticket.UserEmail,
		NewStatus:     "New",
		CurrentStatus: "New",
		TicketsID:     ticket.TrackingID,
	}

	if err := DB.Table("user_tickets").Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// @POST
func UpdateStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	var user struct {
		Email string `json:"email"`
	}

	// @GET Token from Header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// @GET User Email from Token
	if err := DB.Table("users").Select("email").Where("token = ?", token).First(&user).Error; err != nil {
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

	var statusNow = ticket.Status

	if input.Status == "Resolved" || input.Status == "resolved" {
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
		fmt.Println(time.SolvedTime)

		if err := DB.Table("tickets").
			Where("tracking_id = ?", ticket.TrackingID).
			Update("solved_time", time.SolvedTime).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ticket.Status = input.Status

	saveHistory := struct {
		UserEmail     string `json:"user_email"`
		NewStatus     string `json:"new_status"`
		CurrentStatus string `json:"current_status"`
		TicketsID     string `json:"ticket_id"`
	}{
		UserEmail:     user.Email,
		NewStatus:     input.Status,
		CurrentStatus: statusNow,
		TicketsID:     c.Param("tracking_id"),
	}

	if err := DB.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Table("user_tickets").Create(&saveHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "ticket": ticket, "history": saveHistory})
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
