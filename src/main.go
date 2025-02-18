package main

import (
	"my-gin-project/src/database"
	"my-gin-project/src/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {

	database.InitDB()

	r := gin.Default()
	routes.SetupRoutes(r)

	// Menjalankan server
	r.Run(":8080")
}
