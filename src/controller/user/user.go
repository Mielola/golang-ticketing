package user

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"
	"my-gin-project/src/controller/email"
	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var secretKey = []byte("commandcenter2-ticketing")

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
	DB := database.GetDB()
	var users []struct {
		ID       uint    `json:"id"`
		Name     string  `json:"name"`
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Avatar   *string `json:"avatar"`
		Role     string  `json:"role"`
	}

	if err := DB.Model(&models.User{}).
		Select("ID", "Name", "Email", "Password", "Avatar", "role").
		Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to retrieve users: " + err.Error(),
			Data:    nil,
		})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	for i := range users {
		if users[i].Avatar != nil && *users[i].Avatar != "" {
			avatarURL := baseURL + *users[i].Avatar
			users[i].Avatar = &avatarURL
		}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully retrieved users",
		Data:    users,
	})
}

// @UPDATE Users
func UpdateUsers(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	// Cari user berdasarkan id
	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, types.ResponseFormat{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Bind JSON input ke struct
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Update field yang diizinkan
	updates := map[string]interface{}{
		"Name":     input.Name,
		"Email":    input.Email,
		"Password": input.Password,
		"Role":     input.Role,
	}

	if err := DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User updated successfully",
		"user":    user,
	})
}

// @DELETE Users
func DeleteUser(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	// Cari dulu user yang akan dihapus
	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, types.ResponseFormat{
				Success: false,
				Message: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to retrieve user: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Hapus user
	if err := DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete user: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User deleted successfully",
	})
}

func GetUserByID(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var user models.User
	if err := DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, types.ResponseFormat{
				Success: false,
				Message: "User not found",
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to retrieve user: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// @GET Email
func GetEmail(c *gin.Context) {
	DB := database.GetDB()
	var email []struct {
		Id    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := DB.Table("users").Scan(&email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Email" + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: false,
		Message: "Success Get Email",
		Data:    email,
	})
}

// @GET Users
func GetProfile(c *gin.Context) {
	DB := database.GetDB()
	var response types.UserResponseWithoutToken
	token := c.GetHeader("Authorization")

	query := DB.Table("users").Where("users.token = ?", token).First(&response)
	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": query.Error.Error()})
		return
	}

	// Ambil user email
	userEmail := response.Email

	// Buat struct ShiftInfo manual
	type ShiftInfo struct {
		ShiftID     uint    `json:"shift_id"`
		ShiftDate   string  `json:"shift_date"`
		StartTime   string  `json:"start_time"`
		EndTime     string  `json:"end_time"`
		ShiftName   string  `json:"shift_name"`
		ShiftStatus *string `json:"shift_status"`
	}

	var shiftInfo ShiftInfo

	// Sekarang waktu dan tanggal
	now := time.Now()
	currentTime := now.Format("15:04:05")
	currentDate := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// Subquery untuk join shifts
	subQuery := DB.
		Table("employee_shifts").
		Select("employee_shifts.shift_id, employee_shifts.shift_date, shifts.start_time, shifts.end_time, shifts.shift_name").
		Joins("join shifts ON employee_shifts.shift_id = shifts.id").
		Where("employee_shifts.user_email = ?", userEmail).
		Where(DB.Where("shifts.start_time < shifts.end_time AND DATE(employee_shifts.shift_date) = ?", currentDate).
			Or("shifts.start_time > shifts.end_time AND DATE(employee_shifts.shift_date) = ? AND ? <= shifts.end_time", yesterday, currentTime)).
		Limit(1)

	// Execute query dan ambil shiftInfo
	err := DB.Table("(?) as shift_data", subQuery).
		Select("*, CASE WHEN shift_id IS NOT NULL AND ("+
			"(start_time < end_time AND ? BETWEEN start_time AND end_time) OR "+
			"(start_time > end_time AND (? >= start_time OR ? <= end_time))"+
			") THEN 'Active Shift' "+
			"WHEN shift_id IS NOT NULL THEN 'Not On Shift' ELSE NULL END as shift_status",
			currentTime, currentTime, currentTime).
		Scan(&shiftInfo).Error

	if err != nil {
		fmt.Printf("Error fetching shift info: %v\n", err)
	}

	// Tambahkan shift info ke response
	if shiftInfo.ShiftName != "" {
		response.ShiftName = &shiftInfo.ShiftName
		if shiftInfo.ShiftStatus != nil {
			response.ShiftStatus = shiftInfo.ShiftStatus
		}
	}

	// Avatar handler
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
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
	DB := database.GetDB()
	var userLogs []types.UserLogResponse

	if err := DB.Table("user_logs").
		Select("user_logs.login_time, users.avatar, users.email, users.name, users.role, users.avatar, users.status, user_logs.shift_name").
		Joins("JOIN users ON user_logs.user_email = users.email AND DATE()").
		Find(&userLogs).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data User Logs",
		})
		return
	}

	formattedUserLogs := make([]map[string]interface{}, 0)
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)

	for _, user := range userLogs {
		formattedUserLogs = append(formattedUserLogs, map[string]interface{}{
			"email":      user.Email,
			"name":       user.Name,
			"role":       user.Role,
			"avatar":     baseURL + *user.Avatar,
			"shift_name": user.ShiftName,
			"status":     user.Status,
			"login_date": user.LoginTime.Format("2006-01-02"),
			"login_time": user.LoginTime.Format("15:04:05"),
		})
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get Data User Logs",
		Data:    formattedUserLogs,
	})
}

