package repository

import (
	"RealtimeService/internal/domains/connection/model"
	"context"
)

type RedisRepository interface {
	GetLastData(ctx context.Context, unitIDs []int) ([]model.Data, error)
}

type redisRepository struct{}

func NewRedisRepository() RedisRepository {
	return &redisRepository{}
}

func (r *redisRepository) GetLastData(ctx context.Context, unitIDs []int) ([]model.Data, error) {
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
	return result, nil
}
