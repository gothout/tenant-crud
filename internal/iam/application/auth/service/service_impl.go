package service

import (
	"context"
	"tenant-crud/internal/iam/application/auth"
	"tenant-crud/internal/iam/application/auth/model"
	"tenant-crud/internal/iam/application/auth/repository"
	modelUser "tenant-crud/internal/iam/domain/user/model"
	userService "tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/iam/domain/util"
	"tenant-crud/internal/infra/jwt"
)

type impl struct {
	jwt         jwt.TokenGenerator
	userService userService.Service
	repository  repository.Repository
}

func New(userService userService.Service, jwt jwt.TokenGenerator, repository repository.Repository) Service {
	return &impl{
		userService: userService,
		jwt:         jwt,
		repository:  repository,
	}
}

func (s *impl) Login(ctx context.Context, email, pwd string) (model.Login, error) {
	rUser, err := s.userService.Read(ctx, modelUser.User{
		Email: email,
	})
	if err != nil {
		return model.Login{}, auth.ErrPwdWrong
	}
	if err := util.Compare(rUser.PasswordHash, pwd); err != nil {
		return model.Login{}, auth.ErrPwdWrong
	}
	token, expTime, err := s.jwt.GenerateAccessToken(rUser.UUID, *rUser.TenantUUID)
	AcessToken := model.AcessToken{UserUUID: &rUser.UUID, Token: token, Expiry: expTime}
	response := model.Login{
		User:       rUser,
		AcessToken: AcessToken,
	}

	err = s.repository.CreateAcessToken(ctx, AcessToken)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (s *impl) RevokeAcessToken(ctx context.Context, token string) error {
	return s.repository.RevokeAcessToken(ctx, token)
}

func (s *impl) GetAcessToken(ctx context.Context, token string) (model.AcessToken, error) {
	return s.repository.GetAcessToken(ctx, token)
}
