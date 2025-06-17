package places

import (
	"net/http"

	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
)

func CreatePlace(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Name       string `json:"name" binding:"required"`
		ProductsID uint64 `json:"products_id" binding:"required"`
	}

	// Handle invalid input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	// Check if place already exists for the same product
	var existingPlace models.Place
	if err := DB.Where("name = ? AND products_id = ?", input.Name, input.ProductsID).First(&existingPlace).Error; err == nil {
		c.JSON(http.StatusConflict, types.ResponseFormat{
			Success: false,
			Message: "Place already exists for this product",
		})
		return
	}

	// Create new place
	place := models.Place{
		Name:       input.Name,
		ProductsID: input.ProductsID,
	}

	if err := DB.Create(&place).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create place: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "Place created successfully",
		Data:    place,
	})
}

func GetAllPlaces(c *gin.Context) {
	DB := database.GetDB()

	var places []models.Place
	if err := DB.Preload("Product").Find(&places).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "List of places retrieved successfully",
		Data:    places,
	})
}

func UpdatePlace(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var place models.Place
	if err := DB.First(&place, id).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Place not found",
		})
		return
	}

	var input struct {
		Name       string `json:"name" binding:"required"`
		ProductsID uint64 `json:"products_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var existingPlace models.Place
	if err := DB.Where("name = ? AND products_id = ? AND id != ?", input.Name, input.ProductsID, id).First(&existingPlace).Error; err == nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Another place with the same name already exists for this product",
		})
		return
	}

	place.Name = input.Name
	place.ProductsID = input.ProductsID

	if err := DB.Save(&place).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Place updated successfully",
		Data:    place,
	})
}

func DeletePlace(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var place models.Place
	if err := DB.First(&place, id).Error; err != nil {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Place not found",
		})
		return
	}

	if err := DB.Delete(&place).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Place deleted successfully",
	})
}
