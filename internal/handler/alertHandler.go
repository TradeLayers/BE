package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	service service.AlertService
}

func NewAlertHandler(s service.AlertService) *AlertHandler {
	return &AlertHandler{service: s}
}

func (h *AlertHandler) List(c *gin.Context) {
	userCtx := getUserContext(c)

	alerts, err := h.service.List(c.Request.Context(), *userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *AlertHandler) Create(c *gin.Context) {
	userCtx := getUserContext(c)

	var req model.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	alert, domainErr := h.service.Create(c.Request.Context(), *userCtx, req)
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.JSON(http.StatusCreated, alert)
}

func (h *AlertHandler) Delete(c *gin.Context) {
	userCtx := getUserContext(c)

	err := h.service.Delete(*userCtx, c.Param("id"))
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
