package database

import (
	"log"
	"strings"

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

	log.Println("PostgreSQL подключен")
	return db
}
