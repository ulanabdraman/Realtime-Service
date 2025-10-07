package handler

import (
	"log/slog"
	"net/http"

	"RealtimeService/internal/domains/connection/repository"
	"RealtimeService/internal/domains/connection/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	authRepo repository.AuthRepository
	useCase  usecase.ConnectionUseCase
	logger   *slog.Logger
}

func NewWebSocketHandler(
	auth repository.AuthRepository,
	useCase usecase.ConnectionUseCase,
) *WebSocketHandler {
	logger := slog.With(
		slog.String("object", "connection"),
		slog.String("layer", "WShandler"),
	)

	return &WebSocketHandler{
		authRepo: auth,
		useCase:  useCase,
		logger:   logger,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {

		return true
	},
}

func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.logger.Warn("Missing token in query")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.authRepo.ValidateToken(c, token)
	if err != nil {
		h.logger.Warn("Invalid token", slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established", slog.Int64("user_id", claims.UserID))

	if err := h.useCase.HandleConnection(c, claims.UserID, conn); err != nil {
		h.logger.Error("HandleConnection failed", slog.Int64("user_id", claims.UserID), slog.String("error", err.Error()))
	}
}
