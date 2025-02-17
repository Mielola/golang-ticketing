package user

import (
	"fmt"
	"log"
	"math/rand"
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing token"})
			c.Abort()
			return
		}

		// Periksa apakah OTP ada di database
		var user types.User
		if err := DB.Where("OTP = ?", token).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Simpan informasi user dalam context untuk digunakan di endpoint lain
		c.Set("user", user)
		c.Next()
	}
}

// @GET Users
func GetAllUsers(c *gin.Context) {
	var users []types.UserResponse
	tableName := "users"

	// Ambil parameter query 'shift_date' yang diberikan oleh admin
	shiftDate := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// Query dasar
	query := DB.Table(tableName).
		Select("users.id, users.name, users.email, users.password, users.role, users.status, users.OTP, shifts.shift_name").
		Joins("JOIN employee_shifts ON users.email = employee_shifts.user_email").
		Joins("JOIN shifts ON employee_shifts.shift_id = shifts.id")

	// if err := DB.Table(tableName).
	// 	Select("users.email, users.name, users.status, users.role, shifts.shift_name").
	// 	Joins("JOIN employee_shifts ON users.email = employee_shifts.user_email").
	// 	Joins("JOIN shifts ON employee_shifts.shift_id = shifts.id").
	// 	Find(&users).Error; err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// Jika ada parameter 'shift_date', filter berdasarkan tanggal tersebut
	if shiftDate != "" {
		query = query.Where("employee_shifts.shift_date = ?", shiftDate)
	}

	// Eksekusi query
	if err := query.Group("users.id").Order("users.id").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kirimkan respons dengan data yang ditemukan
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All users retrieved successfully",
		"data":    users,
	})
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(999999-100000) + 100000 // generates a 6-digit OTP
	return strconv.Itoa(otp)
}

// @POST Login
func Login(c *gin.Context) {
	var users types.UserBody
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user types.User
	if err := DB.Where("email = ? AND password = ?", users.Email, users.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Update status user menjadi online
	status := "online"
	user.Status = &status

	// Generate dan update OTP
	otp := generateOTP()
	user.OTP = &otp
	user.UpdatedAt = time.Now()

	// Simpan perubahan status dan OTP dengan logging error
	if err := DB.Save(&user).Error; err != nil {
		fmt.Printf("Error saving user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update user OTP",
			"error":   err.Error(),
		})
		return
	}

	loginTime := types.LoginTime{
        Email: users.Email,
        Login: time.Now(),
    }

    // Insert the login time into the login_time table
    if err := DB.Create(&loginTime).Error; err != nil {
        fmt.Printf("Error logging login time: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "message": "Failed to log login time",
            "error":   err.Error(),
        })
        return
    }

	// Response seperti sebelumnya
	response := types.UserResponse{
		Email:  user.Email,
		Name:   user.Name,
		Status: *user.Status,
		Role:   user.Role,
		OTP:    user.OTP,
	}

	// Ambil shift yang berlaku pada tanggal saat ini
	var shifts []string
	currentDate := time.Now().Format("2006-01-02")
	if err := DB.Table("shifts").
		Select("shifts.shift_name").
		Joins("JOIN employee_shifts ON shifts.id = employee_shifts.shift_id").
		Where("employee_shifts.user_email = ? AND employee_shifts.shift_date = ?", user.Email, currentDate).
		Pluck("shifts.shift_name", &shifts).Error; err != nil {
		fmt.Printf("Error fetching shifts: %v\n", err)
	}

	// Jika ada shifts, set shift default
	if len(shifts) > 0 {
		response.ShiftName = &shifts[0]
	}

	// if response.ShiftName == nil {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"status":  false,
	// 		"message": "Bukan Shift Anda",
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User found",
		"user":    response,
	})
}


func Logout(c *gin.Context) {
    var users types.UserResponse
    if err := c.ShouldBindJSON(&users); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find the user by email and verify the OTP
    var user types.User
    if err := DB.Where("email = ? AND otp = ?", users.Email, users.OTP).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Invalid email or OTP"})
        return
    }

    // Update user status to 'offline' and clear OTP
    status := "offline"
    user.Status = &status
    user.OTP = nil // Clear OTP after logout
    user.UpdatedAt = time.Now()

    // Save user status change
    if err := DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "message": "Failed to update user status",
            "error":   err.Error(),
        })
        return
    }

    // Record logout time in the login_time table
    logoutTime := time.Now()
    var loginTimeRecord types.LoginTime

    // Check if a record already exists for this user
    if err := DB.Where("email = ?", users.Email).First(&loginTimeRecord).Error; err != nil {
        // If no record exists, create a new one with login and logout times
        loginTimeRecord = types.LoginTime{
            Email: users.Email,
            Login: user.UpdatedAt, // Assuming the last updated time is when the user logged in
            Logout: &logoutTime,
        }

        if err := DB.Create(&loginTimeRecord).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": "Failed to record logout time",
                "error":   err.Error(),
            })
            return
        }
    } else {
        // If a record exists, update the logout time
        loginTimeRecord.Logout = &logoutTime
        if err := DB.Save(&loginTimeRecord).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": "Failed to update logout time",
                "error":   err.Error(),
            })
            return
        }
    }

    // Send a success response
    c.JSON(http.StatusOK, gin.H{
        "message": "Logout successful",
        "email":   users.Email,
    })
}


// @POST Users
func Registration(c *gin.Context) {
	var users types.User

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

	if err := DB.Create(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": true, "message": "Users added successfully", "users": users})
}

func init() {
	InitDB()
}
