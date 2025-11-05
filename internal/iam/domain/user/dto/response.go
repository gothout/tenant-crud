package dto

import (
	"tenant-crud/internal/iam/domain/user/model"
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	UUID       uuid.UUID      `json:"uuid"`
	TenantUUID *uuid.UUID     `json:"tenant_uuid,omitempty"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Role       model.UserRole `json:"role"`
	Live       bool           `json:"live"`
	CreateAt   time.Time      `json:"create_at"`
	UpdateAt   time.Time      `json:"update_at"`
}
