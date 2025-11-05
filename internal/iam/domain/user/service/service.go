package service

import (
	"context"
	"tenant-crud/internal/iam/domain/user/model"
)

type Service interface {
	Create(ctx context.Context, m model.User, tenantIdentifier string) (model.User, error)
}
