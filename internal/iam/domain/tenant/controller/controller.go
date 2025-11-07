package controller

import (
	domainUserModel "tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/middleware"

	"github.com/gin-gonic/gin"
)

type Controller interface {
	RegisterRoutes(mw middleware.Middleware, routerGroup *gin.RouterGroup)
	Create(c *gin.Context)
	Read(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (ctrl *impl) RegisterRoutes(mw middleware.Middleware, routerGroup *gin.RouterGroup) {
	group := routerGroup.Group("/tenant")
	group.POST("", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin), ctrl.Create)
	group.GET("", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantAdmin), ctrl.Read)
	group.GET("/list", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin), ctrl.List)
	group.PATCH("/:uuid", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantAdmin), ctrl.Update)
	group.DELETE("", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantAdmin), ctrl.Delete)
}
