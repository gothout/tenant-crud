package service

import (
	"context"
	"tenant-crud/internal/iam/application/auth/model"
)

type Service interface {
	Login(ctx context.Context, email, pwd string) (model.Login, error)
}
