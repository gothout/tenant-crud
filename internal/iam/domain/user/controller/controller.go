package controller

import (
	"github.com/gin-gonic/gin"
	domainUserModel "tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/middleware"
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
	group := routerGroup.Group("/user")
	group.POST("/:identifier", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantUser, domainUserModel.RoleTenantAdmin), ctrl.Create)
	group.GET("/:identifier", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantUser, domainUserModel.RoleTenantAdmin), ctrl.Read)
	group.GET("/list", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantUser, domainUserModel.RoleTenantAdmin), ctrl.List)
	group.PATCH("/:identifier", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantUser, domainUserModel.RoleTenantAdmin), ctrl.Update)
	group.DELETE("/:identifier", mw.AuthorizeRole(domainUserModel.RoleSystemAdmin, domainUserModel.RoleTenantUser, domainUserModel.RoleTenantAdmin), ctrl.Delete)
}
