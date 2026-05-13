package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lucsky/cuid"
)

func NewID() string { return cuid.New() }

// OK wraps the payload in the standard `{success:true, data:...}` envelope
// expected by the existing web/mobile clients.
func OK(c *fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{"success": true, "data": data})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": data})
}

// OKPaginated returns `{success:true, data, pagination}` for list endpoints.
func OKPaginated(c *fiber.Ctx, data any, pagination any) error {
	return c.JSON(fiber.Map{"success": true, "data": data, "pagination": pagination})
}

// Raw returns a custom body merged with `success:true`. Use for endpoints
// whose clients expect flat top-level fields (e.g. auth: {user, token}).
func Raw(c *fiber.Ctx, status int, body fiber.Map) error {
	body["success"] = true
	return c.Status(status).JSON(body)
}

func Err(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"success": false, "message": message})
}

func BadRequest(c *fiber.Ctx, message string) error {
	return Err(c, fiber.StatusBadRequest, message)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return Err(c, fiber.StatusUnauthorized, message)
}

func Forbidden(c *fiber.Ctx, message string) error {
	return Err(c, fiber.StatusForbidden, message)
}

func NotFound(c *fiber.Ctx, message string) error {
	return Err(c, fiber.StatusNotFound, message)
}

func Internal(c *fiber.Ctx, message string) error {
	return Err(c, fiber.StatusInternalServerError, message)
}
