package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	JWTSecret         string
	SessionHours      int
	MinPasswordLength int
	DatabasePath      string
	CookieSecure      bool
	SnowflakeNode     int
}

func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret-change-me"),
		SessionHours:      getEnvInt("SESSION_HOURS", 24),
		MinPasswordLength: getEnvInt("MIN_PASSWORD_LENGTH", 3),
		DatabasePath:      getEnv("DATABASE_PATH", "./data/app.db"),
		CookieSecure:      getEnvBool("COOKIE_SECURE", false),
		SnowflakeNode:     getEnvInt("SNOWFLAKE_NODE", 0),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
