package server

import (
	"github.com/jmoiron/sqlx"

	"github.com/hnamhocit/fiber-template/internal/cache"
	"github.com/hnamhocit/fiber-template/internal/config"
	"github.com/hnamhocit/fiber-template/internal/storage"
)

// Deps is the single dependency container handed to every feature module.
// Optional infra fields are nil when not configured; check with .Available().
// Features must NEVER read os.Getenv themselves — everything arrives via Deps.
type Deps struct {
	Cfg   *config.Config
	DB    *sqlx.DB        // always non-nil (required infra)
	Redis *cache.Client   // nil when REDIS_URL unset
	Minio *storage.Client // nil when MINIO_ENDPOINT unset
}
