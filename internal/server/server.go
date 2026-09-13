package server

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/hnamhocit/fiber-template/internal/healthcheck"
	"github.com/hnamhocit/fiber-template/internal/httpx"
)

// New builds the Fiber app. Deps carries cfg + optional infra,
// so the limiter can share the Redis counter when configured.
func New(deps Deps, hc *healthcheck.Registry, features ...Feature) *fiber.App {
	cfg := deps.Cfg

	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ServerHeader: cfg.AppName,

		// Trust proxy headers when behind a load balancer (Fiber v3 API).
		TrustProxy: len(cfg.TrustedProxies) > 0,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: cfg.TrustedProxies,
			Private: len(cfg.TrustedProxies) == 0,
		},
		ProxyHeader: fiber.HeaderXForwardedFor,

		// Global error handler: every failure speaks the same JSON.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			msg := "internal server error"
			if cfg.AppEnv == "dev" {
				msg = err.Error()
			}
			return httpx.Error(c, code, msg)
		},
	})

	// Middleware stack (order matters)
	app.Use(recover.New())
	app.Use(requestid.New(requestid.Config{
		Generator: func() string { return uuid.New().String() },
		Header:    "X-Request-ID",
	}))
	app.Use(helmet.New(helmet.Config{
		CrossOriginEmbedderPolicy: "unsafe-none",
		// Security headers (C)
		PermissionPolicy:      "camera=(), microphone=(), geolocation=()",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' cdn.redoc.ly fonts.googleapis.com; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src fonts.gstatic.com; img-src 'self' data:;",
	}))
	app.Use(corsMiddleware(cfg))

	// Rate limiter: Redis-backed shared counter when Redis is configured,
	// in-memory fallback otherwise (single-node dev).
	var limiterStorage fiber.Storage
	if deps.Redis != nil {
		limiterStorage = newRedisStorage(deps.Redis.Client)
	}
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		Storage:    limiterStorage,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "too many requests, please try again later",
			})
		},
	}))

	// Access log with level derived from status code.
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()

		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency", time.Since(start).String(),
			"request_id", c.Get("X-Request-ID"),
		}
		switch {
		case status >= 500:
			slog.Error("http", attrs...)
		case status >= 400:
			slog.Warn("http", attrs...)
		default:
			slog.Info("http", attrs...)
		}
		return err
	})

	// Metrics endpoint (Prometheus format)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// API docs + health probes + features
	app.Use("/swagger", static.New("./docs"))
	app.Get("/docs", docsHandler)
	hc.Routes(app)

	v1 := app.Group("/v1")
	for _, m := range features {
		m.RegisterRoutes(v1)
	}

	return app
}

func docsHandler(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html")
	return c.SendString(docsHTML)
}

const docsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>API Reference</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Space+Grotesk:wght@500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
<style>
  html, body { height: 100%; margin: 0; padding: 0; }
  #redoc { min-height: 100vh; }
</style>
</head>
<body>
<div id="redoc"></div>
<script>
Redoc.init('/swagger/swagger.json', {
  scrollYOffset: 0,
  hideDownloadButton: false,
  expandResponses: '200',
  jsonSampleExpandLevel: 3,
  theme: {
    colors: {
      primary: { main: '#7c3aed' },
      success: { main: '#16a34a' },
      error:   { main: '#dc2626' },
      warning: { main: '#d97706' },
      text: { primary: '#1e1b2e', secondary: '#6b6580' },
      http: {
        get: '#16a34a', post: '#7c3aed', put: '#d97706',
        patch: '#d97706', delete: '#dc2626',
        basic: '#64748b', link: '#0ea5e9', head: '#0ea5e9'
      }
    },
    typography: {
      fontFamily: '"Inter", -apple-system, BlinkMacSystemFont, sans-serif',
      fontSize: '15px',
      lineHeight: '1.6',
      headings: {
        fontFamily: '"Space Grotesk", "Inter", sans-serif',
        fontWeight: '600',
        color: '#1e1b2e'
      },
      code: {
        fontFamily: '"JetBrains Mono", ui-monospace, monospace',
        fontSize: '13px',
        backgroundColor: 'rgba(124, 58, 237, 0.06)',
        color: '#5b21b6'
      },
      links: { color: '#7c3aed', hover: '#5b21b6' }
    },
    sidebar: {
      width: '280px',
      backgroundColor: '#f6f3fb',
      textColor: '#3f3a52',
      activeTextColor: '#7c3aed'
    },
    rightPanel: {
      backgroundColor: '#170f2b',
      textColor: '#e9e4f5',
      width: '40%'
    },
    codeBlock: { backgroundColor: '#1e1537' },
    schema: {
      linesColor: '#d8d0e8',
      nestedBackground: '#faf8fe',
      typeNameColor: '#6b6580',
      typeTitleColor: '#1e1b2e'
    }
  }
}, document.getElementById('redoc'));
</script>
</body>
</html>`
