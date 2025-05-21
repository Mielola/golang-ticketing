package products

import (
	"my-gin-project/src/database"
	"my-gin-project/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetProducts(c *gin.Context) {
	DB := database.GetDB()

	var products []string

	if err := DB.Table("products").Select("name").Scan(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get Products",
		Data:    products,
	})
}

func CreateProducts(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := DB.Table("products").Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Create Products" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: false,
		Message: "Success Create Products",
		Data:    input,
	})
}

func UpdateProducts(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := DB.Table("products").Where("products.id", c.Param("id")).Save(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Update Products " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Update Tikcet",
		Data:    input,
	})
}

func GetAllProducts(c *gin.Context) {
	DB := database.GetDB()
	var products []struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		TotalTickets int    `json:"total_tickets"`
	}

	if err := DB.Table("products").Select("products.id, products.name, COUNT(tickets.id) AS total_tickets").
		Joins("LEFT JOIN tickets ON tickets.products_name = products.name").
		Group("products.name").
		Scan(&products).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Products " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get All Products",
		Data:    products,
	})
}

func DeleteProducts(c *gin.Context) {
	DB := database.GetDB()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Product ID is required",
		})
		return
	}

	if err := DB.Table("products").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Delete Products",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Delete Products",
	})
}
