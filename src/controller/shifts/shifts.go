package shifts

import (
	"fmt"
	"log"
	"my-gin-project/src/types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "root:@tcp(db:3306)/commandcenter?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	fmt.Println("Connected to MySQL")
}

// @ GET User Shifts
func GetUserShifts(c *gin.Context) {
	var shifts []types.ShiftResponse

	if err := DB.Table("employee_shifts").
		Select("shifts.shift_name, employee_shifts.id, employee_shifts.shift_id, employee_shifts.user_email, employee_shifts.shift_date, users.name").
		Joins("JOIN shifts ON shifts.id = employee_shifts.shift_id").
		Joins("JOIN users ON users.email = employee_shifts.user_email").
		Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	formattedShifts := make([]struct {
		ID        uint   `json:"id"`
		ShiftId   uint   `json:"shift_id"`
		UserEmail string `json:"user_email"`
		UserName  string `json:"name"`
		ShiftName string `json:"shift_name"`
		ShiftDate string `json:"shift_date"`
	}, len(shifts))

	for i, shift := range shifts {
		formattedShifts[i] = struct {
			ID        uint   `json:"id"`
			ShiftId   uint   `json:"shift_id"`
			UserEmail string `json:"user_email"`
			UserName  string `json:"name"`
			ShiftName string `json:"shift_name"`
			ShiftDate string `json:"shift_date"`
		}{
			ID:        shift.ID,
			ShiftId:   shift.ShiftId,
			UserEmail: shift.UserEmail,
			UserName:  shift.Name,
			ShiftName: shift.ShiftName,
			ShiftDate: func() string {
				parsedDate, err := time.Parse(time.RFC3339, shift.ShiftDate)
				if err != nil {
					return shift.ShiftDate
				}
				return parsedDate.Format("2006-01-02")
			}(),
		}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "All shifts retrieved successfully",
		Data:    formattedShifts,
	})
}

// @ GET Shifts By Id
func GetShiftById(c *gin.Context) {
	type rawShiftResponse struct {
		ID        uint      `json:"id"`
		ShiftID   uint      `json:"shift_id"`
		ShiftName string    `json:"shift_name"`
		UserEmail string    `json:"user_email"`
		UserName  string    `json:"name"`
		ShiftDate time.Time `json:"shift_date"`
	}

	var rawShift rawShiftResponse

	if err := DB.Table("employee_shifts").
		Select("shifts.shift_name, employee_shifts.id, employee_shifts.shift_id, employee_shifts.user_email, employee_shifts.shift_date, users.name").
		Joins("JOIN shifts ON shifts.id = employee_shifts.shift_id").
		Joins("JOIN users ON users.email = employee_shifts.user_email").
		Where("employee_shifts.id = ?", c.Param("id")).
		First(&rawShift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Shift: " + err.Error(),
		})
		return
	}

	// Format shift_date ke string YYYY-MM-DD
	formattedShift := types.ShiftResponse{
		ID:        rawShift.ID,
		ShiftId:   rawShift.ShiftID,
		ShiftName: rawShift.ShiftName,
		UserEmail: rawShift.UserEmail,
		Name:      rawShift.UserName,
		ShiftDate: rawShift.ShiftDate.Format("2006-01-02"),
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Shift retrieved successfully",
		Data:    formattedShift,
	})
}

// @ GET ALL SHIFTS
func GetAllShifts(c *gin.Context) {
	var shifts []types.Shift

	if err := DB.Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "All shifts retrieved successfully",
		Data:    shifts,
	})
}

// @ GET Shift Logs
func GetShiftLogs(c *gin.Context) {
	var shiftLogs []types.ShiftLogs

	if err := DB.Find(&shiftLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All shift logs retrieved successfully", "data": shiftLogs})
}

// @ POST Shifts
func AddShift(c *gin.Context) {
	var bodyShift types.ShiftRequest

	if err := c.ShouldBindJSON(&bodyShift); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if bodyShift.ShiftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ShiftID cannot be empty"})
		return
	}

	if bodyShift.ShiftDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ShiftDate cannot be empty"})
		return
	}

	if bodyShift.UserEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserEmail cannot be empty"})
		return
	}

	shiftID, err := strconv.ParseUint(bodyShift.ShiftID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shift_id format"})
		return
	}

	employeeShift := types.EmployeeShift{
		UserEmail: bodyShift.UserEmail,
		ShiftID:   uint(shiftID),
		ShiftDate: bodyShift.ShiftDate,
	}

	if err := DB.Create(&employeeShift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Shift added successfully",
		"data":    employeeShift,
	})
}

