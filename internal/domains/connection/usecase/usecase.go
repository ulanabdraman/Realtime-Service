package usecase

import (
	"context"
	"log/slog"

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
	logger    *slog.Logger
}

func New(unitRepo UnitRepository, redisRepo RedisRepository, h *hub.Hub) ConnectionUseCase {
	logger := slog.With(
		slog.String("object", "connection"),
		slog.String("layer", "UseCase"),
	)
	return &connectionUseCase{
		unitRepo:  unitRepo,
		redisRepo: redisRepo,
		hub:       h,
		logger:    logger,
	}
}

func (uc *connectionUseCase) HandleConnection(ctx context.Context, userID int64, conn *websocket.Conn) error {
	uc.logger.Info("HandleConnection started", slog.Int64("user_id", userID))

	unitIDs, err := uc.unitRepo.GetUserUnits(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user units", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return err
	}
	if len(unitIDs) == 0 {
		uc.logger.Warn("No units found for user", slog.Int64("user_id", userID))
		return nil
	}

	uc.logger.Debug("Subscribing to units", slog.Int64("user_id", userID), slog.Any("unit_ids", unitIDs))
	uc.hub.Subscribe(conn, unitIDs)
	defer uc.hub.Unsubscribe(conn)

	lastData, err := uc.redisRepo.GetLastData(ctx, unitIDs)
	if err != nil {
		uc.logger.Error("Failed to get last data from Redis", slog.String("error", err.Error()))
	} else if len(lastData) > 0 {
		_ = conn.WriteJSON(lastData)
		uc.logger.Debug("Sent last data to client", slog.Int("count", len(lastData)))
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			uc.logger.Info("Connection closed", slog.Int64("user_id", userID), slog.String("error", err.Error()))
			break
		}
	}

	return nil
}
