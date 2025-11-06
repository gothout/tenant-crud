package service

import (
	"context"
	"tenant-crud/internal/iam/application/model"
)

type Service interface {
	Login(ctx context.Context, email, pwd string) (model.Login, error)
}
