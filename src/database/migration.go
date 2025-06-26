package database

import (
	"fmt"
	"log"
	"my-gin-project/src/models"

	"gorm.io/gorm"
)

func seedRoles(db *gorm.DB) {
	roles := []string{"admin", "pegawai"}

	for _, roleName := range roles {
		var count int64
		// Cek apakah role sudah ada
		db.Model(&models.Role{}).Where("name = ?", roleName).Count(&count)
		if count == 0 {
			// Insert jika belum ada
			db.Create(&models.Role{Name: roleName})
			log.Println("Role baru ditambahkan:", roleName)
		}
	}
}

func CleanInvalidCategoryIDs(db *gorm.DB) error {
	var invalidTickets []struct {
		ID         uint64
		CategoryId uint64
	}

	err := db.Raw(`
		SELECT t.id, t.category_id
		FROM tickets t
		LEFT JOIN category c ON t.category_id = c.id
		WHERE c.id IS NULL
	`).Scan(&invalidTickets).Error
	if err != nil {
		return fmt.Errorf("gagal scan ticket invalid: %w", err)
	}

	if len(invalidTickets) == 0 {
		fmt.Println("✔ Tidak ada tickets dengan category_id yang invalid")
		return nil
	}

	// (Opsional) Gunakan category_id default — misalnya ID = 1
	defaultCategoryID := uint64(1)

	// Update semua ticket invalid agar category_id-nya jadi default
	for _, ticket := range invalidTickets {
		if err := db.Model(&models.Ticket{}).
			Where("id = ?", ticket.ID).
			Update("category_id", defaultCategoryID).Error; err != nil {
			return fmt.Errorf("gagal update ticket ID %d: %w", ticket.ID, err)
		}
		fmt.Printf("✔ Ticket ID %d diperbaiki (category_id -> %d)\n", ticket.ID, defaultCategoryID)
	}

	return nil
}

func MigrateDB() {
	db := GetDB()
	if db == nil {
		log.Fatal("Database belum terkoneksi, pastikan ConnectDB() sudah dipanggil")
	}

	err := db.AutoMigrate(
		&models.TestUser{},
		&models.Role{},
		&models.User{},
		&models.TempTickets{},
		&models.TempUserTickets{},
		&models.Shift{},
		&models.EmployeeShift{},
		&models.ShiftLog{},
		&models.Category{},
		&models.Product{},
		&models.CategoryResolved{},
		&models.Ticket{},
		&models.Place{},
		&models.UserLog{},
		&models.UserTicket{},
		&models.ExportLog{},
		&models.Note{},
	)
	if err != nil {
		log.Fatal("Gagal melakukan migrasi:", err)
	}

	seedRoles(db)

	// Contoh safe menjalankan ALTER TABLE, cek error & jalankan hanya sekali dengan mekanisme migrasi versi
	if err := db.Exec("ALTER TABLE shifts MODIFY COLUMN start_time TIME").Error; err != nil {
		log.Println("ALTER TABLE shifts start_time mungkin sudah diubah atau error:", err)
	}
	if err := db.Exec("ALTER TABLE shifts MODIFY COLUMN end_time TIME").Error; err != nil {
		log.Println("ALTER TABLE shifts end_time mungkin sudah diubah atau error:", err)
	}
	if err := db.Exec("ALTER TABLE tickets MODIFY COLUMN waktu_masuk TIME").Error; err != nil {
		log.Println("ALTER TABLE tickets waktu_masuk mungkin sudah diubah atau error:", err)
	}
	if err := db.Exec("ALTER TABLE tickets MODIFY COLUMN waktu_respon TIME DEFAULT NULL").Error; err != nil {
		log.Println("ALTER TABLE tickets waktu_respon mungkin sudah diubah atau error:", err)
	}
	if err := db.Exec("ALTER TABLE temp_tickets MODIFY COLUMN waktu_masuk TIME").Error; err != nil {
		log.Println("ALTER TABLE temp_tickets waktu_masuk mungkin sudah diubah atau error:", err)
	}
	if err := db.Exec("ALTER TABLE temp_tickets MODIFY COLUMN waktu_respon TIME DEFAULT NULL").Error; err != nil {
		log.Println("ALTER TABLE temp_tickets waktu_respon mungkin sudah diubah atau error:", err)
	}

	log.Println("Migrasi database selesai")
}
