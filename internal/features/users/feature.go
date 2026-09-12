package users

import (
	"github.com/gofiber/fiber/v3"

	"github.com/hnamhocit/fiber-template/internal/server"
)

type Feature struct {
	Handler *Handler
}

// NewModule creates the users feature with its dependencies.
func NewModule(deps server.Deps) *Feature {
	repo := NewRepository(deps.DB)
	svc := NewService(repo)
	return &Feature{Handler: NewHandler(svc)}
}

// RegisterRoutes attaches users routes to the given router.
func (f *Feature) RegisterRoutes(router fiber.Router) {
	g := router.Group("/users")
	g.Get("/", f.Handler.List)
	g.Get("/:id", f.Handler.Get)
	g.Post("/", f.Handler.Create)
	g.Put("/:id", f.Handler.Update)
	g.Delete("/:id", f.Handler.Delete)
}
