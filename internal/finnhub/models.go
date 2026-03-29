package finnhub

type WSMessage struct {
	Data []Trade `json:"data"`
	Type string  `json:"type"`
}

type Trade struct {
	Symbol    string  `json:"s"`
	Price     float64 `json:"p"`
	Volume    float64 `json:"v"`
	Timestamp int64   `json:"t"`
}

type SubscribeMsg struct {
	Type   string `json:"type"`
	Symbol string `json:"symbol"`
}

type ProfileResponse struct {
	Name      string  `json:"name"`
	Ticker    string  `json:"ticker"`
	Logo      string  `json:"logo"`
	Industry  string  `json:"finnhubIndustry"`
	Country   string  `json:"country"`
	Exchange  string  `json:"exchange"`
	MarketCap float64 `json:"marketCapitalization"`
	WebURL    string  `json:"weburl"`
}

type SearchResponse struct {
	Result []SearchResult `json:"result"`
}

type SearchResult struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
