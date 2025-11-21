package service

import (
	"context"
	"tenant-crud/internal/iam/application/auth/model"
)

type Service interface {
	Login(ctx context.Context, email, pwd string) (model.Login, error)
	RevokeAcessToken(ctx context.Context, token string) error
	GetAcessToken(ctx context.Context, token string) (model.AcessToken, error)
	CreateOTPCode(ctx context.Context, email string) error
	ValidateOTPCode(ctx context.Context, email, codeDst string) bool
	ChangeUserPwd(ctx context.Context, otpCode, email, pwd string) (bool, error)
}
