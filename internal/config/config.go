package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	JwtSecret        string
	DBConn           string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	WSAllowedOrigin  string
	InternalAPIToken string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Fatal("REDIS_DB должен быть числом")
	}

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		JwtSecret:        getEnv("JWT_SECRET", ""),
		DBConn:           getEnv("DB_URL", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          redisDB,
		WSAllowedOrigin:  getEnv("WS_ALLOWED_ORIGIN", "http://localhost:5173"),
		InternalAPIToken: getEnv("INTERNAL_API_TOKEN", ""),
	}

	if cfg.DBConn == "" {
		log.Fatal("DB_URL пустой")
	}
	if cfg.JwtSecret == "" {
		log.Fatal("JWT_SECRET пустой")
	}
	if cfg.InternalAPIToken == "" {
		log.Fatal("INTERNAL_API_TOKEN пустой")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
