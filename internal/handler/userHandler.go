package handler

import (
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
		c.JSON(400, gin.H{"error": "User context missing"})
		return
	}

	userCtx, ok := userObj.(model.UserContext)
	if !ok {
		c.JSON(400, gin.H{"error": "Invalid user context"})
		return
	}

	user, state, err := h.service.CreateOrFetchUser(userCtx)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	code := 201
	if state == model.UserFetched {
		code = 200
	}

	c.JSON(code, user)
}
