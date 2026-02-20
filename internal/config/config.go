package config

import "os"

type Config struct {
	Port      string
	JwtSecret string
	DBConn    string
}

func LoadConfig() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		JwtSecret: getEnv("JWT_SECRET", "Aaa123"), // Секрет для JWT
		// Строка подключения к Postgres
		DBConn: getEnv("DB_URL", "host=localhost user=postgres password=qwerty123 dbname=user_db port=5432 sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
