package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"goTradingBot/cdl"
	"goTradingBot/external/bybit/models"
	"goTradingBot/httpx"
	"goTradingBot/httpx/ws"
	"log/slog"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"

	"github.com/google/uuid"
)

// GetInstrumentInfo возвращает информацию о торговом инструменте по его символу
// symbol - торговый символ (например, "BTCUSDT")
func (c *Client) GetInstrumentInfo(symbol string) (*models.InstrumentInfo, *Error) {
	params := map[string]any{
		"category": c.category,
		"symbol":   symbol,
	}
	res, err := c.getInstrumentsInfo(params)
	if err != nil {
		return nil, err
	}
	var instrumentInfo models.InstrumentInfo
	if len(res.List) > 0 {
		instrumentInfo = res.List[0]
	}
	return &instrumentInfo, nil
}

// GetTickers возвращает снимок последних цен
// symbol - торговый символ (например, "BTCUSDT")
func (c *Client) GetTickers(symbol string) (*models.Ticker, *Error) {
	params := map[string]any{
		"category": c.category,
		"symbol":   symbol,
	}
	res, err := c.getTickers(params)
	if err != nil {
		return nil, err
	}
	var ticker models.Ticker
	if len(res.List) > 0 {
		ticker = res.List[0]
	}
	return &ticker, nil
}

// GetCandles возвращает исторические свечи с ограничением по количеству
// symbol - торговый символ (например, "BTCUSDT")
// interval - таймфрейм свечей
// limit - максимальное количество возвращаемых свечей
func (c *Client) GetCandles(symbol string, interval cdl.Interval, limit int) ([]cdl.Candle, error) {
	startLimit := min(limit, 1000)
	params := map[string]any{
		"category": c.category,
		"symbol":   symbol,
		"interval": AsLocalInterval(interval),
		"limit":    strconv.Itoa(startLimit),
	}
	res, err := c.getCandles(params)
	if err != nil {
		return nil, err
	}
	candles, extractErr := extractCandleFromRawData(res)
	if extractErr != nil {
		return candles, extractErr
	}
	counter := limit - 1000

	// Дозагрузка оставшихся свечей при необходимости
	for counter > 0 {
		params["limit"] = strconv.Itoa(min(1000, counter+1))
		params["end"] = strconv.FormatInt(candles[len(candles)-1].Time, 10)
		res, err := c.getCandles(params)
		if err != nil {
			return candles, err
		}
		newCandles, extractErr := extractCandleFromRawData(res)
		if extractErr != nil {
			return candles, extractErr
		}
		if len(newCandles) <= 1 {
			break
		}
		candles = append(candles, newCandles[1:]...)
		counter -= 999
		if len(newCandles) < 1000 {
			break
		}
	}
	slices.Reverse(candles)
	return candles, nil
}

// GetAllCandles возвращает все доступные исторические свечи и сохраняет их в CSV
// symbol - торговый символ (например, "BTCUSDT")
// interval - таймфрейм свечей
// Использует кеширование в CSV файлах для ускорения последующих запросов
func (c *Client) GetAllCandles(symbol string, interval cdl.Interval) ([]cdl.Candle, error) {
	filePath := path.Join(CSV_PATH, fmt.Sprintf("%s-%s-%s.csv", symbol, interval.AsDisplayName(), c.category))
	if err := os.MkdirAll(CSV_PATH, 0755); err != nil {
		slog.Warn(errorTitel, "GetAllCandle", "directory creation error", "error", err)
	}
	candles, err := cdl.CandlesFromCsv(filePath)
	lenBeforeUpdate := len(candles)

	defer func() {
		if len(candles) > 0 {
			slices.Reverse(candles)
			if len(candles) != lenBeforeUpdate {
				if err := cdl.SaveCandlesToCsv(filePath, candles[:len(candles)-1]); err != nil {
					slog.Warn(errorTitel, "GetAllCandle", "candle saving error", "error", err)
				}
			}
		}
	}()
	params := map[string]any{
		"category": c.category,
		"symbol":   symbol,
		"interval": AsLocalInterval(interval),
		"limit":    "1000",
	}
	if err != nil {
		res, err := c.getCandles(params)
		if err != nil {
			return nil, err
		}
		candles, _ = extractCandleFromRawData(res)
		if len(candles) < 1000 {
			return candles, nil
		}
	} else {
		slices.Reverse(candles)
	}

	// Загрузка исторических данных
	for {
		params["end"] = strconv.FormatInt(candles[len(candles)-1].Time, 10)
		res, err := c.getCandles(params)
		if err != nil {
			return candles, err
		}
		newCandles, _ := extractCandleFromRawData(res)
		if len(newCandles) <= 1 {
			break
		}
		candles = append(candles, newCandles[1:]...)
		if len(newCandles) < 1000 {
			break
		}
	}

	// Загрузка актуальных данных
	for len(candles) > 0 {
		delete(params, "end")
		params["start"] = strconv.FormatInt(candles[0].Time, 10)
		res, err := c.getCandles(params)
		if err != nil {
			return candles, err
		}
		newCandles, _ := extractCandleFromRawData(res)
		if len(newCandles) <= 1 {
			break
		}
		candles = append(newCandles[:len(newCandles)-1], candles...)
		if len(newCandles) < 1000 {
			break
		}
	}
	return candles, nil
}

