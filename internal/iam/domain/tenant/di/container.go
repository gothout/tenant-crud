package di

import (
	// Tenant layer
	tenantController "tenant-crud/internal/iam/domain/tenant/controller"
	tenantRepository "tenant-crud/internal/iam/domain/tenant/repository"
	tenantService "tenant-crud/internal/iam/domain/tenant/service"

	"gorm.io/gorm"
)

type Container struct {
	// Tenant
	tController tenantController.Controller
	tService    tenantService.Service
	tRepository tenantRepository.Repository
}

func NewContainer(db *gorm.DB) *Container {

	// Tenant
	tenantRepo := tenantRepository.New(db)
	tenantSvc := tenantService.New(tenantRepo)
	tenantCtrl := tenantController.New(tenantSvc)

	return &Container{
		// Tenant
		tController: tenantCtrl,
		tService:    tenantSvc,
		tRepository: tenantRepo,
	}
}

// Tenant
func (c *Container) GetTenantController() tenantController.Controller {
	return c.tController
}
func (c *Container) GetTenantService() tenantService.Service {
	return c.tService
}
func (c *Container) GetTenantRepository() tenantRepository.Repository {
	return c.tRepository
}
