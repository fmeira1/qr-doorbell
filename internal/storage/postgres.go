package storage

import (
	"database/sql"
	"fmt"
)

// DB envuelve la conexión a Postgres.
// Usa database/sql para mantener las dependencias mínimas.
type DB struct {
	SQL *sql.DB
}

// OpenPostgres abre una conexión a Postgres usando un DSN/database URL.
//
// Ejemplos de DSN (con lib/pq):
//
//	postgres://user:pass@localhost:5432/qrdoorbell?sslmode=disable
func OpenPostgres(driverName, dsn string) (*DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping DB: %w", err)
	}
	return &DB{SQL: db}, nil
}
