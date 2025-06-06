package db

import (
	"database/sql"
	"fmt"
	"goTradingBot/db"
	"goTradingBot/trading/types"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbConn *sql.DB
	once   sync.Once
)

const dbPath = "orders.db"

// migrate выполняет необходимые миграции базы данных
func migrate(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS orders (
    linkId TEXT PRIMARY KEY,
    tag TEXT NOT NULL,
    id TEXT NOT NULL,
    symbol TEXT NOT NULL,
    qty REAL NOT NULL,
    price REAL,
    avgPrice REAL NOT NULL,
    execQty REAL NOT NULL,
    execValue REAL NOT NULL,
    fee REAL NOT NULL,
    isClosed INTEGER NOT NULL CHECK (isClosed IN (0, 1)),
    createdAt INTEGER NOT NULL,
    updatedAt INTEGER NOT NULL
);
`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}
	return nil
}

func InsertOrderRequest(r *types.OrderRequest) error {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	if dbConn == nil {
		return fmt.Errorf("база данных не инициализирована")
	}
	if r.Order == nil {
		return fmt.Errorf("отсутствует данные ордера")
	}
	query := `
    INSERT OR REPLACE INTO orders (
		linkId, tag, id, symbol, qty, price, avgPrice, execQty, execValue,
		fee, isClosed, createdAt, updatedAt
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	var isClosed int
	if r.Order.IsClosed {
		isClosed = 1
	}
	_, err := dbConn.Exec(query,
		r.LinkId,
		r.Tag,
		r.Order.ID,
		r.Order.Symbol,
		r.Order.Qty,
		r.Order.Price,
		r.Order.AvgPrice,
		r.Order.ExecQty,
		r.Order.ExecValue,
		r.Order.Fee,
		isClosed,
		r.Order.CreatedAt,
		r.Order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("ошибка сохранения ордера: %w", err)
	}
	return nil
}

// UpdateOrderID обновляет только поле ID ордера в базе данных
func UpdateOrderID(r *types.OrderRequest) error {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	if dbConn == nil {
		return fmt.Errorf("база данных не инициализирована")
	}
	query := `
    UPDATE orders
    SET id = ?,
	updatedAt = ?
    WHERE linkId = ?
    `
	_, err := dbConn.Exec(query,
		r.Order.ID,
		time.Now().UnixMilli(),
		r.LinkId,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления ID ордера: %w", err)
	}
	return nil
}

// GetOrderRequestsByPeriod возвращает список OrderRequest за указанный период времени (в секундах)
// periodSec определяет период времени от текущего момента (например, 24*60*60 для суток)
func GetOrderRequestsByPeriod(periodSec int64) ([]*types.OrderRequest, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	if dbConn == nil {
		return nil, fmt.Errorf("база данных не инициализирована")
	}
	timeBoundary := time.Now().UnixMilli()/1000 - periodSec
	query := `
	SELECT
		linkId, tag, id, symbol, qty, price, avgPrice, execQty, execValue,
		fee, isClosed, createdAt, updatedAt
	FROM orders
	WHERE updatedAt >= ?
	ORDER BY updatedAt DESC
	`
	rows, err := dbConn.Query(query, timeBoundary)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()
	var orders []*types.OrderRequest
	for rows.Next() {
		var (
			linkId string
			tag    string
			order  types.Order
			closed int
		)

		if err := rows.Scan(
			&linkId, &tag, &order.ID, &order.Symbol, &order.Qty, &order.Price, &order.AvgPrice,
			&order.ExecQty, &order.ExecValue, &order.Fee, &closed, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		order.IsClosed = closed == 1
		orderRequest := &types.OrderRequest{
			LinkId: linkId,
			Tag:    tag,
			Order:  &order,
			Reply:  nil,
		}

		orders = append(orders, orderRequest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}
	return orders, nil
}

// UpdateOrder обновляет поля tag, id, symbol, qty, price, avgPrice, execQty, execValue, fee, isClosed, createdAt, updatedAt в таблице orders
func UpdateOrder(r *types.OrderRequest) error {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	if dbConn == nil {
		return fmt.Errorf("база данных не инициализирована")
	}
	if r.Order == nil {
		return fmt.Errorf("отсутствуют данные ордера")
	}
	query := `
    UPDATE orders
    SET
        tag = ?,
        id = ?,
        symbol = ?,
        qty = ?,
        price = ?,
        avgPrice = ?,
        execQty = ?,
        execValue = ?,
        fee = ?,
        isClosed = ?,
        createdAt = ?,
        updatedAt = ?
    WHERE linkId = ?
    `
	var isClosed int
	if r.Order.IsClosed {
		isClosed = 1
	}
	_, err := dbConn.Exec(query,
		r.Tag,
		r.Order.ID,
		r.Order.Symbol,
		r.Order.Qty,
		r.Order.Price,
		r.Order.AvgPrice,
		r.Order.ExecQty,
		r.Order.ExecValue,
		r.Order.Fee,
		isClosed,
		r.Order.CreatedAt,
		r.Order.UpdatedAt,
		r.LinkId,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления ордера: %w", err)
	}
	return nil
}
