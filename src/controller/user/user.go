package user

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"my-gin-project/src/controller/email"
	"my-gin-project/src/types"
	"net/http"
	"strconv"
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

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(999999-100000) + 100000
	return strconv.Itoa(otp)
}

// @GET Users
func GetAllUsers(c *gin.Context) {
	var users []types.UserResponse
	tableName := "users"

	shiftDate := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	query := DB.Table(tableName).
		Select("users.id, users.name, users.email, users.password, users.role, users.status, users.OTP, shifts.shift_name").
		Joins("JOIN employee_shifts ON users.email = employee_shifts.user_email").
		Joins("JOIN shifts ON employee_shifts.shift_id = shifts.id")

	if shiftDate != "" {
		query = query.Where("employee_shifts.shift_date = ?", shiftDate)
	}

	if err := query.Group("users.id").Order("users.id").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All users retrieved successfully",
		"data":    users,
	})
}

// @POST Login
func Login(c *gin.Context) {
	var users types.UserBody

	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingOTP string
	if err := DB.Table("users").Select("OTP").Where("email = ? AND OTP IS NOT NULL", users.Email).Scan(&existingOTP).Error; err == nil && existingOTP != "" {

		// Save History Logout
		if err := DB.Table("user_logs").
			Where("user_email = ? AND logout_time IS NULL", users.Email).
			Order("login_time DESC").
			Limit(1).
			Updates(map[string]interface{}{
				"logout_time": time.Now(),
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log out user", "details": err.Error()})
			return
		}

		// Change Status to Offline
		if err := DB.Table("users").Where("email = ?", users.Email).Update("status", "offline").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status", "details": err.Error()})
			return
		}

		// Clear Existing OTP
		if err := DB.Table("users").Where("email = ?", users.Email).Update("OTP", nil).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing OTP", "details": err.Error()})
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "User is already logged in with an active OTP. Logged out successfully."})
		return

	}

	var user types.User
	if err := DB.Where("email = ? AND password = ?", users.Email, users.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Email or password is incorrect"})
		return
	}

	status := "online"
	user.Status = &status

	otp := generateOTP()
	user.OTP = &otp
	user.UpdatedAt = time.Now()

	htmlTemplate := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>OTP Verification</title>
			<style>
				body {
					font-family: 'Arial', sans-serif;
					background-color: #f4f4f9;
					color: #333;
					margin: 0;
					padding: 0;
				}
				.container {
					width: 100%;
					max-width: 600px;
					margin: 0 auto;
					background-color: #ffffff;
					padding: 20px;
					border-radius: 8px;
					box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
				}
				h1 {
					color: #5A9D5E;
					text-align: center;
				}
				p {
					font-size: 16px;
					line-height: 1.6;
					text-align: center;
				}
				.otp {
					font-size: 24px;
					font-weight: bold;
					color: #5A9D5E;
					text-align: center;
					margin: 20px 0;
				}
				.footer {
					text-align: center;
					font-size: 12px;
					color: #777;
					margin-top: 20px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>OTP Verification</h1>
				<p>Your One-Time Password (OTP) for login is:</p>
				<div class="otp">{{.otp}}</div>
				<p>Please use this OTP to complete your login process. This OTP is valid for 10 minutes.</p>
				<div class="footer">
					<p>Thank you for using our service!</p>
				</div>
			</div>
		</body>
		</html>
	`

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return
	}

	data := map[string]string{
		"otp": otp,
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute email template", "details": err.Error()})
		return
	}

	if err := email.SendEmail(c, "mwildab16@gmail.com", "Login OTP", bodyBuffer.String()); err != nil {
		return
	}

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update user OTP",
			"error":   err.Error(),
		})
		return
	}

	LoginRecord := struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
		Otp       string    `json:"OTP"`
	}{
		UserEmail: users.Email,
		LoginTime: time.Now(),
		Otp:       otp,
	}

	if err := DB.Table("user_logs").Create(&LoginRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Failed to create login record",
			"error":   err.Error(),
		})
		return
	}

	response := types.UserResponse{
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	var shifts struct {
		ShiftName string `json:"shift_name"`
	}
	currentDate := time.Now().Format("2006-01-02")

	if err := DB.Table("shifts").
		Select("shifts.shift_name").
		Joins("JOIN employee_shifts ON shifts.id = employee_shifts.shift_id").
		Where("employee_shifts.user_email = ? AND employee_shifts.shift_date = ?", user.Email, currentDate).
		Scan(&shifts).Error; err != nil {
		fmt.Printf("Error fetching shifts: %v\n", err)
	}

	if shifts.ShiftName != "" {
		response.ShiftName = &shifts.ShiftName
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User found",
		"user":    response,
	})
}

// @POST Logout
func Logout(c *gin.Context) {
	email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Email is required"})
		return
	}

	// Struktur untuk menampung hasil query
	var userInfo struct {
		Status    string
		ShiftName string
	}

	// Query untuk mengambil status dan shift_name dari tabel shifts
	if err := DB.Table("users").
		Select("users.status, shifts.shift_name").
		Joins("JOIN employee_shifts ON users.email = employee_shifts.user_email").
		Joins("JOIN shifts ON employee_shifts.shift_id = shifts.id").
		Where("users.email = ?", email).
		Scan(&userInfo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch user status"})
		return
	}

	if userInfo.Status != "online" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "User cannot log out, status is not 'online'"})
		return
	}

	var user types.User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	status := "offline"
	user.Status = &status
	user.OTP = nil
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update user status"})
		return
	}

	if err := DB.Table("user_logs").
		Where("user_email = ? AND logout_time IS NULL", user.Email).
		Order("login_time DESC").
		Limit(1).
		Updates(map[string]interface{}{
			"logout_time": time.Now(),
			"shift_name":  userInfo.ShiftName,
		}).Error; err != nil {
		fmt.Printf("Error logging user logout: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "User logged out successfully",
		"shiftName": userInfo.ShiftName,
	})
}

// @POST Rehistrasi
func Registration(c *gin.Context) {
	var users types.UserPost

	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if users.Password != users.PasswordRetype {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password and password retype must be the same"})
		return
	}

	if users.Status == nil || *users.Status == "" {
		defaultStatus := "offline"
		users.Status = &defaultStatus // Assign pointer to default value
	}

	if err := DB.Table("users").Create(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": true, "message": "Users added successfully", "users": users})
}

func init() {
	InitDB()
}
