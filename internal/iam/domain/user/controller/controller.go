package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	RegisterRoutes(routerGroup *gin.RouterGroup)
	Create(c *gin.Context)
	Read(c *gin.Context)
}

func (ctrl *impl) RegisterRoutes(routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/user")
	group.POST("/:identifier", ctrl.Create)
	group.GET("/:identifier", ctrl.Read)
}
