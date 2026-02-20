package database

import (
	"log"
	"user-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(connStr string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных: ", err)
	}

	// Автоматическое создание таблицы пользователей
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Ошибка миграции: ", err)
	}

	return db
}
