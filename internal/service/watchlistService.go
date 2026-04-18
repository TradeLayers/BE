package service

import (
	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"gorm.io/gorm"
)

type WatchlistService interface {
	List(userCtx model.UserContext) ([]model.WatchlistItem, appErrors.DomainError)
	Add(userCtx model.UserContext, symbol string) (*model.WatchlistItem, appErrors.DomainError)
	Remove(userCtx model.UserContext, symbol string) appErrors.DomainError
	UpdateThreshold(userCtx model.UserContext, symbol string, thresholdPrice float64) (*model.WatchlistItem, appErrors.DomainError)
}

type watchlistService struct {
	db            *gorm.DB
	stockRepo     repository.StockRepository
	watchlistRepo repository.WatchlistRepository
	finnhubClient finnhub.Client
	priceMap      *finnhub.PriceMap
	wsClient      *finnhub.WSClient
}

func NewWatchlistService(
	db *gorm.DB,
	stockRepo repository.StockRepository,
	watchlistRepo repository.WatchlistRepository,
	finnhubClient finnhub.Client,
	priceMap *finnhub.PriceMap,
	wsClient *finnhub.WSClient,
) WatchlistService {
	return &watchlistService{
		db:            db,
		stockRepo:     stockRepo,
		watchlistRepo: watchlistRepo,
		finnhubClient: finnhubClient,
		priceMap:      priceMap,
		wsClient:      wsClient,
	}
}

func (s *watchlistService) List(userCtx model.UserContext) ([]model.WatchlistItem, appErrors.DomainError) {
	entries, err := s.watchlistRepo.ListByUser(s.db, userCtx.FirebaseId)
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	if len(entries) == 0 {
		return []model.WatchlistItem{}, appErrors.ErrNone
	}

	stocks, err := s.stockRepo.GetAll()
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	byID := make(map[string]*model.Stock, len(stocks))
	for i := range stocks {
		byID[stocks[i].ID.String()] = &stocks[i]
	}

	items := make([]model.WatchlistItem, 0, len(entries))
	for _, e := range entries {
		stock := byID[e.StockID.String()]
		if stock == nil {
			continue
		}
		items = append(items, model.WatchlistItem{
			Symbol:         stock.Symbol,
			Name:           stock.StockName,
			CurrentPrice:   resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol),
			ThresholdPrice: e.ThresholdPrice,
		})
	}

	return items, appErrors.ErrNone
}

func (s *watchlistService) Add(userCtx model.UserContext, symbolRaw string) (*model.WatchlistItem, appErrors.DomainError) {
	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}

	stock, err := ensureStock(s.stockRepo, s.finnhubClient, symbol)
	if err != nil {
		return nil, appErrors.ErrInternal
	}

	exists, err := s.watchlistRepo.Exists(s.db, userCtx.FirebaseId, stock.ID)
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	if exists {
		return nil, appErrors.ErrAlreadyWatched
	}

	if err := s.watchlistRepo.Add(s.db, userCtx.FirebaseId, stock.ID); err != nil {
		return nil, appErrors.ErrInternal
	}

	s.wsClient.Subscribe([]string{symbol})

	return &model.WatchlistItem{
		Symbol:         stock.Symbol,
		Name:           stock.StockName,
		CurrentPrice:   resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol),
		ThresholdPrice: nil,
	}, appErrors.ErrNone
}

func (s *watchlistService) Remove(userCtx model.UserContext, symbolRaw string) appErrors.DomainError {
	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return domainErr
	}

	stock, err := s.stockRepo.GetBySymbol(symbol)
	if err != nil {
		return appErrors.ErrInternal
	}
	if stock == nil {
		return appErrors.ErrNotWatched
	}

	removed, err := s.watchlistRepo.Remove(s.db, userCtx.FirebaseId, stock.ID)
	if err != nil {
		return appErrors.ErrInternal
	}
	if !removed {
		return appErrors.ErrNotWatched
	}

	return appErrors.ErrNone
}

func (s *watchlistService) UpdateThreshold(userCtx model.UserContext, symbolRaw string, thresholdPrice float64) (*model.WatchlistItem, appErrors.DomainError) {
	if thresholdPrice <= 0 {
		return nil, appErrors.ErrInvalidThreshold
	}

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}

	stock, err := s.stockRepo.GetBySymbol(symbol)
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	if stock == nil {
		return nil, appErrors.ErrNotWatched
	}

	updated, err := s.watchlistRepo.UpdateThreshold(s.db, userCtx.FirebaseId, stock.ID, thresholdPrice)
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	if !updated {
		return nil, appErrors.ErrNotWatched
	}

	threshold := thresholdPrice
	return &model.WatchlistItem{
		Symbol:         stock.Symbol,
		Name:           stock.StockName,
		CurrentPrice:   resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol),
		ThresholdPrice: &threshold,
	}, appErrors.ErrNone
}

