package config

import "os"

type Config struct {
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	UploadDir  string
}

func Load() Config {
	return Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "app"),
		DBPassword: getEnv("DB_PASSWORD", "app_password"),
		DBName:     getEnv("DB_NAME", "lost_found"),
		JWTSecret:  getEnv("JWT_SECRET", "local-development-secret"),
		UploadDir:  getEnv("UPLOAD_DIR", "uploads"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
