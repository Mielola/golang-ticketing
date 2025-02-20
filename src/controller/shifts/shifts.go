package shifts

import (
	"fmt"
	"log"
	"my-gin-project/src/types"
	"net/http"
	"strconv"

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
		Select("shifts.shift_name, employee_shifts.id, employee_shifts.user_email, employee_shifts.shift_date").
		Joins("JOIN shifts ON shifts.id = employee_shifts.shift_id").
		Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All shifts retrieved successfully", "data": shifts})
}

// @ GET ALL SHIFTS
func GetAllShifts(c *gin.Context) {
	var shifts []types.Shift

	if err := DB.Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All shifts retrieved successfully", "data": shifts})
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

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Shift updated successfully", "data": shift})
}

func init() {
	InitDB()
}
