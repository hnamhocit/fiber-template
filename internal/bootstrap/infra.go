// Package bootstrap wires infrastructure dependencies into a single struct.
// Main orchestrates; it does not know how each dependency is built.
package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/hnamhocit/fiber-template/internal/cache"
	"github.com/hnamhocit/fiber-template/internal/config"
	"github.com/hnamhocit/fiber-template/internal/database"
	"github.com/hnamhocit/fiber-template/internal/storage"
)

// Close releases all resources in reverse order of init.
func (i *Infra) Close() error {
	if i.Redis != nil {
		i.Redis.Close()
	}
	if i.DB != nil {
		i.DB.Close()
	}
	return nil
}

// Infra holds all infrastructure clients.
// Nil fields mean "not configured" (env var empty).
type Infra struct {
	DB       *sqlx.DB
	DBDriver string // "postgres" | "mysql" | "sqlite"
	Redis    *cache.Client
	Minio    *storage.Client
}

// Init connects to all configured infrastructure.
// Unconfigured (env empty) = skipped silently. Configured but broken = startup error.
func Init(ctx context.Context, cfg *config.Config, svc config.Services) (*Infra, error) {
	infra := &Infra{}

	// Driver: DB_DRIVER overrides, otherwise auto-detected from DATABASE_URL.
	driver, err := database.ResolveDriver(os.Getenv("DB_DRIVER"), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	db, err := database.NewWithDriver(ctx, driver, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database (%s): %w", driver, err)
	}
	infra.DB = db
	infra.DBDriver = driver

	if svc.RedisEnabled() {
		redis, err := cache.New(svc)
		if err != nil {
			return nil, fmt.Errorf("init redis: %w", err)
		}
		infra.Redis = redis
	}

	if svc.MinioEnabled() {
		minio, err := storage.New(svc)
		if err != nil {
			return nil, fmt.Errorf("init minio: %w", err)
		}

		if err := minio.EnsureBucket(ctx); err != nil {
			return nil, fmt.Errorf("init minio: %w", err)
		}
		infra.Minio = minio
	}

	return infra, nil
}
