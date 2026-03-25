package handler

import (
	"errors"
	"net/http"

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

func getUserContext(c *gin.Context) (*model.UserContext, error) {
	userObj, exists := c.Get("userContext")
	if !exists {
		return nil, errors.New("User context missing")
	}

	userCtx, ok := userObj.(model.UserContext)
	if !ok {
		return nil, errors.New("Invalid user context")
	}

	return &userCtx, nil
}

func (h *UserHandler) CreateOrFetchUser(c *gin.Context) {
	userCtx, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, state, err := h.service.CreateOrFetchUser(*userCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user request"})
		return
	}

	code := http.StatusCreated
	if state == model.UserFetched {
		code = http.StatusOK
	}

	c.JSON(code, user)
}

func (h *UserHandler) UpdateFields(c *gin.Context) {
	userCtx, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.UpdateFieldsDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.UpdateFields(*userCtx, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUserAccount(c *gin.Context) {
	userCtx, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.DeleteUser(*userCtx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
