package di

import (
	"tenant-crud/internal/iam/domain/di"
	"tenant-crud/internal/infra/jwt" // Certifique-se que este é o caminho do seu TokenGenerator

	"gorm.io/gorm"
)

type Container struct {
	jwtInstance *jwt.TokenGenerator
	domain      *di.Container
}

func NewContainer(db *gorm.DB, tokenGen *jwt.TokenGenerator) *Container {
	domain := di.NewContainer(db)
	return &Container{
		domain:      domain,
		jwtInstance: tokenGen,
	}
}
func (container *Container) Di() *di.Container                      { return container.domain }
func (container *Container) GetTokenGenerator() *jwt.TokenGenerator { return container.jwtInstance }
