package magang

import (
	"net/http"
	"net/url"
	"time"

	"my-gin-project/src/database"
	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var secretKey = []byte("commandcenter2-ticketing")

func GenerateToken(username string, id uint) (string, error) {
	// Buat klaim token
	claims := jwt.MapClaims{
		"id":       id,
		"username": username,
		"iat":      time.Now().Unix(),
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

func GetAllUsers(c *gin.Context) {
	DB := database.GetDB()

	type TestUser struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Avatar   string `json:"avatar"`
	}

	var users []TestUser

	if err := DB.Table("test_users").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data Users: " + err.Error(),
		})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "No users found",
		})
		return
	}

	// Auto generate avatar menggunakan DiceBear jika kosong
	for i, user := range users {
		if user.Avatar == "" {
			// Encode nama agar spasi diganti %20
			encodedName := url.QueryEscape(user.Name)
			users[i].Avatar = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + encodedName
		}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Get Data Users",
		Data:    users,
	})
}

func Login(c *gin.Context) {
	DB := database.GetDB()
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})

		return
	}

	var user struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := DB.Table("test_users").
		Where("email = ? AND password = ?", req.Email, req.Password).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, types.ResponseFormat{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	token, err := GenerateToken(user.Name, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Login successful",
		Data:    gin.H{"token": token},
	})
}

func CreateUsers(c *gin.Context) {
	DB := database.GetDB()
	type TestUser struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input TestUser

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	var existingUser TestUser
	if err := DB.Table("test_users").Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
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

	newUser := TestUser{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	if err := DB.Table("test_users").Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "User created successfully",
		Data:    newUser,
	})
}

func DeleteUsers(c *gin.Context) {
	DB := database.GetDB()
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	if err := DB.Table("test_users").Where("id = ?", userID).Delete(&struct{}{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "User deleted successfully",
	})
}
