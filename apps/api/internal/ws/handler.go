package ws

import (
	"log"

	"github.com/devnolife/umkm-api/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Upgrade middleware to allow only WebSocket connections.
func Upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// Handler authenticates via ?token=... and joins rooms per client request.
// Client message protocol:
//
//	{"type":"join","room":"order:abc"}
//	{"type":"leave","room":"order:abc"}
func Handler(c *websocket.Conn) {
	defer func() {
		Default().LeaveAll(c)
		_ = c.Close()
	}()

	token := c.Query("token")
	if token == "" {
		_ = c.WriteJSON(fiber.Map{"event": "error", "data": "missing token"})
		return
	}
	claims, err := services.ParseToken(token)
	if err != nil {
		_ = c.WriteJSON(fiber.Map{"event": "error", "data": "invalid token"})
		return
	}

	// Auto-join user-specific room.
	userRoom := "user:" + claims.UserID
	Default().Join(userRoom, c)
	_ = c.WriteJSON(fiber.Map{"event": "connected", "data": fiber.Map{"userId": claims.UserID}})

	for {
		var msg struct {
			Type string `json:"type"`
			Room string `json:"room"`
		}
		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("[ws] read: %v", err)
			return
		}
		switch msg.Type {
		case "join":
			if msg.Room != "" {
				Default().Join(msg.Room, c)
			}
		case "leave":
			if msg.Room != "" {
				Default().Leave(msg.Room, c)
			}
		case "ping":
			_ = c.WriteJSON(fiber.Map{"event": "pong"})
		}
	}
}
