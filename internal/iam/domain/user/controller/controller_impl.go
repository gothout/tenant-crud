package controller

import (
	"errors"
	"net/http"
	"strconv"
	"tenant-crud/internal/iam/domain/user"
	"tenant-crud/internal/iam/domain/user/dto"
	"tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/domain/user/service"
	"tenant-crud/internal/iam/middleware"
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

const (
	defaultPage = 1   // Página padrão se nenhuma for informada
	defaultSize = 10  // Tamanho padrão se nenhum for informado
	maxSize     = 100 // Tamanho máximo permitido
)

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

	userAuth, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		restError := rest_err.NewBadRequestError("user is not authenticated")
		c.JSON(restError.Code, restError)
		return
	}

	if userAuth.User.Role == model.RoleTenantAdmin {
		tenantIdentifier = userAuth.User.Tenant.UUID.String()
		if req.Role == model.RoleSystemAdmin {
			restError := rest_err.NewBadRequestError("you cannot create this function.")
			c.JSON(restError.Code, restError)
			return
		}
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
		restError := rest_err.NewBadRequestError("user identifier is required in URL path")
		c.JSON(restError.Code, restError)
		return
	}
	var rUser model.User
	if err := uuid.Validate(userIdentifier); err != nil {
		rUser = model.User{Email: userIdentifier}
	} else {
		rUser = model.User{UUID: uuid.MustParse(userIdentifier)}
	}
	authUser, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		restError := rest_err.NewBadRequestError("user is not authenticated")
		c.JSON(restError.Code, restError)
		return
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
	switch authUser.User.Role {
	case model.RoleSystemAdmin:
		// full access
	case model.RoleTenantAdmin:
		if uRead.TenantUUID.String() != authUser.User.TenantUUID.String() {
			restError := rest_err.NewNotFoundError("user does not belong to this tenant")
			c.JSON(restError.Code, restError)
			return
		}
	case model.RoleTenantUser:
		if uRead.UUID != authUser.User.UUID {
			restError := rest_err.NewNotFoundError("user does not belong to this user")
			c.JSON(restError.Code, restError)
			return
		}
	default:
		restError := rest_err.NewForbiddenError("insufficient permissions to read user")
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

// @Summary      Lista usuários
// @Description  Retorna uma lista paginada de usuários.
// @Tags         User
// @Produce      json
//
// @Param        page query int false "Número da página para paginação." default(1)
// @Param        size query int false "Número de itens por página." default(10) maximum(100)
//
// @Success      200  {object}  dto.UserListResponse  "Lista de usuários recuperada com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (parâmetros 'page' ou 'size' inválidos, ou erro ao listar)."
//
// @Router       /api/v1/user/list [get]
func (ctrl *impl) List(c *gin.Context) {
	var (
		page int = defaultPage // Define o padrão
		size int = defaultSize // Define o padrão
		err  error
	)
	pageQuery := c.Query("page")
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil || page <= 0 {
			restError := rest_err.NewBadRequestError("parâmetro 'page' inválido: deve ser um inteiro positivo")
			c.JSON(restError.Code, restError)
			return
		}
	}

	sizeQuery := c.Query("size")
	if sizeQuery != "" {
		size, err = strconv.Atoi(sizeQuery)
		if err != nil || size <= 0 {
			restError := rest_err.NewBadRequestError("parâmetro 'size' inválido: deve ser um inteiro positivo")
			c.JSON(restError.Code, restError)
			return
		}
	}
	if size > maxSize {
		size = maxSize
	}

	users, err := ctrl.service.List(c.Request.Context(), page, size)
	if err != nil {
		restError := rest_err.NewBadRequestError("erro ao listar usuarios")
		c.JSON(restError.Code, restError)
	}

	lUser := make([]dto.UserResponse, len(users))
	for i, u := range users {
		lUser[i] = dto.UserResponse{
			UUID:       u.UUID,
			TenantUUID: u.TenantUUID,
			Name:       u.Name,
			Email:      u.Email,
			Role:       u.Role,
			CreateAt:   u.CreateAt,
			UpdateAt:   u.UpdateAt,
		}
	}
	c.JSON(http.StatusOK, &dto.UserListResponse{lUser})
}

