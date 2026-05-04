package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StockHandler struct {
	service service.StockService
}

func NewStockHandler(s service.StockService) *StockHandler {
	return &StockHandler{service: s}
}

func (h *StockHandler) GetAllStocks(c *gin.Context) {
	stocks, err := h.service.GetAllStocks(c.Request.Context())
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, stocks)
}

func (h *StockHandler) GetQuote(c *gin.Context) {
	symbol := c.Param("symbol")

	quote, err := h.service.GetQuote(c.Request.Context(), symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, quote)
}

func (h *StockHandler) GetQuotes(c *gin.Context) {
	log := requestLogger(c)

	var req model.QuotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("failed to bind stock quotes request", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidSymbol)
		return
	}

	quotes, err := h.service.GetQuotes(c.Request.Context(), req.Symbols)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, quotes)
}

func (h *StockHandler) SearchStocks(c *gin.Context) {
	query := c.Query("q")

	results, err := h.service.SearchStocks(c.Request.Context(), query)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *StockHandler) GetProfile(c *gin.Context) {
	symbol := c.Param("symbol")

	profile, err := h.service.GetProfile(c.Request.Context(), symbol)
	if err != appErrors.ErrNone {
		appErrors.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *StockHandler) GetCandles(c *gin.Context) {
	log := requestLogger(c)
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		log.Warn("missing symbols query parameter")
		appErrors.ReturnError(c, appErrors.ErrInvalidSymbol)
		return
	}

	symbols := splitSymbols(symbolsParam)
	resolution := c.DefaultQuery("resolution", "D")

	from, err := strconv.ParseInt(c.Query("from"), 10, 64)
	if err != nil {
		log.Warn("invalid candles from query parameter", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}
	to, err := strconv.ParseInt(c.Query("to"), 10, 64)
	if err != nil {
		log.Warn("invalid candles to query parameter", zap.Error(err))
		appErrors.ReturnError(c, appErrors.ErrInvalidFieldInformation)
		return
	}

	candles, domainErr := h.service.GetCandles(c.Request.Context(), symbols, resolution, from, to)
	if domainErr != appErrors.ErrNone {
		appErrors.ReturnError(c, domainErr)
		return
	}

	c.JSON(http.StatusOK, candles)
}

func splitSymbols(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
