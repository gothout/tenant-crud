package controller

import (
	"errors"
	"net/http"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/dto"
	"tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/pkg/rest_err"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// @Summary      Busca um usuário
// @Description  Obtém os detalhes de um usuário específico, buscando pelo UUID ou pelo email.
// @Tags         User
// @Produce      json
//
// @Param        identifier path string true "Identificador (UUID ou Email) do Usuário a ser buscado."
//
// @Success      200  {object}  dto.UserResponse  "Usuário encontrado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (o 'identifier' do usuário está ausente ou é inválido)."
// @Failure      404  {object}  rest_err.RestErr    "Não encontrado (o recurso solicitado não foi localizado)."
// @Failure      409  {object}  rest_err.RestErr    "Conflito (o serviço retornou um erro de conflito, ex: 'email duplicado')."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/user/{identifier} [get]
func (ctrl *impl) Read(c *gin.Context) {
	userIdentifier := c.Param("identifier")
	if userIdentifier == "" {
		restError := rest_err.NewBadRequestError("tenant identifier is required in URL path")
		c.JSON(restError.Code, restError)
		return
	}
	var rUser model.User
	err := uuid.Validate(userIdentifier)
	if err != nil {
		rUser = model.User{
			Email: userIdentifier,
		}
	} else {
		uUUIDParsed := uuid.MustParse(userIdentifier)
		rUser = model.User{
			UUID: uUUIDParsed,
		}
	}

	uRead, err := ctrl.service.Read(c.Request.Context(), rUser)
	if err != nil {
		var restError *rest_err.RestErr
		switch {
		case errors.Is(err, user.ErrNotFound):
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
		UUID:       uRead.UUID,
		TenantUUID: uRead.TenantUUID,
		Name:       uRead.Name,
		Email:      uRead.Email,
		Role:       uRead.Role,
		Live:       uRead.Live,
		CreateAt:   uRead.CreateAt,
		UpdateAt:   uRead.UpdateAt,
	}
	c.JSON(http.StatusOK, response)
}
