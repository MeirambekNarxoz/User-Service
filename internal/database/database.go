package database

import (
	"log"
	"strings"
	"user-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(connStr string) *gorm.DB {
	if strings.TrimSpace(connStr) == "" {
		log.Fatal("DB_URL пустой: строка подключения к PostgreSQL не передана")
	}

	log.Println("DB CONN STR =", connStr)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных: ", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Ошибка миграции: ", err)
	}

	log.Println("PostgreSQL подключен")
	return db
}
