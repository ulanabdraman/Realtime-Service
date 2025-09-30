package hub

import (
	"RealtimeService/internal/domains/connection/model" // ты потом подставишь свой путь
	"context"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Hub struct {
	mu sync.RWMutex
	// unitID → соединения
	subscribers map[int]map[*websocket.Conn]struct{}
}

func New() *Hub {
	return &Hub{
		subscribers: make(map[int]map[*websocket.Conn]struct{}),
	}
}

// Subscribe добавляет conn ко всем нужным unitID
func (h *Hub) Subscribe(conn *websocket.Conn, unitIDs []int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, id := range unitIDs {
		if h.subscribers[id] == nil {
			h.subscribers[id] = make(map[*websocket.Conn]struct{})
		}
		h.subscribers[id][conn] = struct{}{}
	}
}

// Unsubscribe удаляет conn из всех подписок
func (h *Hub) Unsubscribe(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for unitID, conns := range h.subscribers {
		if _, ok := conns[conn]; ok {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(h.subscribers, unitID)
			}
		}
	}
}

// Broadcast рассылает сообщение всем, кто подписан на unitID
func (h *Hub) Broadcast(unitID int, data model.Data) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.subscribers[unitID]
	if !ok {
		return
	}

	for conn := range conns {
		err := conn.WriteJSON(data)
		if err != nil {
			log.Printf("hub: ошибка отправки: %v", err)
			conn.Close()
		}
	}
}

// Optional: Graceful shutdown
func (h *Hub) Shutdown(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, conns := range h.subscribers {
		for conn := range conns {
			conn.Close()
		}
	}
	h.subscribers = make(map[int]map[*websocket.Conn]struct{})
}
