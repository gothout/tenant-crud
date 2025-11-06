package repository

import (
	"context"
	"errors"
	"fmt"
	"tenant-crud/internal/iam/application/auth"
	"tenant-crud/internal/iam/application/auth/model"
	"time"

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

func (r *impl) CreateAcessToken(ctx context.Context, m model.AcessToken) error {
	query := r.db.WithContext(ctx).Create(&m)

	if query.Error == nil {
		return nil
	}
	var pgErr *pgconn.PgError

	if errors.As(query.Error, &pgErr) {
		if pgErr.Code == "23505" {
			return auth.ErrTokenDuplicated
		}
		if pgErr.Code == "23503" {
			return auth.ErrTokenDuplicated
		}
		return pgErr
	}
	return query.Error
}

func (r *impl) RevokeAcessToken(ctx context.Context, token string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.AcessToken{}).
		Where("token = ?", token).
		Update("expire_date", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
