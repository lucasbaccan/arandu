package config

import (
	"os"
	"strconv"
)

const (
	DefaultHost = "0.0.0.0"
	DefaultPort = "8080"
	DefaultDB   = "./data/app.db"
)

type Config struct {
	Host              string
	Port              string
	JWTSecret         string
	SessionHours      int
	MinPasswordLength int
	DatabasePath      string
	CookieSecure      bool
	CookieSameSite    string
	// Nenhuma das duas envs veio preenchida: em vez de chutar uma política fixa,
	// o cookie de sessão se adapta a cada requisição (ver cookieAttrs em
	// internal/api). Chutar errado aqui é caro — Secure/None sobre HTTP puro faz
	// o navegador descartar o cookie em silêncio (login "funciona" e a próxima
	// requisição volta sem sessão), e Lax sem Secure derruba o cenário
	// cross-site (frontend na Vercel + backend em outro domínio).
	CookieAuto bool
	CORSOrigins       string
	SnowflakeNode     int
	ViteDevURL        string
	FrontendDir       string
}

func Load() Config {
	return Config{
		Host:              getEnv("HOST", DefaultHost),
		Port:              getEnv("PORT", DefaultPort),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret-change-me"),
		SessionHours:      getEnvInt("SESSION_HOURS", 24),
		MinPasswordLength: getEnvInt("MIN_PASSWORD_LENGTH", 3),
		DatabasePath:      getEnv("DATABASE_PATH", DefaultDB),
		CookieSecure:      getEnvBool("COOKIE_SECURE", false),
		CookieSameSite:    getEnv("COOKIE_SAMESITE", "lax"),
		CookieAuto:        os.Getenv("COOKIE_SECURE") == "" && os.Getenv("COOKIE_SAMESITE") == "",
		CORSOrigins:       getEnv("CORS_ORIGINS", ""),
		SnowflakeNode:     getEnvInt("SNOWFLAKE_NODE", 0),
		ViteDevURL:        getEnv("VITE_DEV_URL", ""),
		FrontendDir:       getEnv("FRONTEND_DIR", ""),
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
