package category

import (
	"my-gin-project/src/database"
	"my-gin-project/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func GetCategory(c *gin.Context) {
	DB := database.GetDB()

	var category []struct {
		ID           int    `json:"id"`
		CategoryName string `json:"category_name"`
		ProductsName string `json:"products_name"`
	}

	if err := DB.Table("category").
		Select("category.category_name, products.name AS products_name, category.id").
		Joins("LEFT JOIN products ON category.products_id = products.id").
		Order("products_name").
		Scan(&category).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Category " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: false,
		Message: "Success Get Category",
		Data:    category,
	})
}

func GetCategoryById(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var category struct {
		ID           int    `json:"id"`
		CategoryName string `json:"category_name"`
		ProductsName string `json:"products_name"`
	}

	if err := DB.Table("category").
		Select("category.id, category.category_name, products.name AS products_name").
		Joins("LEFT JOIN products ON category.products_id = products.id").
		Where("category.id = ?", id).
		Scan(&category).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Category by ID: " + err.Error(),
		})
		return
	}

	if category.ID == 0 {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get Category by ID",
		Data:    category,
	})
}

func CreateCategory(c *gin.Context) {
	DB := database.GetDB()

	var input struct {
		CategoryName string `json:"category_name"`
		ProductsID   string `json:"products_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	var conflict bool
	if err := DB.Table("category").
		Select("count(*) > 0").
		Where("category_name = ? AND products_id = ?", input.CategoryName, input.ProductsID).
		Find(&conflict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to check conflict: " + err.Error(),
		})
		return
	}

	if conflict {
		c.JSON(http.StatusConflict, types.ResponseFormat{
			Success: false,
			Message: "Conflict: category name already exists for this product",
		})
		return
	}

	if err := DB.Table("category").Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to create category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "Successfully created category",
		Data:    input,
	})
}

func DeleteCategory(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var exists bool
	if err := DB.Table("category").Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).Error; err != nil || !exists {
		c.JSON(http.StatusNotFound, types.ResponseFormat{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	if err := DB.Table("category").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to delete category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully deleted category",
	})
}

func UpdateCategory(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var input struct {
		CategoryName string `json:"category_name"`
		ProductsID   string `json:"products_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid input: " + err.Error(),
		})
		return
	}

	var conflict bool
	if err := DB.Table("category").
		Select("count(*) > 0").
		Where("category_name = ? AND products_id = ? AND id != ?", input.CategoryName, input.ProductsID, id).
		Find(&conflict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to check conflict: " + err.Error(),
		})
		return
	}

	if conflict {
		c.JSON(http.StatusConflict, types.ResponseFormat{
			Success: false,
			Message: "Conflict: category name already used under this product",
		})
		return
	}

	if err := DB.Table("category").Where("id = ?", id).
		Updates(map[string]interface{}{
			"category_name": input.CategoryName,
			"products_id":   input.ProductsID,
		}).Scan(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to update category: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully updated category",
		Data:    input,
	})
}
