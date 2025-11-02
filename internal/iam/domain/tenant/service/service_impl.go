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

func (s *impl) Create(ctx context.Context, tenant model.Tenant) (model.Tenant, error) {
	return s.repository.Create(ctx, tenant)
}

func (s *impl) Read(ctx context.Context, m model.Tenant) (model.Tenant, error) {
	return s.repository.Read(ctx, m)
}

func (s *impl) List(ctx context.Context, page, pageSize int) ([]model.Tenant, error) {
	return s.repository.List(ctx, page, pageSize)
}
