package database

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib" // registers sql driver "pgx"
	"github.com/jmoiron/sqlx"
)

// NewPostgres opens a PostgreSQL pool with an explicit driver choice.
func NewPostgres(ctx context.Context, dsn string) (*sqlx.DB, error) {
	return NewWithDriver(ctx, "postgres", dsn)
}
