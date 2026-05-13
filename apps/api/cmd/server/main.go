package main

import (
	"log"
	"strings"

	"github.com/devnolife/umkm-api/internal/config"
	"github.com/devnolife/umkm-api/internal/database"
	"github.com/devnolife/umkm-api/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()
	database.Connect()

	// Optional auto-migrate in dev. Disable in production (Prisma is source of truth).
	if cfg.AppEnv == "development" {
		database.AutoMigrate()
	}

	app := fiber.New(fiber.Config{
		AppName:      "umkm-api",
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Production: strict allowlist only (config.Load() already refuses '*' outside dev).
			// Development: still respect CORS_ORIGINS but also allow common localhost ports for DX.
			for _, o := range cfg.CORSOrigins {
				if o == origin {
					return true
				}
			}
			if cfg.AppEnv == "development" {
				if strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") {
					return true
				}
			}
			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: false,
	}))

	handlers.RegisterRoutes(app)

	addr := ":" + cfg.Port
	log.Printf("[server] listening on http://localhost%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("[server] %v", err)
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": err.Error(),
	})
}
