package service

import (
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AlertService interface {
	List(userCtx model.UserContext) ([]model.AlertView, appErrors.DomainError)
	Create(userCtx model.UserContext, req model.AlertRequest) (*model.AlertView, appErrors.DomainError)
	Delete(userCtx model.UserContext, id string) appErrors.DomainError
}

type alertService struct {
	db            *gorm.DB
	stockRepo     repository.StockRepository
	alertRepo     repository.AlertRepository
	finnhubClient finnhub.Client
	priceMap      *finnhub.PriceMap
	wsClient      *finnhub.WSClient
}

func NewAlertService(
	db *gorm.DB,
	stockRepo repository.StockRepository,
	alertRepo repository.AlertRepository,
	finnhubClient finnhub.Client,
	priceMap *finnhub.PriceMap,
	wsClient *finnhub.WSClient,
) AlertService {
	return &alertService{
		db:            db,
		stockRepo:     stockRepo,
		alertRepo:     alertRepo,
		finnhubClient: finnhubClient,
		priceMap:      priceMap,
		wsClient:      wsClient,
	}
}

func (s *alertService) List(userCtx model.UserContext) ([]model.AlertView, appErrors.DomainError) {
	alerts, err := s.alertRepo.ListByUser(s.db, userCtx.FirebaseId)
	if err != nil {
		return nil, appErrors.ErrInternal
	}
	if len(alerts) == 0 {
		return []model.AlertView{}, appErrors.ErrNone
	}

	stocksByID, err := stocksByID(s.stockRepo)
	if err != nil {
		return nil, appErrors.ErrInternal
	}

	views := make([]model.AlertView, 0, len(alerts))
	for _, alert := range alerts {
		stock := stocksByID[alert.StockID]
		if stock == nil {
			continue
		}

		currentPrice := resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol)
		if alert.TriggeredAt == nil && alertCrossed(alert, currentPrice) {
			updated, err := s.alertRepo.MarkTriggered(s.db, alert.ID)
			if err != nil {
				return nil, appErrors.ErrInternal
			}
			alert = *updated
		}

		views = append(views, alertView(alert, stock, currentPrice))
	}

	return views, appErrors.ErrNone
}

func (s *alertService) Create(userCtx model.UserContext, req model.AlertRequest) (*model.AlertView, appErrors.DomainError) {
	symbol, domainErr := normalizeSymbol(req.Symbol)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}
	if req.ThresholdPrice <= 0 {
		return nil, appErrors.ErrInvalidFieldInformation
	}

	direction := model.AlertDirection(strings.ToLower(strings.TrimSpace(string(req.Direction))))
	if direction != model.AlertDirectionAbove && direction != model.AlertDirectionBelow {
		return nil, appErrors.ErrInvalidFieldInformation
	}

	stock, err := ensureStock(s.stockRepo, s.finnhubClient, symbol)
	if err != nil {
		return nil, appErrors.ErrInternal
	}

	alert := &model.Alert{
		UserID:         userCtx.FirebaseId,
		StockID:        stock.ID,
		ThresholdPrice: decimal.NewFromFloat(req.ThresholdPrice),
		Direction:      direction,
	}
	if err := s.alertRepo.Create(s.db, alert); err != nil {
		return nil, appErrors.ErrInternal
	}

	s.wsClient.Subscribe([]string{symbol})

	currentPrice := resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol)
	return ptr(alertView(*alert, stock, currentPrice)), appErrors.ErrNone
}

func (s *alertService) Delete(userCtx model.UserContext, id string) appErrors.DomainError {
	alertID, err := uuid.Parse(id)
	if err != nil {
		return appErrors.ErrInvalidFieldInformation
	}

	deleted, err := s.alertRepo.DeleteByUser(s.db, userCtx.FirebaseId, alertID)
	if err != nil {
		return appErrors.ErrInternal
	}
	if !deleted {
		return appErrors.ErrAlertNotFound
	}

	return appErrors.ErrNone
}

func alertCrossed(alert model.Alert, currentPrice float64) bool {
	if currentPrice <= 0 {
		return false
	}
	threshold, _ := alert.ThresholdPrice.Float64()
	switch alert.Direction {
	case model.AlertDirectionAbove:
		return currentPrice >= threshold
	case model.AlertDirectionBelow:
		return currentPrice <= threshold
	default:
		return false
	}
}

func alertView(alert model.Alert, stock *model.Stock, currentPrice float64) model.AlertView {
	threshold, _ := alert.ThresholdPrice.Float64()
	return model.AlertView{
		ID:             alert.ID,
		Symbol:         stock.Symbol,
		Name:           stock.StockName,
		ThresholdPrice: threshold,
		Direction:      alert.Direction,
		CurrentPrice:   currentPrice,
		TriggeredAt:    alert.TriggeredAt,
		CreatedAt:      alert.CreatedAt,
	}
}

func ptr[T any](v T) *T {
	return &v
}
