package server

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"github.com/hnamhocit/fiber-template/internal/config"
)

// corsMiddleware returns CORS middleware configured for the given origins.
// In production with empty origins, it fails fast instead of falling back to wildcard (*).
func corsMiddleware(cfg *config.Config) fiber.Handler {
	if cfg.IsProduction() && len(cfg.CORSOrigins) == 0 {
		log.Fatal("CORS_ORIGINS must be set in production (comma-separated list of allowed origins)")
	}

	allowOrigins := cfg.CORSOrigins
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: !containsWildcard(allowOrigins),
		MaxAge:           86400,
	})
}

func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if strings.TrimSpace(o) == "*" {
			return true
		}
	}
	return false
}
