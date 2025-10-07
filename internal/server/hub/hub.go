package hub

import (
	"RealtimeService/internal/domains/connection/model"
	"context"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
)

type Hub struct {
	mu          sync.RWMutex
	subscribers map[int]map[*websocket.Conn]struct{}
	logger      *slog.Logger
}

func New() *Hub {
	logger := slog.With(
		slog.String("object", "server"),
		slog.String("layer", "Hub"),
	)

	return &Hub{
		subscribers: make(map[int]map[*websocket.Conn]struct{}),
		logger:      logger,
	}
}

func (h *Hub) Subscribe(conn *websocket.Conn, unitIDs []int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, id := range unitIDs {
		if h.subscribers[id] == nil {
			h.subscribers[id] = make(map[*websocket.Conn]struct{})
		}
		h.subscribers[id][conn] = struct{}{}
	}

	h.logger.Info("Subscribed connection", slog.Any("unit_ids", unitIDs))
}

func (h *Hub) Unsubscribe(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	removed := 0
	for unitID, conns := range h.subscribers {
		if _, ok := conns[conn]; ok {
			delete(conns, conn)
			removed++
			if len(conns) == 0 {
				delete(h.subscribers, unitID)
			}
		}
	}

	h.logger.Info("Unsubscribed connection", slog.Int("removed_from_units", removed))
}

func (h *Hub) Broadcast(unitID int, data model.Data) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.subscribers[unitID]
	if !ok {
		return
	}

	for conn := range conns {
		if err := conn.WriteJSON(data); err != nil {
			h.logger.Error("Failed to send data", slog.Int("unit_id", unitID), slog.String("error", err.Error()))
			conn.Close()
		}
	}
}

func (h *Hub) Shutdown(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	count := 0
	for _, conns := range h.subscribers {
		for conn := range conns {
			conn.Close()
			count++
		}
	}
	h.subscribers = make(map[int]map[*websocket.Conn]struct{})

	h.logger.Info("Hub shutdown completed", slog.Int("connections_closed", count))
}
