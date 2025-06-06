package cryptos

import (
	"context"
	"encoding/json"
	"fmt"
	"goTradingBot/external/cryptos/models"
	"goTradingBot/httpx"
	"time"
)

const MAINNET = "https://api.coinmarketcap.com/data-api/v3"

// ServerResponse представляет структуру стандартного ответа от API CoinMarketCap
type ServerResponse struct {
	// Data - полезные данные ответа сервера
	Data any `json:"data"`
	// Status - информация о статусе запроса (ошибка, код ошибки...)
	Status models.Status `json:"status"`
}

// Client представляет клиент для работы с REST API CoinMarketCap
type Client struct {
	// baseURL - базовый URL API
	baseURL string
	// ctx - контекст для выполнения запросов
	ctx context.Context
	// timeout - таймаут HTTP-запросов
	timeout time.Duration
}

// NewClient создает новый экземпляр клиента для работы с API CoinMarketCap
// Принимает опциональные параметры конфигурации через Option функции
func NewClient(opts ...Option) *Client {
	client := &Client{
		baseURL: MAINNET,
		timeout: 5 * time.Second,
	}
	for _, option := range opts {
		option(client)
	}
	return client
}

// Option определяет тип функции для настройки Client
// Используется для применения параметров конфигурации
type Option func(*Client)

// WithContext устанавливает контекст для выполнения запросов
func WithContext(ctx context.Context) Option {
	return func(c *Client) {
		c.ctx = ctx
	}
}

// WithTimeout устанавливает таймаут для HTTP-запросов
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func (c *Client) callAPI(req httpx.RequestBuilder, result any) error {
	if c.ctx != nil {
		req = req.WithContext(c.ctx)
	}
	if c.timeout > 0 {
		req = req.WithTimeout(c.timeout)
	}
	res := req.Do()
	defer res.Close()
	if err := res.Error(); err != nil {
		return fmt.Errorf("CoinMarketCapAPI: не удалось выполнить запрос: %w", err)
	}
	if res.IsServerError() {
		return fmt.Errorf("CoinMarketCapAPI: статус код ошибки сервера: %d", res.StatusCode())
	}
	var serverResponse ServerResponse
	if err := res.UnmarshalBody(&serverResponse); err != nil {
		return fmt.Errorf("CoinMarketCapAPI: не удалось разобрать ответ сервера: %w", err)
	}
	status := serverResponse.Status
	if serverResponse.Status.ErrorCode != "0" {
		return fmt.Errorf("CoinMarketCapAPI: ошибка ответа сервера: %s (код %s)", status.ErrorMessage, status.ErrorCode)
	}
	data, err := json.Marshal(serverResponse.Data)
	if err != nil {
		return fmt.Errorf("CoinMarketCapAPI: не удалось преобразовать данные результата: %w", err)
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("CoinMarketCapAPI: не удалось десериализовать результат: %w", err)
	}
	return nil
}
