package repository

import (
	"RealtimeService/internal/domains/connection/model"
	"context"
	"log/slog"
)

type RedisRepository interface {
	GetLastData(ctx context.Context, unitIDs []int) ([]model.Data, error)
}

type redisRepository struct {
	logger *slog.Logger
}

func NewRedisRepository() RedisRepository {
	logger := slog.With(
		slog.String("object", "connection"),
		slog.String("layer", "RedisRepository"),
	)
	return &redisRepository{logger: logger}
}

func (r *redisRepository) GetLastData(ctx context.Context, unitIDs []int) ([]model.Data, error) {
	r.logger.Info("GetLastData called", slog.Int("unit_count", len(unitIDs)))

	r.logger.Debug("Fetching last data for unitIDs", slog.Any("unit_ids", unitIDs))

	var result []model.Data
	for _, id := range unitIDs {
		result = append(result, model.Data{
			ID:      int64(id),
			T:       1,
			ST:      0,
			Address: "г. Алматы",
			Pos: model.Pos{
				X:  77.0,
				Y:  43.0,
				Z:  0,
				A:  0,
				S:  0,
				St: 0,
			},
			Params: map[string]interface{}{
				"speed": 52.3,
			},
		})
	}

	r.logger.Debug("Returning mock result", slog.Int("result_count", len(result)))
	return result, nil
}