// @Summary      Atualiza um usuário
// @Description  Atualiza parcialmente os detalhes de um usuário existente (nome, email, senha ou role).
// @Description  Apenas os campos fornecidos no corpo do JSON serão atualizados.
// @Tags         User
// @Accept       json
// @Produce      json
//
// @Param        identifier path string true "Identificador (UUID ou Documento) do usuarioa ser atualizado."
// @Param        request body dto.UpdateUserRequest true "Campos do usuário que precisam ser atualizados."
//
// @Success      200  {object}  dto.UserResponse  "Usuário atualizado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (UUID mal formatado, corpo JSON mal formatado ou falha na validação)."
// @Failure      404  {object}  rest_err.RestErr    "Não encontrado (o 'identifier' do usuário não corresponde a nenhum usuário existente)."
// @Failure      409  {object}  rest_err.RestErr    "Conflito (o 'email' fornecido já está em uso por outro usuário)."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/user/{identifier} [patch]
func (ctrl *impl) Update(c *gin.Context) {
	userIdentifier := c.Param("identifier")
	if userIdentifier == "" {
		restError := rest_err.NewBadRequestError("user identifier is required in URL path")
		c.JSON(restError.Code, restError)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restError := rest_err.NewBadRequestError("invalid json body or validation failed")
		c.JSON(restError.Code, restError)
		return
	}

	if req.Name == "" && req.Email == "" && req.Password == "" && req.Role == "" {
		restError := rest_err.NewBadRequestError("at least one field must be provided for update")
		c.JSON(restError.Code, restError)
		return
	}
	uptUser := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         req.Role,
	}

	newUser, err := ctrl.service.Update(c.Request.Context(), *uptUser, userIdentifier)
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

	respose := dto.UserResponse{
		UUID:       newUser.UUID,
		TenantUUID: newUser.TenantUUID,
		Name:       newUser.Name,
		Email:      newUser.Email,
		Role:       newUser.Role,
		CreateAt:   newUser.CreateAt,
		UpdateAt:   newUser.UpdateAt,
	}

	c.JSON(http.StatusOK, respose)
}

// @Summary      Remove um usuário
// @Description  Exclui permanentemente um usuário existente do sistema, identificado pelo UUID ou Email.
// @Description  Esta operação não é reversível. Se o usuário não existir, retorna erro 404.
// @Tags         User
// @Produce      json
//
// @Param        identifier path string true "Identificador (UUID ou Email) do usuário a ser removido."
//
// @Success      204  "Usuário removido com sucesso (sem conteúdo no corpo da resposta)."
// @Failure      400  {object}  rest_err.RestErr  "Requisição inválida (identificador ausente ou mal formatado)."
// @Failure      404  {object}  rest_err.RestErr  "Não encontrado (o usuário especificado não existe)."
// @Failure      500  {object}  rest_err.RestErr  "Erro interno do servidor."
//
// @Router       /api/v1/user/{identifier} [delete]
func (ctrl *impl) Delete(c *gin.Context) {
	userIdentifier := c.Param("identifier")
	if userIdentifier == "" {
		restError := rest_err.NewBadRequestError("user identifier is required in URL path")
		c.JSON(restError.Code, restError)
		return
	}

	err := ctrl.service.Delete(c.Request.Context(), userIdentifier)
	if err != nil {
		var restError *rest_err.RestErr
		switch {
		case errors.Is(err, user.ErrNotFound):
			restError = rest_err.NewNotFoundError(err.Error())

		case errors.Is(err, user.ErrInvalidInput):
			restError = rest_err.NewBadRequestError(err.Error())

		default:
			restError = rest_err.NewInternalServerError("internal server error", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
