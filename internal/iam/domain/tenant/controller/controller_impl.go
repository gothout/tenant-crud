package controller

import (
	"net/http"
	"tenant-crud/internal/iam/domain/tenant"
	"tenant-crud/internal/iam/domain/tenant/dto"
	"tenant-crud/internal/iam/domain/tenant/model"
	"tenant-crud/internal/iam/domain/tenant/service"
	domainUserModel "tenant-crud/internal/iam/domain/user/model"
	"tenant-crud/internal/iam/middleware"
	"tenant-crud/internal/pkg/rest_err"
	"time"

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

// @Summary      Cria um novo Tenant
// @Description  Registra um novo tenant (empresa/organização) no sistema.
// @Tags         Tenant
// @Accept       json
// @Produce      json
//
// @Param        request body dto.CreateTenantRequest true "Objeto do tenant que precisa ser criado."
//
// @Success      201  {object}  dto.TenantResponse  "Tenant criado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (corpo JSON mal formatado ou dados de entrada inválidos)."
// @Failure      409  {object}  rest_err.RestErr    "Conflito (o 'document' fornecido já está em uso)."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/tenant [post]
func (ctrl *impl) Create(c *gin.Context) {
	var request dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		restError := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restError.Code, restError)
		return
	}
	cTenant := model.Tenant{
		UUID:     uuid.New(),
		Name:     request.Name,
		Document: request.Document,
		Live:     request.Live,
		CreateAt: time.Now().UTC(),
		UpdateAt: time.Now().UTC(),
	}
	tenantCreated, err := ctrl.service.Create(c.Request.Context(), cTenant)
	if err != nil {
		var restError *rest_err.RestErr

		switch err {
		case tenant.ErrDocumentDuplicated:

			restError = rest_err.NewConflictValidationError("document is duplicated", nil)
		case tenant.ErrInvalidInput:
			restError = rest_err.NewBadRequestError("invalid input data")
		default:
			restError = rest_err.NewInternalServerError("failed to create tenant", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}
	c.JSON(http.StatusCreated, &dto.TenantResponse{
		UUID:     tenantCreated.UUID,
		Name:     tenantCreated.Name,
		Document: tenantCreated.Document,
		Live:     tenantCreated.Live,
		CreateAt: tenantCreated.CreateAt,
		UpdateAt: tenantCreated.UpdateAt,
	})
}

// @Summary      Busca um Tenant
// @Description  Busca um tenant no sistema usando o UUID ou o Documento (CNPJ/CPF). Pelo menos um dos dois campos deve ser fornecido.
// @Tags         Tenant
// @Produce      json
//
// @Param        uuid query string false "UUID do tenant a ser buscado. (Ex: 8871abf3-ed11-4770-b986-e8d98d022d4f)"
// @Param        document query string false "Documento (CNPJ/CPF) do tenant a ser buscado. (Ex: 12345678901234)"
//
// @Success      200  {object}  dto.TenantResponse  "Tenant encontrado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (UUID inválido, ou nenhum dos campos 'uuid'/'document' fornecido)."
// @Failure      404  {object}  rest_err.RestErr    "Tenant não encontrado com os dados fornecidos."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/tenant [get]
func (ctrl *impl) Read(c *gin.Context) {
	var req dto.ReadTenantRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		restError := rest_err.NewBadRequestError("Parâmetros de busca inválidos.")
		c.JSON(restError.Code, restError)
		return
	}

	if req.UUID == "" && req.Document == "" {
		restError := rest_err.NewBadRequestError("É necessário fornecer o 'uuid' OU o 'document' para a busca.")
		c.JSON(restError.Code, restError)
		return
	}

	var tenantUUID uuid.UUID
	if req.UUID != "" {
		// Tentativa de conversão
		parsedUUID, err := uuid.Parse(req.UUID)
		if err != nil {
			// Captura erro de UUID mal formatado (retorna 400)
			restError := rest_err.NewBadRequestError("O UUID fornecido não é um formato válido.")
			c.JSON(restError.Code, restError)
			return
		}
		tenantUUID = parsedUUID
	}

	rTenant, err := ctrl.service.Read(c.Request.Context(), model.Tenant{
		UUID:     tenantUUID,
		Document: req.Document,
	})
	if err != nil {
		var restError *rest_err.RestErr
		switch err {
		case tenant.ErrNotFound:
			// Tratamento para 404
			restError = rest_err.NewNotFoundError(tenant.ErrNotFound.Error())
		case tenant.ErrInvalidInput:
			// Tratamento para 400 (Assumindo que InvalidInput no Read é uma falha na query/dados)
			restError = rest_err.NewBadRequestError(tenant.ErrInvalidInput.Error())
		default:
			// Tratamento para 500
			restError = rest_err.NewInternalServerError("Falha ao buscar tenant", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	authUser, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		restError := rest_err.NewForbiddenError("Usuário não autenticado.")
		c.JSON(restError.Code, restError)
		return
	}

	switch authUser.User.Role {
	case domainUserModel.RoleSystemAdmin:
		// full access
	case domainUserModel.RoleTenantAdmin:
		if authUser.User.TenantUUID == nil || authUser.User.TenantUUID.String() != rTenant.UUID.String() {
			restError := rest_err.NewForbiddenError("insufficient permissions to read tenant")
			c.JSON(restError.Code, restError)
			return
		}
	default:
		restError := rest_err.NewForbiddenError("insufficient permissions to read tenant")
		c.JSON(restError.Code, restError)
		return
	}

	c.JSON(http.StatusOK, &dto.TenantResponse{
		UUID:     rTenant.UUID,
		Name:     rTenant.Name,
		Document: rTenant.Document,
		Live:     rTenant.Live,
		CreateAt: rTenant.CreateAt,
		UpdateAt: rTenant.UpdateAt,
	})
}

// @Summary      Lista Tenants
// @Description  Retorna uma lista paginada de todos os tenants registrados no sistema.
// @Tags         Tenant
// @Produce      json
//
// @Param        page query int true "O número da página a ser retornada (deve ser >= 1)." default(1)
// @Param        pageSize query int true "O número de itens por página (máximo 100)." default(10)
//
// @Success      200  {object}  dto.TenantsResponse  "Lista de tenants retornada com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (parâmetros de paginação ausentes ou inválidos, ou pageSize > 100)."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/tenant/list [get]
func (ctrl *impl) List(c *gin.Context) {
	var req dto.ListTenantRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		restError := rest_err.NewBadRequestError("Parâmetros de busca inválidos. Verifique 'page' e 'pageSize'.")
		c.JSON(restError.Code, restError)
		return
	}

	if req.PageSize > 100 {
		restError := rest_err.NewBadRequestError("É permitido um máximo de 100 listagens por página.")
		c.JSON(restError.Code, restError)
		return // Adicionado o 'return' para evitar execução posterior
	}

	lTenants, err := ctrl.service.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		var restError *rest_err.RestErr
		restError = rest_err.NewInternalServerError("Falha ao buscar tenants", nil)
		c.JSON(restError.Code, restError)
		return
	}

	tenantResponses := make([]dto.TenantResponse, len(lTenants))
	for i, t := range lTenants {
		tenantResponses[i] = dto.TenantResponse{
			UUID:     t.UUID,
			Name:     t.Name,
			Document: t.Document,
			Live:     t.Live,
			CreateAt: t.CreateAt,
			UpdateAt: t.UpdateAt,
		}
	}
	c.JSON(http.StatusOK, &dto.TenantsResponse{
		Tenants: tenantResponses,
		Page:    req.Page,
		Size:    req.PageSize,
	})
}

