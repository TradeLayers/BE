package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WSClient struct {
	apiKey        string
	wsURL         string
	priceMap      *PriceMap
	logger        *zap.Logger
	mu            sync.Mutex
	conn          *websocket.Conn
	subscriptions map[string]bool
}

func NewWSClient(apiKey string, wsURL string, priceMap *PriceMap, logger *zap.Logger) *WSClient {
	return &WSClient{
		apiKey:        apiKey,
		wsURL:         wsURL,
		priceMap:      priceMap,
		logger:        logger,
		subscriptions: make(map[string]bool),
	}
}

func (ws *WSClient) Run(ctx context.Context) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			ws.close()
			return
		default:
		}

		err := ws.connect()
		if err != nil {
			ws.logger.Error("finnhub ws connect failed", zap.Error(err))
			ws.sleep(ctx, backoff)
			backoff = ws.nextBackoff(backoff)
			continue
		}

		backoff = time.Second
		ws.logger.Info("finnhub ws connected")
		ws.resubscribe()
		ws.readLoop(ctx)
	}
}

func (ws *WSClient) Subscribe(symbols []string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, s := range symbols {
		ws.subscriptions[s] = true
		ws.sendSubscribe(s)
	}
}

func (ws *WSClient) connect() error {
	url := fmt.Sprintf("%s?token=%s", ws.wsURL, ws.apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.mu.Unlock()

	return nil
}

func (ws *WSClient) close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
}

func (ws *WSClient) resubscribe() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for symbol := range ws.subscriptions {
		ws.sendSubscribe(symbol)
	}
}

func (ws *WSClient) sendSubscribe(symbol string) {
	if ws.conn == nil {
		return
	}

	msg := SubscribeMsg{
		Type:   "subscribe",
		Symbol: symbol,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		ws.logger.Error("failed to marshal subscribe msg", zap.Error(err))
		return
	}

	if err := ws.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		ws.logger.Error("failed to send subscribe msg", zap.String("symbol", symbol), zap.Error(err))
	}
}

func (ws *WSClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ws.mu.Lock()
		conn := ws.conn
		ws.mu.Unlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			ws.logger.Error("finnhub ws read error", zap.Error(err))
			ws.close()
			return
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			ws.logger.Error("finnhub ws unmarshal error", zap.Error(err))
			continue
		}

		if wsMsg.Type != "trade" {
			continue
		}

		for _, trade := range wsMsg.Data {
			ws.priceMap.Set(trade.Symbol, trade.Price, trade.Volume, trade.Timestamp)
		}
	}
}

func (ws *WSClient) sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func (ws *WSClient) nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > 30*time.Second {
		return 30 * time.Second
	}
	return next
}
