package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"tenant-crud/internal/iam/domain/tenant"
	"tenant-crud/internal/iam/domain/tenant/model"
	"time"

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

func (r *impl) Update(ctx context.Context, m *model.Tenant) (model.Tenant, error) {
	if m.UUID == uuid.Nil {
		return model.Tenant{}, tenant.ErrInvalidInput
	}

	updateModel := model.Tenant{
		Name:     m.Name,
		Document: m.Document,
		Live:     m.Live,
		UpdateAt: time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).
		Where("uuid = ?", m.UUID).
		Select("Name", "Document", "Live", "UpdateAt").
		Updates(updateModel)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			return model.Tenant{}, tenant.ErrDocumentDuplicated
		}
		return model.Tenant{}, result.Error
	}

	if result.RowsAffected == 0 {
		existingTenant, err := r.Read(ctx, model.Tenant{UUID: m.UUID})
		if err != nil {
			if errors.Is(err, tenant.ErrNotFound) {
				return model.Tenant{}, tenant.ErrNotFound
			}
			return model.Tenant{}, fmt.Errorf("falha ao verificar se o tenant existe: %w", err)
		}

		return existingTenant, nil
	}
	updatedTenant, err := r.Read(ctx, model.Tenant{UUID: m.UUID})
	if err != nil {
		return updatedTenant, fmt.Errorf("falha ao ler tenant atualizado: %w", err)
	}

	return updatedTenant, nil
}

func (r *impl) Delete(ctx context.Context, m model.Tenant) error {
	if m.UUID == uuid.Nil && m.Document == "" {
		return tenant.ErrInvalidInput
	}
	query := r.db.WithContext(ctx)
	if m.UUID != uuid.Nil {
		query = query.Where("uuid = ?", m.UUID)
	} else {
		query = query.Where("document = ?", m.Document)
	}
	result := query.Delete(&model.Tenant{})
	if result.Error != nil {
		return fmt.Errorf("falha ao deletar tenant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return tenant.ErrNotFound
	}
	return nil
}
