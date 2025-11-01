package service

import (
	"context"
	"tenant-crud/internal/iam/domain/tenant/model"
)

type Service interface {
	Create(ctx context.Context, tenant model.Tenant) (model.Tenant, error)
}
