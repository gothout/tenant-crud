package di

import (
	authDi "tenant-crud/internal/iam/application/auth/di"
	domainContainer "tenant-crud/internal/iam/domain/di"
	"tenant-crud/internal/infra/jwt"

	"gorm.io/gorm"
)

type Container struct {
	domainContainer *domainContainer.Container
	authContainer   *authDi.Container
	jwtInstance     *jwt.TokenGenerator
}

func NewContainer(db *gorm.DB, domainContainer *domainContainer.Container, jwtInstance *jwt.TokenGenerator) *Container {
	authContainer := authDi.NewContainer(db, domainContainer, jwtInstance)
	return &Container{
		domainContainer: domainContainer,
		authContainer:   authContainer,
		jwtInstance:     jwtInstance,
	}
}

func (c *Container) GetAuthContainer() *authDi.Container {
	return c.authContainer
}
