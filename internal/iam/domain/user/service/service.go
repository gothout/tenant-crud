package service

import (
	"context"
	"tenant-crud/internal/iam/domain/user/model"
)

type Service interface {
	Create(ctx context.Context, m model.User, tenantIdentifier string) (model.User, error)
	Read(ctx context.Context, m model.User) (model.User, error)
	List(ctx context.Context, page, pageSize int) ([]model.User, error)
	Update(ctx context.Context, m model.User, userIdentifier string) (model.User, error)
	Delete(ctx context.Context, userIdentifier string) error
}
