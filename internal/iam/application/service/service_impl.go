package service

import (
	"context"
	"tenant-crud/internal/iam/application"
	"tenant-crud/internal/iam/application/model"
	modelUser "tenant-crud/internal/iam/domain/user/model"
	userService "tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/iam/domain/util"
	"tenant-crud/internal/infra/jwt"
)

type impl struct {
	jwt         jwt.TokenGenerator
	userService userService.Service
}

func New(userService userService.Service, jwt jwt.TokenGenerator) Service {
	return &impl{
		userService: userService,
		jwt:         jwt,
	}
}

func (s *impl) Login(ctx context.Context, email, pwd string) (model.Login, error) {
	rUser, err := s.userService.Read(ctx, modelUser.User{
		Email: email,
	})
	if err != nil {
		return model.Login{}, application.ErrPwdWrong
	}
	if util.Compare(rUser.PasswordHash, pwd); err != nil {
		return model.Login{}, application.ErrPwdWrong
	}
	token, expTime, err := s.jwt.GenerateAccessToken(rUser.UUID, *rUser.TenantUUID)
	response := model.Login{
		User:  rUser,
		Token: model.Token{Token: token, Expiration: expTime},
	}
	return response, nil
}
