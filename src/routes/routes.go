package routes

import (
	"my-gin-project/src/controller/dashboard"
	"my-gin-project/src/controller/notes"
	"my-gin-project/src/controller/shifts"
	"my-gin-project/src/controller/ticket"
	"my-gin-project/src/controller/user"
	"my-gin-project/src/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/V1")

	// --------------------------------------------
	// @ Public routes
	// --------------------------------------------

	v1.POST("/login", user.SendOTP)
	v1.POST("/register", user.Registration)
	v1.GET("/get-profile", user.GetProfile)
	v1.POST("/verifys-otp", user.VerifyOTP)

	protected_V1 := v1.Group("/")
	protected_V1.Use(middleware.AuthMiddleware())

	// --------------------------------------------
	// @ V1 Routes
	// --------------------------------------------

	// Dashboard
	protected_V1.GET("/dashboard", dashboard.GetDashboard)

	// Tickets
	protected_V1.GET("/tickets", ticket.GetAllTickets)
	protected_V1.POST("/tickets", ticket.AddTicket)
	protected_V1.DELETE("/tickets/:tracking_id", ticket.DeleteTicket)
	protected_V1.POST("/tickets/:tracking_id", ticket.UpdateStatus)
	protected_V1.GET("/tickets/category", ticket.GetTicketsByCategory)
	protected_V1.GET("/tickets/status", ticket.GetTicketsByStatus)
	protected_V1.GET("/tickets/priority", ticket.GetTicketsByPriority)
	protected_V1.GET("/tickets-logs", ticket.GetTicketsLogs)

	// Auth
	protected_V1.POST("/logout", user.Logout)

	// Users
	protected_V1.GET("/users", user.GetAllUsers)
	protected_V1.POST("/users", user.Registration)
	protected_V1.GET("/users-logs", user.GetUsersLogs)
	protected_V1.POST("/edit-profile", user.EditProfile)
	protected_V1.POST("/edit-status-user", user.UpdateStatusUser)

	// Notes
	protected_V1.GET("/notes", notes.GetAllNotes)
	protected_V1.GET("/notesByEmail", notes.FindByEmail)
	protected_V1.POST("/notes", notes.CreateNote)

	// Shifts
	protected_V1.POST("/shifts", shifts.AddShift)
	protected_V1.POST("/shifts/:id", shifts.UpdateShift)
	protected_V1.GET("/shifts", shifts.GetAllShifts)
	protected_V1.GET("/shifts-logs", shifts.GetShiftLogs)
	protected_V1.GET("/shifts-users", shifts.GetUserShifts)
	protected_V1.DELETE("/shifts/:id", shifts.DeleteShift)
}
