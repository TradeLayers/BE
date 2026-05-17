package finnhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://finnhub.io/api/v1"

// ErrNoData is returned by GetCandles when Finnhub responds with status
// "no_data". Callers may choose to skip the symbol instead of failing the
// whole batch.
var ErrNoData error = errors.New("finnhub: no candle data")

type Client interface {
	GetProfile(symbol string) (*ProfileResponse, error)
	Search(query string) (*SearchResponse, error)
	GetQuote(symbol string) (*QuoteResponse, error)
	GetCandles(symbol, resolution string, from, to int64) (*CandleResponse, error)
}

type finnhubClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) Client {
	return &finnhubClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *finnhubClient) GetProfile(symbol string) (*ProfileResponse, error) {
	endpoint := fmt.Sprintf("%s/stock/profile2?symbol=%s", baseURL, url.QueryEscape(symbol))

	var resp ProfileResponse = ProfileResponse{}
	if err := c.doGet(endpoint, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *finnhubClient) Search(query string) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s", baseURL, url.QueryEscape(query))

	var resp SearchResponse = SearchResponse{}
	if err := c.doGet(endpoint, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *finnhubClient) GetQuote(symbol string) (*QuoteResponse, error) {
	endpoint := fmt.Sprintf("%s/quote?symbol=%s", baseURL, url.QueryEscape(symbol))

	var resp QuoteResponse = QuoteResponse{}
	if err := c.doGet(endpoint, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetCandles fetches historical OHLC data. Finnhub's free tier no longer
// exposes /stock/candle (returns 403), so we source candles from Yahoo's
// public chart API instead. The resolution param uses Finnhub notation and
// is mapped to Yahoo's interval strings.
func (c *finnhubClient) GetCandles(symbol, resolution string, from, to int64) (*CandleResponse, error) {
	interval := yahooInterval(resolution)
	endpoint := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s",
		url.PathEscape(symbol),
		from,
		to,
		interval,
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo returned status %d", resp.StatusCode)
	}

	var body yahooChartResponse = yahooChartResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(body.Chart.Result) == 0 {
		return nil, ErrNoData
	}

	result := body.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 || len(result.Timestamp) == 0 {
		return nil, ErrNoData
	}

	quote := result.Indicators.Quote[0]
	return &CandleResponse{
		Status:     "ok",
		Timestamps: result.Timestamp,
		Close:      quote.Close,
		High:       quote.High,
		Low:        quote.Low,
		Open:       quote.Open,
		Volume:     quote.Volume,
	}, nil
}

func yahooInterval(resolution string) string {
	switch resolution {
	case "1":
		return "1m"
	case "5":
		return "5m"
	case "15":
		return "15m"
	case "30":
		return "30m"
	case "60":
		return "60m"
	case "D", "":
		return "1d"
	case "W":
		return "1wk"
	case "M":
		return "1mo"
	default:
		return "1d"
	}
}

func (c *finnhubClient) doGet(endpoint string, result interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Finnhub-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("finnhub request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("finnhub rate limited")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("finnhub returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}
