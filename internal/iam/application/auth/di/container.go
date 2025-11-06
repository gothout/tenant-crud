package di

import (
	"tenant-crud/internal/iam/application/auth/controller"
	"tenant-crud/internal/iam/application/auth/repository"
	"tenant-crud/internal/iam/application/auth/service"
	domainContainer "tenant-crud/internal/iam/domain/di"
	"tenant-crud/internal/infra/jwt"

	"gorm.io/gorm"
)

type Container struct {
	Repository repository.Repository
	Service    service.Service
	Controller controller.Controller
	jwt        *jwt.TokenGenerator
}

func NewContainer(db *gorm.DB, domainContainer *domainContainer.Container, jwt *jwt.TokenGenerator) *Container {
	repo := repository.New(db)
	svc := service.New(domainContainer.GetUserContainer().GetUserService(), *jwt, repo)
	ctrl := controller.New(svc)

	return &Container{
		Repository: repo,
		Service:    svc,
		Controller: ctrl,
		jwt:        jwt,
	}
}

func (c *Container) GetRepository() repository.Repository { return c.Repository }
func (c *Container) GetService() service.Service          { return c.Service }
func (c *Container) GetController() controller.Controller { return c.Controller }
