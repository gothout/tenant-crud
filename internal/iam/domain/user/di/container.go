package di

import (
	tenantContainer "tenant-crud/internal/iam/domain/tenant/di"
	// User layer
	userController "tenant-crud/internal/iam/domain/user/controller"
	userRepository "tenant-crud/internal/iam/domain/user/repository"
	userService "tenant-crud/internal/iam/domain/user/service"

	"gorm.io/gorm"
)

type Container struct {
	// User
	uController userController.Controller
	uService    userService.Service
	uRepository userRepository.Repository
}

func NewContainer(db *gorm.DB, tenantContainer *tenantContainer.Container) *Container {

	// User
	userRepo := userRepository.New(db)
	userService := userService.New(userRepo, tenantContainer.GetTenantService())
	userController := userController.New(userService)

	return &Container{
		// User
		uController: userController,
		uService:    userService,
		uRepository: userRepo,
	}
}

// User
func (c *Container) GetUserController() userController.Controller {
	return c.uController
}
func (c *Container) GetUserService() userService.Service {
	return c.uService
}
func (c *Container) GetUserRepository() userRepository.Repository {
	return c.uRepository
}
