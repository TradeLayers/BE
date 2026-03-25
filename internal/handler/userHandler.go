package handler

import (
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

func (h *UserHandler) CreateOrFetchUser(c *gin.Context) {
	userObj, exists := c.Get("userContext")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User context missing"})
		return
	}

	userCtx, ok := userObj.(model.UserContext)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user context"})
		return
	}

	user, state, err := h.service.CreateOrFetchUser(userCtx)
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
	userObj, exists := c.Get("userContext")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User context missing"})
		return
	}

	userCtx, ok := userObj.(model.UserContext)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user context"})
		return
	}

	var req model.UpdateFieldsDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.UpdateFields(userCtx, req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUserAccount(c *gin.Context) {
	userObj, exists := c.Get("userContext")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User context missing"})
		return
	}

	userCtx, ok := userObj.(model.UserContext)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user context"})
		return
	}

	err := h.service.DeleteUser(userCtx)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user context"})
		return
	}

	c.Status(http.StatusNoContent)
}
