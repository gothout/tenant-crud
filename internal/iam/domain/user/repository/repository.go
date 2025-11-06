package repository

import (
	"context"
	"tenant-crud/internal/iam/domain/user/model"
)

type Repository interface {
	Create(ctx context.Context, m model.User) (model.User, error)
	Read(ctx context.Context, m model.User) (model.User, error)
	List(ctx context.Context, page, pageSize int) ([]model.User, error)
	Update(ctx context.Context, m model.User) (model.User, error)
	Delete(ctx context.Context, m model.User) error
}
