// Package healthcheck provides liveness and readiness probes.
//
//	GET /health  liveness:  process is up; checks NO dependencies on purpose
//	GET /ready   readiness: every registered dependency check must pass (else 503)
package healthcheck

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Check reports the health of a single dependency.
type Check struct {
	Name string
	Fn   func(ctx context.Context) error
}

// CheckResult is the per-dependency outcome exposed by GET /ready.
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ReadyResponse is the payload of GET /ready.
type ReadyResponse struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

// Registry holds all registered dependency checks.
type Registry struct {
	mu      sync.RWMutex
	checks  []Check
	timeout time.Duration
}

// NewRegistry creates a registry; timeout bounds every individual check.
func NewRegistry(timeout time.Duration) *Registry {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Registry{timeout: timeout}
}

// Register adds a dependency check. Not configured? Simply don't register it.
func (r *Registry) Register(c Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks = append(r.checks, c)
}

// Run executes all checks concurrently, each bounded by the registry timeout.
func (r *Registry) Run(ctx context.Context) ReadyResponse {
	r.mu.RLock()
	checks := make([]Check, len(r.checks))
	copy(checks, r.checks)
	r.mu.RUnlock()

	results := make([]CheckResult, len(checks))
	var wg sync.WaitGroup
	for i, c := range checks {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, r.timeout)
			defer cancel()

			start := time.Now()
			err := c.Fn(cctx)
			res := CheckResult{Name: c.Name, Latency: time.Since(start).Round(time.Microsecond).String()}
			if err != nil {
				res.Status = "down"
				res.Error = err.Error()
			} else {
				res.Status = "up"
			}
			results[i] = res
		}(i, c)
	}
	wg.Wait()

	status := "ready"
	for _, res := range results {
		if res.Status != "up" {
			status = "not_ready"
			break
		}
	}
	return ReadyResponse{Status: status, Checks: results}
}

// Routes mounts the probes at the ROOT path (probes are not API surface).
func (r *Registry) Routes(rt fiber.Router) {
	rt.Get("/health", r.liveness)
	rt.Get("/ready", r.readiness)
}

// liveness godoc
//
//	@Summary		Liveness probe
//	@Description	Reports whether the process is alive. Deliberately checks NO dependencies: orchestrators use this probe to decide restarts, and a transient dependency hiccup must not kill the process.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (r *Registry) liveness(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "alive"})
}

// readiness godoc
//
//	@Summary		Readiness probe
//	@Description	Reports whether the service can take traffic. Runs every registered dependency check concurrently (2s timeout each); returns 503 if any dependency is down.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ReadyResponse	"all dependencies up"
//	@Failure		503	{object}	ReadyResponse	"at least one dependency down"
//	@Router			/ready [get]
func (r *Registry) readiness(c fiber.Ctx) error {
	resp := r.Run(c.Context())
	code := fiber.StatusOK
	if resp.Status != "ready" {
		code = fiber.StatusServiceUnavailable
	}
	return c.Status(code).JSON(resp)
}
