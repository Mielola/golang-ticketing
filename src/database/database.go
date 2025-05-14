package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	DBUser     = "root"
	DBPassword = ""
	DBHost     = "localhost"
	DBPort     = "3306"
	DBName     = "commandcenter4"
)

func InitDB() {
	// Step 1: Connect without database name
	dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, DBPort)
	tempDB, err := gorm.Open(mysql.Open(dsnNoDB), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to MySQL server: %v", err)
	}

	// Step 2: Create the database if it doesn't exist
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", DBName)
	if err := tempDB.Exec(createDBSQL).Error; err != nil {
		log.Fatalf("Could not create database: %v", err)
	}
	log.Printf("Database %s ensured to exist.\n", DBName)

	// Step 3: Connect to the actual database
	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DBUser, DBPassword, DBHost, DBPort, DBName)
	DB, err = gorm.Open(mysql.Open(dsnWithDB), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database %s: %v", DBName, err)
	}
	log.Println("Connected to MySQL database:", DBName)
}

func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database belum terkoneksi, pastikan InitDB() sudah dipanggil")
	}
	return DB
}
