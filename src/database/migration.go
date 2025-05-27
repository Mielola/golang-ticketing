package database

import (
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

func MigrateDB() {
	db := GetDB()
	if db == nil {
		log.Fatal("Database belum terkoneksi, pastikan ConnectDB() sudah dipanggil")
	}

	err := db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Shift{},
		&models.EmployeeShift{},
		&models.ShiftLog{},
		&models.Category{},
		&models.Product{},
		&models.Ticket{},
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

	log.Println("Migrasi database selesai")
}
