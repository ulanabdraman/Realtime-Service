package repository

import (
	"RealtimeService/internal/domains/connection/model"
	"context"
	"log/slog"
)

type AuthRepository interface {
	ValidateToken(ctx context.Context, token string) (model.UserClaims, error)
}

type authRepository struct {
	logger *slog.Logger
}

func NewAuthRepository() AuthRepository {
	logger := slog.With(
		slog.String("object", "connection"),
		slog.String("layer", "AuthRepository"),
	)
	return &authRepository{logger: logger}
}

func (r *authRepository) ValidateToken(ctx context.Context, token string) (model.UserClaims, error) {
	r.logger.Info("ValidateToken called", slog.String("token", token))
	// Заглушка: всегда возвращает userID = 1
	return model.UserClaims{UserID: 1}, nil
}
