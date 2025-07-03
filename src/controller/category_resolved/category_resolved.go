package categoryresolved

import (
	"errors"
	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetResolvedCategory(c *gin.Context) {
	DB := database.GetDB()

	var category []models.CategoryResolved
	if err := DB.Table("category_resolved").Find(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Resolved Category " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: false,
		Message: "Success Get Resolved Category",
		Data:    category,
	})
}

func CreateResolvedCategory(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	// Cek apakah nama kategori sudah ada (conflict)
	var existingCategory models.CategoryResolved
	if err := DB.Table("category_resolved").Where("name = ?", input.Name).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict, types.ResponseFormat{
			Success: false,
			Message: "Category with the same name already exists",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}

	category := models.CategoryResolved{
		Name: input.Name,
	}

	if err := DB.Table("category_resolved").Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create resolved category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully created resolved category",
		Data:    category,
	})
}
