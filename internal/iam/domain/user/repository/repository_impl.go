package repository

import (
	"context"
	"errors"
	"fmt"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/model"

	"github.com/google/uuid"
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

func (r *impl) Read(ctx context.Context, m model.User) (model.User, error) {
	query := r.db.WithContext(ctx).Last(&model.User{})
	if m.UUID != uuid.Nil {
		query = query.Where("uuid = ?", m.UUID).First(&m)
	} else if m.Email != "" {
		query = query.Where("email = ?", m.Email).First(&m)
	} else {
		return model.User{}, user.ErrInvalidInput
	}
	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return model.User{}, user.ErrNotFound
		}
		return model.User{}, query.Error
	}
	return m, nil
}

func (r *impl) List(ctx context.Context, page, pageSize int) ([]model.User, error) {
	var users []model.User
	query := r.db.WithContext(ctx).Model(&users)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	result := query.Limit(pageSize).Offset(offset).Find(&users)
	if result.Error != nil {
		return users, result.Error
	}
	return users, nil
}

func (r *impl) Update(ctx context.Context, m model.User) (model.User, error) {
	updateFields := make(map[string]interface{})

	// Monta apenas os campos enviados
	if m.Name != "" {
		updateFields["name"] = m.Name
	}
	if m.Email != "" {
		updateFields["email"] = m.Email
	}
	if m.PasswordHash != "" {
		updateFields["password_hash"] = m.PasswordHash
	}
	if m.Role != "" {
		updateFields["role"] = m.Role
	}
	// Cuidado: se Live for false, ele não entra aqui (a menos que use ponteiro)
	if m.Live {
		updateFields["live"] = m.Live
	}
	if !m.UpdateAt.IsZero() {
		updateFields["update_at"] = m.UpdateAt
	}

	if len(updateFields) == 0 {
		return model.User{}, errors.New("nenhum campo válido para atualização")
	}

	// Executa o update
	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uuid = ?", m.UUID).
		Updates(updateFields)

	// --- Tratamento de erros do PostgreSQL ---
	if query.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(query.Error, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique violation
				if pgErr.ConstraintName == "users_email_key" {
					return model.User{}, user.ErrEmailDuplicated
				}
				return model.User{}, fmt.Errorf("violação de unicidade (%s): %w", pgErr.ConstraintName, query.Error)
			case "23503": // Foreign key violation
				return model.User{}, user.ErrTenantNotFound
			default:
				return model.User{}, fmt.Errorf("erro do banco (%s): %w", pgErr.Code, query.Error)
			}
		}
		// Erro genérico (não é PgError)
		return model.User{}, query.Error
	}

	// --- Verifica se o registro realmente existe ---
	if query.RowsAffected == 0 {
		existingUser, err := r.Read(ctx, model.User{UUID: m.UUID})
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				return model.User{}, user.ErrNotFound
			}
			return model.User{}, err
		}
		return existingUser, nil
	}

	// --- Retorna o usuário atualizado ---
	updatedUser, err := r.Read(ctx, model.User{UUID: m.UUID})
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return model.User{}, user.ErrNotFound
		}
		return model.User{}, err
	}

	return updatedUser, nil
}

func (r *impl) Delete(ctx context.Context, m model.User) error {
	query := r.db.WithContext(ctx).Delete(&m).Where("uuid = ?", m.UUID)
	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return user.ErrNotFound
		}
		return query.Error
	}
	return nil
}
