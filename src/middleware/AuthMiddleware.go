package middleware

import (
	"my-gin-project/src/database"
	"my-gin-project/src/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header Authorization
		token := c.GetHeader("Authorization")

		// Hapus "Bearer " jika ada
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		// Periksa apakah OTP ada di database
		var user types.User
		if err := database.DB.Where("OTP = ?", token).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Simpan user dalam context agar bisa diakses di endpoint
		c.Set("user", user)
		c.Next()
	}
}
