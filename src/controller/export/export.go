package export

import (
	"fmt"
	"log"
	"my-gin-project/src/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "root:@tcp(localhost:3306)/commandcenter?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	fmt.Println("Connected to MySQL")
}

func init() {
	InitDB()
}

func ExportTickets(c *gin.Context) {
	var input struct {
		StartDate    string   `json:"start_date" binding:"required"`
		EndDate      string   `json:"end_date" binding:"required"`
		StartTime    string   `json:"start_time" binding:"required"`
		EndTime      string   `json:"end_time" binding:"required"`
		ProductsName string   `json:"products_name" binding:"required"`
		Email        []string `json:"email" binding:"required"`
	}

	var tickets []types.TicketsResponseAll

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid StartDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid EndDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if input.StartDate == "" || input.EndDate == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Start date, end date, status are required",
			Data:    nil,
		})
		return
	}

	var startDateTime = input.StartDate + " " + input.StartTime
	var endDateTime = input.EndDate + " " + input.EndTime

	if err := DB.Table("tickets").Select("*, tickets.status").
		Joins("LEFT JOIN users ON tickets.user_email = users.email").
		Where("users.email IN (?) AND tickets.created_at BETWEEN ? AND ? AND products_name = ?", input.Email, startDateTime, endDateTime, input.ProductsName).
		Order("priority DESC").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	type StatusTimestamp struct {
		Status    string     `json:"status"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	type FormattedTicket struct {
		CreatedAt         string               `json:"created_at"`
		CreatedTicket     types.TicketsCreator `json:"created_ticket"`
		FirstResponAt     string               `json:"firs_respon_at"`
		FirstResponInDate string               `json:"first_respon_in_date"`
		FirstResponInTime string               `json:"first_respon_in_time"`
		LastReplier       interface{}          `json:"last_replier"`
		NoClient          string               `json:"no_client"`
		Priority          string               `json:"priority"`
		ProductsName      string               `json:"products_name"`
		SolvedTime        *string              `json:"solved_time"`
		Status            string               `json:"status"`
		TrackingID        string               `json:"tracking_id"`
		StatusTimestamps  []StatusTimestamp    `json:"status_timestamps"`
	}

	baseURL := "http://localhost:8080/storage/images/"
	var formattedTickets []FormattedTicket

	for _, ticket := range tickets {
		var ticketCreator types.TicketsCreator

		if err := DB.Table("users").
			Select("email, name, avatar").
			Where("email = ?", ticket.UserEmail).
			Scan(&ticketCreator).Error; err != nil {
			return
		}

		if ticketCreator.Avatar != "" {
			ticketCreator.Avatar = baseURL + ticketCreator.Avatar
		}

		var lastReply struct {
			UserEmail string    `json:"user_email"`
			NewStatus string    `json:"new_status"`
			UpdatedAt time.Time `json:"update_at"`
		}

		if err := DB.Table("user_tickets").
			Select("user_email, new_status, update_at").
			Where("tickets_id = ?", ticket.TrackingID).
			Order("update_at DESC").
			Limit(1).
			Scan(&lastReply).Error; err != nil {
			lastReply = struct {
				UserEmail string    `json:"user_email"`
				NewStatus string    `json:"new_status"`
				UpdatedAt time.Time `json:"update_at"`
			}{}
		}

		// Get last replier's information
		var lastReplier *struct {
			Email  string `json:"email"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}

		if lastReply.UserEmail != "" {
			var replierInfo struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			}

			if err := DB.Table("users").
				Select("email, name, avatar").
				Where("email = ?", lastReply.UserEmail).
				Scan(&replierInfo).Error; err == nil {

				if replierInfo.Avatar != "" {
					replierInfo.Avatar = baseURL + replierInfo.Avatar
				}
				lastReplier = &replierInfo
			}
		}

		parsedHariRespon, err := time.Parse("2006-01-02T15:04:05-07:00", ticket.HariRespon)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: "Failed to parse HariRespon: " + err.Error(),
				Data:    nil,
			})
			return
		}

		// Get status timestamps using the provided SQL query
		var statusTimestamps []StatusTimestamp
		rows, err := DB.Raw(`
			SELECT s.status, MIN(ut.update_at) AS update_at
			FROM (
				SELECT 'New' AS status
				UNION ALL
				SELECT 'On Progress'
				UNION ALL
				SELECT 'Resolved'
			) AS s
			LEFT JOIN user_tickets ut ON ut.new_status = s.status AND ut.tickets_id = ?
			GROUP BY s.status
			ORDER BY FIELD(s.status, 'New', 'On Progress', 'Resolved')
		`, ticket.TrackingID).Rows()

		if err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: "Failed to get status timestamps: " + err.Error(),
				Data:    nil,
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var status string
			var updatedAt *time.Time
			if err := rows.Scan(&status, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{
					Success: false,
					Message: "Failed to scan status timestamps: " + err.Error(),
					Data:    nil,
				})
				return
			}
			statusTimestamps = append(statusTimestamps, StatusTimestamp{
				Status:    status,
				UpdatedAt: updatedAt,
			})
		}

		// Append to formattedTickets
		formattedTickets = append(formattedTickets, FormattedTicket{
			TrackingID:        ticket.TrackingID,
			NoClient:          ticket.NoWhatsapp,
			CreatedTicket:     ticketCreator,
			ProductsName:      ticket.ProductsName,
			Status:            ticket.Status,
			CreatedAt:         ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			Priority:          ticket.Priority,
			FirstResponAt:     parsedHariRespon.Format("2006-01-02 15:04:05"),
			FirstResponInDate: parsedHariRespon.Format("2006-01-02"),
			FirstResponInTime: ticket.WaktuRespon,
			LastReplier:       lastReplier,
			SolvedTime:        ticket.SolvedTime,
			StatusTimestamps:  statusTimestamps,
		})
	}

	data := formattedTickets
	if data == nil {
		data = []FormattedTicket{}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data:    data,
	})
}

