package di

import (
	application "tenant-crud/internal/iam/application/di"
	domain "tenant-crud/internal/iam/domain/di"
	"tenant-crud/internal/infra/jwt"

	"gorm.io/gorm"
)

type Container struct {
	jwtInstance *jwt.TokenGenerator
	domain      *domain.Container
	application *application.Container
}

func NewContainer(db *gorm.DB, tokenGen *jwt.TokenGenerator) *Container {
	domain := domain.NewContainer(db)
	application := application.NewContainer(db, domain, tokenGen)
	return &Container{
		domain:      domain,
		application: application,
		jwtInstance: tokenGen,
	}
}
func (container *Container) Di() *domain.Container                  { return container.domain }
func (container *Container) GetTokenGenerator() *jwt.TokenGenerator { return container.jwtInstance }
func (container *Container) GetApplicationContainer() *application.Container {
	return container.application
}
