package bybit

import (
	"context"
	"encoding/json"
	"goTradingBot/cdl"
	"goTradingBot/utils/numeric"
	"strconv"
)

//	type TradingClient interface {
//		PlaceOrder(symbol string, amount float64, price *float64) (string, error)
//		CancelOrder(symbol, orderId string) (string, error)
//		GetOrder(orderId string) ([]byte, error)
//		GetInstrumentInfo(symbol string) ([]byte, error)
//	}
type TradingClientImpl struct {
	cli *Client
}

//	type DataProvider interface {
//		cdl.CandleProvider
//		GetInstrumentInfo(symbol string) ([]byte, error)
//	}
type DataProvider struct {
	cli *Client
}

func (c *Client) TradingClientImpl() *TradingClientImpl {
	return &TradingClientImpl{cli: c}
}

func (c *Client) DataProviderImpl() *DataProvider {
	return &DataProvider{cli: c}
}

func (i *TradingClientImpl) PlaceOrder(symbol string, amount float64, price *float64) (string, error) {
	res, err := i.cli.PlaceOrder(symbol, amount, price)
	if err != nil {
		return "", err
	}
	return res.OrderId, nil
}

func (i *TradingClientImpl) CancelOrder(symbol, orderId string) (string, error) {
	res, err := i.cli.CancelOrder(symbol, orderId)
	if err != nil {
		return "", err
	}
	return res.OrderId, nil
}

// GetOrder получении детальной информации об ордере.
// Возвращает данные ордера в формате JSON со следующими полями:
//   - id:         string  - ID ордера в системе Bybit
//   - symbol:     string  - Торговая пара (например "BTCUSDT")
//   - isClosed:   bool    - Флаг завершенности ордера (true - завершен/отменен, false - активен)
//   - avgPrice:   float64 - Средняя цена исполнения ордера
//   - execQty:    float64 - Исполненное количество базовой валюты:
//     >0 для покупок (Buy), <0 для продаж (Sell)
//   - execValue:  float64 - Стоимость исполненного объема в котируемой валюте:
//     >0 для покупок, <0 для продаж
//   - fee:        float64 - Сумма уплаченной комиссии (всегда положительное значение)
//   - createdAt   int64   - Время последнего обновления ордера (Unix timestamp в миллисекундах)
//   - updatedAt:   int64   - Время последнего обновления ордера (Unix timestamp в миллисекундах)
func (i *TradingClientImpl) GetOrder(orderId string) ([]byte, error) {
	detail, err := i.cli.GetOrderHistoryDetail(orderId)
	if err != nil {
		return nil, err
	}
	createdAt, parseErr := strconv.ParseInt(detail.CreatedTime, 10, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	updatedAt, parseErr := strconv.ParseInt(detail.UpdatedTime, 10, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	qty, parseErr := strconv.ParseFloat(detail.Qty, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	price, parseErr := strconv.ParseFloat(detail.Price, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	avgPrice, parseErr := strconv.ParseFloat(detail.AvgPrice, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	execQty, parseErr := strconv.ParseFloat(detail.CumExecQty, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	execValue, parseErr := strconv.ParseFloat(detail.CumExecValue, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	if detail.Side == "Sell" {
		qty = -qty
		execQty = -execQty
		execValue = -execValue
	}
	fee, parseErr := strconv.ParseFloat(detail.CumExecFee, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	isClosed := true
	switch detail.OrderStatus {
	case "New", "PartiallyFilled", "Untriggered":
		isClosed = false
	}
	orderData := map[string]any{
		"id":        detail.OrderId,
		"symbol":    detail.Symbol,
		"qty":       qty,
		"price":     price,
		"avgPrice":  avgPrice,
		"execQty":   execQty,
		"execValue": execValue,
		"fee":       fee,
		"isClosed":  isClosed,
		"createdAt": createdAt,
		"updatedAt": updatedAt,
	}
	return json.Marshal(orderData)
}

// GetInstrumentInfo получении детальной информации об инструменте.
// Возвращает данные ордера в формате JSON со следующими полями:
//   - minOrderQty: float64  - Минимальное количество для ордера
//   - tickSize:    float64  - Шаг изменения цены
func (i *DataProvider) GetInstrumentInfo(symbol string) ([]byte, error) {
	info, err := i.cli.GetInstrumentInfo(symbol)
	if err != nil {
		return nil, err
	}
	var minOrderAmt float64
	var qtyPrecision int
	if i.cli.category == "spot" {
		v, parseErr := strconv.ParseFloat(info.LotSizeFilter.MaxOrderAmt, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		minOrderAmt = v
		v, parseErr = strconv.ParseFloat(info.LotSizeFilter.BasePrecision, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		qtyPrecision = numeric.DecimalPlaces(v)
	} else {
		v, parseErr := strconv.ParseFloat(info.LotSizeFilter.MinNotionalValue, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		minOrderAmt = v
		v, parseErr = strconv.ParseFloat(info.LotSizeFilter.QtyStep, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		qtyPrecision = numeric.DecimalPlaces(v)
	}
	tickSize, parseErr := strconv.ParseFloat(info.PriceFilter.TickSize, 64)
	if parseErr != nil {
		return nil, parseErr
	}
	infoData := map[string]any{
		"qtyPrecision": qtyPrecision,
		"minOrderAmt":  minOrderAmt,
		"tickSize":     tickSize,
	}
	return json.Marshal(infoData)
}

func (i *DataProvider) GetCandles(symbol string, interval cdl.Interval, limit int) ([]cdl.Candle, error) {
	return i.cli.GetCandles(symbol, interval, limit)
}

func (i *DataProvider) CandleStream(ctx context.Context, symbol string, interval cdl.Interval) (<-chan *cdl.CandleStreamData, error) {
	return i.cli.CandleStream(ctx, symbol, interval)
}
