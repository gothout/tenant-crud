package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	RegisterRoutes(routerGroup *gin.RouterGroup)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	CreateOTP(c *gin.Context)
	ResetPassword(c *gin.Context)
}

func (ctrl *impl) RegisterRoutes(routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/auth")
	group.POST("/login", ctrl.Login)
	group.POST("/logout/:token", ctrl.Logout)
	group.POST("/otp", ctrl.CreateOTP)
	group.POST("/password/reset", ctrl.ResetPassword)
}
