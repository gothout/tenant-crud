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

func (ctrl *impl) Create(c *gin.Context) {
	var request dto.CreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		restError := rest_err.NewBadRequestError("invalid json body")
		c.JSON(restError.Code, restError)
		return
	}
	cTenant := model.Tenant{
		UUID:       uuid.New(),
		Name:       request.Name,
		Document:   request.Document,
		Live:       request.Live,
		CreateDate: time.Now().UTC(),
		UpdateDate: time.Now().UTC(),
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
	c.JSON(http.StatusCreated, tenantCreated)
}
