package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PortfolioService interface {
	Buy(ctx context.Context, userCtx model.UserContext, symbol string, qty float64) (*model.TradeResult, appErrors.DomainError)
	Sell(ctx context.Context, userCtx model.UserContext, symbol string, qty float64) (*model.TradeResult, appErrors.DomainError)
	GetHoldings(ctx context.Context, userCtx model.UserContext) ([]model.HoldingView, appErrors.DomainError)
	GetTransactions(ctx context.Context, userCtx model.UserContext, symbol *string, from, to *time.Time) ([]model.TransactionView, appErrors.DomainError)
	GetHistory(ctx context.Context, userCtx model.UserContext, interval string) (*model.PortfolioHistoryResponse, appErrors.DomainError)
}

type portfolioService struct {
	db            *gorm.DB
	stockRepo     repository.StockRepository
	holdingsRepo  repository.HoldingsRepository
	txRepo        repository.TransactionsRepository
	finnhubClient finnhub.Client
	priceMap      *finnhub.PriceMap
	wsClient      *finnhub.WSClient
}

// errDomainRollback aborts the GORM transaction so the caller-captured
// appErrors.DomainError can be returned to the client.
var errDomainRollback error = errors.New("domain rollback")

func NewPortfolioService(
	db *gorm.DB,
	stockRepo repository.StockRepository,
	holdingsRepo repository.HoldingsRepository,
	txRepo repository.TransactionsRepository,
	finnhubClient finnhub.Client,
	priceMap *finnhub.PriceMap,
	wsClient *finnhub.WSClient,
) PortfolioService {
	return &portfolioService{
		db:            db,
		stockRepo:     stockRepo,
		holdingsRepo:  holdingsRepo,
		txRepo:        txRepo,
		finnhubClient: finnhubClient,
		priceMap:      priceMap,
		wsClient:      wsClient,
	}
}

func (s *portfolioService) Buy(ctx context.Context, userCtx model.UserContext, symbolRaw string, qtyF float64) (*model.TradeResult, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}
	if qtyF <= 0 {
		return nil, appErrors.ErrInvalidQuantity
	}

	qty := decimal.NewFromFloat(qtyF)

	stock, err := ensureStock(ctx, s.stockRepo, s.finnhubClient, symbol)
	if err != nil {
		log.Error("failed to ensure stock for buy", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	priceF := resolveCurrentPrice(s.priceMap, s.finnhubClient, symbol)
	if priceF <= 0 {
		return nil, appErrors.ErrStockNotFound
	}
	price := decimal.NewFromFloat(priceF)
	totalCost := price.Mul(qty)

	s.wsClient.Subscribe([]string{symbol})

	var resultTx *model.StockTransaction = nil
	var newBalance decimal.Decimal = decimal.Decimal{}
	domainFailure := appErrors.ErrNone

	txErr := s.db.WithContext(ctx).Transaction(func(txdb *gorm.DB) error {
		var user model.User = model.User{}
		if err := txdb.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("firebase_id = ?", userCtx.FirebaseId).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				domainFailure = appErrors.ErrUserNotFound
				return errDomainRollback
			}
			return err
		}

		if user.Balance.LessThan(totalCost) {
			domainFailure = appErrors.ErrInsufficientBalance
			return errDomainRollback
		}

		if err := s.holdingsRepo.Upsert(ctx, txdb, userCtx.FirebaseId, stock.ID, qty); err != nil {
			return err
		}

		txn := &model.StockTransaction{
			UserID:          userCtx.FirebaseId,
			StockID:         stock.ID,
			Price:           price,
			Quantity:        qty,
			TransactionType: model.TransactionTypeBought,
		}
		if err := s.txRepo.Create(ctx, txdb, txn); err != nil {
			return err
		}

		newBal := user.Balance.Sub(totalCost)
		if err := txdb.Model(&model.User{}).
			Where("firebase_id = ?", userCtx.FirebaseId).
			Update("balance", newBal).Error; err != nil {
			return err
		}

		resultTx = txn
		newBalance = newBal
		return nil
	})

	if domainFailure != appErrors.ErrNone {
		return nil, domainFailure
	}
	if txErr != nil {
		log.Error("buy transaction failed", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(txErr))
		return nil, appErrors.ErrInternal
	}

	return buildTradeResult(resultTx, stock, priceF, qtyF, newBalance), appErrors.ErrNone
}

