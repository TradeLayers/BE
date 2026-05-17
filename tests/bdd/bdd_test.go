//go:build bdd

package bdd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/service"
	"github.com/cucumber/godog"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testUserID = "bdd-user"

type fakeMarketClient struct {
	prices map[string]float64
}

func (c fakeMarketClient) GetProfile(symbol string) (*finnhub.ProfileResponse, error) {
	return &finnhub.ProfileResponse{Name: symbol, Ticker: symbol}, nil
}

func (c fakeMarketClient) Search(query string) (*finnhub.SearchResponse, error) {
	return &finnhub.SearchResponse{}, nil
}

func (c fakeMarketClient) GetQuote(symbol string) (*finnhub.QuoteResponse, error) {
	return &finnhub.QuoteResponse{CurrentPrice: c.prices[symbol], Timestamp: 1}, nil
}

func (c fakeMarketClient) GetCandles(symbol, resolution string, from, to int64) (*finnhub.CandleResponse, error) {
	return &finnhub.CandleResponse{Timestamps: []int64{from, to}, Close: []float64{c.prices[symbol], c.prices[symbol]}}, nil
}

type world struct {
	db               *gorm.DB
	portfolioService service.PortfolioService
	watchlistService service.WatchlistService
	userCtx          model.UserContext
	lastErr          appErrors.DomainError
	initialBalance   decimal.Decimal
	initialHolding   decimal.Decimal
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			w := newWorld(t)
			ctx.Step(`^a user with balance \$(\d+) and live price of AAPL is \$(\d+)$`, w.aUserWithBalanceAndLivePrice)
			ctx.Step(`^the user buys (\d+) AAPL$`, w.theUserBuysAAPL)
			ctx.Step(`^their balance is \$(\d+), they hold (\d+) AAPL, and a BOUGHT transaction is recorded$`, w.balanceHoldingAndBoughtTransaction)
			ctx.Step(`^the request fails with 400 and their balance is unchanged$`, w.requestFailsWith400AndBalanceUnchanged)
			ctx.Step(`^a user already holds (\d+) AAPL$`, w.aUserAlreadyHoldsAAPL)
			ctx.Step(`^the user buys (\d+) more AAPL$`, w.theUserBuysAAPL)
			ctx.Step(`^they hold (\d+) AAPL$`, w.theyHoldAAPL)
			ctx.Step(`^a user holds (\d+) AAPL$`, w.aUserAlreadyHoldsAAPL)
			ctx.Step(`^the user sells (\d+) AAPL$`, w.theUserSellsAAPL)
			ctx.Step(`^the holdings row is deleted and a SOLD transaction is recorded$`, w.holdingsDeletedAndSoldTransaction)
			ctx.Step(`^the user tries to sell (\d+) AAPL$`, w.theUserSellsAAPL)
			ctx.Step(`^the request fails with 400 and their holdings are unchanged$`, w.requestFailsWith400AndHoldingsUnchanged)
			ctx.Step(`^a user has no watchlist entries$`, w.aUserHasNoWatchlistEntries)
			ctx.Step(`^they add AAPL to the watchlist$`, w.theyAddAAPLToWatchlist)
			ctx.Step(`^GET /api/watchlist returns AAPL$`, w.watchlistReturnsAAPL)
			ctx.Step(`^a user already watches AAPL$`, w.aUserAlreadyWatchesAAPL)
			ctx.Step(`^they add AAPL again$`, w.theyAddAAPLToWatchlist)
			ctx.Step(`^the request fails with 409$`, w.requestFailsWith409)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}

func newWorld(t *testing.T) *world {
	db := openDB(t)
	applyMigrations(t, db)
	priceMap := finnhub.NewPriceMap()
	priceMap.Set("AAPL", 200, 0, 1)
	client := fakeMarketClient{prices: map[string]float64{"AAPL": 200}}
	stockRepo := repository.NewStockRepository(db)
	holdingsRepo := repository.NewHoldingsRepository()
	txRepo := repository.NewTransactionsRepository()
	watchRepo := repository.NewWatchlistRepository()
	wsClient := finnhub.NewWSClient("", "", priceMap, zap.NewNop())
	return &world{
		db:               db,
		portfolioService: service.NewPortfolioService(db, stockRepo, holdingsRepo, txRepo, client, priceMap, wsClient),
		watchlistService: service.NewWatchlistService(db, stockRepo, watchRepo, client, priceMap, wsClient),
		userCtx: model.UserContext{
			FirebaseId: testUserID,
			Email:      "bdd@example.com",
			Name:       "BDD User",
		},
	}
}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("BDD_DATABASE_URL")
	if dsn == "" {
		ensureBDDDatabase(t)
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=tradelayers_bdd sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func ensureBDDDatabase(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres db: %v", err)
	}
	var count int64 = 0
	if err := db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", "tradelayers_bdd").Scan(&count).Error; err != nil {
		t.Fatalf("check bdd database: %v", err)
	}
	if count == 0 {
		if err := db.Exec("CREATE DATABASE tradelayers_bdd").Error; err != nil {
			t.Fatalf("create bdd database: %v", err)
		}
	}
}

func applyMigrations(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public").Error; err != nil {
		t.Fatalf("reset bdd schema: %v", err)
	}
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("find migrations: %v", err)
	}
	sort.Strings(files)
	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read migration: %v", err)
		}
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			t.Fatalf("apply migration %s: %v", file, err)
		}
	}
}

