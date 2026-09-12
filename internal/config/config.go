package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	AppEnv   string // development | production
	Port     string
	LogLevel string

	DatabaseURL string

	RedisURL string

	CORSOrigins    []string
	TrustedProxies []string `env:"TRUSTED_PROXIES" envSeparator:","`
}

var (
	instance *Config
	once     sync.Once
)

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// Load returns the singleton config instance.
func Load() *Config {
	once.Do(func() {
		instance = loadFromEnv()
	})
	return instance
}

func loadFromEnv() *Config {
	appEnv := os.Getenv("APP_ENV")

	if appEnv != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		AppName:        getEnv("APP_NAME", "fiber-app"),
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", ":8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		CORSOrigins:    parseOrigins(getEnv("CORS_ORIGINS", "")),
		TrustedProxies: parseCommaSeparated(getEnv("TRUSTED_PROXIES", "")),
	}

	validateRequired([]string{"DATABASE_URL"})
	return cfg
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

func validateRequired(keys []string) {
	for _, key := range keys {
		if os.Getenv(key) == "" {
			fmt.Fprintf(os.Stderr, "config: %s is required\n", key)
			os.Exit(1)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseCommaSeparated splits a comma-separated env value into a trimmed slice.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
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
