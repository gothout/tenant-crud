package controller

import (
	"net/http"
	"tenant-crud/internal/iam/domain/tenant"
	"tenant-crud/internal/iam/domain/tenant/dto"
	"tenant-crud/internal/iam/domain/tenant/model"
	"tenant-crud/internal/iam/domain/tenant/service"
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
