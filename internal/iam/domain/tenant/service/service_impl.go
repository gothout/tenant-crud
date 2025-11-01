package service

import (
	"context"
	"tenant-crud/internal/iam/domain/tenant/model"
	"tenant-crud/internal/iam/domain/tenant/repository"
)

type impl struct {
	repository repository.Repository
}

func New(repository repository.Repository) Service {
	return &impl{
		repository: repository,
	}
}

func (i *impl) Create(ctx context.Context, tenant model.Tenant) (model.Tenant, error) {
	return i.repository.Create(ctx, tenant)
}
