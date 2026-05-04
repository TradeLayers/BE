package service

import (
	"context"
	"fmt"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const unreadNotificationsLimit = 50

type NotificationService interface {
	ListUnread(ctx context.Context, userCtx model.UserContext) ([]model.ThresholdNotificationView, appErrors.DomainError)
	MarkRead(ctx context.Context, userCtx model.UserContext, notificationID string) appErrors.DomainError
}

type notificationService struct {
	db               *gorm.DB
	watchlistRepo    repository.WatchlistRepository
	stockRepo        repository.StockRepository
	notificationRepo repository.NotificationRepository
	finnhubClient    finnhub.Client
	priceMap         *finnhub.PriceMap
	wsClient         *finnhub.WSClient
}

func NewNotificationService(
	db *gorm.DB,
	watchlistRepo repository.WatchlistRepository,
	stockRepo repository.StockRepository,
	notificationRepo repository.NotificationRepository,
	finnhubClient finnhub.Client,
	priceMap *finnhub.PriceMap,
	wsClient *finnhub.WSClient,
) NotificationService {
	return &notificationService{
		db:               db,
		watchlistRepo:    watchlistRepo,
		stockRepo:        stockRepo,
		notificationRepo: notificationRepo,
		finnhubClient:    finnhubClient,
		priceMap:         priceMap,
		wsClient:         wsClient,
	}
}

func (s *notificationService) ListUnread(ctx context.Context, userCtx model.UserContext) ([]model.ThresholdNotificationView, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	if err := s.evaluateThresholdCrossings(ctx, userCtx.FirebaseId); err != nil {
		log.Error("failed to evaluate notification thresholds", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	notifications, err := s.notificationRepo.ListUnreadByUser(ctx, s.db, userCtx.FirebaseId, unreadNotificationsLimit)
	if err != nil {
		log.Error("failed to list unread notifications", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	views := make([]model.ThresholdNotificationView, 0, len(notifications))
	for _, n := range notifications {
		views = append(views, model.ThresholdNotificationView{
			ID:             n.ID.String(),
			Symbol:         n.Symbol,
			ThresholdPrice: n.ThresholdPrice,
			TriggerPrice:   n.TriggerPrice,
			TriggeredAt:    n.CreatedAt,
			Message: fmt.Sprintf(
				"%s fell to or below your threshold %.2f (current %.2f)",
				n.Symbol,
				n.ThresholdPrice,
				n.TriggerPrice,
			),
		})
	}

	return views, appErrors.ErrNone
}

func (s *notificationService) MarkRead(ctx context.Context, userCtx model.UserContext, notificationID string) appErrors.DomainError {
	id, err := uuid.Parse(notificationID)
	if err != nil {
		return appErrors.ErrInvalidFieldInformation
	}

	_, err = s.notificationRepo.MarkReadByUser(ctx, s.db, userCtx.FirebaseId, id)
	if err != nil {
		requestlog.FromContext(ctx).Error("failed to mark notification as read", zap.String("firebase_id", userCtx.FirebaseId), zap.String("notification_id", notificationID), zap.Error(err))
		return appErrors.ErrInternal
	}

	return appErrors.ErrNone
}

func (s *notificationService) evaluateThresholdCrossings(ctx context.Context, userID string) error {
	entries, err := s.watchlistRepo.ListByUser(ctx, s.db, userID)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	stocks, err := s.stockRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	stockByID := make(map[string]*model.Stock, len(stocks))
	for i := range stocks {
		stockByID[stocks[i].ID.String()] = &stocks[i]
	}

	symbolsToSubscribe := make([]string, 0, len(entries))
	for _, entry := range entries {
		stock := stockByID[entry.StockID.String()]
		if stock == nil || entry.ThresholdPrice == nil {
			continue
		}

		symbolsToSubscribe = append(symbolsToSubscribe, stock.Symbol)
		currentPrice := resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol)
		if currentPrice <= 0 {
			continue
		}

		isReached := currentPrice <= *entry.ThresholdPrice
		if isReached && !entry.ThresholdReached {
			notification := &model.ThresholdNotification{
				UserID:         userID,
				StockID:        entry.StockID,
				Symbol:         stock.Symbol,
				ThresholdPrice: *entry.ThresholdPrice,
				TriggerPrice:   currentPrice,
			}
			if err := s.notificationRepo.Create(ctx, s.db, notification); err != nil {
				return err
			}
			if _, err := s.watchlistRepo.UpdateThresholdReached(ctx, s.db, userID, entry.StockID, true); err != nil {
				return err
			}
			continue
		}

		if !isReached && entry.ThresholdReached {
			if _, err := s.watchlistRepo.UpdateThresholdReached(ctx, s.db, userID, entry.StockID, false); err != nil {
				return err
			}
		}
	}

	if len(symbolsToSubscribe) > 0 {
		s.wsClient.Subscribe(symbolsToSubscribe)
	}

	return nil
}