func (s *portfolioService) Sell(ctx context.Context, userCtx model.UserContext, symbolRaw string, qtyF float64) (*model.TradeResult, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	symbol, domainErr := normalizeSymbol(symbolRaw)
	if domainErr != appErrors.ErrNone {
		return nil, domainErr
	}
	if qtyF <= 0 {
		return nil, appErrors.ErrInvalidQuantity
	}

	qty := decimal.NewFromFloat(qtyF)

	stock, err := s.stockRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		log.Error("failed to fetch stock for sell", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if stock == nil {
		return nil, appErrors.ErrInsufficientHoldings
	}

	priceF := resolveCurrentPrice(s.priceMap, s.finnhubClient, symbol)
	if priceF <= 0 {
		return nil, appErrors.ErrStockNotFound
	}
	price := decimal.NewFromFloat(priceF)
	proceeds := price.Mul(qty)

	var resultTx *model.StockTransaction = nil
	var newBalance decimal.Decimal = decimal.Decimal{}
	domainFailure := appErrors.ErrNone

	txErr := s.db.WithContext(ctx).Transaction(func(txdb *gorm.DB) error {
		var user model.User = model.User{}
		if err := txdb.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("firebase_id = ?", userCtx.FirebaseId).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				domainFailure = appErrors.ErrUserNotFound
				return errDomainRollback
			}
			return err
		}

		holding, err := s.holdingsRepo.GetOne(ctx, txdb, userCtx.FirebaseId, stock.ID)
		if err != nil {
			return err
		}
		if holding == nil || holding.Quantity.LessThan(qty) {
			domainFailure = appErrors.ErrInsufficientHoldings
			return errDomainRollback
		}

		remaining := holding.Quantity.Sub(qty)
		if remaining.IsZero() {
			if err := s.holdingsRepo.Delete(ctx, txdb, userCtx.FirebaseId, stock.ID); err != nil {
				return err
			}
		} else {
			if err := s.holdingsRepo.SetQuantity(ctx, txdb, userCtx.FirebaseId, stock.ID, remaining); err != nil {
				return err
			}
		}

		txn := &model.StockTransaction{
			UserID:          userCtx.FirebaseId,
			StockID:         stock.ID,
			Price:           price,
			Quantity:        qty,
			TransactionType: model.TransactionTypeSold,
		}
		if err := s.txRepo.Create(ctx, txdb, txn); err != nil {
			return err
		}

		newBal := user.Balance.Add(proceeds)
		if err := txdb.Model(&model.User{}).
			Where("firebase_id = ?", userCtx.FirebaseId).
			Update("balance", newBal).Error; err != nil {
			return err
		}

		resultTx = txn
		newBalance = newBal
		return nil
	})

	if domainFailure != appErrors.ErrNone {
		return nil, domainFailure
	}
	if txErr != nil {
		log.Error("sell transaction failed", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", symbol), zap.Error(txErr))
		return nil, appErrors.ErrInternal
	}

	return buildTradeResult(resultTx, stock, priceF, qtyF, newBalance), appErrors.ErrNone
}

