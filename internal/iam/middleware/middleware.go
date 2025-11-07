package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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

func (mw *impl) SetContextAutorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := extractBearerToken(authHeader)

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: token ausente."})
			return
		}

		ctx := c.Request.Context()
		loginResult, err := mw.applicationAuthService.GetAcessToken(ctx, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: token inválido."})
			return
		}

		if loginResult.UserUUID == nil || *loginResult.UserUUID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: usuário não associado ao token."})
			return
		}

		if time.Now().UTC().After(loginResult.Expiry) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: token expirado."})
			return
		}

		user, err := mw.userDomainServiceInstance.Read(ctx, domainUserModel.User{UUID: *loginResult.UserUUID})
		if err != nil || user.UUID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: usuário inválido."})
			return
		}

		login := Login{
			User: user,
			AcessToken: AcessToken{
				UserUUID: loginResult.UserUUID,
				Token:    loginResult.Token,
				Expiry:   loginResult.Expiry,
			},
		}

		SetAuthenticatedUser(c, &login)
		c.Next()
	}
}

func (mw *impl) AuthorizeRole(requiredRoles ...domainUserModel.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		lUser, ok := GetAuthenticatedUser(c)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso não autorizado: Usuário não autenticado."})
			return
		}

		if !isRoleAuthorized(lUser.User.Role, requiredRoles) {
			requiredRoleStrings := make([]string, len(requiredRoles))
			for i, role := range requiredRoles {
				requiredRoleStrings[i] = string(role)
			}

			fmt.Printf("Resultado da Autorização: FALSE - Negado. Role do Usuário: %s. Requer: %v\n", lUser.User.Role, requiredRoleStrings)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Acesso negado. Requer uma das roles: %v", requiredRoleStrings)})
			return
		}

		fmt.Printf("Resultado da Autorização: TRUE - Autorizado. Role do Usuário: %s\n", lUser.User.Role)
		c.Next() // Permite o acesso à rota
	}
}

func isRoleAuthorized(userRole domainUserModel.UserRole, requiredRoles []domainUserModel.UserRole) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	for _, requiredRole := range requiredRoles {
		if userRole == requiredRole {
			return true
		}
	}

	return false
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}
