package routes

import (
	export "my-gin-project/src/controller/Export"
	"my-gin-project/src/controller/dashboard"
	"my-gin-project/src/controller/notes"
	"my-gin-project/src/controller/products"
	"my-gin-project/src/controller/shifts"
	"my-gin-project/src/controller/statistik"
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
	v1.POST("/verify-otp", user.VerifyOTP)
	v1.GET("/mikrotik", user.ConnectMikrotik)

	protected_V1 := v1.Group("/")
	protected_V1.Use(middleware.AuthMiddleware())

	// --------------------------------------------
	// @ V1 Routes
	// --------------------------------------------

	// Dashboard
	protected_V1.GET("/dashboard", dashboard.GetDashboard)
	protected_V1.POST("/get-data-form", dashboard.GetForm)

	// Tickets
	protected_V1.POST("/tickets", ticket.AddTicket)
	protected_V1.GET("/tickets", ticket.GetAllTickets)
	protected_V1.POST("/tickets/:tracking_id", ticket.UpdateTicket)
	protected_V1.POST("/ticket-status/:tracking_id", ticket.UpdateStatus)
	protected_V1.GET("/check-tickets-deadline", ticket.CheckTicketsDeadline)
	protected_V1.POST("/report", ticket.GenerateReport)
	protected_V1.GET("/tickets/:tracking_id", ticket.GetTicketByID)
	protected_V1.GET("/tickets/category", ticket.GetTicketsByCategory)
	protected_V1.GET("/tickets/status", ticket.GetTicketsByStatus)
	protected_V1.GET("/tickets/priority", ticket.GetTicketsByPriority)
	protected_V1.GET("/tickets-logs", ticket.GetTicketsLogs)
	protected_V1.GET("/tickets-date", ticket.GetTicketsByDateRange)
	protected_V1.DELETE("/tickets/:tracking_id", ticket.DeleteTicket)

	// Export
	protected_V1.POST("/export/tickets", export.ExportTickets)
	protected_V1.POST("/export/users", export.ExportUsers)
	protected_V1.POST("/export-logs", export.ExportLogs)

	// Auth
	protected_V1.POST("/logout", user.Logout)

	// Users
	protected_V1.GET("/users", user.GetAllUsers)
	protected_V1.GET("/email", user.GetEmail)
	protected_V1.POST("/users", user.Registration)
	protected_V1.GET("/users-logs", user.GetUsersLogs)
	protected_V1.POST("/edit-profile", user.EditProfile)
	protected_V1.POST("/edit-status-user", user.UpdateStatusUser)
	protected_V1.GET("/get-profile", user.GetProfile)

	// Notes
	protected_V1.GET("/notes", notes.GetAllNotes)
	protected_V1.GET("/notesByEmail", notes.FindByEmail)
	protected_V1.POST("/notes", notes.CreateNote)

	// Shifts
	protected_V1.POST("/shifts", shifts.AddShift)
	protected_V1.POST("/shifts/:id", shifts.UpdateShift)
	protected_V1.GET("/shifts/:id", shifts.GetShiftById)
	protected_V1.GET("/shifts", shifts.GetAllShifts)
	protected_V1.GET("/shifts-logs", shifts.GetShiftLogs)
	protected_V1.GET("/shifts-users", shifts.GetUserShifts)
	protected_V1.DELETE("/shifts/:id", shifts.DeleteShift)
	protected_V1.POST("/shifts/export", shifts.ExportShifts)

	// Products
	protected_V1.GET("/products", products.GetProducts)

	// Statistik
	protected_V1.POST("/statistik", statistik.GetStatistik)
}
