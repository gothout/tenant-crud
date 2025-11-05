package service

import (
	"context"
	"errors"
	"fmt"
	tModel "tenant-crud/internal/iam/domain/tenant/model"
	"tenant-crud/internal/iam/domain/tenant/service"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/domain/user/repository"
	"tenant-crud/internal/iam/domain/util"
	"time"

	"github.com/google/uuid"
)

type impl struct {
	svcTenant  service.Service
	repository repository.Repository
}

func New(repo repository.Repository, svcTenant service.Service) Service {
	return &impl{
		repository: repo,
		svcTenant:  svcTenant,
	}
}

func (s *impl) Create(ctx context.Context, m model.User, tenantIdentifier string) (model.User, error) {
	var mTenant tModel.Tenant
	var err error
	if parsedUUID, errUUID := uuid.Parse(tenantIdentifier); errUUID == nil {
		mTenant, err = s.svcTenant.Read(ctx, tModel.Tenant{UUID: parsedUUID})
		if err != nil {
			mTenant, err = s.svcTenant.Read(ctx, tModel.Tenant{Document: tenantIdentifier})
			if err != nil {
				return model.User{}, user.ErrTenantNotFound
			}
		}
	} else {
		mTenant, err = s.svcTenant.Read(ctx, tModel.Tenant{Document: tenantIdentifier})
		if err != nil {
			return model.User{}, user.ErrTenantNotFound
		}
	}
	hashedPassword, err := util.Hash(m.PasswordHash)
	if err != nil {
		return model.User{}, fmt.Errorf("erro ao gerar hash de senha: %w", err)
	}

	newUser := model.User{
		UUID:         uuid.New(),
		TenantUUID:   &mTenant.UUID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: hashedPassword,
		Role:         m.Role,
		Live:         m.Live,
		CreateAt:     time.Now().UTC(),
		UpdateAt:     time.Now().UTC(),
		Tenant:       mTenant,
	}
	createdUser, err := s.repository.Create(ctx, newUser)
	if err != nil {
		if errors.Is(err, user.ErrEmailDuplicated) || errors.Is(err, user.ErrTenantNotFound) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("erro interno ao criar usuário: %w", err)
	}
	return createdUser, nil
}

func (s *impl) Read(ctx context.Context, m model.User) (model.User, error) {
	return s.repository.Read(ctx, m)
}

func (s *impl) List(ctx context.Context, page, pageSize int) ([]model.User, error) {
	return s.repository.List(ctx, page, pageSize)
}
