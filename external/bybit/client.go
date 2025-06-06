package bybit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goTradingBot/httpx"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	CSV_PATH = "csv/bybit"
	PUBLICWS = "wss://stream.bybit.com/v5/public"
	// Базовые конечные точки REST API Bybit
	MAINNET     = "https://api.bybit.com"         // Основная конечная точка
	MAINNET_ALT = "https://api.bytick.com"        // Альтернативная основная конечная точка
	TESTNET     = "https://api-testnet.bybit.com" // Конечная точка для тестовой сети
	// Региональные конечные точки для соответствия локальным требованиям
	NETHERLAND_ENV = "https://api.bybit.nl"
	HONGKONG_ENV   = "https://api.byhkbit.com"
	TURKEY_ENV     = "https://api.bybit-tr.com"
	KAZAKHSTAN_ENV = "https://api.bybit.kz"
)

// ServerResponse представляет структуру стандартного ответа от API Bybit
type ServerResponse struct {
	// RetCode - код возврата, где 0 означает успешный запрос
	RetCode int `json:"retCode"`
	// RetMsg - сообщение от сервера ("OK", "SUCCESS" или описание ошибки)
	RetMsg string `json:"retMsg"`
	// Result - основные данные ответа, тип зависит от конкретного запроса
	Result any `json:"result"`
	// RetExtInfo - дополнительная информация (обычно пустой объект)
	RetExtInfo struct{} `json:"retExtInfo"`
	// Time - временная метка сервера в миллисекундах
	Time int64 `json:"time"`
}

// Client представляет клиент для работы с REST API Bybit
type Client struct {
	baseURL    string          // базовый URL API (тестовая или основная сеть)
	apiKey     string          // публичный API-ключ для аутентификации
	apiSecret  string          // секретный ключ для подписи запросов (HMAC)
	recvWindow int             // временное окно валидности запроса в миллисекундах (по умолчанию 5000)
	category   string          // spot/linear/inverse
	ctx        context.Context // контекст для выполнения запросов
	timeout    time.Duration   // таймаут HTTP-запросов
}

// NewClient создает новый экземпляр клиента для работы с API Bybit
// Загружает учетные данные из .env файла (API_KEY и API_SECRET)
// Принимает опциональные параметры конфигурации через Option функции
func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	client := &Client{
		baseURL:    MAINNET,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: 5000,
		category:   "spot",
		timeout:    5 * time.Second,
	}
	for _, option := range opts {
		option(client)
	}
	return client
}

func NewClientFromEnv(opts ...Option) *Client {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("%s: NewClientFromEnv: ошибка загрузки .env файла", errorTitel)
	}
	apiKey := os.Getenv("BYBIT_API_KEY")
	if apiKey == "" {
		log.Fatalf("%s: NewClientFromEnv: не указан BYBIT_API_KEY", errorTitel)
	}
	apiSecret := os.Getenv("BYBIT_API_SECRET")
	if apiSecret == "" {
		log.Fatalf("%s: NewClientFromEnv: не указан BYBIT_API_SECRET", errorTitel)
	}
	return NewClient(apiKey, apiSecret, opts...)
}

// Option определяет тип функции для настройки Client
type Option func(*Client)

// WithRecvWindow устанавливает пользовательское значение recvWindow
// recvWindow - временное окно валидности запроса в миллисекундах
// Меньшие значения безопаснее, но могут приводить к ошибкам при медленных соединениях
func WithRecvWindow(recvWindow int) Option {
	return func(c *Client) {
		c.recvWindow = recvWindow
	}
}

// WithBaseURL устанавливает пользовательский базовый URL API
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithCategory устанавливает категорию (spot, linear, inverse)
func WithCategory(category string) Option {
	return func(c *Client) {
		c.category = category
	}
}

// WithTimeout устанавливает таймаут для HTTP-запросов
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithContext устанавливает контекст для выполнения запросов
func WithContext(ctx context.Context) Option {
	return func(c *Client) {
		c.ctx = ctx
	}
}

func (c *Client) callAPI(req httpx.RequestBuilder, queryString string, result any) *Error {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	sigString := fmt.Sprintf("%s%s%d%s", timestamp, c.apiKey, c.recvWindow, queryString)
	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	if _, err := mac.Write([]byte(sigString)); err != nil {
		err := fmt.Errorf("error when creating the request signature: %w", err)
		return NewUnknownError(err)
	}
	signature := hex.EncodeToString(mac.Sum(nil))
	header := make(http.Header)
	header.Add("X-BAPI-API-KEY", c.apiKey)
	header.Add("X-BAPI-TIMESTAMP", timestamp)
	header.Add("X-BAPI-SIGN", signature)
	header.Add("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	header.Add("Content-Type", "application/json")
	header.Add("Accept", "application/json")
	if c.ctx != nil {
		req = req.WithContext(c.ctx)
	}
	if c.timeout > 0 {
		req = req.WithTimeout(c.timeout)
	}
	res := req.SetHeader(header).Do()
	defer res.Close()
	if err := ErrorFromResponse(res); err != nil {
		return err
	}
	var serverResponse ServerResponse
	if err := res.UnmarshalBody(&serverResponse); err != nil {
		return NewError(SerDeErrorT, err)
	}
	if err := ErrorFromServerResponse(&serverResponse); err.ServerResponseCode() != 0 {
		return err
	}
	data, err := json.Marshal(serverResponse.Result)
	if err != nil {
		return NewError(SerDeErrorT, err)
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return NewError(SerDeErrorT, err)
		}
	}
	return nil
}
