package routes

import (
	iamDomainContainer "tenant-crud/internal/iam/domain/di"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "tenant-crud/docs"
)

func SetupRouter(tContainer *iamDomainContainer.Container) *gin.Engine {
	r := gin.Default()

	// Acessível em http://localhost:8080/doc/index.html
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Configuração das rotas da API
	apiGroup := r.Group("/api")
	SetupApiRoutes(apiGroup, tContainer)

	return r
}

// Assinatura atualizada (sem o 'r *gin.Engine')
func SetupApiRoutes(routerGroup *gin.RouterGroup, iamDomainContainer *iamDomainContainer.Container) {
	v1Group := routerGroup.Group("/v1")

	// Tenants
	tController := iamDomainContainer.GetTenantContainer().GetTenantController()
	tController.RegisterRoutes(v1Group)
}
