package service

import (
	"context"
	"errors"
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/requestlog"
	"go.uber.org/zap"
)

var DefaultStocks = map[string]string{
	"AAPL":  "Apple Inc",
	"MSFT":  "Microsoft Corp",
	"GOOGL": "Alphabet Inc",
	"AMZN":  "Amazon.com Inc",
	"TSLA":  "Tesla Inc",
	"META":  "Meta Platforms Inc",
	"NVDA":  "NVIDIA Corp",
	"JPM":   "JPMorgan Chase",
	"V":     "Visa Inc",
	"JNJ":   "Johnson & Johnson",
	"WMT":   "Walmart Inc",
	"DIS":   "Walt Disney Co",
	"NFLX":  "Netflix Inc",
	"INTC":  "Intel Corp",
	"AMD":   "Advanced Micro Devices",
}

func DefaultSymbols() []string {
	symbols := make([]string, 0, len(DefaultStocks))
	for s := range DefaultStocks {
		symbols = append(symbols, s)
	}
	return symbols
}

type StockService interface {
	GetAllStocks(ctx context.Context) ([]model.StockListItem, appErrors.DomainError)
	GetQuote(ctx context.Context, symbol string) (*model.StockQuote, appErrors.DomainError)
	GetQuotes(ctx context.Context, symbols []string) ([]model.StockQuote, appErrors.DomainError)
	SearchStocks(ctx context.Context, query string) ([]model.StockSearchResult, appErrors.DomainError)
	GetProfile(ctx context.Context, symbol string) (*model.StockProfile, appErrors.DomainError)
	GetCandles(ctx context.Context, symbols []string, resolution string, from, to int64) (*model.CandlesResponse, appErrors.DomainError)
}

type stockService struct {
	finnhubClient finnhub.Client
	priceMap      *finnhub.PriceMap
	repo          repository.StockRepository
	wsClient      *finnhub.WSClient
}

func NewStockService(client finnhub.Client, priceMap *finnhub.PriceMap, repo repository.StockRepository, wsClient *finnhub.WSClient) StockService {
	return &stockService{
		finnhubClient: client,
		priceMap:      priceMap,
		repo:          repo,
		wsClient:      wsClient,
	}
}

func (s *stockService) getPrice(symbol string) float64 {
	return resolveCurrentPrice(s.priceMap, s.finnhubClient, symbol)
}

func (s *stockService) GetAllStocks(_ context.Context) ([]model.StockListItem, appErrors.DomainError) {
	items := make([]model.StockListItem, 0, len(DefaultStocks))
	for symbol, name := range DefaultStocks {
		items = append(items, model.StockListItem{
			Symbol: symbol,
			Name:   name,
			Price:  s.getPrice(symbol),
		})
	}
	return items, appErrors.ErrNone
}

func (s *stockService) GetQuote(_ context.Context, symbol string) (*model.StockQuote, appErrors.DomainError) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, appErrors.ErrInvalidSymbol
	}

	tp, ok := s.priceMap.Get(symbol)
	if !ok {
		return nil, appErrors.ErrStockNotFound
	}

	return &model.StockQuote{
		Symbol:    symbol,
		Price:     tp.Price,
		Volume:    tp.Volume,
		Timestamp: tp.Timestamp,
	}, appErrors.ErrNone
}

func (s *stockService) GetQuotes(_ context.Context, symbols []string) ([]model.StockQuote, appErrors.DomainError) {
	if len(symbols) == 0 {
		return nil, appErrors.ErrInvalidSymbol
	}

	prices := s.priceMap.GetMulti(symbols)

	quotes := make([]model.StockQuote, 0, len(prices))
	for symbol, tp := range prices {
		quotes = append(quotes, model.StockQuote{
			Symbol:    symbol,
			Price:     tp.Price,
			Volume:    tp.Volume,
			Timestamp: tp.Timestamp,
		})
	}

	return quotes, appErrors.ErrNone
}

func (s *stockService) SearchStocks(ctx context.Context, query string) ([]model.StockSearchResult, appErrors.DomainError) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, appErrors.ErrInvalidSymbol
	}

	resp, err := s.finnhubClient.Search(query)
	if err != nil {
		requestlog.FromContext(ctx).Error("failed to search stocks", zap.String("query", query), zap.Error(err))
		return nil, appErrors.ErrFinnhubUnavailable
	}

	results := make([]model.StockSearchResult, 0, len(resp.Result))
	for _, r := range resp.Result {
		results = append(results, model.StockSearchResult{
			Symbol:      r.Symbol,
			Description: r.Description,
			Type:        r.Type,
		})
	}

	return results, appErrors.ErrNone
}

func (s *stockService) GetProfile(ctx context.Context, symbol string) (*model.StockProfile, appErrors.DomainError) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, appErrors.ErrInvalidSymbol
	}

	resp, err := s.finnhubClient.GetProfile(symbol)
	if err != nil {
		requestlog.FromContext(ctx).Error("failed to fetch stock profile", zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrFinnhubUnavailable
	}

	if resp.Name == "" {
		return nil, appErrors.ErrStockNotFound
	}

	return &model.StockProfile{
		Symbol:    resp.Ticker,
		Name:      resp.Name,
		Logo:      resp.Logo,
		Industry:  resp.Industry,
		Country:   resp.Country,
		Exchange:  resp.Exchange,
		MarketCap: resp.MarketCap,
		WebURL:    resp.WebURL,
		Price:     s.getPrice(symbol),
	}, appErrors.ErrNone
}

func (s *stockService) GetCandles(ctx context.Context, symbols []string, resolution string, from, to int64) (*model.CandlesResponse, appErrors.DomainError) {
	if len(symbols) == 0 {
		return nil, appErrors.ErrInvalidSymbol
	}
	if resolution == "" {
		resolution = "D"
	}

	series := make(map[string]model.CandleSeries)
	sawError := false
	log := requestlog.FromContext(ctx)
	for _, raw := range symbols {
		symbol := strings.ToUpper(strings.TrimSpace(raw))
		if symbol == "" {
			continue
		}
		resp, err := s.finnhubClient.GetCandles(symbol, resolution, from, to)
		if err != nil {
			if errors.Is(err, finnhub.ErrNoData) {
				continue
			}
			log.Warn("failed to fetch stock candles", zap.String("symbol", symbol), zap.Error(err))
			sawError = true
			continue
		}
		series[symbol] = model.CandleSeries{
			Timestamps: resp.Timestamps,
			Close:      resp.Close,
			High:       resp.High,
			Low:        resp.Low,
			Open:       resp.Open,
			Volume:     resp.Volume,
		}
	}

	if len(series) == 0 {
		if sawError {
			return nil, appErrors.ErrHistoricalDataUnavailable
		}
		return nil, appErrors.ErrStockNotFound
	}

	return &model.CandlesResponse{Series: series}, appErrors.ErrNone
}
