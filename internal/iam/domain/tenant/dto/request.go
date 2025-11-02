package dto

type CreateTenantRequest struct {
	Name     string `json:"name" binding:"required"`
	Document string `json:"document" binding:"required"`
	Live     bool   `json:"live" binding:"required"`
}

type ReadTenantRequest struct {
	UUID     string `form:"uuid"`
	Document string `form:"document"`
}

type ListTenantRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"size"`
}
