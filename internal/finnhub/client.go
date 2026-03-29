package finnhub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://finnhub.io/api/v1"

type Client interface {
	GetProfile(symbol string) (*ProfileResponse, error)
	Search(query string) (*SearchResponse, error)
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

	var resp ProfileResponse
	if err := c.doGet(endpoint, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *finnhubClient) Search(query string) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s", baseURL, url.QueryEscape(query))

	var resp SearchResponse
	if err := c.doGet(endpoint, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
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
