package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	applicationAuthServiceInstance "tenant-crud/internal/iam/application/auth/service"
	domainUserModel "tenant-crud/internal/iam/domain/user/model"
	userDomainServiceInstance "tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/pkg/rest_err"
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
			err := rest_err.NewForbiddenError("Token ausente ou inválido.")
			c.AbortWithStatusJSON(err.Code, err)
			return
		}

		ctx := c.Request.Context()
		loginResult, err := mw.applicationAuthService.GetAcessToken(ctx, token)
		if err != nil {
			e := rest_err.NewForbiddenError("Falha ao validar token de acesso.")
			c.AbortWithStatusJSON(e.Code, e)
			return
		}

		if loginResult.UserUUID == nil || *loginResult.UserUUID == uuid.Nil {
			e := rest_err.NewForbiddenError("Token não associado a nenhum usuário válido.")
			c.AbortWithStatusJSON(e.Code, e)
			return
		}

		if time.Now().UTC().After(loginResult.Expiry) {
			e := rest_err.NewForbiddenError("Token expirado. Efetue login novamente.")
			c.AbortWithStatusJSON(e.Code, e)
			return
		}

		user, err := mw.userDomainServiceInstance.Read(ctx, domainUserModel.User{UUID: *loginResult.UserUUID})
		if err != nil || user.UUID == uuid.Nil {
			e := rest_err.NewForbiddenError("Usuário associado ao token não encontrado ou inválido.")
			c.AbortWithStatusJSON(e.Code, e)
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
			e := rest_err.NewForbiddenError("Usuário não autenticado.")
			c.AbortWithStatusJSON(e.Code, e)
			return
		}

		if !isRoleAuthorized(lUser.User.Role, requiredRoles) {
			requiredRoleStrings := make([]string, len(requiredRoles))
			for i, role := range requiredRoles {
				requiredRoleStrings[i] = string(role)
			}

			e := rest_err.NewForbiddenError(fmt.Sprintf(
				"Acesso negado. É necessário possuir uma das permissões: %v.",
				requiredRoleStrings,
			))
			c.AbortWithStatusJSON(e.Code, e)
			return
		}

		c.Next()
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
