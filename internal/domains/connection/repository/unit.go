package repository

import (
	"context"
	"log/slog"
)

type UnitRepository interface {
	GetUserUnits(ctx context.Context, userID int64) ([]int, error)
}

type unitRepository struct {
	logger *slog.Logger
}

func NewUnitRepository() UnitRepository {
	logger := slog.With(
		slog.String("object", "connection"),
		slog.String("layer", "UnitRepository"),
	)
	return &unitRepository{logger: logger}
}

func (r *unitRepository) GetUserUnits(ctx context.Context, userID int64) ([]int, error) {
	r.logger.Info("GetUserUnits called", slog.Int64("user_id", userID))

	unitIDs := []int{586, 102, 103}

	r.logger.Debug("Returning unit IDs", slog.Any("unit_ids", unitIDs))
	return unitIDs, nil
}
