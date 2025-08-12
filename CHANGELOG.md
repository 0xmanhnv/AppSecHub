## Changelog

### Unreleased

- security,jwt: always set `typ=JWT` header; enforce KID when multiple RS256/EdDSA public keys (fallback only when a single key is configured).
- http,auth: standardize 401 responses using response helpers; attach `WWW-Authenticate` details.
- rate-limit: standardize 429 envelope `{ error: { code: "too_many_requests", message: "too many requests" } }` with `Retry-After` and `X-RateLimit-*` headers for both in-memory and Redis limiters; add `CodeTooManyRequests` and helper.
- limiter(redis): per-email rate limit after normalization (trim/lower) and SHA256 hashing to protect PII; optional fail-closed via `HTTP_LOGIN_RATELIMIT_FAIL_CLOSED`.
- i18n: YAML-driven catalogs (`configs/locales`), locale middleware (`Accept-Language`), thread-safe rendering with fallback.
- validation: custom tags `strict_email`, `strong_password`; JSON field names in errors; rich `error.details` mapping; 413 handling for oversized bodies.
- router: split into `internal/interfaces/http/router/*` by concern (base/health/auth/admin); keep `internal/interfaces/http/route.go` as thin facade (`http.NewRouter`).
- jwt: configurable algorithms via env (`JWT_ALG` HS256/RS256/EdDSA), KID, key paths/dirs; verify against configured algorithms & public keys.
- refresh-tokens: map `redis.Nil` → `ErrInvalidRefreshToken`; idempotent revoke.
- bcrypt: `Hash` returns `(string, error)`; cost via `BCRYPT_COST`.

## Changelog

## Unreleased
- Observability: add `/metrics` (Prometheus), enable pprof in dev, optional OpenTelemetry tracing for HTTP and DB.
- CI/CD: add GitHub Actions (build, lint, unit + integration tests, govulncheck), optional Trivy image scan.

## 1.0.0 - 2025-08-10

### Added
- Layered architecture with composition root wiring (`cmd/api`).
- Health and readiness endpoints: `GET /healthz`, `GET /readyz` (DB ping with timeout).
- Auth: JWT with role in claims, auth middleware, RBAC with YAML policy (`RBAC_POLICY_PATH`).
- Auth endpoints: `POST /v1/auth/register`, `POST /v1/auth/login`, `GET /v1/auth/me`, `POST /v1/auth/change-password`.
- Feature-flagged refresh flow (Redis-backed): `POST /v1/auth/refresh`, `POST /v1/auth/logout` when `AUTH_REFRESH_ENABLED=true`.
- Rate limit for `/v1/auth/login` (in-memory); optional Redis distributed limiter when `REDIS_ADDR` set; 429 includes `Retry-After`, `X-RateLimit-*`.
- Security headers middleware (CSP, HSTS when HTTPS), CORS from env, trusted proxies config.
- Response envelope and error mapping helpers; friendly JSON binding/validation errors; request body size limit via `HTTP_MAX_BODY_BYTES`.
- Postgres migrations and repository; connection pool tuning via env.
- Seeding initial admin via `SEED_*` env.
- OpenAPI 3.0 schema (`/openapi.json`) and ReDoc (`/swagger`) in dev-only.
- DevEx: `Makefile` targets, `.golangci.yml`, `scripts/test_integration.sh`.
- Docker: multi-stage `build/Dockerfile` (distroless runtime), `docker-compose.dev.yml`, `docker-compose.yml`, `docker-compose.test.yml`.

### Changed
- Global structured logger initialization (`logger.Init`), use `logger.L()` across the app.
- JWT hardening: set `nbf=now`, validate `iss/aud/nbf/exp` with leeway; restrict to HS256.
- Router consolidated to a single `NewRouter(...)`; graceful shutdown and HTTP server timeouts.

### Fixed
- Prevent double route registration; clearer JSON error messages; corrected `NotBefore` validation.

### Security
- Do not expose `/swagger` or `/openapi.json` in production (only when `ENV=dev`).
- Notes on secrets management for `JWT_SECRET` and `DB_PASSWORD`.

### Docs
- README updated (architecture, ports glossary, feature flags, DB pool tuning, DevEx shortcuts).
- OpenAPI `info.version` bumped to `1.0.0`.


