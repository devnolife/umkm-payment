package middleware

import (
	"strings"

	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/services"
	"github.com/devnolife/umkm-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

const (
	CtxUserID = "userId"
	CtxRole   = "role"
)

// JWTAuth verifies Bearer token and stores userId & role in context.
func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			return utils.Unauthorized(c, "missing or malformed Authorization header")
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		claims, err := services.ParseToken(tokenStr)
		if err != nil {
			return utils.Unauthorized(c, "invalid or expired token")
		}
		c.Locals(CtxUserID, claims.UserID)
		c.Locals(CtxRole, claims.Role)
		return c.Next()
	}
}

// RequireRole guards endpoint by one or more user roles.
func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(CtxRole).(models.UserRole)
		if !ok {
			return utils.Unauthorized(c, "missing role")
		}
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return utils.Forbidden(c, "insufficient role")
	}
}

func GetUserID(c *fiber.Ctx) string {
	if v, ok := c.Locals(CtxUserID).(string); ok {
		return v
	}
	return ""
}

func GetRole(c *fiber.Ctx) models.UserRole {
	if v, ok := c.Locals(CtxRole).(models.UserRole); ok {
		return v
	}
	return ""
}
