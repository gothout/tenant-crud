package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	RegisterRoutes(routerGroup *gin.RouterGroup)
	Create(c *gin.Context)
	Read(c *gin.Context)
	List(c *gin.Context)
}

func (ctrl *impl) RegisterRoutes(routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/tenant")
	group.POST("", ctrl.Create)
	group.GET("", ctrl.Read)
	group.GET("/list", ctrl.List)
}
