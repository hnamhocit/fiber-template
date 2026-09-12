# Go Fiber Production Template

A production-ready, highly robust, and developer-friendly template for building RESTful APIs using [Go Fiber v3](https://gofiber.io/). This template is structured following Clean Architecture principles, ensuring scalability, maintainability, and enterprise-grade security.

## 🌟 Key Features

- **Blazing Fast Framework**: Powered by Fiber v3.
- **Clean Architecture**: Domain-driven feature design (`handler` -> `service` -> `repository`).
- **Database Support**: Built-in support for PostgreSQL, MySQL, and SQLite (using `sqlx`), equipped with Atlas migrations.
- **Advanced Rate Limiting**: Global Redis-backed rate limiting (safe for multi-node deployments).
- **Graceful Shutdown**: Zero-downtime deployments with proper termination contexts.
- **Production-grade Logging**: Structured JSON logging (`slog`) for production (Loki/Datadog friendly) and pretty-colored console logging for development.
- **Security First**: 
  - Strict CORS validation (`IsProduction()` checks).
  - Reverse Proxy IP Trust (`TrustProxy` with explicit CIDR blocks) to prevent IP spoofing.
  - Helmet middleware for HTTP header security.
- **Validation Layer**: Built-in generic Request Body validation via `go-playground/validator`.
- **Standardized Error Responses**: Centralized JSON error handling (including 422 Validation Errors with field-level details).
- **Kubernetes Ready**: Dedicated `Liveness` and `Readiness` probes built-in.
- **API Documentation**: Live interactive OpenAPI spec via Redoc (`/docs`).
- **Infrastructure Integrations**: Seamless integration with Redis (Caching) and MinIO (S3-compatible Object Storage).

## 📁 Project Structure

```text
.
├── cmd/
│   └── server/          # Application entrypoint (main.go)
├── internal/
│   ├── bootstrap/       # Infrastructure initialization (DB, Redis, MinIO)
│   ├── cache/           # Redis client wrapper
│   ├── config/          # Environment variable loading & validation
│   ├── database/        # Database connection pool & driver detection
│   ├── features/        # Domain business logic (e.g., users)
│   ├── healthcheck/     # Liveness/Readiness probes (K8s friendly)
│   ├── httpx/           # HTTP helpers, generic Bind, and standard JSON responses
│   ├── logger/          # Env-aware slog configuration (JSON vs Pretty)
│   ├── server/          # Fiber app setup, middleware, and routing
│   ├── storage/         # MinIO/S3 client wrapper
│   └── validator/       # Struct validation instance
├── migrations/          # SQL Database migration files (Atlas)
├── schema.sql           # Database schema definition
└── docker-compose.yaml  # Local development infrastructure
```

## 🚀 Getting Started

### 1. Prerequisites
- **Go**: 1.26 or newer
- **Docker** & **Docker Compose**

### 2. Environment Setup
Clone the repository and set up your environment variables:
```bash
cp .env.example .env
```
*(Fill in the passwords and secrets in your `.env` file)*

### 3. Spin up Infrastructure
Start the required databases and services (PostgreSQL, Redis, MinIO) locally:
```bash
docker-compose up -d
```

### 4. Run the Application
```bash
go run ./cmd/server/main.go
```
The server will start at `http://localhost:8080`. 

## 📖 API Documentation
Once the server is running, you can view the beautiful interactive API documentation by visiting:
**[http://localhost:8080/docs](http://localhost:8080/docs)**

## 🛡️ Best Practices Built-in

### Global Error Handling
Forget plain text errors! Every panic, 404, or validation failure is cleanly intercepted and formatted into a consistent JSON response:
```json
{
  "success": false,
  "error": "validation failed",
  "errors": [
    { "field": "email", "message": "invalid format" }
  ]
}
```

### Generic Body Binding
In your controllers, parsing and validating JSON is simplified to 4 lines of code:
```go
req, err := httpx.Bind[CreateRequest](c)
if err != nil {
    return err // Auto-responds with 400 or 422 JSON
}
```

### Smart Logger
HTTP 500s are automatically logged as `ERROR`, 400s as `WARN`, and everything else as `INFO`. 

## ☁️ Deployment Notes

When deploying to a Production environment (e.g., Kubernetes, AWS ECS, DigitalOcean):
1. **Set `APP_ENV=production`**: This enables JSON logging and enforces strict security checks.
2. **Configure CORS**: You MUST set `CORS_ORIGINS=https://your-domain.com`. Wildcards (`*`) are blocked in production to prevent CSRF attacks.
3. **Configure Proxies**: Ensure `TRUSTED_PROXIES` is populated with the CIDR blocks of your Load Balancer / WAF to accurately identify Client IPs and prevent the Rate Limiter from blocking the proxy itself.
