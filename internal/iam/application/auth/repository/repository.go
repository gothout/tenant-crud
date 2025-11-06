package repository

import (
	"context"
	"tenant-crud/internal/iam/application/auth/model"
)

type Repository interface {
	CreateAcessToken(ctx context.Context, m model.AcessToken) error
	RevokeAcessToken(ctx context.Context, token string) error
}
