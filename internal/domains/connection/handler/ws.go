package handler

import (
	"log"
	"net/http"

	"RealtimeService/internal/domains/connection/repository"
	"RealtimeService/internal/domains/connection/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	authRepo repository.AuthRepository
	useCase  usecase.ConnectionUseCase
}

func NewWebSocketHandler(
	auth repository.AuthRepository,
	useCase usecase.ConnectionUseCase,
) *WebSocketHandler {
	return &WebSocketHandler{
		authRepo: auth,
		useCase:  useCase,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: заменить на строгую проверку
		return true
	},
}

func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// Извлекаем токен
	token := c.Query("token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Парсим токен
	claims, err := h.authRepo.ValidateToken(c, token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Апгрейд до WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Обработка через UseCase
	err = h.useCase.HandleConnection(c, claims.UserID, conn)
	if err != nil {
		log.Println("handle connection error:", err)
	}
}
