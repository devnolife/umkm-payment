package handlers

import (
	"github.com/devnolife/umkm-api/internal/middleware"
	"github.com/devnolife/umkm-api/internal/models"
	"github.com/devnolife/umkm-api/internal/ws"
	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Auth
	auth := api.Group("/auth")
	auth.Post("/register", Register)
	auth.Post("/login", Login)
	auth.Post("/refresh", middleware.JWTAuth(), Refresh)
	auth.Get("/profile", middleware.JWTAuth(), Profile)
	auth.Put("/profile", middleware.JWTAuth(), UpdateProfile)

	// Stores (public read, seller/admin write)
	stores := api.Group("/stores")
	stores.Get("/", ListStores)
	stores.Get("/mine", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller), GetMyStore)
	stores.Get("/:id", GetStore)
	stores.Get("/:id/menu", GetStoreMenu)
	stores.Get("/:id/queue", GetStoreQueue)
	stores.Post("/", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller), CreateStore)
	stores.Put("/:id", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller, models.RoleAdmin), UpdateStore)
	stores.Patch("/:id/toggle", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller, models.RoleAdmin), ToggleStore)

	// Categories
	cat := api.Group("/categories", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller, models.RoleAdmin))
	cat.Post("/", CreateCategory)
	cat.Put("/:id", UpdateCategory)
	cat.Delete("/:id", DeleteCategory)

	// Menu items
	menu := api.Group("/menu", middleware.JWTAuth(), middleware.RequireRole(models.RoleSeller, models.RoleAdmin))
	menu.Post("/", CreateMenuItem)
	menu.Put("/:id", UpdateMenuItem)
	menu.Delete("/:id", DeleteMenuItem)

	// Orders
	orders := api.Group("/orders", middleware.JWTAuth())
	orders.Post("/", middleware.RequireRole(models.RoleBuyer), CreateOrder)
	orders.Get("/", ListOrders)
	orders.Get("/:id", GetOrder)
	orders.Patch("/:id/status", UpdateOrderStatus)

	// Payments
	pay := api.Group("/payments")
	pay.Post("/create", middleware.JWTAuth(), middleware.RequireRole(models.RoleBuyer), CreatePayment)
	pay.Post("/webhook", PaymentWebhook) // no auth — verified via signature

	// Admin
	admin := api.Group("/admin", middleware.JWTAuth(), middleware.RequireRole(models.RoleAdmin))
	admin.Get("/stats", AdminStats)
	admin.Get("/users", AdminListUsers)
	admin.Patch("/users/:id/toggle", AdminToggleUser)

	// WebSocket
	app.Use("/ws", ws.Upgrade)
	app.Get("/ws", fws.New(ws.Handler))
}