func (w *world) reset(IBalance int) error {
	if err := w.db.Exec("TRUNCATE alerts, watchlist, users_holdings, stock_transactions, stocks, users RESTART IDENTITY CASCADE").Error; err != nil {
		return err
	}
	w.initialBalance = decimal.NewFromInt(int64(IBalance))
	w.initialHolding = decimal.Zero
	return w.db.Create(&model.User{
		FirebaseId: testUserID,
		Name:       "BDD User",
		Email:      "bdd@example.com",
		Balance:    w.initialBalance,
	}).Error
}

func (w *world) aUserWithBalanceAndLivePrice(IBalance, IPrice int) error {
	_ = IPrice
	return w.reset(IBalance)
}

func (w *world) theUserBuysAAPL(IQty int) error {
	_, w.lastErr = w.portfolioService.Buy(w.userCtx, "AAPL", float64(IQty))
	return nil
}

func (w *world) balanceHoldingAndBoughtTransaction(IBalance, IHolding int) error {
	if err := w.expectBalance(IBalance); err != nil {
		return err
	}
	if err := w.expectHolding(IHolding); err != nil {
		return err
	}
	return w.expectTransaction(model.TransactionTypeBought)
}

func (w *world) requestFailsWith400AndBalanceUnchanged() error {
	if w.lastErr != appErrors.ErrInsufficientBalance {
		return fmt.Errorf("expected insufficient balance, got %v", w.lastErr)
	}
	return w.expectBalanceDecimal(w.initialBalance)
}

func (w *world) aUserAlreadyHoldsAAPL(IQty int) error {
	if err := w.reset(1000); err != nil {
		return err
	}
	_, w.lastErr = w.portfolioService.Buy(w.userCtx, "AAPL", float64(IQty))
	w.initialHolding = decimal.NewFromInt(int64(IQty))
	return nil
}

func (w *world) theyHoldAAPL(IQty int) error {
	return w.expectHolding(IQty)
}

func (w *world) theUserSellsAAPL(IQty int) error {
	_, w.lastErr = w.portfolioService.Sell(w.userCtx, "AAPL", float64(IQty))
	return nil
}

func (w *world) holdingsDeletedAndSoldTransaction() error {
	if err := w.expectHolding(0); err != nil {
		return err
	}
	return w.expectTransaction(model.TransactionTypeSold)
}

func (w *world) requestFailsWith400AndHoldingsUnchanged() error {
	if w.lastErr != appErrors.ErrInsufficientHoldings {
		return fmt.Errorf("expected insufficient holdings, got %v", w.lastErr)
	}
	holding, _ := w.initialHolding.Float64()
	return w.expectHolding(int(holding))
}

func (w *world) aUserHasNoWatchlistEntries() error {
	return w.reset(1000)
}

func (w *world) theyAddAAPLToWatchlist() error {
	_, w.lastErr = w.watchlistService.Add(w.userCtx, "AAPL")
	return nil
}

func (w *world) watchlistReturnsAAPL() error {
	items, err := w.watchlistService.List(w.userCtx)
	if err != appErrors.ErrNone {
		return fmt.Errorf("list watchlist: %v", err)
	}
	if len(items) != 1 || items[0].Symbol != "AAPL" {
		return fmt.Errorf("expected AAPL watchlist entry, got %+v", items)
	}
	return nil
}

func (w *world) aUserAlreadyWatchesAAPL() error {
	if err := w.reset(1000); err != nil {
		return err
	}
	_, w.lastErr = w.watchlistService.Add(w.userCtx, "AAPL")
	return nil
}

func (w *world) requestFailsWith409() error {
	if w.lastErr != appErrors.ErrAlreadyWatched {
		return fmt.Errorf("expected duplicate watchlist error, got %v", w.lastErr)
	}
	return nil
}

func (w *world) expectBalance(IBalance int) error {
	return w.expectBalanceDecimal(decimal.NewFromInt(int64(IBalance)))
}

func (w *world) expectBalanceDecimal(expected decimal.Decimal) error {
	var user model.User = model.User{}
	if err := w.db.Where("firebase_id = ?", testUserID).First(&user).Error; err != nil {
		return err
	}
	if !user.Balance.Equal(expected) {
		return fmt.Errorf("expected balance %s, got %s", expected, user.Balance)
	}
	return nil
}

func (w *world) expectHolding(IQty int) error {
	holdings, err := w.portfolioService.GetHoldings(w.userCtx)
	if err != appErrors.ErrNone {
		return fmt.Errorf("get holdings: %v", err)
	}
	if IQty == 0 {
		if len(holdings) != 0 {
			return fmt.Errorf("expected no holdings, got %+v", holdings)
		}
		return nil
	}
	if len(holdings) != 1 || holdings[0].Symbol != "AAPL" || int(holdings[0].Quantity) != IQty {
		return fmt.Errorf("expected %d AAPL, got %+v", IQty, holdings)
	}
	return nil
}

func (w *world) expectTransaction(txType model.TransactionType) error {
	txs, err := w.portfolioService.GetTransactions(w.userCtx, nil, nil, nil)
	if err != appErrors.ErrNone {
		return fmt.Errorf("get transactions: %v", err)
	}
	for _, tx := range txs {
		if tx.TransactionType == txType {
			return nil
		}
	}
	return fmt.Errorf("expected %s transaction, got %+v", txType, txs)
}
