package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	AppEnv          string
	CORSOrigins     []string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiresHours int

	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
}

var cfg *Config

func Load() *Config {
	if cfg != nil {
		return cfg
	}
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env not found, using OS env")
	}

	cfg = &Config{
		Port:                 getEnv("PORT", "4000"),
		AppEnv:               getEnv("APP_ENV", "development"),
		CORSOrigins:          splitCSV(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8081")),
		DatabaseURL:          mustEnv("DATABASE_URL"),
		JWTSecret:            mustEnv("JWT_SECRET"),
		JWTExpiresHours:      getEnvInt("JWT_EXPIRES_HOURS", 168),
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransIsProduction: getEnv("MIDTRANS_IS_PRODUCTION", "false") == "true",
	}

	if cfg.AppEnv != "development" {
		for _, o := range cfg.CORSOrigins {
			if o == "*" {
				log.Fatalf("[config] CORS_ORIGINS cannot contain '*' when APP_ENV=%s; set an explicit allowlist", cfg.AppEnv)
			}
		}
		if len(cfg.CORSOrigins) == 0 {
			log.Println("[config] WARNING: CORS_ORIGINS is empty in production — all cross-origin requests will be blocked")
		}
	}

	return cfg
}

func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("[config] required env %s is empty", key)
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