// @ DELETE Shifts
func DeleteShift(c *gin.Context) {
	shiftID := c.Param("id")
	var shift types.EmployeeShift

	if shiftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shift ID cannot be empty"})
		return
	}

	if err := DB.Where("id = ?", shiftID).First(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Shift not found"})
		return
	}

	if err := DB.Delete(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shift deleted successfully", "data": shift})
}

// @ PUT Shifts
func UpdateShift(c *gin.Context) {
	shiftID := c.Param("id")
	var shift types.EmployeeShift

	if err := DB.Where("id = ?", shiftID).First(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var bodyShift types.ShiftRequest
	if err := c.ShouldBindJSON(&bodyShift); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate ShiftID - ensure it is not empty
	if bodyShift.UserEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User Email cannot be empty"})
		return
	}

	// Validate ShiftID - ensure it is not empty
	if bodyShift.ShiftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shift ID cannot be empty"})
		return
	}

	// Validate ShiftDate - ensure it is not empty
	if bodyShift.ShiftDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shift Date cannot be empty"})
		return
	}

	if bodyShift.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason cannot be empty"})
		return
	}

	shiftIDUint, err := strconv.ParseUint(bodyShift.ShiftID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shift_id format"})
		return
	}

	shift.ShiftID = uint(shiftIDUint)
	shift.ShiftDate = bodyShift.ShiftDate
	shift.UserEmail = bodyShift.UserEmail

	if err := DB.Save(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shiftLog := struct {
		ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
		UserEmail string `json:"user_email"`
		ShiftID   uint   `json:"shift_id"`
		ShiftDate string `json:"shift_date"`
		Reason    string `json:"reason"`
	}{
		UserEmail: bodyShift.UserEmail,
		ShiftID:   shift.ShiftID,
		ShiftDate: shift.ShiftDate,
		Reason:    bodyShift.Reason,
	}

	if err := DB.Table("shift_logs").Save(&shiftLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Shift updated successfully",
		Data:    shiftLog,
	})
}

// @POST Export Shift
func ExportShifts(c *gin.Context) {
	var input struct {
		Email     []string `json:"email" binding:"required"`
		StartDate string   `json:"start_date" binding:"required"`
		EndDate   string   `json:"end_date" binding:"required"`
	}

	type rawShiftResponse struct {
		ShiftID   *uint      `json:"shift_id"`
		ShiftName *string    `json:"shift_name"`
		UserEmail string     `json:"user_email"`
		Name      string     `json:"name"`
		ShiftDate *time.Time `json:"shift_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	startDate, err1 := time.Parse("2006-01-02", input.StartDate)
	endDate, err2 := time.Parse("2006-01-02", input.EndDate)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid date format, use YYYY-MM-DD",
			Data:    nil,
		})
		return
	}

	var rows []rawShiftResponse
	if err := DB.Table("users").
		Select("employee_shifts.shift_id, shifts.shift_name, users.email AS user_email, users.name, employee_shifts.shift_date").
		Joins("LEFT JOIN employee_shifts ON employee_shifts.user_email = users.email AND employee_shifts.shift_date BETWEEN ? AND ?", input.StartDate, input.EndDate).
		Joins("LEFT JOIN shifts ON employee_shifts.shift_id = shifts.id").
		Where("users.email IN ?", input.Email).
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed to get shifts: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Buat map [email][date] = shift
	shiftMap := map[string]map[string]rawShiftResponse{}
	for _, r := range rows {
		dateStr := ""
		if r.ShiftDate != nil {
			dateStr = r.ShiftDate.Format("2006-01-02")
		}
		if _, ok := shiftMap[r.UserEmail]; !ok {
			shiftMap[r.UserEmail] = map[string]rawShiftResponse{}
		}
		shiftMap[r.UserEmail][dateStr] = r
	}

	// Generate kombinasi email & tanggal
	formattedShifts := []map[string]interface{}{}
	for _, email := range input.Email {
		name := ""
		for _, r := range rows {
			if r.UserEmail == email {
				name = r.Name
				break
			}
		}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			shiftData, found := shiftMap[email][dateStr]

			if found && shiftData.ShiftID != nil {
				formattedShifts = append(formattedShifts, map[string]interface{}{
					"user_email": email,
					"name":       name,
					"shift_id":   *shiftData.ShiftID,
					"shift_name": *shiftData.ShiftName,
					"shift_date": dateStr,
				})
			} else {
				formattedShifts = append(formattedShifts, map[string]interface{}{
					"user_email": email,
					"name":       name,
					"shift_id":   nil,
					"shift_name": "Libur",
					"shift_date": dateStr,
				})
			}
		}
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: true,
		Message: "Successfully Get Shifts",
		Data:    formattedShifts,
	})
}

func init() {
	InitDB()
}
