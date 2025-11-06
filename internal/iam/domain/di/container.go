package di

import (
	tenantContainer "tenant-crud/internal/iam/domain/tenant/di"
	userContainer "tenant-crud/internal/iam/domain/user/di"

	// Tenant layer
	"gorm.io/gorm"
)

type Container struct {

	// Tenant
	tenantContainer *tenantContainer.Container
	// User
	userContainer *userContainer.Container
}

func NewContainer(db *gorm.DB) *Container {

	// Tenant
	tenantContainer := tenantContainer.NewContainer(db)

	// User
	userContainer := userContainer.NewContainer(db, tenantContainer)
	return &Container{
		tenantContainer: tenantContainer,
		userContainer:   userContainer,
	}
}

func (c *Container) GetContainer() *tenantContainer.Container {
	return c.tenantContainer
}

// Tenant
func (c *Container) GetTenantContainer() *tenantContainer.Container { return c.tenantContainer }

// User
func (c *Container) GetUserContainer() *userContainer.Container { return c.userContainer }
