package handler

import (
	"net/http"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	service service.StockService
}

func NewStockHandler(s service.StockService) *StockHandler {
	return &StockHandler{service: s}
}

func (h *StockHandler) GetQuote(c *gin.Context) {
	symbol := c.Param("symbol")

	quote, err := h.service.GetQuote(symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, quote)
}

func (h *StockHandler) GetQuotes(c *gin.Context) {
	var req model.QuotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.ReturnError(c, appErrors.ErrInvalidSymbol)
		return
	}

	quotes, err := h.service.GetQuotes(req.Symbols)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, quotes)
}

func (h *StockHandler) SearchStocks(c *gin.Context) {
	query := c.Query("q")

	results, err := h.service.SearchStocks(query)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *StockHandler) GetProfile(c *gin.Context) {
	symbol := c.Param("symbol")

	profile, err := h.service.GetProfile(symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}
