package database

import (
	"context"

	_ "github.com/go-sql-driver/mysql" // registers sql driver "mysql"
	"github.com/jmoiron/sqlx"
)

// NewMySQL opens a MySQL/MariaDB pool with an explicit driver choice.
func NewMySQL(ctx context.Context, dsn string) (*sqlx.DB, error) {
	return NewWithDriver(ctx, "mysql", dsn)
}