// CandleStream устанавливает WebSocket соединение для потокового получения свечей
// symbol - торговый символ (например, "BTCUSDT")
// interval - таймфрейм свечей
// Возвращает канал для получения данных и функцию для остановки соединения
func (c *Client) CandleStream(ctx context.Context, symbol string, interval cdl.Interval) (<-chan *cdl.CandleStreamData, error) {
	arg := fmt.Sprintf("kline.%s.%s", AsLocalInterval(interval), symbol)
	subMessage := map[string]any{
		"req_id": uuid.NewString(),
		"op":     "subscribe",
		"args":   []string{arg},
	}
	handshakeMessage, _ := json.Marshal(subMessage)
	outChan, err := ws.NewClient(
		ctx,
		ws.WithHandshake(handshakeMessage),
	).Connect(fmt.Sprintf("%s/%s", PUBLICWS, c.category))
	if err != nil {
		err = fmt.Errorf("couldn't create websocket connection: %w", err)
		return nil, NewInternalError(err).SetEndpoint("CandleStream")
	}
	stream := make(chan *cdl.CandleStreamData, 100)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(stream)
				return
			case data := <-outChan:
				var candleStreamRawData models.CandleStreamRawData
				if err := json.Unmarshal(data, &candleStreamRawData); err != nil {
					continue
				}
				candleStreamData, err := candleStreamFromRawData(&candleStreamRawData)
				if err != nil {
					continue
				}
				select {
				case stream <- candleStreamData:
				default:
				}
			}
		}
	}()
	return stream, nil
}

// getInstrumentsInfo выполняет запрос информации о торговых инструментах
// params - параметры запроса
func (c *Client) getInstrumentsInfo(params map[string]any) (*models.InstrumentsInfo, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf(
		"%s%s?%s",
		c.baseURL,
		"/v5/market/instruments-info",
		queryString,
	)
	req := httpx.Get(fullURL)
	var instrumentsInfo models.InstrumentsInfo
	if err := c.callAPI(req, queryString, &instrumentsInfo); err != nil {
		return &instrumentsInfo, err.SetEndpoint("getInstrumentsInfo")
	}
	return &instrumentsInfo, nil
}

// getTickers выполняет запрос к API для получения тикеров
// params - параметры запроса
func (c *Client) getTickers(params map[string]any) (*models.Tickers, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf(
		"%s%s?%s",
		c.baseURL,
		"/v5/market/tickers",
		queryString,
	)
	req := httpx.Get(fullURL)
	var tickers models.Tickers
	if err := c.callAPI(req, queryString, &tickers); err != nil {
		return &tickers, err.SetEndpoint("getTickers")
	}
	return &tickers, nil
}

// getCandles выполняет запрос исторических данных свечей
// params - параметры запроса (symbol, interval, limit и др.)
func (c *Client) getCandles(params map[string]any) (*models.CandleRawData, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf(
		"%s%s?%s",
		c.baseURL,
		"/v5/market/kline",
		queryString,
	)
	req := httpx.Get(fullURL)
	var candleRawData models.CandleRawData
	if err := c.callAPI(req, queryString, &candleRawData); err != nil {
		return &candleRawData, err.SetEndpoint("getCandle")
	}
	return &candleRawData, nil
}
