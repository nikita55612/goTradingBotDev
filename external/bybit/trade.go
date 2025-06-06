package bybit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goTradingBot/external/bybit/models"
	"goTradingBot/httpx"
	"math"
	"net/url"
	"strconv"
)

// PlaceOrder создает рыночный или лимитный ордер
// symbol - торговый символ (например "BTCUSDT")
// qty - объем: положительный - покупка, отрицательный - продажа
// price - цена (если указан - лимитный ордер, иначе - рыночный)
func (c *Client) PlaceOrder(symbol string, qty float64, price *float64) (*models.PlaceOrderResult, *Error) {
	side := "Buy"
	if qty < 0 {
		side = "Sell"
	}
	params := map[string]any{
		"category":   c.category,
		"symbol":     symbol,
		"side":       side,
		"orderType":  "Market",
		"isLeverage": 1,
	}
	qty = math.Abs(qty)
	params["qty"] = strconv.FormatFloat(qty, 'f', -1, 64)
	if price != nil {
		params["price"] = strconv.FormatFloat(*price, 'f', -1, 64)
		params["orderType"] = "Limit"
	}
	res, err := c.placeOrder(params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ClosePosition закрывает позицию по указанной монете, конвертируя весь баланс в quote-монету
// baseCoin - базовая монета (например "BTC")
// quoteCoin - котируемая монета (например "USDT")
// precision - точность округления суммы ордера
// func (c *Client) ClosePosition(baseCoin, quoteCoin string, precision int) (*models.PlaceOrderResult, *Error) {
// 	coinInfo, err := c.GetCoinBalance(baseCoin)
// 	if err != nil {
// 		return nil, err
// 	}
// 	amount, parseErr := strconv.ParseFloat(coinInfo.UsdValue, 64)
// 	if parseErr != nil {
// 		e := fmt.Errorf("converting quoteCoin Value to float64: %w", parseErr)
// 		return nil, NewInternalError(e).SetEndpoint("ClosePosition")
// 	}

// 	amount = numeric.FloorFloat(amount, precision)
// 	if amount == 0 {
// 		e := fmt.Errorf("sum of the coin position %s is 0", baseCoin)
// 		return nil, NewInternalError(e).SetEndpoint("ClosePosition")
// 	}

// 	return c.PlaceOrder(baseCoin+quoteCoin, -amount, nil)
// }

// CancelOrder отменяет активный ордер
// symbol - торговый символ (например "BTCUSDT")
// orderId - ID ордера для отмены
func (c *Client) CancelOrder(symbol, orderId string) (*models.CancelOrderResult, *Error) {
	params := map[string]any{
		"category": c.category,
		"symbol":   symbol,
		"orderId":  orderId,
	}
	res, err := c.cancelOrder(params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetOrderHistoryDetail возвращает детали ордера по его ID
// orderId - ID ордера для поиска
func (c *Client) GetOrderHistoryDetail(orderId string) (*models.OrderHistoryDetail, *Error) {
	params := map[string]any{
		"category": c.category,
		"orderId":  orderId,
	}
	res, err := c.getOrderHistory(params)
	if err != nil {
		return nil, err
	}
	if len(res.List) == 0 {
		e := fmt.Errorf("order with id %s not found", orderId)
		return nil, NewInternalError(e).SetEndpoint("GetOrderHistoryDetail")
	}
	return &res.List[0], nil
}

// placeOrder отправляет запрос на создание ордера (внутренний метод)
func (c *Client) placeOrder(params map[string]any) (*models.PlaceOrderResult, *Error) {
	jsonData, _ := json.Marshal(params)
	body := bytes.NewBuffer(jsonData)
	fullURL := fmt.Sprintf("%s%s", c.baseURL, "/v5/order/create")
	req := httpx.Post(fullURL).WithBody(body)
	var placeOrderResult models.PlaceOrderResult
	if err := c.callAPI(req, string(jsonData), &placeOrderResult); err != nil {
		return &placeOrderResult, err.SetEndpoint("placeOrder")
	}
	return &placeOrderResult, nil
}

// cancelOrder отправляет запрос на отмену ордера (внутренний метод)
func (c *Client) cancelOrder(params map[string]any) (*models.CancelOrderResult, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/order/cancel", queryString)
	req := httpx.Post(fullURL)
	var cancelOrderResult models.CancelOrderResult
	if err := c.callAPI(req, queryString, &cancelOrderResult); err != nil {
		return &cancelOrderResult, err.SetEndpoint("cancelOrder")
	}
	return &cancelOrderResult, nil
}

// getOrderHistory получает историю ордеров (внутренний метод)
func (c *Client) getOrderHistory(params map[string]any) (*models.OrderHistoryResult, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, "/v5/order/history", queryString)
	req := httpx.Get(fullURL)
	var orderHistoryResult models.OrderHistoryResult
	if err := c.callAPI(req, queryString, &orderHistoryResult); err != nil {
		return &orderHistoryResult, err.SetEndpoint("getOrderHistory")
	}
	return &orderHistoryResult, nil
}
