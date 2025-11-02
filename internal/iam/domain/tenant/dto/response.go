package dto

import (
	"time"

	"github.com/google/uuid"
)

type TenantResponse struct {
	UUID     uuid.UUID `json:"uuid"`
	Name     string    `json:"name"`
	Document string    `json:"document"`
	Live     bool      `json:"live"`
	CreateAt time.Time `json:"createAt"`
	UpdateAt time.Time `json:"updateAt"`
}

type TenantsResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
}
