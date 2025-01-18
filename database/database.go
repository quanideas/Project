package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var DB DbInstance

func ConnectDb() {
	// Connect to Postgres db via Gorm
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(mysql.Open(connectionStr), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true})

	// Handle connection fail and ping
	if err != nil {
		log.Fatal("Failed to connect.\n", err)
	} else {
		log.Println("connected")
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql db.\n", err)
	}

	if err := sqlDb.Ping(); err != nil {
		log.Fatal("Failed to ping Postgres db.\n", err)
	}

	// Set global db instance to connected db
	DB = DbInstance{
		Db: db,
	}
}
