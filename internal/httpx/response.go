package httpx

import (
	"github.com/gofiber/fiber/v3"

	"github.com/hnamhocit/fiber-template/internal/validator"
)

type response struct {
	Success bool                   `json:"success"`
	Data    any                    `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Errors  []validator.FieldError `json:"errors,omitempty"`
}

// OK returns a 200 response with data.
func OK(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(response{Success: true, Data: data})
}

// Created returns a 201 response with data.
func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(response{Success: true, Data: data})
}

// Error returns a single-error response (generic failures).
func Error(c fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(response{Success: false, Error: msg})
}

// ValidationError returns a 422 response with per-field error details.
// This is the standard response when request body validation fails.
func ValidationError(c fiber.Ctx, errors []validator.FieldError) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(response{
		Success: false,
		Error:   "validation failed",
		Errors:  errors,
	})
}

// NoContent returns a 204 response with no body.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
