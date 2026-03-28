package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
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

	user, state, err := h.service.CreateOrFetchUser(*userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	code := http.StatusCreated
	if state == model.UserFetched {
		code = http.StatusOK
	}

	c.JSON(code, user)
}

func (h *UserHandler) UpdateFields(c *gin.Context) {
	userCtx := getUserContext(c)

	var updatedFields model.UpdateFieldsDto
	if err := c.ShouldBindJSON(&updatedFields); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	user, err := h.service.UpdateFields(*userCtx, updatedFields)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUserAccount(c *gin.Context) {
	userCtx := getUserContext(c)

	err := h.service.DeleteUser(*userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
