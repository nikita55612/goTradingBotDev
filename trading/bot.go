package trading

import (
	"context"
	"encoding/json"
	"goTradingBot/trading/config"
	orderdb "goTradingBot/trading/db"
	"goTradingBot/trading/types"
	"goTradingBot/utils/slogx"
	"os"
	"time"

	"log/slog"
)

// TradingBot реализует логику торгового бота
type TradingBot struct {
	ctx                context.Context
	tradingClient      types.TradingClient
	dataProvider       types.DataProvider
	subData            *types.SubData
	logger             *slogx.AsyncSlog
	ch                 chan *types.OrderRequest
	strategysCtx       context.Context
	cancelStrategys    context.CancelFunc
	placeOrderInterval time.Duration
}

// NewTradingBot создает новый экземпляр TradingBot
func NewTradingBot(
	ctx context.Context,
	tradingClient types.TradingClient,
	dataProvider types.DataProvider,
	logger *slog.Logger,
	cfg *config.TradingBotConfig,
) *TradingBot {

	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	if cfg == nil {
		cfg = config.DefaultTradingBotConfig()
	}
	strategysCtx, cancelStrategys := context.WithCancel(context.Background())
	b := &TradingBot{
		ctx:                ctx,
		tradingClient:      tradingClient,
		dataProvider:       dataProvider,
		subData:            types.NewSubData(ctx, dataProvider, cfg.SubDataBufferSize),
		logger:             slogx.NewAsyncSlog(context.Background(), logger),
		ch:                 make(chan *types.OrderRequest, cfg.ChannelBufferSize),
		strategysCtx:       strategysCtx,
		cancelStrategys:    cancelStrategys,
		placeOrderInterval: 200 * time.Millisecond,
	}

	go b.runPolling()

	return b
}

// runPolling запускает основной цикл обработки ордеров
func (b *TradingBot) runPolling() {
	b.logger.Log(slog.LevelInfo, "trading bot start polling")

	go func() {
		<-b.ctx.Done()
		b.logger.Log(slog.LevelInfo, "trading bot stopped")
		b.cancelStrategys()
		b.subData.Clear()

		time.Sleep(3 * time.Second)
		close(b.ch)
	}()

	for order := range b.ch {
		go b.handleOrder(order)
	}
}

// replyOrder отправляет ответ с обновлением ордера
func (b *TradingBot) replyOrder(req *types.OrderRequest) {
	if req.Reply == nil {
		return
	}
	select {
	case req.Reply <- &types.OrderUpdate{
		LinkId: req.LinkId,
		Order:  req.Order,
	}:
	default:
		b.logger.Log(
			slog.LevelError,
			"failed to send order update",
			"orderRequest", req,
		)
	}
}

// handleOrder обрабатывает новый ордер
func (b *TradingBot) handleOrder(req *types.OrderRequest) {
	if req.Order == nil {
		return
	}
	reqClone := req.Clone()
	b.logger.Log(slog.LevelInfo, "new order request", "orderRequest", reqClone)

	isReg := req.Order.GetID() != ""
	if !isReg {
		isReg = b.placeOrderWithRetry(req)
	}
	orderdb.InsertOrderRequest(reqClone) // отложенное сохранение старой копии
	if isReg {
		b.replyOrder(req)
		reqClone := req.Clone()
		b.logger.Log(slog.LevelInfo, "order is registered", "orderRequest", reqClone)
		orderdb.UpdateOrderID(reqClone)
		if b.waitForOrderClosed(req) {
			b.replyOrder(req)
			reqClone := req.Clone()
			b.logger.Log(slog.LevelInfo, "order is closed", "orderRequest", reqClone)
			orderdb.UpdateOrder(reqClone)
			return
		}
		b.replyOrder(req)
		b.cancelOrderWithRetry(req)
	}
}

// placeOrderWithRetry пытается разместить ордер с повторными попытками
func (b *TradingBot) placeOrderWithRetry(req *types.OrderRequest) bool {
	if req.Delay > 0 {
		time.Sleep(req.Delay)
	}

	timeout := time.After(time.Second)
	for {
		req.Order.Lock()
		orderId, err := b.tradingClient.PlaceOrder(
			req.Order.Symbol,
			req.Order.Qty,
			req.Order.Price,
		)
		req.Order.Unlock()
		if err == nil {
			req.Order.SetID(orderId)
			return true
		}
		select {
		case <-time.After(b.placeOrderInterval):
		case <-timeout:
			b.logger.Log(
				slog.LevelError,
				"order registration deadline has expired",
				"orderRequest", req.Clone(),
			)
			return false
		}
	}
}

// waitForOrderClosed ожидает закрытия ордера
func (b *TradingBot) waitForOrderClosed(req *types.OrderRequest) bool {
	time.Sleep(200 * time.Millisecond)

	timeoutDuration := max(time.Second, req.CloseTimeout)
	ticker := time.NewTicker(timeoutDuration / 10)
	defer ticker.Stop()

	timeout := time.After(timeoutDuration)
	for {
		if b.checkOrderClosed(req) {
			return true
		}
		select {
		case <-b.ctx.Done():
			return false
		case <-ticker.C:
		case <-timeout:
			b.logger.Log(
				slog.LevelError,
				"waiting time for order closing has expired",
				"orderRequest", req,
			)
			return b.checkOrderClosed(req)
		}
	}
}

// checkOrderClosed проверяет статус ордера
func (b *TradingBot) checkOrderClosed(req *types.OrderRequest) bool {
	data, err := b.tradingClient.GetOrder(req.Order.GetID())
	if err != nil {
		return false
	}
	var updOrder types.Order
	if err := json.Unmarshal(data, &updOrder); err != nil {
		return false
	}
	if !updOrder.IsClosed {
		return false
	}
	req.Order.Replace(&updOrder)
	return true
}

// cancelOrderWithRetry отменяет ордер
func (b *TradingBot) cancelOrderWithRetry(req *types.OrderRequest) {
	req.Order.Lock()
	symbol := req.Order.Symbol
	orderId := req.Order.ID
	req.Order.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Hour)
	for {
		_, err := b.tradingClient.CancelOrder(symbol, orderId)
		if err == nil {
			b.logger.Log(
				slog.LevelInfo,
				"unclosed order cancelled",
				"orderRequest", req,
			)
			return
		}
		select {
		case <-ticker.C:
		case <-timeout:
			return
		}
	}
}

// AddStrategy добавляет новую стратегию к торговому боту
func (b *TradingBot) AddStrategys(strategys ...types.Strategy) {
	for _, s := range strategys {
		s.Init(b.strategysCtx, b.subData, b.ch)
		if err := s.Go(); err != nil {
			b.logger.Log(slog.LevelError, "launching strategy", "error", err)
		}
	}
}
