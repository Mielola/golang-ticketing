package main

import (
	"my-gin-project/src/database"
	"my-gin-project/src/routes"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	database.InitDB()
	database.MigrateDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	// API duluan
	routes.SetupRoutes(r)

	// Static files
	r.Static("/storage", "./storage")

	// Sajikan semua file static Angular dari folder dist/ticketing
	r.Static("/app", "./dist/ticketing")

	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path

		// Cek apakah file statis ada
		filePath := filepath.Join("./dist/ticketing", requestPath)

		if _, err := os.Stat(filePath); err == nil && !isDir(filePath) {
			c.File(filePath)
			return
		}

		// Fallback ke index.html untuk Angular routing
		c.File("./dist/ticketing/index.html")
	})

	r.Run(":8089")
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
