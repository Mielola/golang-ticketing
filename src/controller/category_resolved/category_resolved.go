package categoryresolved

import (
	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"my-gin-project/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
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
