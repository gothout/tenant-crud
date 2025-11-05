package repository

import (
	"context"
	"errors"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/model"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type impl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &impl{
		db: db,
	}
}

func (r *impl) Create(ctx context.Context, m model.User) (model.User, error) {
	result := r.db.WithContext(ctx).Create(&m)
	if result.Error == nil {
		return m, nil
	}
	var pgErr *pgconn.PgError

	if errors.As(result.Error, &pgErr) {
		if pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return model.User{}, user.ErrEmailDuplicated
			default:
				return model.User{}, result.Error
			}
		}
		if pgErr.Code == "23503" {
			return model.User{}, user.ErrTenantNotFound
		}
	}
	return model.User{}, result.Error
}
