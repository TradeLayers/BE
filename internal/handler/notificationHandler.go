package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

func (h *NotificationHandler) ListUnread(c *gin.Context) {
	userCtx := getUserContext(c)

	notifications, domainErr := h.service.ListUnread(c.Request.Context(), *userCtx)
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userCtx := getUserContext(c)

	domainErr := h.service.MarkRead(c.Request.Context(), *userCtx, c.Param("id"))
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.Status(http.StatusNoContent)
}
