package service

import (
	"context"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/requestlog"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WatchlistService interface {
	List(ctx context.Context, userCtx model.UserContext) ([]model.WatchlistItem, appErrors.DomainError)
	Add(ctx context.Context, userCtx model.UserContext, symbol string) (*model.WatchlistItem, appErrors.DomainError)
	Remove(ctx context.Context, userCtx model.UserContext, symbol string) appErrors.DomainError
	UpdateThreshold(ctx context.Context, userCtx model.UserContext, symbol string, thresholdPrice float64) (*model.WatchlistItem, appErrors.DomainError)
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

func (s *watchlistService) List(ctx context.Context, userCtx model.UserContext) ([]model.WatchlistItem, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	entries, err := s.watchlistRepo.ListByUser(ctx, s.db, userCtx.FirebaseId)
	if err != nil {
		log.Error("failed to list watchlist entries", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if len(entries) == 0 {
		return []model.WatchlistItem{}, appErrors.ErrNone
	}

	stocks, err := s.stockRepo.GetAll(ctx)
	if err != nil {
		log.Error("failed to load stocks for watchlist", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	byID := make(map[string]*model.Stock, len(stocks))
	for IIndex := range stocks {
		byID[stocks[IIndex].ID.String()] = &stocks[IIndex]
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

func (s *watchlistService) Add(ctx context.Context, userCtx model.UserContext, symbolRaw string) (*model.WatchlistItem, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}

	stock, err := ensureStock(ctx, s.stockRepo, s.finnhubClient, symbol)
	if err != nil {
		log.Error("failed to ensure stock for watchlist", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	exists, err := s.watchlistRepo.Exists(ctx, s.db, userCtx.FirebaseId, stock.ID)
	if err != nil {
		log.Error("failed to check existing watchlist entry", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if exists {
		return nil, appErrors.ErrAlreadyWatched
	}

	if err := s.watchlistRepo.Add(ctx, s.db, userCtx.FirebaseId, stock.ID); err != nil {
		log.Error("failed to add watchlist entry", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
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

func (s *watchlistService) Remove(ctx context.Context, userCtx model.UserContext, symbolRaw string) appErrors.DomainError {
	log := requestlog.FromContext(ctx)

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return domainErr
	}

	stock, err := s.stockRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		log.Error("failed to fetch stock for watchlist removal", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return appErrors.ErrInternal
	}
	if stock == nil {
		return appErrors.ErrNotWatched
	}

	removed, err := s.watchlistRepo.Remove(ctx, s.db, userCtx.FirebaseId, stock.ID)
	if err != nil {
		log.Error("failed to remove watchlist entry", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return appErrors.ErrInternal
	}
	if !removed {
		return appErrors.ErrNotWatched
	}

	return appErrors.ErrNone
}

func (s *watchlistService) UpdateThreshold(ctx context.Context, userCtx model.UserContext, symbolRaw string, thresholdPrice float64) (*model.WatchlistItem, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	if thresholdPrice <= 0 {
		return nil, appErrors.ErrInvalidThreshold
	}

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}

	stock, err := s.stockRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		log.Error("failed to fetch stock for watchlist threshold update", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if stock == nil {
		return nil, appErrors.ErrNotWatched
	}

	updated, err := s.watchlistRepo.UpdateThreshold(ctx, s.db, userCtx.FirebaseId, stock.ID, thresholdPrice)
	if err != nil {
		log.Error("failed to update watchlist threshold", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
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
