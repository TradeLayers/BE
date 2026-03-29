package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	LogLevel     string
	LogDir       string
	InfoLogFile  string
	ErrorLogFile  string
	FinnhubAPIKey string
	FinnhubWSURL  string
}

func Load() *Config {
	loadDotenv()

	return &Config{
		AppPort:      getEnv("APP_PORT", "5000"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "tradelayers"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		LogLevel:     getEnv("LOG_LEVEL", "debug"),
		LogDir:       getEnv("LOG_DIR", "logs"),
		InfoLogFile:  getEnv("INFO_LOG_FILE", "app.log"),
		ErrorLogFile:  getEnv("ERROR_LOG_FILE", "error.log"),
		FinnhubAPIKey: getEnv("FINNHUB_API_KEY", ""),
		FinnhubWSURL:  getEnv("FINNHUB_WS_URL", "wss://ws.finnhub.io"),
	}
}

func loadDotenv() {
	paths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Overload(path) //nolint:errcheck // .env is optional
			return
		}
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
