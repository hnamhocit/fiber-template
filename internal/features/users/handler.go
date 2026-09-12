package users

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/hnamhocit/fiber-template/internal/httpx"
)

// Handler contains users HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a Handler with the given service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List returns all users.
func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		slog.Error("list users failed", "error", err)
		return httpx.Error(c, fiber.StatusInternalServerError, "failed to list users")
	}
	return httpx.OK(c, items)
}

// Get returns a user by ID.
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	item, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return httpx.Error(c, fiber.StatusNotFound, "user not found")
		}
		slog.Error("get user failed", "error", err, "id", id)
		return httpx.Error(c, fiber.StatusInternalServerError, "failed to get user")
	}
	return httpx.OK(c, item)
}

// Create creates a new user.
func (h *Handler) Create(c fiber.Ctx) error {
	req, err := httpx.Bind[CreateRequest](c)
	if err != nil {
		return err // already sent 400/422 response
	}

	item, err := h.svc.Create(c.Context(), req)
	if err != nil {
		slog.Error("create user failed", "error", err)
		return httpx.Error(c, fiber.StatusInternalServerError, "failed to create user")
	}
	return httpx.Created(c, item)
}

// Update updates a user.
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	req, err := httpx.Bind[UpdateRequest](c)
	if err != nil {
		return err
	}

	item, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return httpx.Error(c, fiber.StatusNotFound, "user not found")
		}
		slog.Error("update user failed", "error", err, "id", id)
		return httpx.Error(c, fiber.StatusInternalServerError, "failed to update user")
	}
	return httpx.OK(c, item)
}

// Delete deletes a user.
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return httpx.Error(c, fiber.StatusNotFound, "user not found")
		}
		slog.Error("delete user failed", "error", err, "id", id)
		return httpx.Error(c, fiber.StatusInternalServerError, "failed to delete user")
	}
	return httpx.NoContent(c)
}
