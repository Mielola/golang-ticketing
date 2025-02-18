package notes

import (
	"os"
	"fmt"
	"log"
	"net/http"

	"my-gin-project/src/types"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("root:@tcp(%s:3306)/commandcenter?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DB_HOST"))
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	fmt.Println("DB_HOST:", os.Getenv("DB_HOST"))
}

// --------------------------------------------
// @ GET
// --------------------------------------------

func GetAllNotes(c *gin.Context) {
	var notes []struct {
		ID      uint   `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}

	if err := DB.Table("note").
		Select("note.id, note.title, note.content, users.email, users.name").
		Joins("JOIN users ON note.user_email = users.email").
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All notes retrieved successfully", "data": notes})
}

// func GetAllNotes(c *gin.Context) {
// 	var notes []types.Note
// 	var notesResponse []types.NoteResponse
// 	if err := DB.Table("note").
// 		Select("note.id, note.title, note.content, users.email, users.name").
// 		Joins("JOIN users ON note.user_id = users.id").
// 		Find(&notes).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// Map untuk mengelompokkan notes berdasarkan email
// 	groupedNotes := make(map[string]types.NoteResponse)
// 	// Iterasi melalui hasil query
// 	for _, note := range notes {
// 		key := note.Email // Gunakan email sebagai kunci unik
// 		// Jika belum ada entri untuk email ini, buat entri baru
// 		if _, exists := groupedNotes[key]; !exists {
// 			groupedNotes[key] = types.NoteResponse{
// 				Email: note.Email,
// 				Name:  note.Name,
// 				Notes: []types.NoteDetail{},
// 			}
// 		}
// 		// Tambahkan catatan ke dalam array Notes
// 		noteResponse := groupedNotes[key]
// 		noteResponse.Notes = append(noteResponse.Notes, types.NoteDetail{
// 			ID:      note.ID,
// 			Title:   note.Title,
// 			Content: note.Content,
// 		})
// 		groupedNotes[key] = noteResponse
// 	}
// 	// Konversi map menjadi slice untuk response
// 	for _, response := range groupedNotes {
// 		notesResponse = append(notesResponse, response)
// 	}
// 	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All notes retrieved successfully", "data": notesResponse})
// }

func FindByEmail(c *gin.Context) {
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
	if err := DB.Table("note").
		Select("note.id, note.title, note.content").
		Where("note.user_email = ?", email).
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
	var input types.NoteBody

	// switch {
	// case input.Title == "":
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
	// 	return
	// case input.Content == "":
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
	// 	return
	// case input.Email == "":
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
	// 	return
	// }

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Table("note").Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": true, "message": "Notes added successfully", "users": input})
}

func init() {
	InitDB()
}
