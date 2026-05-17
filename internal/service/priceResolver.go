package service

import (
	"context"

	"github.com/TradeLayers/BE/internal/finnhub"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/google/uuid"
)

func resolveCurrentPrice(priceMap *finnhub.PriceMap, client finnhub.Client, symbol string) float64 {
	if tp, ok := priceMap.Get(symbol); ok && tp.Price > 0 {
		return tp.Price
	}

	quote, err := client.GetQuote(symbol)
	if err == nil && quote.CurrentPrice > 0 {
		priceMap.Set(symbol, quote.CurrentPrice, 0, quote.Timestamp)
		return quote.CurrentPrice
	}

	return 0
}

// ensureStock finds the stock row by symbol or creates it, using the default
// catalogue name, then a Finnhub profile lookup, then the symbol itself as a
// last resort. Trading and watchlist features both need this because the
// stocks table is only populated on demand.
func ensureStock(ctx context.Context, repo repository.StockRepository, client finnhub.Client, symbol string) (*model.Stock, error) {
	stock, err := repo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if stock != nil {
		return stock, nil
	}

	name := DefaultStocks[symbol]
	if name == "" {
		if profile, err := client.GetProfile(symbol); err == nil && profile.Name != "" {
			name = profile.Name
		} else {
			name = symbol
		}
	}

	newStock := &model.Stock{StockName: name, Symbol: symbol}
	if err := repo.Create(ctx, newStock); err != nil {
		return nil, err
	}
	return newStock, nil
}

func stocksByID(ctx context.Context, repo repository.StockRepository) (map[uuid.UUID]*model.Stock, error) {
	all, err := repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*model.Stock, len(all))
	for IIndex := range all {
		result[all[IIndex].ID] = &all[IIndex]
	}
	return result, nil
}
