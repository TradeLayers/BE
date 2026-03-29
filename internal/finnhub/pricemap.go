package finnhub

import "sync"

type TradePrice struct {
	Price     float64
	Volume    float64
	Timestamp int64
}

type PriceMap struct {
	mu   sync.RWMutex
	data map[string]TradePrice
}

func NewPriceMap() *PriceMap {
	return &PriceMap{
		data: make(map[string]TradePrice),
	}
}

func (pm *PriceMap) Set(symbol string, price float64, volume float64, timestamp int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.data[symbol] = TradePrice{
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
}

func (pm *PriceMap) Get(symbol string) (*TradePrice, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	tp, ok := pm.data[symbol]
	if !ok {
		return nil, false
	}

	return &tp, true
}

func (pm *PriceMap) GetMulti(symbols []string) map[string]TradePrice {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]TradePrice, len(symbols))
	for _, s := range symbols {
		if tp, ok := pm.data[s]; ok {
			result[s] = tp
		}
	}

	return result
}
