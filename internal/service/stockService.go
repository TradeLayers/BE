package service

import (
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
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
	GetAllStocks() ([]model.StockListItem, appErrors.DomainError)
	GetQuote(symbol string) (*model.StockQuote, appErrors.DomainError)
	GetQuotes(symbols []string) ([]model.StockQuote, appErrors.DomainError)
	SearchStocks(query string) ([]model.StockSearchResult, appErrors.DomainError)
	GetProfile(symbol string) (*model.StockProfile, appErrors.DomainError)
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
	if tp, ok := s.priceMap.Get(symbol); ok && tp.Price > 0 {
		return tp.Price
	}

	quote, err := s.finnhubClient.GetQuote(symbol)
	if err == nil && quote.CurrentPrice > 0 {
		s.priceMap.Set(symbol, quote.CurrentPrice, 0, quote.Timestamp)
		return quote.CurrentPrice
	}

	return 0
}

func (s *stockService) GetAllStocks() ([]model.StockListItem, appErrors.DomainError) {
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

func (s *stockService) GetQuote(symbol string) (*model.StockQuote, appErrors.DomainError) {
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

func (s *stockService) GetQuotes(symbols []string) ([]model.StockQuote, appErrors.DomainError) {
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

func (s *stockService) SearchStocks(query string) ([]model.StockSearchResult, appErrors.DomainError) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, appErrors.ErrInvalidSymbol
	}

	resp, err := s.finnhubClient.Search(query)
	if err != nil {
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

func (s *stockService) GetProfile(symbol string) (*model.StockProfile, appErrors.DomainError) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, appErrors.ErrInvalidSymbol
	}

	resp, err := s.finnhubClient.GetProfile(symbol)
	if err != nil {
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
