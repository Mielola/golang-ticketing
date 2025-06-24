package routes

import (
	"my-gin-project/src/controller/category"
	"my-gin-project/src/controller/dashboard"
	"my-gin-project/src/controller/export"
	"my-gin-project/src/controller/notes"
	"my-gin-project/src/controller/places"
	"my-gin-project/src/controller/products"
	"my-gin-project/src/controller/role"
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
	v1.POST("/resend-otp", user.ResendOTP)
	v1.POST("/delete-otp", user.DeleteOTP)

	protected_V1 := v1.Group("/")
	protected_V1.Use(middleware.AuthMiddleware())

	// --------------------------------------------
	// @ V1 Routes
	// --------------------------------------------

	// Dashboard
	protected_V1.GET("/dashboard", dashboard.GetDashboard)
	protected_V1.POST("/get-data-form", dashboard.GetForm)

	// Tickets
	protected_V1.GET("/handover", shifts.GetHandoverTickets)
	protected_V1.POST("/tickets", ticket.AddTicket)
	protected_V1.POST("/tickets/excel", ticket.ImportTicketsArray)
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
	protected_V1.POST("/tickets/restore/:tracking_id", ticket.RestoreTicket)
	protected_V1.GET("/tickets/handover", ticket.HandOverTicket)
	protected_V1.GET("/tickets/temp", ticket.GetDeletedTickets)
	protected_V1.DELETE("/tickets/temp/:tracking_id", ticket.DeleteTempTickets)

	// Places
	protected_V1.POST("/places", places.CreatePlace)
	protected_V1.GET("/places", places.GetAllPlaces)
	protected_V1.DELETE("/places/:id", places.DeletePlace)
	protected_V1.POST("/places/:id", places.UpdatePlace)

	// Export
	protected_V1.POST("/export/tickets", export.ExportTickets)
	protected_V1.POST("/export/users", export.ExportUsers)
	protected_V1.POST("/export-logs", export.ExportLogs)

	// Role
	protected_V1.GET("/role", role.GetRole)

	// Auth
	protected_V1.POST("/logout", user.Logout)

	// Users
	protected_V1.GET("/users", user.GetAllUsers)
	protected_V1.GET("/email", user.GetEmail)
	protected_V1.POST("/users", user.Registration)
	protected_V1.POST("/users/:id", user.UpdateUsers)
	protected_V1.GET("/users-logs", user.GetUsersLogs)
	protected_V1.POST("/edit-profile", user.EditProfile)
	protected_V1.POST("/edit-status-user", user.UpdateStatusUser)
	protected_V1.GET("/get-profile", user.GetProfile)
	protected_V1.DELETE("/users/:id", user.DeleteUser)
	protected_V1.GET("/users/:id", user.GetUserByID)

	// Notes
	protected_V1.GET("/notes", notes.GetAllNotes)
	protected_V1.GET("/notesByEmail", notes.FindByEmail)
	protected_V1.POST("/notes", notes.CreateNote)
	protected_V1.POST("/notes/:id", notes.UpdateNote)
	protected_V1.DELETE("/notes/:id", notes.DeleteNote)

	// Shifts
	protected_V1.POST("/shifts", shifts.AddShift)
	protected_V1.POST("/shifts/:id", shifts.UpdateShift)
	protected_V1.GET("/shifts/:id", shifts.GetShiftById)
	protected_V1.GET("/shifts", shifts.GetAllShifts)
	protected_V1.GET("/shifts-logs", shifts.GetShiftLogs)
	protected_V1.GET("/shifts-users", shifts.GetUserShifts)
	protected_V1.DELETE("/shifts/:id", shifts.DeleteShift)
	protected_V1.POST("/shifts/export", shifts.ExportShifts)
	protected_V1.GET("/shifts-time", shifts.GetShiftTime)
	protected_V1.GET("/shifts-time/:id", shifts.GetShiftTimeById)
	protected_V1.POST("/shifts-time", shifts.CreateShiftTime)
	protected_V1.POST("/shifts-time/:id", shifts.UpdateShiftTime)
	protected_V1.DELETE("/shifts-time/:id", shifts.DeleteShiftTime)

	// Products
	protected_V1.GET("/list-products", products.GetProducts)
	protected_V1.GET("/products", products.GetAllProducts)
	protected_V1.GET("/products/:id", products.GetProductByID)
	protected_V1.POST("/products", products.CreateProducts)
	protected_V1.POST("/products/:id", products.UpdateProducts)
	protected_V1.DELETE("/products/:id", products.DeleteProducts)

	// Category
	protected_V1.GET("/category", category.GetCategory)
	protected_V1.GET("/category/:id", category.GetCategoryById)
	protected_V1.POST("/category", category.CreateCategory)
	protected_V1.POST("/category/:id", category.UpdateCategory)
	protected_V1.DELETE("/category/:id", category.DeleteCategory)

	// Statistik
	protected_V1.POST("/statistik", statistik.GetStatistik)
}
