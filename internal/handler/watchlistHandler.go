package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type WatchlistHandler struct {
	service service.WatchlistService
}

func NewWatchlistHandler(s service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{service: s}
}

func (h *WatchlistHandler) List(c *gin.Context) {
	userCtx := getUserContext(c)

	items, err := h.service.List(*userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *WatchlistHandler) Add(c *gin.Context) {
	userCtx := getUserContext(c)

	var req model.WatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidSymbol)
		return
	}

	item, err := h.service.Add(*userCtx, req.Symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *WatchlistHandler) Remove(c *gin.Context) {
	userCtx := getUserContext(c)

	err := h.service.Remove(*userCtx, c.Param("symbol"))
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
