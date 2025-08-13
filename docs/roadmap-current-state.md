# AppSecHub — Current State & Implementation Roadmap

## 1) Đã đạt
- Clean Architecture/DDD rõ ràng; router tách theo concern; composition root `cmd/api`.
- Domain identity (thay user): VO Email/Role; use cases Register/Login/GetMe/ChangePassword/Refresh/Logout.
- JWT: HS256/RS256/EdDSA, `typ=JWT`, enforce `kid` khi multi-key, leeway/nbf/iss/aud.
- Refresh tokens (Redis), map `redis.Nil`; bcrypt Hash trả error.
- Validation/i18n: 3 lớp, tags `strict_email`, `strong_password`; i18n YAML + LocaleMiddleware.
- HTTP: error envelope chuẩn (kèm 401/429), helpers đầy đủ; limiter IP/email (Redis, hash).
- OIDC: Entra ID skeleton (login/callback, state/nonce/PKCE [plain], verify ID token), use case LoginOIDC.
- Docs: README (flows, JWT, 401/429), domain-structure, migration-old-project, project-identity-rbac.

## 2) Còn thiếu
- Identity hoàn thiện:
  - Rà soát import, xóa hẳn `internal/domain/user`, admin endpoints (list/update role), audit logs.
- OIDC production:
  - State store → Redis, PKCE S256, policy provision (whitelist domain, auto-provision optional).
  - JWKS HTTP route (hiện có `PublicJWKs()`).
- Observability:
  - `/metrics` (Prometheus), tracing (OpenTelemetry), pprof (dev).
- CI/CD & Security:
  - GH Actions: build/test/lint/govulncheck; Trivy scan; i18n key-sync check.
- Tests:
  - Unit/handler cho identity; OIDC mocked provider; repo sqlc integration rộng hơn.
- Domain tiếp theo:
  - project + project_memberships + RBAC scoped (manager/validator/dev) + middleware RequireProjectPermissions.
  - sourcecontrol + asset + GitLab sync (read-only).
  - sbom + ingest CycloneDX; finding + ingest SARIF/Trivy.

## 3) Lộ trình triển khai (ưu tiên)

### P0 — Hoàn tất identity (1–2 ngày)
- Xóa `internal/domain/user` sau khi `rg "internal/domain/user"` không còn kết quả → build/test xanh.
- Expose JWKS:
  - Route `GET /.well-known/jwks.json` → `jwtSvc.PublicJWKs()`.
- Admin endpoints:
  - `GET /v1/admin/users` (list), `PATCH /v1/admin/users/:id/role` (update) — RequirePermissions.

### P1 — Observability & CI (1–2 ngày)
- `/metrics`: http requests, db pool stats, rate-limit hits, login attempts/fails.
- GH Actions: build/test, golangci-lint, govulncheck; Docker build + Trivy.
- i18n key-sync check (script compare en/vi).

### P2 — Project RBAC scoped (3–4 ngày)
- Migrations: `projects`, `project_memberships`.
- Domain `project` + Ports/Repos (sqlc).
- PolicyProvider YAML (role→perms); Authorizer; Middleware `RequireProjectPermissions(perms, extractProjectID)`.
- Áp dụng lên routes quan trọng (`/v1/projects/:id/...`).

### P3 — SourceControl + Asset (1–2 tuần)
- Migrations: `source_controls`, `repos`, `assets`.
- GitLab client + sync repos; map → assets/services; store `first_seen/last_seen`.
- API: list assets (filter owner/env/criticality/tag/text).

### P4 — SBOM + Findings (2–3 tuần)
- SBOM schema (`components`, `deps`, `component_vulns`), ingest CycloneDX; KB NVD tối thiểu.
- Findings schema + ingest SARIF/Trivy; dedupe; triage endpoints; SLA.

### P5 — OIDC productionize (song song)
- State/nonce PKCE S256 bằng Redis; domain allowlist; (optional) auto-provision.

## 4) Tiêu chí hoàn thành mỗi phase
- Build/test xanh, linter pass.
- Docs cập nhật (README snippets, usage, env).
- Metrics hiện số liệu chính; logs có request_id.
- Errors map đúng envelope; 401/429 chuẩn; rate-limit có headers.

## 5) Checklist ngắn
- [ ] Xóa `internal/domain/user`, expose JWKS
- [ ] /metrics + GH Actions + Trivy + i18n sync
- [ ] Migrations `projects`, `project_memberships`; Authorizer + middleware
- [ ] Migrations `source_controls`, `repos`, `assets`; GitLab sync; list assets
- [ ] SBOM ingest; Findings ingest; triage
- [ ] OIDC S256 + Redis state; provision policy