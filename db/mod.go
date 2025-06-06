package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB инициализирует соединение с SQLite базой данных
// path - путь к файлу базы данных
// migrates - применить миграции
func InitDB(path string, migrates ...func(db *sql.DB) error) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_sync=NORMAL&_foreign_keys=on", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return db, fmt.Errorf("не удалось открыть базу данных: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		return db, fmt.Errorf("не удалось проверить соединение с базой данных: %w", err)
	}
	for _, m := range migrates {
		if err := m(db); err != nil {
			return db, fmt.Errorf("не удалось применить миграции: %w", err)
		}
	}
	return db, nil
}
