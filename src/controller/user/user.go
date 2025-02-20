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
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

var secretKey = []byte("commandcenter-ticketing")

func GenerateToken(username string) (string, error) {
	// Buat klaim token
	claims := jwt.MapClaims{
		"username": username,
	}

	// Buat token dengan algoritma HMAC
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tangani token dengan secret key
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
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

// @GET Users
func GetProfile(c *gin.Context) {
	var response types.UserResponseWithoutToken
	token := c.GetHeader("Authorization")
	query := DB.Table("users").Where("users.token = ?", token).First(&response)

	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": query.Error.Error()})
		return
	}

	// Cek shift saat ini
	var shifts struct {
		ShiftName string `json:"shift_name"`
	}
	currentDate := time.Now().Format("2006-01-02")

	if err := DB.Table("shifts").
		Select("shifts.shift_name").
		Joins("JOIN employee_shifts ON shifts.id = employee_shifts.shift_id").
		Where("employee_shifts.user_email = ? AND employee_shifts.shift_date = ?", response.Email, currentDate).
		Scan(&shifts).Error; err != nil {
		fmt.Printf("Error fetching shifts: %v\n", err)
	}

	if shifts.ShiftName != "" {
		response.ShiftName = &shifts.ShiftName
	}

	// Avatar Base Url
	baseURL := "http://localhost:8080/storage/images/"
	if response.Avatar != nil && *response.Avatar != "" {
		photoURL := baseURL + *response.Avatar
		response.Avatar = &photoURL
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User retrieved successfully",
		"data":    response,
	})
}

// @GET Users Logs
func GetUsersLogs(c *gin.Context) {
	var rawLogs []struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
	}

	var users []types.UserResponseWithoutToken

	// Ambil data dari user_logs
	if err := DB.Table("user_logs").Select("*").Scan(&rawLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Ambil data user dari tabel users yang memiliki log aktivitas
	if err := DB.Table("users").
		Select("users.*").
		Joins("JOIN user_logs ON user_logs.user_email = users.email").
		Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Buat response dengan menggabungkan user info dan log waktu login
	var response []struct {
		types.UserResponseWithoutToken
		LoginDate string `json:"login_date"`
		LoginTime string `json:"login_time"`
	}

	// Buat mapping email ke user agar lebih cepat saat pencocokan
	userMap := make(map[string]types.UserResponseWithoutToken)
	for _, user := range users {
		userMap[user.Email] = user
	}

	// Gabungkan rawLogs dengan userMap
	for _, log := range rawLogs {
		userData, exists := userMap[log.UserEmail]
		if !exists {
			// Jika user tidak ditemukan, skip
			continue
		}

		response = append(response, struct {
			types.UserResponseWithoutToken
			LoginDate string `json:"login_date"`
			LoginTime string `json:"login_time"`
		}{
			UserResponseWithoutToken: userData,
			LoginDate:                log.LoginTime.Format("2006-01-02"),
			LoginTime:                log.LoginTime.Format("15:04:05"),
		})
	}

	// Kirim response ke client
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User Logs retrieved successfully",
		"users":   response,
	})
}

// @POST Send OTP
func SendOTP(c *gin.Context) {
	var users types.UserBody

	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi Email dan Password
	var user types.User
	if err := DB.Where("email = ? AND password = ?", users.Email, users.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Email or password is incorrect"})
		return
	}

	otp := generateOTP()
	user.OTP = &otp
	user.UpdatedAt = time.Now()

	htmlTemplate := `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>OTP Verification</title>
		<style>
			body { font-family: 'Arial', sans-serif; background-color: #f4f4f9; color: #333; }
			.container { width: 100%; max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1); }
			h1 { color: #5A9D5E; text-align: center; }
			p { font-size: 16px; text-align: center; }
			.otp { font-size: 24px; font-weight: bold; color: #5A9D5E; text-align: center; margin: 20px 0; }
			.footer { text-align: center; font-size: 12px; color: #777; margin-top: 20px; }
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
	</html>`

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse email template"})
		return
	}

	data := map[string]string{"otp": otp}
	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute email template"})
		return
	}

	if err := email.SendEmail(c, user.Email, "Login OTP", bodyBuffer.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user OTP", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "OTP sent to email",
	})
}

// @POST Verify OTP
func VerifyOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user types.User
	if err := DB.Where("email = ? AND OTP = ?", req.Email, req.OTP).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid OTP"})
		return
	}
	// Give Token
	token, err := GenerateToken(req.OTP)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}

	// Reset OTP setelah verifikasi berhasil
	user.OTP = nil
	status := "online"
	user.Status = &status
	user.UpdatedAt = time.Now()
	user.Token = token

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user status", "error": err.Error()})
		return
	}

	// Buat record login
	LoginRecord := struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
	}{
		UserEmail: req.Email,
		LoginTime: time.Now(),
	}

	if err := DB.Table("user_logs").Create(&LoginRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create login record"})
		return
	}

	response := types.UserResponse{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		Status: *user.Status,
		Token:  user.Token,
	}

	// Cek shift saat ini
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
		"message": "User verified",
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

// @PUT User Profile
func EditProfile(c *gin.Context) {
	var response types.UserResponseWithoutToken
	token := c.GetHeader("Authorization")
	query := DB.Table("users").Where("users.token = ?", token).First(&response)

	// Cek User Ada atau Tidak
	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "User Not Found"})
		return
	}

	// Struct untuk menangkap inputan form-data
	var input struct {
		Name  string `form:"name"`
		Email string `form:"email"`
	}

	// Bind form-data (bukan JSON)
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not found"})
		return
	}

	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, allowed types: .jpg, .jpeg, .png"})
		return
	}

	filePath := "storage/images/" + file.Filename

	// Simpan file ke folder yang diinginkan
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Update data user di database
	if err := DB.Table("users").Where("token = ?", token).Updates(map[string]interface{}{
		"name":   input.Name,
		"email":  input.Email,
		"avatar": file.Filename,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Respon sukses
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Profile updated successfully",
		"photo_url": file.Filename,
	})
}

func init() {
	InitDB()
}
