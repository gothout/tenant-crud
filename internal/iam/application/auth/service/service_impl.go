package service

import (
	"context"
	"fmt"
	"tenant-crud/internal/iam/application/auth"
	"tenant-crud/internal/iam/application/auth/cache"
	"tenant-crud/internal/iam/application/auth/model"
	"tenant-crud/internal/iam/application/auth/repository"
	modelUser "tenant-crud/internal/iam/domain/user/model"
	userService "tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/iam/domain/util"
	"tenant-crud/internal/infra/jwt"
	"tenant-crud/internal/pkg/mailer"
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

func (s *impl) CreateOTPCode(ctx context.Context, email string) error {
	_, found := cache.GetOTP(email)
	if found {
		return auth.OTPCodeExist
	}
	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		return err
	}
	cache.SaveOTP(email, otpCode)
	err = mailer.Use().SendRaw(
		email,
		"OTP Code",
		fmt.Sprintf("<h1>Seu código OTP é: %s</h1>", otpCode),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *impl) ValidateOTPCode(ctx context.Context, email, codeDst string) bool {
	otpExist, found := cache.GetOTP(email)
	if !found {
		return false
	}
	if otpExist != codeDst {
		return false
	}

	return true
}

func (s *impl) ChangeUserPwd(ctx context.Context, otpCode, email, pwd string) (bool, error) {
	if !s.ValidateOTPCode(ctx, email, otpCode) {
		return false, auth.OTPCodeWrong
	}
	cache.DeleteOTP(email)
	updUser := modelUser.User{PasswordHash: pwd}
	_, err := s.userService.Update(ctx, updUser, email)
	if err != nil {
		return false, err
	}
	return true, nil
}
