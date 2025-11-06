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

func (s *impl) Update(ctx context.Context, m model.User, userIdentifier string) (model.User, error) {
	err := uuid.Validate(userIdentifier)
	var srcUser model.User
	if err != nil {
		srcUser = model.User{
			Email: userIdentifier,
		}
	} else {
		srcUser = model.User{
			UUID: uuid.MustParse(userIdentifier),
		}
	}
	oldUser, err := s.Read(ctx, srcUser)
	if err != nil {
		return model.User{}, err
	}
	var pwdHash string
	if m.PasswordHash != "" {
		pwdHash, err = util.Hash(m.PasswordHash)
		if err != nil {
			return model.User{}, fmt.Errorf("erro ao gerar hash de senha: %w", err)
		}
	} else {
		pwdHash = oldUser.PasswordHash
	}
	newUser := model.User{
		UUID:         oldUser.UUID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: pwdHash,
		Role:         m.Role,
		Live:         m.Live,
		UpdateAt:     time.Now().UTC(),
	}
	updatedUser, err := s.repository.Update(ctx, newUser)
	if err != nil {
		return updatedUser, err
	}
	return updatedUser, nil
}

func (s *impl) Delete(ctx context.Context, userIdentifier string) error {
	err := uuid.Validate(userIdentifier)
	var srcUser model.User
	if err != nil {
		srcUser = model.User{
			Email: userIdentifier,
		}
	} else {
		srcUser = model.User{
			UUID: uuid.MustParse(userIdentifier),
		}
	}
	oldUser, err := s.Read(ctx, srcUser)
	if err != nil {
		return err
	}
	return s.repository.Delete(ctx, oldUser)
}
