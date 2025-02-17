package routes

import (
	"my-gin-project/src/controller/notes"
	"my-gin-project/src/controller/user"
	"my-gin-project/src/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")

	// 🔹 Public routes (tanpa OTP)
	v1.POST("/login", user.Login)
	v1.POST("/register", user.Registration)

	// 🔒 Protected routes (wajib OTP)
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// Users
	protected.GET("/users", user.GetAllUsers)
	protected.POST("/users", user.Registration)

	// Notes
	protected.GET("/notes", notes.GetAllNotes)
	protected.GET("/notesByEmail", notes.FindByEmail)
	protected.POST("/notes", notes.CreateNote)
}
