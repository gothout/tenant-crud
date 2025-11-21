package controller

import (
	"errors"
	"net/http"
	"tenant-crud/internal/iam/application/auth"
	"tenant-crud/internal/iam/application/auth/dto"
	"tenant-crud/internal/iam/application/auth/service"
	userDto "tenant-crud/internal/iam/domain/user/dto"
	"tenant-crud/internal/pkg/rest_err"

	"github.com/gin-gonic/gin"
)

type impl struct {
	service service.Service
}

func New(service service.Service) Controller {
	return &impl{
		service: service,
	}
}

// @Summary Efetua o login do usuário
// @Description Recebe email e senha, autentica o usuário e retorna o token de acesso.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credenciais do Usuário (Email e Senha)"
// @Success 200 {object} dto.LoginResponse "Login bem-sucedido"
// @Failure 400 {object} rest_err.RestErr "Requisição inválida (JSON mal formatado)"
// @Failure 404 {object} rest_err.RestErr "Credenciais inválidas (usuário/senha errados)"
// @Failure 409 {object} rest_err.RestErr "Token duplicado ou conflito"
// @Failure 500 {object} rest_err.RestErr "Erro interno do servidor"
// @Router /api/auth/login [post]
func (ctrl *impl) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restErr.Code, restErr)
		return
	}

	uLogin, err := ctrl.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		var restError *rest_err.RestErr
		switch {
		case errors.Is(err, auth.ErrPwdWrong):
			restError = rest_err.NewNotFoundError(err.Error())

		case errors.Is(err, auth.ErrTokenDuplicated):
			restError = rest_err.NewConflictValidationError(err.Error(), nil)

		default:
			restError = rest_err.NewInternalServerError("internal server error", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	response := dto.LoginResponse{
		User: userDto.UserResponse{
			UUID:       uLogin.User.UUID,
			TenantUUID: uLogin.User.TenantUUID,
			Name:       uLogin.User.Name,
			Email:      uLogin.User.Email,
			Role:       uLogin.User.Role,
			Live:       uLogin.User.Live,
			CreateAt:   uLogin.User.CreateAt,
			UpdateAt:   uLogin.User.UpdateAt,
		},
		Token:  uLogin.AcessToken.Token,
		Expire: uLogin.AcessToken.Expiry,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Revoga o token de acesso
// @Description Invalida o token de acesso atual do usuário.
// @Tags Auth
// @Accept json
// @Produce json
// @Param token path string true "Token de acesso a ser revogado"
// @Success 202 "Token revogado com sucesso"
// @Failure 404 {object} rest_err.RestErr "Token não encontrado"
// @Failure 500 {object} rest_err.RestErr "Erro interno do servidor"
// @Router /api/auth/logout/{token} [post]
func (ctrl *impl) Logout(c *gin.Context) {
	token := c.Param("token")
	if err := ctrl.service.RevokeAcessToken(c.Request.Context(), token); err != nil {
		restErr := rest_err.NewForbiddenError("user not authorized")
		c.JSON(restErr.Code, restErr)
		return
	}
	c.JSON(http.StatusAccepted, nil)
}

// @Summary Solicita um código OTP
// @Description Gera um OTP vinculado ao e-mail e envia por e-mail.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.OTPRequest true "Email para envio do OTP"
// @Success 202 "OTP enviado com sucesso"
// @Failure 400 {object} rest_err.RestErr "JSON inválido"
// @Failure 409 {object} rest_err.RestErr "OTP já existente"
// @Failure 500 {object} rest_err.RestErr "Erro interno"
// @Router /api/auth/otp [post]
func (ctrl *impl) CreateOTP(c *gin.Context) {
	var req dto.OTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restErr.Code, restErr)
		return
	}

	if err := ctrl.service.CreateOTPCode(c.Request.Context(), req.Email); err != nil {
		var restErr *rest_err.RestErr

		switch {
		case errors.Is(err, auth.OTPCodeExist):
			restErr = rest_err.NewConflictValidationError(err.Error(), nil)
		default:
			restErr = rest_err.NewInternalServerError("internal server error", nil)
		}

		c.JSON(restErr.Code, restErr)
		return
	}

	c.Status(http.StatusAccepted)
}

// @Summary Troca a senha usando OTP
// @Description Valida o OTP e troca a senha do usuário.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.OTPResetPasswordRequest true "Email, OTP e nova senha"
// @Success 200 "Senha alterada com sucesso"
// @Failure 400 {object} rest_err.RestErr "JSON inválido"
// @Failure 403 {object} rest_err.RestErr "OTP inválido"
// @Failure 500 {object} rest_err.RestErr "Erro interno"
// @Router /api/auth/password/reset [post]
func (ctrl *impl) ResetPassword(c *gin.Context) {
	var req dto.OTPResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restErr.Code, restErr)
		return
	}

	ok, err := ctrl.service.ChangeUserPwd(
		c.Request.Context(),
		req.OTPCode,
		req.Email,
		req.Password,
	)
	if err != nil {
		var restErr *rest_err.RestErr

		switch {
		case errors.Is(err, auth.OTPCodeWrong):
			restErr = rest_err.NewForbiddenError(err.Error())
		default:
			restErr = rest_err.NewInternalServerError("internal server error", nil)
		}

		c.JSON(restErr.Code, restErr)
		return
	}

	if !ok {
		// fallback defensivo, teoricamente não deveria cair aqui
		restErr := rest_err.NewInternalServerError("could not change password", nil)
		c.JSON(restErr.Code, restErr)
		return
	}

	c.Status(http.StatusOK)
}