// @Summary      Atualiza um Tenant
// @Description  Atualiza dados de um tenant existente. O tenant a ser atualizado é identificado pelo UUID no path
// @Tags         Tenant
// @Accept       json
// @Produce      json
//
// @Param        uuid path string true "UUID do tenant a ser atualizado."
// @Param        request body dto.UpdateTenantRequest true "Campos do tenant a serem atualizados. Apenas os campos presentes serão modificados."
//
// @Success      200  {object}  dto.TenantResponse  "Tenant atualizado com sucesso."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (corpo JSON mal formatado, UUID inválido ou dados de entrada inválidos)."
// @Failure      404  {object}  rest_err.RestErr    "Tenant não encontrado para o UUID fornecido."
// @Failure      409  {object}  rest_err.RestErr    "Conflito (o novo 'document' fornecido já está em uso por outro tenant)."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/tenant/{uuid} [patch]
func (ctrl *impl) Update(c *gin.Context) {
	uuidStr := c.Param("uuid")
	tenantUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		restError := rest_err.NewBadRequestError("O UUID fornecido na URL não é um formato válido.")
		c.JSON(restError.Code, restError)
		return
	}

	authUser, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		restError := rest_err.NewForbiddenError("Usuário não autenticado.")
		c.JSON(restError.Code, restError)
		return
	}

	switch authUser.User.Role {
	case domainUserModel.RoleSystemAdmin:
		// full access
	case domainUserModel.RoleTenantAdmin:
		if authUser.User.TenantUUID == nil || authUser.User.TenantUUID.String() != tenantUUID.String() {
			restError := rest_err.NewForbiddenError("insufficient permissions to update tenant")
			c.JSON(restError.Code, restError)
			return
		}
	default:
		restError := rest_err.NewForbiddenError("insufficient permissions to update tenant")
		c.JSON(restError.Code, restError)
		return
	}

	var request dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		restError := rest_err.NewBadRequestError("Corpo JSON inválido ou mal formatado.")
		c.JSON(restError.Code, restError)
		return
	}

	uTenant := model.Tenant{
		UUID:     tenantUUID,
		Document: request.Document,
		Live:     *request.Live,
		Name:     request.Name,
		UpdateAt: time.Now().UTC(),
	}

	if request.Live != nil {
		uTenant.Live = *request.Live
	}

	tenantUpdated, err := ctrl.service.Update(c.Request.Context(), &uTenant)

	if err != nil {
		var restError *rest_err.RestErr

		switch err {
		case tenant.ErrNotFound:
			restError = rest_err.NewNotFoundError(tenant.ErrNotFound.Error())
		case tenant.ErrDocumentDuplicated:
			restError = rest_err.NewConflictValidationError("O novo documento fornecido já está em uso por outro tenant.", nil)
		case tenant.ErrInvalidInput:
			restError = rest_err.NewBadRequestError(tenant.ErrInvalidInput.Error())
		default:
			restError = rest_err.NewInternalServerError("Falha ao atualizar tenant", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	c.JSON(http.StatusOK, &dto.TenantResponse{
		UUID:     tenantUpdated.UUID,
		Name:     tenantUpdated.Name,
		Document: tenantUpdated.Document,
		Live:     tenantUpdated.Live,
		CreateAt: tenantUpdated.CreateAt,
		UpdateAt: tenantUpdated.UpdateAt,
	})
}

// @Summary      Deleta um Tenant
// @Description  Exclui permanentemente um tenant no sistema usando o UUID ou o Documento (CNPJ/CPF). Pelo menos um dos dois campos deve ser fornecido.
// @Tags         Tenant
// @Produce      json
//
// @Param        uuid query string false "UUID do tenant a ser excluído. (Ex: 8871abf3-ed11-4770-b986-e8d98d022d4f)"
// @Param        document query string false "Documento (CNPJ/CPF) do tenant a ser excluído. (Ex: 12345678901234)"
//
// @Success      204  {string} string "Tenant excluído com sucesso (No Content)."
// @Failure      400  {object}  rest_err.RestErr    "Requisição inválida (UUID inválido, ou nenhum dos campos 'uuid'/'document' fornecido)."
// @Failure      404  {object}  rest_err.RestErr    "Tenant não encontrado com os dados fornecidos."
// @Failure      500  {object}  rest_err.RestErr    "Erro interno do servidor."
//
// @Router       /api/v1/tenant [delete]
func (ctrl *impl) Delete(c *gin.Context) {
	var req dto.ReadTenantRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		restError := rest_err.NewBadRequestError("Parâmetros de busca inválidos.")
		c.JSON(restError.Code, restError)
		return
	}

	if req.UUID == "" && req.Document == "" {
		restError := rest_err.NewBadRequestError("É necessário fornecer o 'uuid' OU o 'document' para a exclusão.")
		c.JSON(restError.Code, restError)
		return
	}

	var tenantUUID uuid.UUID
	if req.UUID != "" {
		parsedUUID, err := uuid.Parse(req.UUID)
		if err != nil {
			restError := rest_err.NewBadRequestError("O UUID fornecido não é um formato válido.")
			c.JSON(restError.Code, restError)
			return
		}
		tenantUUID = parsedUUID
	}

	targetTenant, err := ctrl.service.Read(c.Request.Context(), model.Tenant{
		UUID:     tenantUUID,
		Document: req.Document,
	})
	if err != nil {
		var restError *rest_err.RestErr
		switch err {
		case tenant.ErrNotFound:
			restError = rest_err.NewNotFoundError(tenant.ErrNotFound.Error())
		case tenant.ErrInvalidInput:
			restError = rest_err.NewBadRequestError(tenant.ErrInvalidInput.Error())
		default:
			restError = rest_err.NewInternalServerError("Falha ao excluir tenant", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}

	authUser, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		restError := rest_err.NewForbiddenError("Usuário não autenticado.")
		c.JSON(restError.Code, restError)
		return
	}

	switch authUser.User.Role {
	case domainUserModel.RoleSystemAdmin:
		// full access
	case domainUserModel.RoleTenantAdmin:
		if authUser.User.TenantUUID == nil || authUser.User.TenantUUID.String() != targetTenant.UUID.String() {
			restError := rest_err.NewForbiddenError("insufficient permissions to delete tenant")
			c.JSON(restError.Code, restError)
			return
		}
	default:
		restError := rest_err.NewForbiddenError("insufficient permissions to delete tenant")
		c.JSON(restError.Code, restError)
		return
	}

	err = ctrl.service.Delete(c.Request.Context(), model.Tenant{
		UUID:     tenantUUID,
		Document: req.Document,
	})

	if err != nil {
		var restError *rest_err.RestErr
		switch err {
		case tenant.ErrNotFound:
			restError = rest_err.NewNotFoundError(tenant.ErrNotFound.Error())
		case tenant.ErrInvalidInput:
			restError = rest_err.NewBadRequestError(tenant.ErrInvalidInput.Error())
		default:
			restError = rest_err.NewInternalServerError("Falha ao excluir tenant", nil)
		}

		c.JSON(restError.Code, restError)
		return
	}
	c.Status(http.StatusNoContent)
}
