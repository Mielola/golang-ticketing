package main

import (
	"my-gin-project/src/database"
	"my-gin-project/src/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	database.InitDB()
	database.MigrateDB()

	r := gin.Default()

	// ✅ Tambahkan Middleware CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Ganti dengan URL FE-mu
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	// Setup routes
	routes.SetupRoutes(r)
	r.Static("/storage", "./storage")

	// Menjalankan server
	r.Run(":8081")
}
