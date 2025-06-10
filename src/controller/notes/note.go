package notes

import (
	"net/http"
	"time"

	"my-gin-project/src/database"
	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
)

// --------------------------------------------
// @ GET
// --------------------------------------------

func GetAllNotes(c *gin.Context) {
	DB := database.GetDB()

	var rawNotes []struct {
		ID        uint
		Title     string
		Content   string
		Email     string
		Name      string
		UpdatedAt time.Time
	}

	if err := DB.Table("notes").
		Select("notes.id, notes.title, notes.content, users.email, users.name, notes.created_at, notes.updated_at").
		Joins("LEFT JOIN users ON notes.user_email = users.email").
		Order("notes.updated_at DESC").
		Find(&rawNotes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: "Failed Get Data Notes : " + err.Error(),
		})
		return
	}

	// Format hasilnya
	var notes []map[string]interface{}
	for _, note := range rawNotes {
		notes = append(notes, map[string]interface{}{
			"id":         note.ID,
			"title":      note.Title,
			"content":    note.Content,
			"email":      note.Email,
			"name":       note.Name,
			"updated_at": note.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notes retrieved successfully",
		"data":    notes,
	})
}

func FindByEmail(c *gin.Context) {
	DB := database.GetDB()
	var user types.User
	var notes []types.NoteDetail
	var noteResponse types.NoteResponse

	email := c.Query("email")

	// Check if email is provided
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Email query parameter is required"})
		return
	}

	// Find user based on email
	if err := DB.Table("users").Select("email, name").Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// Find notes for the user
	if err := DB.Table("notes").
		Select("notes.id, notes.title, notes.content").
		Where("notes.user_email = ?", email).
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Construct the response
	noteResponse.Email = user.Email
	noteResponse.Name = user.Name
	noteResponse.Notes = notes

	// Return the response as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notes retrieved successfully",
		"data":    noteResponse,
	})
}

// --------------------------------------------
// @ POST
// --------------------------------------------

func CreateNote(c *gin.Context) {
	DB := database.GetDB()
	var input types.NoteBody
	var user struct {
		Email string
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: "Invalid Token",
		})
		return
	}

	if err := DB.Table("users").Select("email").Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	type NewNotes struct {
		Title     string `json:"title"`
		Content   string `json:"content"`
		UserEmail string `json:"user_email"`
	}

	formattedNotes := NewNotes{
		Title:     input.Title,
		Content:   input.Content,
		UserEmail: user.Email,
	}

	if err := DB.Table("notes").Create(&formattedNotes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, types.ResponseFormat{
		Success: true,
		Message: "Notes added successfully",
		Data:    input,
	})
}

func DeleteNote(c *gin.Context) {
	DB := database.GetDB()

	id := c.Param("id")

	if err := DB.Table("notes").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to delete note: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Note deleted successfully",
	})
}

func UpdateNote(c *gin.Context) {
	DB := database.GetDB()
	id := c.Param("id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := DB.Table("notes").Where("id = ?", id).Updates(map[string]interface{}{
		"title":   input.Title,
		"content": input.Content,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.ResponseFormat{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ResponseFormat{
		Success: false,
		Message: "Note updated successfully",
	})
}
