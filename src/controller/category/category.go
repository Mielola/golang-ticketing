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
		CategoryName string `json:"category_name"`
		ProductsName string `json:"products_name"`
	}

	if err := DB.Table("category").
		Select("category.category_name, products.name AS products_name").
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
