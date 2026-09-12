package database

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// normalizeDriver maps friendly names to registered sql driver names.
func normalizeDriver(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "postgres", "postgresql", "pgx":
		return "pgx", nil
	case "mysql", "mariadb":
		return "mysql", nil
	case "sqlite", "sqlite3":
		return "sqlite", nil
	default:
		return "", fmt.Errorf("unknown database driver %q (want postgres|mysql|sqlite)", name)
	}
}

// friendly is the human-facing name used in logs and health checks.
func friendly(sqlDriver string) string {
	if sqlDriver == "pgx" {
		return "postgres"
	}
	return sqlDriver
}

// DetectDriver infers the sql driver from a DSN/URL so callers never
// pass the driver by hand:
//
//	postgres://…   postgresql://…   host=… dbname=…        -> pgx
//	mysql://…      user:pass@tcp(host:port)/db              -> mysql
//	sqlite://…     file:…           ./data/app.db           -> sqlite
func DetectDriver(dsn string) (string, error) {
	// 1. URL with an explicit scheme.
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		switch u.Scheme {
		case "postgres", "postgresql":
			return "pgx", nil
		case "mysql", "mariadb":
			return "mysql", nil
		case "sqlite", "sqlite3", "file":
			return "sqlite", nil
		}
	}

	// 2. Scheme-less DSNs (both drivers accept these).
	switch {
	case strings.Contains(dsn, "@tcp("), strings.Contains(dsn, "@unix("):
		return "mysql", nil // user:pass@tcp(host:3306)/db
	case strings.HasPrefix(dsn, "host="), strings.Contains(dsn, " dbname="):
		return "pgx", nil // libpq key=value form
	}

	// 3. Bare file path -> sqlite.
	lower := strings.ToLower(dsn)
	if strings.HasSuffix(lower, ".db") || strings.HasSuffix(lower, ".sqlite") || strings.HasSuffix(lower, ".sqlite3") {
		return "sqlite", nil
	}

	return "", fmt.Errorf("cannot detect database driver from DSN %q — set DB_DRIVER to override", dsn)
}

// ResolveDriver returns the effective friendly driver name,
// honoring an explicit override (DB_DRIVER) over auto-detection.
func ResolveDriver(override, dsn string) (string, error) {
	if override != "" {
		raw, err := normalizeDriver(override)
		if err != nil {
			return "", err
		}
		return friendly(raw), nil
	}
	raw, err := DetectDriver(dsn)
	if err != nil {
		return "", err
	}
	return friendly(raw), nil
}

// New opens a pool with the driver auto-detected from the DSN.
func New(ctx context.Context, dsn string) (*sqlx.DB, error) {
	driver, err := DetectDriver(dsn)
	if err != nil {
		return nil, err
	}
	return NewWithDriver(ctx, driver, dsn)
}

// NewWithDriver opens a pool with an explicit driver (friendly names accepted).
func NewWithDriver(ctx context.Context, driver, dsn string) (*sqlx.DB, error) {
	name, err := normalizeDriver(driver)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open(name, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db (%s): %w", name, err)
	}

	// (cores * 2) + 1, floor at 10 — safe for small VPS.
	maxOpen := runtime.NumCPU()*2 + 1
	if maxOpen < 10 {
		maxOpen = 10
	}
	// SQLite is single-writer: a big pool just manufactures SQLITE_BUSY.
	if name == "sqlite" {
		maxOpen = 1
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db (%s): %w", name, err)
	}

	return db, nil
}

// HealthCheck verifies database connectivity.
func HealthCheck(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
