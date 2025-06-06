package types

import (
	"goTradingBot/utils/seqs"
	"sync"
	"time"
)

type Order struct {
	sync.Mutex `json:"-"`
	ID         string   `json:"id"`        // ID ордера
	Symbol     string   `json:"symbol"`    // Торговая пара
	Qty        float64  `json:"qty"`       // Исходное количество
	Price      *float64 `json:"price"`     // Цена для лимитного ордера
	AvgPrice   float64  `json:"avgPrice"`  // Средняя цена исполнения
	ExecQty    float64  `json:"execQty"`   // Исполненное количество
	ExecValue  float64  `json:"execValue"` // Стоимость исполненного объема
	Fee        float64  `json:"fee"`       // Сумма комиссии
	CreatedAt  int64    `json:"createdAt"` // Время создания (мс)
	UpdatedAt  int64    `json:"updatedAt"` // Время обновления (мс)
	IsClosed   bool     `json:"isClosed"`  // Флаг завершенности
}

func NewOrder(symbol string, qty float64, price *float64) *Order {
	return &Order{
		Symbol:    symbol,
		Qty:       qty,
		Price:     price,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (o *Order) GetID() string {
	o.Lock()
	defer o.Unlock()

	return o.ID
}

func (o *Order) SetID(id string) {
	o.Lock()
	defer o.Unlock()

	o.ID = id
}

func (o *Order) WithLock(f func(order *Order)) {
	o.Lock()
	defer o.Unlock()

	f(o)
}

func (o *Order) Replace(newOrder *Order) {
	o.Lock()
	defer o.Unlock()

	o.AvgPrice = newOrder.AvgPrice
	o.Qty = newOrder.Qty
	o.Price = newOrder.Price
	o.ExecQty = newOrder.ExecQty
	o.ExecValue = newOrder.ExecValue
	o.Fee = newOrder.Fee
	o.CreatedAt = newOrder.CreatedAt
	o.UpdatedAt = newOrder.UpdatedAt
	o.ID = newOrder.ID
	o.Symbol = newOrder.Symbol
	o.IsClosed = newOrder.IsClosed
}

func (o *Order) Clone() *Order {
	o.Lock()
	defer o.Unlock()

	return &Order{
		AvgPrice:  o.AvgPrice,
		Qty:       o.Qty,
		Price:     o.Price,
		ExecQty:   o.ExecQty,
		ExecValue: o.ExecValue,
		Fee:       o.Fee,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		ID:        o.ID,
		Symbol:    o.Symbol,
		IsClosed:  o.IsClosed,
	}
}

type OrderUpdate struct {
	LinkId string `json:"linkId"`
	Order  *Order `json:"order"`
}

type OrderRequest struct {
	LinkId       string              `json:"linkId"`
	Tag          string              `json:"tag"`
	Order        *Order              `json:"order"`
	Delay        time.Duration       `json:"-"`
	CloseTimeout time.Duration       `json:"-"`
	Reply        chan<- *OrderUpdate `json:"-"`
}

func (r *OrderRequest) Clone() *OrderRequest {
	var clonedOrder *Order
	if r.Order != nil {
		clonedOrder = r.Order.Clone()
	}

	return &OrderRequest{
		LinkId: r.LinkId,
		Tag:    r.Tag,
		Order:  clonedOrder,
		Reply:  r.Reply,
	}
}

type OrderLog struct {
	*seqs.OrderedMap[string, *Order]
}

func NewOrderLog() *OrderLog {
	return &OrderLog{
		OrderedMap: seqs.NewOrderedMap[string, *Order](32),
	}
}
