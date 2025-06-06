package db

import (
	"database/sql"
	"fmt"
	"goTradingBot/db"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbConn *sql.DB
	once   sync.Once
)

const dbPath = "cryptos.db"

// migrate выполняет необходимые миграции базы данных
func migrate(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS cryptos (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	symbol TEXT NOT NULL UNIQUE,
	logo BLOB
);

CREATE INDEX IF NOT EXISTS idx_cryptos_symbol ON cryptos(symbol);
`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}
	return nil
}

type Crypto struct {
	ID     int
	Name   string
	Symbol string
	Logo   []byte
}

// InsertCrypto добавляет новую криптовалюту в базу данных
func InsertCrypto(c *Crypto) error {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `INSERT INTO cryptos (id, name, symbol, logo) VALUES (?, ?, ?, ?)`
	_, err := dbConn.Exec(query, c.ID, c.Name, c.Symbol, c.Logo)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении криптовалюты: %w", err)
	}
	return nil
}

// FindCrypto ищет криптовалюту сначала по символу, затем по имени
func FindCrypto(search string) (*Crypto, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `SELECT id, name, symbol, logo FROM cryptos WHERE symbol = ?`
	row := dbConn.QueryRow(query, search)
	var c Crypto
	err := row.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo)
	if err == nil {
		return &c, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("ошибка поиска по символу '%s': %w", search, err)
	}
	query = `SELECT id, name, symbol, logo FROM cryptos WHERE name = ?`
	row = dbConn.QueryRow(query, search)
	err = row.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("криптовалюта не найдена ни по символу, ни по имени '%s'", search)
		}
		return nil, fmt.Errorf("ошибка поиска по имени '%s': %w", search, err)
	}
	return &c, nil
}

// GetCryptoByID возвращает криптовалюту по её id
func GetCryptoByID(id int) (*Crypto, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `SELECT id, name, symbol, logo FROM cryptos WHERE id = ?`
	row := dbConn.QueryRow(query, id)
	var c Crypto
	err := row.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("криптовалюта с id '%d' не найдена", id)
		}
		return nil, fmt.Errorf("ошибка при сканировании данных: %w", err)
	}
	return &c, nil
}

// GetCryptoByName возвращает криптовалюту по её имени
func GetCryptoByName(name string) (*Crypto, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `SELECT id, name, symbol, logo FROM cryptos WHERE name = ?`
	row := dbConn.QueryRow(query, name)
	var c Crypto
	err := row.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("криптовалюта с именем '%s' не найдена", name)
		}
		return nil, fmt.Errorf("ошибка при сканировании данных: %w", err)
	}
	return &c, nil
}

// GetCryptoBySymbol возвращает криптовалюту по её символу
func GetCryptoBySymbol(symbol string) (*Crypto, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `SELECT id, name, symbol, logo FROM cryptos WHERE symbol = ?`
	row := dbConn.QueryRow(query, symbol)
	var c Crypto
	err := row.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("криптовалюта с символом '%s' не найдена", symbol)
		}
		return nil, fmt.Errorf("ошибка при сканировании данных: %w", err)
	}
	return &c, nil
}

// GetAllCryptos возвращает список всех криптовалют
func GetAllCryptos() ([]Crypto, error) {
	once.Do(func() { dbConn, _ = db.InitDB(dbPath, migrate) })
	query := `SELECT id, name, symbol, logo FROM cryptos`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()
	var cryptos []Crypto
	for rows.Next() {
		var c Crypto
		if err := rows.Scan(&c.ID, &c.Name, &c.Symbol, &c.Logo); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных: %w", err)
		}
		cryptos = append(cryptos, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}
	return cryptos, nil
}
