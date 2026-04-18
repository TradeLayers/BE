package handler

import (
	"net/http"
	"time"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	service service.PortfolioService
}

func NewPortfolioHandler(s service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{service: s}
}

func (h *PortfolioHandler) Buy(c *gin.Context) {
	userCtx := getUserContext(c)

	var req model.TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	result, err := h.service.Buy(*userCtx, req.Symbol, req.Quantity)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *PortfolioHandler) Sell(c *gin.Context) {
	userCtx := getUserContext(c)

	var req model.TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	result, err := h.service.Sell(*userCtx, req.Symbol, req.Quantity)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *PortfolioHandler) GetHoldings(c *gin.Context) {
	userCtx := getUserContext(c)

	holdings, err := h.service.GetHoldings(*userCtx)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, holdings)
}

func (h *PortfolioHandler) GetTransactions(c *gin.Context) {
	userCtx := getUserContext(c)

	var symbolFilter *string
	if s := c.Query("stock"); s != "" {
		symbolFilter = &s
	}

	from, fromErr := parseOptionalTime(c.Query("from"))
	if fromErr != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}
	to, toErr := parseOptionalTime(c.Query("to"))
	if toErr != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	txs, err := h.service.GetTransactions(*userCtx, symbolFilter, from, to)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, txs)
}

func (h *PortfolioHandler) GetHistory(c *gin.Context) {
	userCtx := getUserContext(c)

	history, err := h.service.GetHistory(*userCtx, c.Query("interval"))
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, history)
}

func parseOptionalTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
