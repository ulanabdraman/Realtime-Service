package usecase

import (
	"context"
	"log"

	"RealtimeService/internal/domains/connection/model"
	"RealtimeService/internal/server/hub"

	"github.com/gorilla/websocket"
)

type AuthRepository interface {
	ValidateToken(ctx context.Context, token string) (model.UserClaims, error)
}

type UnitRepository interface {
	GetUserUnits(ctx context.Context, userID int64) ([]int, error)
}

type RedisRepository interface {
	GetLastData(ctx context.Context, unitIDs []int) ([]model.Data, error)
}

type ConnectionUseCase interface {
	HandleConnection(ctx context.Context, userID int64, conn *websocket.Conn) error
}

type connectionUseCase struct {
	unitRepo  UnitRepository
	redisRepo RedisRepository
	hub       *hub.Hub
}

func New(unitRepo UnitRepository, redisRepo RedisRepository, h *hub.Hub) ConnectionUseCase {
	return &connectionUseCase{
		unitRepo:  unitRepo,
		redisRepo: redisRepo,
		hub:       h,
	}
}

func (uc *connectionUseCase) HandleConnection(ctx context.Context, userID int64, conn *websocket.Conn) error {
	// Получаем unit_id
	unitIDs, err := uc.unitRepo.GetUserUnits(ctx, userID)
	if err != nil {
		log.Println("unitRepo error:", err)
		return err
	}
	if len(unitIDs) == 0 {
		log.Println("No units found for user:", userID)
		return nil
	}

	// Подписываемся в hub
	uc.hub.Subscribe(conn, unitIDs)
	defer uc.hub.Unsubscribe(conn)

	// Получаем последние данные
	lastData, err := uc.redisRepo.GetLastData(ctx, unitIDs)
	if err == nil && len(lastData) > 0 {
		_ = conn.WriteJSON(lastData)
	}

	// Держим соединение открытым
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("connection closed:", err)
			break
		}
	}

	return nil
}
