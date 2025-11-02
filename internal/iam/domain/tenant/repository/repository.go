package repository

import (
	"context"
	"tenant-crud/internal/iam/domain/tenant/model"
)

type Repository interface {
	Create(ctx context.Context, tenant model.Tenant) (model.Tenant, error)
	Read(ctx context.Context, m model.Tenant) (model.Tenant, error)
	List(ctx context.Context, page, pageSize int) ([]model.Tenant, error)
}
