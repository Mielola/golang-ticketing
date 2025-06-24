package magang

import (
	"net/http"

	"my-gin-project/src/database"
	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(c *gin.Context) {
	DB := database.GetDB()

	type TestUser struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var users []TestUser

	if err := DB.Table("test_users").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data Users : " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Get Data Users",
		Data:    users,
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