func (s *portfolioService) GetHoldings(ctx context.Context, userCtx model.UserContext) ([]model.HoldingView, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	holdings, err := s.holdingsRepo.GetByUser(ctx, s.db, userCtx.FirebaseId)
	if err != nil {
		log.Error("failed to load holdings", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	if len(holdings) == 0 {
		return []model.HoldingView{}, appErrors.ErrNone
	}

	stocksByID, err := s.stocksByID(ctx)
	if err != nil {
		log.Error("failed to map stocks for holdings", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	views := make([]model.HoldingView, 0, len(holdings))
	for _, h := range holdings {
		stock := stocksByID[h.StockID]
		if stock == nil {
			continue
		}
		qtyF, _ := h.Quantity.Float64()
		views = append(views, model.HoldingView{
			StockID:      stock.ID,
			Symbol:       stock.Symbol,
			Name:         stock.StockName,
			Quantity:     qtyF,
			CurrentPrice: resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol),
		})
	}

	return views, appErrors.ErrNone
}

func (s *portfolioService) GetTransactions(ctx context.Context, userCtx model.UserContext, symbolFilter *string, from, to *time.Time) ([]model.TransactionView, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	var stockIDFilter *uuid.UUID = nil
	if symbolFilter != nil {
		sym, domainErr := normalizeSymbol(*symbolFilter)
		if domainErr != appErrors.ErrNone {
			return nil, domainErr
		}
		stock, err := s.stockRepo.GetBySymbol(ctx, sym)
		if err != nil {
			log.Error("failed to fetch stock for transaction filter", zap.String("firebase_id", userCtx.FirebaseId), zap.String("symbol", sym), zap.Error(err))
			return nil, appErrors.ErrInternal
		}
		if stock == nil {
			return []model.TransactionView{}, appErrors.ErrNone
		}
		stockIDFilter = &stock.ID
	}

	txs, err := s.txRepo.ListByUser(ctx, s.db, userCtx.FirebaseId, stockIDFilter, from, to)
	if err != nil {
		log.Error("failed to load transactions", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if len(txs) == 0 {
		return []model.TransactionView{}, appErrors.ErrNone
	}

	stocksByID, err := s.stocksByID(ctx)
	if err != nil {
		log.Error("failed to map stocks for transactions", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	views := make([]model.TransactionView, 0, len(txs))
	for _, t := range txs {
		stock := stocksByID[t.StockID]
		if stock == nil {
			continue
		}
		priceF, _ := t.Price.Float64()
		qtyF, _ := t.Quantity.Float64()
		views = append(views, model.TransactionView{
			ID:              t.ID,
			Symbol:          stock.Symbol,
			Name:            stock.StockName,
			Price:           priceF,
			Quantity:        qtyF,
			TransactionDate: t.TransactionDate,
			TransactionType: t.TransactionType,
		})
	}

	return views, appErrors.ErrNone
}

func (s *portfolioService) GetHistory(ctx context.Context, userCtx model.UserContext, interval string) (*model.PortfolioHistoryResponse, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	from := intervalStart(interval)

	allTxs, err := s.txRepo.ListByUser(ctx, s.db, userCtx.FirebaseId, nil, nil, nil)
	if err != nil {
		log.Error("failed to load all portfolio history transactions", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}
	if len(allTxs) == 0 {
		currentValue, err := s.currentPortfolioValue(ctx, userCtx.FirebaseId)
		if err != nil {
			log.Error("failed to calculate current portfolio value", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
			return nil, appErrors.ErrInternal
		}
		return &model.PortfolioHistoryResponse{
			Points:       []model.PortfolioHistoryPoint{},
			MarketValue:  []model.PortfolioMarketValuePoint{},
			CurrentValue: currentValue,
		}, appErrors.ErrNone
	}

	txs, err := s.txRepo.ListByUser(ctx, s.db, userCtx.FirebaseId, nil, from, nil)
	if err != nil {
		log.Error("failed to load portfolio history transactions", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	// For cumulative math we must start from the user's baseline invested at
	// the start of the range, so pull everything before "from" and fold it in.
	var baseline decimal.Decimal = decimal.Decimal{}
	if from != nil {
		earlier, err := s.txRepo.ListByUser(ctx, s.db, userCtx.FirebaseId, nil, nil, from)
		if err != nil {
			log.Error("failed to load baseline portfolio transactions", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
			return nil, appErrors.ErrInternal
		}
		for IIndex := range earlier {
			baseline = applyInvestmentDelta(baseline, &earlier[IIndex])
		}
	}

	running := baseline
	points := make([]model.PortfolioHistoryPoint, 0, len(txs))
	for IIndex := range txs {
		running = applyInvestmentDelta(running, &txs[IIndex])
		investedF, _ := running.Float64()
		points = append(points, model.PortfolioHistoryPoint{
			Date:            txs[IIndex].TransactionDate,
			InvestedCapital: investedF,
		})
	}

	currentValue, err := s.currentPortfolioValue(ctx, userCtx.FirebaseId)
	if err != nil {
		log.Error("failed to calculate current portfolio value", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	marketValue, err := s.marketValueHistory(ctx, allTxs, from)
	if err != nil {
		log.Error("failed to calculate market value history", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	return &model.PortfolioHistoryResponse{
		Points:       points,
		MarketValue:  marketValue,
		CurrentValue: currentValue,
	}, appErrors.ErrNone
}

func (s *portfolioService) marketValueHistory(ctx context.Context, txs []model.StockTransaction, from *time.Time) ([]model.PortfolioMarketValuePoint, error) {
	if len(txs) == 0 {
		return []model.PortfolioMarketValuePoint{}, nil
	}

	stocksByID, err := s.stocksByID(ctx)
	if err != nil {
		return nil, err
	}

	start := txs[0].TransactionDate
	if from != nil && from.Before(start) {
		start = *from
	}
	to := time.Now()

	candlesByStock := make(map[uuid.UUID]model.CandleSeries)
	for _, tx := range txs {
		if _, ok := candlesByStock[tx.StockID]; ok {
			continue
		}
		stock := stocksByID[tx.StockID]
		if stock == nil {
			continue
		}
		resp, err := s.finnhubClient.GetCandles(stock.Symbol, "D", start.Unix(), to.Unix())
		if err != nil {
			continue
		}
		candlesByStock[tx.StockID] = model.CandleSeries{
			Timestamps: resp.Timestamps,
			Close:      resp.Close,
			High:       resp.High,
			Low:        resp.Low,
			Open:       resp.Open,
			Volume:     resp.Volume,
		}
	}

	holdings := make(map[uuid.UUID]decimal.Decimal)
	points := make([]model.PortfolioMarketValuePoint, 0, len(txs))
	for _, tx := range txs {
		if tx.TransactionType == model.TransactionTypeSold {
			holdings[tx.StockID] = holdings[tx.StockID].Sub(tx.Quantity)
			if !holdings[tx.StockID].GreaterThan(decimal.Zero) {
				delete(holdings, tx.StockID)
			}
		} else {
			holdings[tx.StockID] = holdings[tx.StockID].Add(tx.Quantity)
		}

		if from != nil && tx.TransactionDate.Before(*from) {
			continue
		}

		total := decimal.Zero
		for stockID, qty := range holdings {
			series := candlesByStock[stockID]
			closePrice := closeAtOrBefore(&series, tx.TransactionDate)
			if closePrice <= 0 {
				continue
			}
			total = total.Add(qty.Mul(decimal.NewFromFloat(closePrice)))
		}
		value, _ := total.Float64()
		points = append(points, model.PortfolioMarketValuePoint{
			Date:  tx.TransactionDate,
			Value: value,
		})
	}

	return points, nil
}

func closeAtOrBefore(series *model.CandleSeries, at time.Time) float64 {
	target := at.Unix()
	for IIndex := len(series.Timestamps) - 1; IIndex >= 0; IIndex-- {
		if series.Timestamps[IIndex] <= target && IIndex < len(series.Close) {
			return series.Close[IIndex]
		}
	}
	return 0
}

func (s *portfolioService) currentPortfolioValue(ctx context.Context, userID string) (float64, error) {
	holdings, err := s.holdingsRepo.GetByUser(ctx, s.db, userID)
	if err != nil {
		return 0, err
	}
	if len(holdings) == 0 {
		return 0, nil
	}

	stocksByID, err := s.stocksByID(ctx)
	if err != nil {
		return 0, err
	}

	total := decimal.Zero
	for _, h := range holdings {
		stock := stocksByID[h.StockID]
		if stock == nil {
			continue
		}
		price := resolveCurrentPrice(s.priceMap, s.finnhubClient, stock.Symbol)
		if price <= 0 {
			continue
		}
		total = total.Add(h.Quantity.Mul(decimal.NewFromFloat(price)))
	}
	totalF, _ := total.Float64()
	return totalF, nil
}

func (s *portfolioService) stocksByID(ctx context.Context) (map[uuid.UUID]*model.Stock, error) {
	all, err := s.stockRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*model.Stock, len(all))
	for IIndex := range all {
		result[all[IIndex].ID] = &all[IIndex]
	}
	return result, nil
}

func applyInvestmentDelta(running decimal.Decimal, t *model.StockTransaction) decimal.Decimal {
	amount := t.Price.Mul(t.Quantity)
	if t.TransactionType == model.TransactionTypeSold {
		return running.Sub(amount)
	}
	return running.Add(amount)
}

func intervalStart(interval string) *time.Time {
	now := time.Now()
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "daily":
		t := now.Add(-24 * time.Hour)
		return &t
	case "weekly":
		t := now.Add(-7 * 24 * time.Hour)
		return &t
	case "monthly":
		t := now.Add(-30 * 24 * time.Hour)
		return &t
	default:
		return nil
	}
}

func normalizeSymbol(raw string) (string, appErrors.DomainError) {
	symbol := strings.ToUpper(strings.TrimSpace(raw))
	if symbol == "" || len(symbol) > 32 {
		return "", appErrors.ErrInvalidSymbol
	}
	return symbol, appErrors.ErrNone
}

func buildTradeResult(tx *model.StockTransaction, stock *model.Stock, price, qty float64, balance decimal.Decimal) *model.TradeResult {
	balF, _ := balance.Float64()
	return &model.TradeResult{
		Transaction: model.TransactionView{
			ID:              tx.ID,
			Symbol:          stock.Symbol,
			Name:            stock.StockName,
			Price:           price,
			Quantity:        qty,
			TransactionDate: tx.TransactionDate,
			TransactionType: tx.TransactionType,
		},
		Balance: balF,
	}
}
