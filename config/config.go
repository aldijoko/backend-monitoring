package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	Database DatabaseConfig

	JWTSecret        string
	JWTExpiry        string
	JWTRefreshExpiry string

	AdminUsername string
	AdminEmail    string
	AdminPassword string

	MediaMTXAPIURL     string
	MediaMTXHLSBaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8000"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "monitoring_cctv"),
			User:     getEnv("DB_USER", "cctv_user"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		JWTRefreshExpiry:   getEnv("JWT_REFRESH_EXPIRY", "168h"),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminEmail:         getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "changeme123"),
		MediaMTXAPIURL:     getEnv("MEDIAMTX_API_URL", "http://localhost:9997"),
		MediaMTXHLSBaseURL: getEnv("MEDIAMTX_HLS_BASE_URL", "http://localhost:8888"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
