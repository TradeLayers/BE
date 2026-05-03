package model

import (
	"time"

	"github.com/google/uuid"
)

type HoldingView struct {
	StockID      uuid.UUID `json:"stockId"`
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	Quantity     float64   `json:"quantity"`
	CurrentPrice float64   `json:"currentPrice"`
}

type TransactionView struct {
	ID              uuid.UUID       `json:"id"`
	Symbol          string          `json:"symbol"`
	Name            string          `json:"name"`
	Price           float64         `json:"price"`
	Quantity        float64         `json:"quantity"`
	TransactionDate time.Time       `json:"transactionDate"`
	TransactionType TransactionType `json:"transactionType"`
}

type TradeRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`
}

type TradeResult struct {
	Transaction TransactionView `json:"transaction"`
	Balance     float64         `json:"balance"`
}

type PortfolioHistoryPoint struct {
	Date            time.Time `json:"date"`
	InvestedCapital float64   `json:"investedCapital"`
}

type PortfolioHistoryResponse struct {
	Points       []PortfolioHistoryPoint `json:"points"`
	CurrentValue float64                 `json:"currentValue"`
}

type WatchlistItem struct {
	Symbol         string   `json:"symbol"`
	Name           string   `json:"name"`
	CurrentPrice   float64  `json:"currentPrice"`
	ThresholdPrice *float64 `json:"thresholdPrice"`
}

type WatchlistRequest struct {
	Symbol string `json:"symbol" binding:"required"`
}

type WatchlistThresholdRequest struct {
	ThresholdPrice float64 `json:"thresholdPrice" binding:"required"`
}

type ThresholdNotificationView struct {
	ID             string    `json:"id"`
	Symbol         string    `json:"symbol"`
	ThresholdPrice float64   `json:"thresholdPrice"`
	TriggerPrice   float64   `json:"triggerPrice"`
	TriggeredAt    time.Time `json:"triggeredAt"`
	Message        string    `json:"message"`
}

type CandleSeries struct {
	Timestamps []int64   `json:"t"`
	Close      []float64 `json:"c"`
	High       []float64 `json:"h"`
	Low        []float64 `json:"l"`
	Open       []float64 `json:"o"`
	Volume     []float64 `json:"v"`
}

type CandlesResponse struct {
	Series map[string]CandleSeries `json:"series"`
}
