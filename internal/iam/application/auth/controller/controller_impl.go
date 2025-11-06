package controller

import (
	"tenant-crud/internal/iam/application/auth/service"

	"github.com/gin-gonic/gin"
)

type impl struct {
	service service.Service
}

func New(service service.Service) Controller {
	return &impl{
		service: service,
	}
}

func (impl *impl) Login(c *gin.Context) {

}
