package controller

import (
	"errors"
	"net/http"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/dto"
	"tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/pkg/rest_err" // <-- 1. Importado o rest_err

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

// @Summary      Cria um novo Usuário
// @Description  Registra um novo usuário no sistema, associado a um tenant (empresa/organização).
// @Tags         User
// @Accept       json
// @Produce      json
//
// @Param        identifier path string true "Identificador (UUID ou Documento) do Tenant ao qual o usuário será associado."
// @Param        request body dto.CreateUserRequest true "Objeto do usuário que precisa ser criado."
//
// @Success      201  {object}  dto.UserResponse  "Usuário criado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (corpo JSON mal formatado, dados de entrada inválidos, ou 'identifier' do tenant ausente)."
// @Failure      404  {object}  rest_err.RestErr    "Tenant não encontrado (o 'identifier' fornecido não corresponde a nenhum tenant existente)."
// @Failure      409  {object}  rest_err.RestErr    "Conflito (o 'email' fornecido já está em uso)."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/user/{identifier} [post]
func (ctrl *impl) Create(c *gin.Context) {
	tenantIdentifier := c.Param("identifier")
	if tenantIdentifier == "" {
		restError := rest_err.NewBadRequestError("tenant identifier is required in URL path")
		c.JSON(restError.Code, restError)
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restError := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restError.Code, restError)
		return
	}

	newUser := model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         req.Role,
		Live:         true,
	}

	userCreated, err := ctrl.service.Create(c.Request.Context(), newUser, tenantIdentifier)

	if err != nil {
		var restError *rest_err.RestErr
		switch {
		case errors.Is(err, user.ErrTenantNotFound):
			restError = rest_err.NewNotFoundError(err.Error())

		case errors.Is(err, user.ErrEmailDuplicated):
			restError = rest_err.NewConflictValidationError(err.Error(), nil)

		case errors.Is(err, user.ErrInvalidInput):
			restError = rest_err.NewBadRequestError(err.Error())

		default:
			restError = rest_err.NewInternalServerError("internal server error", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	response := dto.UserResponse{
		UUID:       userCreated.UUID,
		TenantUUID: userCreated.TenantUUID,
		Name:       userCreated.Name,
		Email:      userCreated.Email,
		Role:       userCreated.Role,
		Live:       userCreated.Live,
		CreateAt:   userCreated.CreateAt,
		UpdateAt:   userCreated.UpdateAt,
	}
	c.JSON(http.StatusCreated, response)
}
