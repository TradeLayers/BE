package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type CreatePortfolioRequest struct {
	Name           string   `json:"name"`
	InitialBalance *float64 `json:"initial_balance"`
}

type PortfolioHandler struct {
	service service.PortfolioService
}

func NewPortfolioHandler(s service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{service: s}
}

func (h *PortfolioHandler) CreatePortfolio(c *gin.Context) {
	var req CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.InitialBalance == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and initial_balance are required"})
		return
	}

	portfolio, err := h.service.CreatePortfolio(c.Request.Context(), req.Name, *req.InitialBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create portfolio", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, portfolio)
}

