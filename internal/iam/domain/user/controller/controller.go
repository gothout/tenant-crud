package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	RegisterRoutes(routerGroup *gin.RouterGroup)
	Create(c *gin.Context)
	Read(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (ctrl *impl) RegisterRoutes(routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/user")
	group.POST("/:identifier", ctrl.Create)
	group.GET("/:identifier", ctrl.Read)
	group.GET("/list", ctrl.List)
	group.PATCH("/:identifier", ctrl.Update)
	group.DELETE("/:identifier", ctrl.Delete)
}
