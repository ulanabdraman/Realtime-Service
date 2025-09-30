package repository

import (
	"context"
)

type UnitRepository interface {
	GetUserUnits(ctx context.Context, userID int64) ([]int, error)
}

type unitRepository struct{}

func NewUnitRepository() UnitRepository {
	return &unitRepository{}
}

func (r *unitRepository) GetUserUnits(ctx context.Context, userID int64) ([]int, error) {
	// ⚠️ Заглушка: всегда 3 unit_id
	return []int{586, 102, 103}, nil
}
