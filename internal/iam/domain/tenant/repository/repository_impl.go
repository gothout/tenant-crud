package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"tenant-crud/internal/iam/domain/tenant"
	"tenant-crud/internal/iam/domain/tenant/model"

	"github.com/google/uuid"
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

func (r *impl) Create(ctx context.Context, m model.Tenant) (model.Tenant, error) {
	result := r.db.WithContext(ctx).Create(&m)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.Tenant{}, tenant.ErrDocumentDuplicated
		}
		return model.Tenant{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.Tenant{}, fmt.Errorf("no rows affected")
	}
	return m, nil
}

func (r *impl) Read(ctx context.Context, m model.Tenant) (model.Tenant, error) {
	query := r.db.WithContext(ctx).Model(&model.Tenant{})

	if m.UUID != uuid.Nil {
		query = query.First(&m, "uuid = ?", m.UUID)
	} else if m.Document != "" {
		query = query.Where("document = ?", m.Document).First(&m)
	} else {
		return model.Tenant{}, tenant.ErrInvalidInput
	}

	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return model.Tenant{}, tenant.ErrNotFound
		}
		return model.Tenant{}, fmt.Errorf("erro ao ler tenant: %w", query.Error)
	}

	return m, nil
}

func (r *impl) List(ctx context.Context, page, pageSize int) ([]model.Tenant, error) {
	var listTenant []model.Tenant
	query := r.db.WithContext(ctx).Model(&model.Tenant{})

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	result := query.Limit(pageSize).Offset(offset).Find(&listTenant)
	if result.Error != nil {
		return nil, result.Error
	}

	return listTenant, nil
}