func ExportUsers(c *gin.Context) {
	var input struct {
		StartDate    string   `json:"start_date" binding:"required"`
		EndDate      string   `json:"end_date" binding:"required"`
		StartTime    string   `json:"start_time" binding:"required"`
		EndTime      string   `json:"end_time" binding:"required"`
		ProductsName string   `json:"products_name" binding:"required"`
		Email        []string `json:"email" binding:"required"`
	}

	var tickets []types.TicketsResponseAll

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid StartDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid EndDate format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	if input.StartDate == "" || input.EndDate == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Start date, end date, status are required",
			Data:    nil,
		})
		return
	}

	var startDateTime = input.StartDate + " " + input.StartTime
	var endDateTime = input.EndDate + " " + input.EndTime

	fmt.Println(startDateTime + " " + endDateTime)

	if err := DB.Table("tickets").Select("*, tickets.status").
		Joins("LEFT JOIN users ON tickets.user_email = users.email").
		Where("users.email IN (?) AND tickets.created_at BETWEEN ? AND ? AND products_name = ?", input.Email, startDateTime, endDateTime, input.ProductsName).
		Order("users.email").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	type StatusTimestamp struct {
		Status    string     `json:"status"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	type FormattedTicket struct {
		CreatedAt         string               `json:"created_at"`
		CreatedTicket     types.TicketsCreator `json:"created_ticket"`
		FirstResponAt     string               `json:"firs_respon_at"`
		FirstResponInDate string               `json:"first_respon_in_date"`
		FirstResponInTime string               `json:"first_respon_in_time"`
		LastReplier       interface{}          `json:"last_replier"`
		NoClient          string               `json:"no_client"`
		Priority          string               `json:"priority"`
		ProductsName      string               `json:"products_name"`
		SolvedTime        *string              `json:"solved_time"`
		Status            string               `json:"status"`
		TrackingID        string               `json:"tracking_id"`
		StatusTimestamps  []StatusTimestamp    `json:"status_timestamps"`
	}

	baseURL := "http://localhost:8080/storage/images/"
	var formattedTickets []FormattedTicket

	for _, ticket := range tickets {
		var ticketCreator types.TicketsCreator

		if err := DB.Table("users").
			Select("email, name, avatar").
			Where("email = ?", ticket.UserEmail).
			Scan(&ticketCreator).Error; err != nil {
			return
		}

		if ticketCreator.Avatar != "" {
			ticketCreator.Avatar = baseURL + ticketCreator.Avatar
		}

		var lastReply struct {
			UserEmail string    `json:"user_email"`
			NewStatus string    `json:"new_status"`
			UpdatedAt time.Time `json:"update_at"`
		}

		if err := DB.Table("user_tickets").
			Select("user_email, new_status, update_at").
			Where("tickets_id = ?", ticket.TrackingID).
			Order("update_at DESC").
			Limit(1).
			Scan(&lastReply).Error; err != nil {
			lastReply = struct {
				UserEmail string    `json:"user_email"`
				NewStatus string    `json:"new_status"`
				UpdatedAt time.Time `json:"update_at"`
			}{}
		}

		// Get last replier's information
		var lastReplier *struct {
			Email  string `json:"email"`
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}

		if lastReply.UserEmail != "" {
			var replierInfo struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			}

			if err := DB.Table("users").
				Select("email, name, avatar").
				Where("email = ?", lastReply.UserEmail).
				Scan(&replierInfo).Error; err == nil {

				if replierInfo.Avatar != "" {
					replierInfo.Avatar = baseURL + replierInfo.Avatar
				}
				lastReplier = &replierInfo
			}
		}

		parsedHariRespon, err := time.Parse("2006-01-02T15:04:05-07:00", ticket.HariRespon)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: "Failed to parse HariRespon: " + err.Error(),
				Data:    nil,
			})
			return
		}

		var statusTimestamps []StatusTimestamp
		rows, err := DB.Raw(`
			SELECT s.status, MIN(ut.update_at) AS update_at
			FROM (
				SELECT 'New' AS status
				UNION ALL
				SELECT 'On Progress'
				UNION ALL
				SELECT 'Resolved'
			) AS s
			LEFT JOIN user_tickets ut ON ut.new_status = s.status AND ut.tickets_id = ?
			GROUP BY s.status
			ORDER BY FIELD(s.status, 'New', 'On Progress', 'Resolved')
		`, ticket.TrackingID).Rows()

		if err != nil {
			c.JSON(http.StatusInternalServerError, types.ResponseFormat{
				Success: false,
				Message: "Failed to get status timestamps: " + err.Error(),
				Data:    nil,
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var status string
			var updatedAt *time.Time
			if err := rows.Scan(&status, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, types.ResponseFormat{
					Success: false,
					Message: "Failed to scan status timestamps: " + err.Error(),
					Data:    nil,
				})
				return
			}
			statusTimestamps = append(statusTimestamps, StatusTimestamp{
				Status:    status,
				UpdatedAt: updatedAt,
			})
		}

		// Kemudian saat append
		formattedTickets = append(formattedTickets, FormattedTicket{
			TrackingID:        ticket.TrackingID,
			NoClient:          ticket.NoWhatsapp,
			CreatedTicket:     ticketCreator,
			ProductsName:      ticket.ProductsName,
			Status:            ticket.Status,
			CreatedAt:         ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			Priority:          ticket.Priority,
			FirstResponAt:     parsedHariRespon.Format("2006-01-02 15:04:05"),
			FirstResponInDate: parsedHariRespon.Format("2006-01-02"),
			FirstResponInTime: ticket.WaktuRespon,
			LastReplier:       lastReplier,
			SolvedTime:        ticket.SolvedTime,
			StatusTimestamps:  statusTimestamps,
		})
	}

	data := formattedTickets
	if data == nil {
		data = []FormattedTicket{}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data:    data,
	})
}

func ExportLogs(c *gin.Context) {
	var input struct {
		HistoryType string `json:"history_type"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	var exportLogs []struct {
		FileName  string    `json:"file_name"`
		UserEmail string    `json:"user_email"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := DB.Table("export_logs AS el").
		Select("*").
		Where("el.history_type = ?", input.HistoryType).
		Scan(&exportLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data Export Logs" + err.Error(),
		})
		return
	}

	var formattedLogs []map[string]interface{}

	for _, logs := range exportLogs {
		formattedLogs = append(formattedLogs, map[string]interface{}{
			"file_name":  logs.FileName,
			"user_email": logs.UserEmail,
			"created_at": logs.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	data := formattedLogs
	if data == nil {
		data = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Success Get Data Export Logs",
		Data:    data,
	})
}
