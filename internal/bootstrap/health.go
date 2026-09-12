package bootstrap

import (
	"time"

	"github.com/hnamhocit/fiber-template/internal/healthcheck"
)

// Health builds a registry from the provided infra.
// Unconfigured dependencies (nil fields) are simply not registered.
func Health(infra *Infra) *healthcheck.Registry {
	hc := healthcheck.NewRegistry(2 * time.Second)

	if infra.DB != nil {
		hc.Register(healthcheck.Check{Name: "postgres", Fn: infra.DB.PingContext})
	}
	if infra.Redis != nil {
		hc.Register(healthcheck.Check{Name: "redis", Fn: infra.Redis.HealthCheck})
	}
	if infra.Minio != nil {
		hc.Register(healthcheck.Check{Name: "minio", Fn: infra.Minio.HealthCheck})
	}

	return hc
}
