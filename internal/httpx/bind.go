package httpx

import (
	"github.com/gofiber/fiber/v3"

	"github.com/hnamhocit/fiber-template/internal/validator"
)

// Bind parses the request body into T and validates it.
// Returns (T, nil) on success, or (zero-value, error-response) on failure.
//
// Usage in handler:
//
//	req, err := httpx.Bind[CreateRequest](c)
//	if err != nil {
//		return err  // already sent 400/422 response
//	}
//	// req is validated, safe to use
func Bind[T any](c fiber.Ctx) (T, error) {
	var req T
	if err := c.Bind().Body(&req); err != nil {
		return req, Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if errs := validator.Validate(&req); errs != nil {
		return req, ValidationError(c, errs)
	}
	return req, nil
}
