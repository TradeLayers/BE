package model

type StockQuote struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

type StockSearchResult struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type StockProfile struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Logo      string  `json:"logo"`
	Industry  string  `json:"industry"`
	Country   string  `json:"country"`
	Exchange  string  `json:"exchange"`
	MarketCap float64 `json:"marketCap"`
	WebURL    string  `json:"webUrl"`
}

type QuotesRequest struct {
	Symbols []string `json:"symbols" binding:"required"`
}

type StockListItem struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}
