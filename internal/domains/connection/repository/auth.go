package repository

import (
	"RealtimeService/internal/domains/connection/model"
	"context"
)

type AuthRepository interface {
	ValidateToken(ctx context.Context, token string) (model.UserClaims, error)
}

type authRepository struct{}

func NewAuthRepository() AuthRepository {
	return &authRepository{}
}

func (r *authRepository) ValidateToken(ctx context.Context, token string) (model.UserClaims, error) {
	// ⚠️ Заглушка: всегда возвращает userID = 1
	return model.UserClaims{UserID: 1}, nil
}
