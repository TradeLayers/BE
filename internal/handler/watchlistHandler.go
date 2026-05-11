package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WatchlistHandler struct {
	service service.WatchlistService
}

func NewWatchlistHandler(s service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{service: s}
}

func (h *WatchlistHandler) List(c *gin.Context) {
	userCtx := getUserContext(c)

	items, err := h.service.List(c.Request.Context(), *userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *WatchlistHandler) Add(c *gin.Context) {
	userCtx := getUserContext(c)
	ctx := c.Request.Context()
	log := requestLogger(c)

	var req model.WatchlistRequest = model.WatchlistRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("failed to bind watchlist add request", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidSymbol)
		return
	}

	item, err := h.service.Add(ctx, *userCtx, req.Symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *WatchlistHandler) Remove(c *gin.Context) {
	userCtx := getUserContext(c)

	err := h.service.Remove(c.Request.Context(), *userCtx, c.Param("symbol"))
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WatchlistHandler) UpdateThreshold(c *gin.Context) {
	userCtx := getUserContext(c)
	ctx := c.Request.Context()
	log := requestLogger(c)

	var req model.WatchlistThresholdRequest = model.WatchlistThresholdRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("failed to bind watchlist threshold request", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidThreshold)
		return
	}

	item, domainErr := h.service.UpdateThreshold(ctx, *userCtx, c.Param("symbol"), req.ThresholdPrice)
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.JSON(http.StatusOK, item)
}
