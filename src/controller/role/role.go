package role

import (
	"my-gin-project/src/database"
	"my-gin-project/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRole(c *gin.Context) {
	DB := database.GetDB()

	var role []struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}
	if err := DB.Table("role").Find(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get roles",
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Get roles",
		Data:    role,
	})
}
