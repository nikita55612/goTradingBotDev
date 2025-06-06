package types

import (
	"context"
	"goTradingBot/cdl"
)

type Signal int

const (
	Sell Signal = iota - 1
	Hold
	Buy
)

type Strategy interface {
	Init(ctx context.Context, subData *SubData, req chan<- *OrderRequest)
	Go() error
}

type TradingClient interface {
	PlaceOrder(symbol string, amount float64, price *float64) (string, error)
	CancelOrder(symbol, orderId string) (string, error)
	GetOrder(orderId string) ([]byte, error)
}

type DataProvider interface {
	cdl.CandleProvider
	GetInstrumentInfo(symbol string) ([]byte, error)
}
