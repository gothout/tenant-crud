package di

import (
	// Tenant layer
	"gorm.io/gorm"
	tenantContainer "tenant-crud/internal/iam/domain/tenant/di"
)

type Container struct {
	// Tenant
	tenantContainer *tenantContainer.Container
}

func NewContainer(db *gorm.DB) *Container {

	// Tenant
	tenantContainer := tenantContainer.NewContainer(db)
	return &Container{
		tenantContainer: tenantContainer,
	}
}

func (c *Container) GetContainer() *tenantContainer.Container {
	return c.tenantContainer
}

// Tenant
func (c *Container) GetTenantContainer() *tenantContainer.Container {
	return c.tenantContainer
}
