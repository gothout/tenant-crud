package repository

import (
	"context"
	"tenant-crud/internal/iam/domain/user/model"
)

type Repository interface {
	Create(ctx context.Context, m model.User) (model.User, error)
	Read(ctx context.Context, m model.User) (model.User, error)
}
