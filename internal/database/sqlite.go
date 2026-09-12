package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // registers sql driver "sqlite" (pure Go, no cgo)
)

// NewSQLite opens a SQLite pool with an explicit driver choice.
func NewSQLite(ctx context.Context, dsn string) (*sqlx.DB, error) {
	return NewWithDriver(ctx, "sqlite", dsn)
}
