package telebot

import (
	"context"
	"fmt"
	"goTradingBot/httpx"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	MAINNET    = "https://api.telegram.org"
	errorTitel = "TelebotAPI"
)

type Bot struct {
	baseURL      string
	apiKey       string
	writeChatIDs []string
	parseMode    string
	ctx          context.Context
	timeout      time.Duration
}

func NewBot(apiKey string, opts ...Option) *Bot {
	bot := &Bot{
		apiKey:    apiKey,
		baseURL:   MAINNET,
		parseMode: "HTML",
		timeout:   5 * time.Second,
	}
	for _, option := range opts {
		option(bot)
	}
	return bot
}

func NewBotFromEnv(opts ...Option) *Bot {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("%s: NewClientFromEnv: ошибка загрузки .env файла", errorTitel)
	}
	apiKey := os.Getenv("TELEBOT_API_KEY")
	if apiKey == "" {
		log.Fatalf("%s: NewClientFromEnv: не указан TELEBOT_API_KEY", errorTitel)
	}
	return NewBot(apiKey, opts...)
}

// Option определяет тип функции для настройки Bot
type Option func(*Bot)

// WithContext устанавливает контекст для выполнения запросов
func WithContext(ctx context.Context) Option {
	return func(b *Bot) {
		b.ctx = ctx
	}
}

// WithTimeout устанавливает таймаут для HTTP-запросов
func WithTimeout(timeout time.Duration) Option {
	return func(b *Bot) {
		b.timeout = timeout
	}
}

// WithWriteChatIDs устанавливает список id приемников для отправки сообщений через метод Write
func WithWriteChatIDs(chatIDs []string) Option {
	return func(b *Bot) {
		b.writeChatIDs = chatIDs
	}
}

// WithWriteChatID добавляет id приемника для отправки сообщений через метод Write
func WithWriteChatID(chatID string) Option {
	return func(b *Bot) {
		b.writeChatIDs = append(b.writeChatIDs, chatID)
	}
}

// WithParseMode устанавливает режим парсинга сообщений
func WithParseMode(parseMode string) Option {
	return func(c *Bot) {
		c.parseMode = parseMode
	}
}

func (b *Bot) callAPI(req httpx.RequestBuilder, result any) error {
	if b.ctx != nil {
		req = req.WithContext(b.ctx)
	}
	if b.timeout > 0 {
		req = req.WithTimeout(b.timeout)
	}
	res := req.Do()
	defer res.Close()
	if err := res.Error(); err != nil {
		return fmt.Errorf("%s: не удалось выполнить запрос: %w", errorTitel, err)
	}
	if res.IsServerError() {
		return fmt.Errorf("%s: статус код ошибки сервера: %d", errorTitel, res.StatusCode())
	}
	if result != nil {
		if err := res.UnmarshalBody(&result); err != nil {
			return fmt.Errorf("%s: не удалось разобрать ответ сервера: %w", errorTitel, err)
		}
	}
	return nil
}

func (b *Bot) SendMessage(chatId, text string) (any, error) {
	fullURL := fmt.Sprintf("%s/bot%s/sendMessage", b.baseURL, b.apiKey)
	params := map[string]any{
		"chat_id":    chatId,
		"text":       text,
		"parse_mode": b.parseMode,
	}
	req := httpx.Post(fullURL).WithJsonData(params).
		AddHeader("Content-Type", "application/json")
	var result any
	if err := b.callAPI(req, nil /*&result*/); err != nil {
		return &result, err
	}
	return result, nil
}

// Write отправка сообщений для всех в списке writeChatIDs
func (b *Bot) Write(p []byte) (n int, err error) {
	if len(b.writeChatIDs) == 0 {
		return n, fmt.Errorf("%s: Write: список writeChatIDs пуст", errorTitel)
	}
	for _, c := range b.writeChatIDs {
		_, err = b.SendMessage(c, string(p))
	}
	if err != nil {
		return n, fmt.Errorf("%s: Write: ошибка отправки сообщения", errorTitel)
	}
	return len(p), err
}