// @POST Send OTP
func SendOTP(c *gin.Context) {
	DB := database.GetDB()
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
	DB := database.GetDB()
	var input struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user types.User
	if err := DB.Where("email = ? AND OTP = ?", input.Email, input.OTP).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid OTP"})
		return
	}
	// Give Token
	token, err := GenerateToken(input.OTP)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}

	// Reset OTP setelah verifikasi berhasil
	var otpNow = user.OTP
	user.OTP = nil
	status := "online"
	user.Status = &status
	user.UpdatedAt = time.Now()
	user.Token = token

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user status", "error": err.Error()})
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
		Where("employee_shifts.user_email = ? AND employee_shifts.shift_date = ?", user.Email, currentDate).
		Scan(&shifts).Error; err != nil {
		fmt.Printf("Error fetching shifts: %v\n", err)
	}

	// Buat record login
	LoginRecord := struct {
		UserEmail string    `json:"user_email"`
		LoginTime time.Time `json:"login_time"`
		ShiftName string    `json:"shift_name"`
		OTP       string    `json:"OTP"`
	}{
		UserEmail: user.Email,
		LoginTime: time.Now(),
		ShiftName: shifts.ShiftName,
		OTP:       *otpNow,
	}

	if err := DB.Table("user_logs").Create(&LoginRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	response := types.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		Status:    *user.Status,
		Token:     user.Token,
		ShiftName: &shifts.ShiftName,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User verified",
		"user":    response,
	})
}

// @POST Logout
func Logout(c *gin.Context) {
	DB := database.GetDB()
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
	DB := database.GetDB()

	// Struct input untuk binding JSON
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	// Bind input JSON ke struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Cek apakah email sudah ada
	var existingUser models.User
	if err := DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, types.ResponseFormat{
			Success: false,
			Message: "Email already exists",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}

	// Hash password dulu (gunakan bcrypt)
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, types.ResponseFormat{
	// 		Success: false,
	// 		Message: "Failed to hash password: " + err.Error(),
	// 	})
	// 	return
	// }

	// Buat object user baru
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		Role:     input.Role,
		Status:   "offline",
	}

	// Simpan user ke database
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}

	// Response sukses
	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "User added successfully",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// @PUT User Profile
func EditProfile(c *gin.Context) {
	DB := database.GetDB()
	var response types.UserResponseWithoutRole
	token := c.GetHeader("Authorization")
	query := DB.Table("users").Where("users.token = ?", token).First(&response)

	// Cek User Ada atau Tidak
	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "User Not Found"})
		return
	}

	var input struct {
		Name  string `form:"name"`
		Email string `form:"email"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData := map[string]interface{}{}

	if input.Name != "" {
		updateData["name"] = input.Name
	}
	if input.Email != "" {
		updateData["email"] = input.Email
	}

	// Cek apakah file avatar dikirimkan
	file, err := c.FormFile("avatar")
	if err == nil {
		allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		ext := filepath.Ext(file.Filename)

		if !allowedExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, allowed types: .jpg, .jpeg, .png"})
			return
		}

		filePath := "storage/images/" + file.Filename

		// Simpan file ke folder
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Tambahkan avatar ke updateData hanya jika file dikirim
		updateData["avatar"] = file.Filename
	}

	// Perbarui data user di database hanya dengan field yang tersedia
	if err := DB.Table("users").Where("token = ?", token).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := DB.Table("users").Where("users.token = ?", token).First(&response).Error
	if user != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Avatar Base Url
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	if response.Avatar != nil && *response.Avatar != "" {
		photoURL := baseURL + *response.Avatar
		response.Avatar = &photoURL
	}

	// Respon sukses
	responseData := gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data": gin.H{
			"user": &response,
		},
	}

	c.JSON(http.StatusOK, responseData)
}

// @POST Edit Status User
func UpdateStatusUser(c *gin.Context) {
	DB := database.GetDB()
	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	if input.Status == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Status is required",
			Data:    nil,
		})
		return
	}

	if input.Status != "online" && input.Status != "offline" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid status",
			Data:    nil,
		})
		return
	}

	var response types.UserResponseWithoutRole
	token := c.GetHeader("Authorization")
	query := DB.Table("users").Where("users.token = ?", token).First(&response)

	if query.Error != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "User Not Found",
			Data:    nil,
		})
		return
	}

	if err := DB.Table("users").Where("token = ?", token).Update("status", input.Status).Scan(&response).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Avatar Base Url
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/storage/images/", scheme, c.Request.Host)
	if response.Avatar != nil && *response.Avatar != "" {
		photoURL := baseURL + *response.Avatar
		response.Avatar = &photoURL
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Status updated successfully",
		Data: gin.H{
			"user": response,
		},
	})
}
