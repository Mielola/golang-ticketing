package database

import (
	"log"
	"my-gin-project/src/models"
)

func MigrateDB() {
	db := GetDB()
	if db == nil {
		log.Fatal("Database belum terkoneksi, pastikan ConnectDB() sudah dipanggil")
	}

	err := DB.AutoMigrate(
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
	db.Exec("ALTER TABLE shifts MODIFY COLUMN start_time TIME")
	db.Exec("ALTER TABLE shifts MODIFY COLUMN end_time TIME")
	db.Exec("ALTER TABLE tickets MODIFY COLUMN waktu_masuk TIME")
	db.Exec("ALTER TABLE tickets MODIFY COLUMN waktu_respon TIME DEFAULT NULL")
	if err != nil {
		log.Fatal("Gagal melakukan migrasi:", err)
	}
	log.Println("Migrasi database selesai")
}
