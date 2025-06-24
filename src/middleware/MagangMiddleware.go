package middleware

import (
	"my-gin-project/src/database"
	"my-gin-project/src/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func MagangMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Token not provided"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("commandcenter2-ticketing"), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["id"] == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Invalid token claims"})
			c.Abort()
			return
		}

		userIDFloat, ok := claims["id"].(float64) // JWT numeric values are float64
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: Invalid user ID in token"})
			c.Abort()
			return
		}

		userID := uint(userIDFloat)

		var user models.TestUser
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: User not found"})
			c.Abort()
			return
		}

		// Simpan user ke context agar bisa diakses di handler
		c.Set("user", user)
		c.Next()
	}
}
