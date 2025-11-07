package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"net/http"
	"strings"
	applicationAuthServiceInstance "tenant-crud/internal/iam/application/auth/service"
	domainUserModel "tenant-crud/internal/iam/domain/user/model"
	userDomainServiceInstance "tenant-crud/internal/iam/domain/user/service"
)

type Middleware interface {
	SetContextAutorization() gin.HandlerFunc
	AuthorizeRole(requiredRoles ...domainUserModel.UserRole) gin.HandlerFunc
}

type impl struct {
	applicationAuthService    applicationAuthServiceInstance.Service
	userDomainServiceInstance userDomainServiceInstance.Service
}

func New(applicationAuthService applicationAuthServiceInstance.Service, userDomainServiceInstance userDomainServiceInstance.Service) Middleware {
	return &impl{
		applicationAuthService:    applicationAuthService,
		userDomainServiceInstance: userDomainServiceInstance,
	}
}

func checkAuthorization(requiredRoles []string, userRoles []string) bool {
	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role] = true
	}
	for _, requiredRole := range requiredRoles {
		if roleMap[requiredRole] {
			return true
		}
	}

	return false
}

func (mw *impl) SetContextAutorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		loginResult, err := mw.applicationAuthService.GetAcessToken(c, token)

		login := Login{AcessToken: AcessToken{UserUUID: loginResult.UserUUID, Token: loginResult.Token, Expiry: loginResult.Expiry}}

		if err != nil {
			return
		}
		SetAuthenticatedUser(c, &login)
		c.Next()
	}
}

func (mw *impl) AuthorizeRole(requiredRoles ...domainUserModel.UserRole) gin.HandlerFunc {
	requiredRoleStrings := make([]string, len(requiredRoles))
	for i, role := range requiredRoles {
		requiredRoleStrings[i] = string(role)
	}

	return func(c *gin.Context) {
		lUser, ok := GetAuthenticatedUser(c)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: Usuário não autenticado."})
			return
		}

		ctx := c.Request.Context()
		loggedUser, err := mw.userDomainServiceInstance.Read(ctx, domainUserModel.User{UUID: lUser.User.UUID})

		if err != nil || loggedUser.UUID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: Falha ao carregar dados do usuário."})
			return
		}

		currentUserRole := string(loggedUser.Role)
		userRoleStrings := []string{currentUserRole}

		isAuthorized := checkAuthorization(requiredRoleStrings, userRoleStrings)

		if isAuthorized {
			fmt.Printf("Resultado da Autorização: TRUE - Autorizado. Roles do Usuário: %v\n", userRoleStrings)
			c.Next() // Permite o acesso à rota
		} else {
			fmt.Printf("Resultado da Autorização: FALSE - Negado. Roles do Usuário: %v. Requer: %v\n", userRoleStrings, requiredRoleStrings)
			// Usa 403 Forbidden, pois o usuário está autenticado, mas não tem permissão
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Acesso negado. Requer uma das roles: %v", requiredRoleStrings)})
		}
	}
}
