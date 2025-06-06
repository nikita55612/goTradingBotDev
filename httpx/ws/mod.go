package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client представляет WebSocket клиент с поддержкой реконнекта
type Client struct {
	// Подключение и базовые параметры
	conn   *websocket.Conn
	dialer websocket.Dialer
	header http.Header

	// Каналы управления
	outChan   chan []byte     // Канал для исходящих сообщений
	reconnect chan bool       // Сигнал для реконнекта
	ctx       context.Context // Контекст
	wg        sync.WaitGroup  // Группа ожидания горутин

	// Таймауты и интервалы
	writeWait    time.Duration // Таймаут записи
	pongWait     time.Duration // Ожидание понга
	pingInterval time.Duration // Интервал пингов

	// Дополнительные параметры
	handshake []byte // Данные для начального рукопожатия
}

// NewClient создает новый WebSocket клиент с опциональными настройками
func NewClient(ctx context.Context, opts ...Option) *Client {
	c := &Client{
		outChan:   make(chan []byte),
		ctx:       ctx,
		header:    make(http.Header),
		reconnect: make(chan bool),
		dialer: websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		writeWait:    15 * time.Second,
		pongWait:     30 * time.Second,
		pingInterval: (30 * time.Second * 9) / 10,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option определяет тип функции для настройки клиента
type Option func(*Client)

// WithHandshake устанавливает данные для начального рукопожатия
func WithHandshake(h []byte) Option {
	return func(c *Client) { c.handshake = h }
}

// WithHeader добавляет заголовки для подключения
func WithHeader(h http.Header) Option {
	return func(c *Client) { c.header = h }
}

// WithWriteTimeout устанавливает таймаут записи
func WithWriteTimeout(d time.Duration) Option {
	return func(c *Client) { c.writeWait = d }
}

// WithPongTimeout устанавливает таймаут ожидания pong
// и автоматически вычисляет pingInterval как 90% от pongWait
func WithPongTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 5*time.Second {
			c.pongWait = d
			c.pingInterval = time.Duration(float64(d) * 0.9)
		}
	}
}

// Connect устанавливает соединение и возвращает канал для чтения
func (c *Client) Connect(url string) (<-chan []byte, error) {
	conn, _, err := c.dialer.Dial(url, c.header)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения: %w", err)
	}
	c.conn = conn

	if len(c.handshake) > 0 {
		if err := c.writeMessage(websocket.TextMessage, c.handshake); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ошибка рукопожатия: %w", err)
		}
	}

	go c.runPumps(url)

	return c.outChan, nil
}

// runPumps запускает горутины чтения/записи и обработку реконнекта
func (c *Client) runPumps(url string) {
	c.wg.Add(2)
	go c.readPump()
	go c.writePump()
	c.wg.Wait()

	c.conn.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			close(c.outChan)
			close(c.reconnect)
			return
		case <-ticker.C:
			if _, err := c.Connect(url); err == nil {
				return
			}
		}
	}
}

// readPump обрабатывает входящие сообщения
func (c *Client) readPump() {
	defer c.signalReconnect()

	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		select {
		case <-c.ctx.Done():
			return
		case c.outChan <- msg:
		}
	}
}

// writePump отправляет ping-сообщения
func (c *Client) writePump() {
	defer c.signalReconnect()

	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.writeMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// signalReconnect уведомляет о необходимости переподключения
func (c *Client) signalReconnect() {
	c.wg.Done()
	select {
	case c.reconnect <- true:
	default:
	}
}

// writeMessage отправляет сообщение с учетом контекста
func (c *Client) writeMessage(msgType int, data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
	return c.conn.WriteMessage(msgType, data)
}
