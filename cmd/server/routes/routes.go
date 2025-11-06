package routes

import (
	"fmt"
	"os"
	iamDomainContainer "tenant-crud/internal/iam/di"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "tenant-crud/docs"
)

func SetupRouter(tContainer *iamDomainContainer.Container) *gin.Engine {
	env := viper.GetString("app.env")
	// 1. Configuração do modo Gin
	switch env {
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	case "":
		fmt.Println("WARNING: 'app.env' not set in config. Defaulting to 'dev' mode.")
		gin.SetMode(gin.DebugMode)
	default:
		fmt.Printf("ERROR: Invalid environment value '%s'. Must be 'dev' or 'prod'.\n", env)
		os.Exit(1)
	}

	r := gin.Default()

	// Acessível em http://localhost:8080/doc/index.html
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Configuração das rotas da API
	apiGroup := r.Group("/api")
	SetupApiRoutes(apiGroup, tContainer)
	SetupAuthRoutes(apiGroup, tContainer)

	return r
}

func SetupApiRoutes(routerGroup *gin.RouterGroup, iamDomainContainer *iamDomainContainer.Container) {
	v1Group := routerGroup.Group("/v1")
	// iamDomainsRoutes
	tController := iamDomainContainer.Di().GetTenantContainer().GetTenantController()
	uController := iamDomainContainer.Di().GetUserContainer().GetUserController()
	tController.RegisterRoutes(v1Group)
	uController.RegisterRoutes(v1Group)
}

func SetupAuthRoutes(routerGroup *gin.RouterGroup, iamDomainContainer *iamDomainContainer.Container) {
	authController := iamDomainContainer.GetApplicationContainer().GetAuthContainer().Controller
	authController.RegisterRoutes(routerGroup)
}
