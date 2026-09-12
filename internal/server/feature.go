package server

import "github.com/gofiber/fiber/v3"

// Feature is implemented by each feature module to register its own routes.
type Feature interface {
	RegisterRoutes(router fiber.Router)
}
