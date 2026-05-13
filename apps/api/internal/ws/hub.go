package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Hub manages WebSocket clients grouped by room (e.g., "user:{id}", "store:{id}", "order:{id}").
type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*websocket.Conn]bool
}

var defaultHub = &Hub{rooms: map[string]map[*websocket.Conn]bool{}}

func Default() *Hub { return defaultHub }

func (h *Hub) Join(room string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = map[*websocket.Conn]bool{}
	}
	h.rooms[room][c] = true
}

func (h *Hub) Leave(room string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.rooms[room]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.rooms, room)
		}
	}
}

func (h *Hub) LeaveAll(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for room, conns := range h.rooms {
		if conns[c] {
			delete(conns, c)
			if len(conns) == 0 {
				delete(h.rooms, room)
			}
		}
	}
}

// Emit broadcasts a payload as JSON to all clients in the room.
func (h *Hub) Emit(room, event string, payload any) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.rooms[room]))
	for c := range h.rooms[room] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	msg, err := json.Marshal(fiberMap{"event": event, "data": payload})
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}
	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[ws] write error: %v", err)
		}
	}
}

type fiberMap map[string]any
