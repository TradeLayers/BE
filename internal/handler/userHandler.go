package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func getUserContext(c *gin.Context) *model.UserContext {
	userObj, _ := c.Get("userContext")
	userCtx, _ := userObj.(model.UserContext)

	return &userCtx
}

func (h *UserHandler) CreateOrFetchUser(c *gin.Context) {
	userCtx := getUserContext(c)
	ctx := c.Request.Context()

	user, state, err := h.service.CreateOrFetchUser(ctx, *userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	ICode := http.StatusCreated
	if state == model.UserFetched {
		ICode = http.StatusOK
	}

	c.JSON(ICode, user)
}

func (h *UserHandler) UpdateFields(c *gin.Context) {
	userCtx := getUserContext(c)
	ctx := c.Request.Context()
	log := requestLogger(c)

	var updatedFields model.UpdateFieldsDto = model.UpdateFieldsDto{}
	if err := c.ShouldBindJSON(&updatedFields); err != nil {
		log.Warn("failed to bind user update request", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	user, err := h.service.UpdateFields(ctx, *userCtx, updatedFields)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUserAccount(c *gin.Context) {
	userCtx := getUserContext(c)
	ctx := c.Request.Context()

	err := h.service.DeleteUser(ctx, *userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
