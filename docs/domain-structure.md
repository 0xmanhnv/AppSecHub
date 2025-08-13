# Domain and Value Objects (Clean Architecture + DDD)

## Mục tiêu
- Chuẩn hóa cách tách domain theo bounded context, bám sát kiến trúc hiện tại: `domain` (Entities/VOs) → `application` (UseCases/Ports) → `interfaces/http` (DTO/handlers) & `infras` (adapters).
- Mỗi domain sở hữu dữ liệu và bất biến (invariants), không phụ thuộc framework.
- Dùng Value Object (VO) để gói logic hợp lệ của giá trị có ngữ nghĩa.

## Quy ước kiến trúc
- `internal/domain/<context>`
  - Entities, Value Objects, Domain errors
  - Không import `application/*`, `interfaces/*`, `infras/*`
- `internal/application/usecase/<context>`
  - Use cases (dịch vụ ứng dụng), chỉ dùng Ports và Domain
- `internal/application/ports/<context>`
  - Interfaces (repo/client) mà UseCases dùng
- `internal/infras/...`
  - Adapters/implementations cho Ports (postgres/sqlc, redis, http clients)
- `internal/interfaces/http`
  - DTO/binding/validation, handlers, router, middleware, response
- Wire toàn cục tại `cmd/api` (composition root)

## Value Object (VO)
- Đại diện cho giá trị có ngữ nghĩa (Email, URL chuẩn hóa, Severity, PURL…).
- Bất biến, so sánh theo giá trị, tự đảm bảo hợp lệ khi tạo.
- Lợi ích: gom logic hợp lệ 1 chỗ, an toàn khi truyền giữa tầng/lớp.

VO dùng/đề xuất:
- `user.Email`, `user.Role`
- `sourcecontrol.SourceType` (gitlab/github/bitbucket), `sourcecontrol.NormalizedURL`
- `asset.AssetType` (domain/host/ip/service/application), `asset.Environment`, `asset.Criticality`
- `sbom.PURL`
- `finding.Severity`, `finding.Status`, `finding.DedupeKey`

## Bounded contexts (Domains)

### 1) identity
- Ý nghĩa: danh tính, quyền và phân quyền; nền tảng cho mọi domain khác.
- Entities: `User`, `Role`, `Permission` (policy YAML mapping), `AuditLog` (append-only)
- VOs: `Email`, `RoleName`
- Bất biến: email duy nhất theo org; role chỉ là nhãn, quyền đến từ policy
- Ports: `UserRepository`, `RoleRepository`
- Use cases: đăng ký/đăng nhập, cấp JWT/refresh (đã có), tra cứu user, audit ghi nhận hành động

### 2) sourcecontrol
- Ý nghĩa: quản lý endpoint SCM và kho nguồn (repo) từ GitLab/GitHub/Bitbucket
- Entities: `SourceControl` (url gốc + loại), `Repo` (repo từ provider)
- VOs: `SourceType`, `NormalizedURL`
- Bất biến: `unique (normalized_url, type)`; `unique (source_control, provider_repo_id)`
- Ports: `SourceControlRepository`, `RepoRepository`, `GitLabClient`
- Use cases: `ListSourceControls`, `UpsertSourceControl`, `SyncGitLabRepos`

### 3) asset
- Ý nghĩa: Asset Inventory (domain/host/ip/service/application) + ownership/env/lifecycle
- Entities: `Asset` (có `first_seen/last_seen`, `internet_exposed`), `Ownership`
- VOs: `AssetType`, `Environment`, `Criticality`
- Bất biến: `last_seen` tăng dần; `first_seen` không giảm; mapping owner/team tồn tại
- Ports: `AssetRepository`
- Use cases: `UpsertAsset`, `ListAssets`, `TouchSeen`

### 4) sbom
- Ý nghĩa: quản lý thành phần phần mềm (components/packages) và phụ thuộc, nguồn SBOM/SCA
- Entities: `Component`, `License`, `DependencyEdge`
- VOs: `PURL`
- Bất biến: `unique (asset_id, purl)`; component immutable theo checksum
- Ports: `ComponentRepository`
- Use cases: `ImportCycloneDX`, `ListComponents`

### 5) vulnkb (vulnerability knowledge base)
- Ý nghĩa: kho tri thức CVE/CWE/CVSS/EPSS để enrich findings/components
- Entities: `Advisory`, `Score`
- Bất biến: CVE duy nhất; nhiều nguồn score theo thời gian
- Ports: `AdvisoryRepository`
- Use cases: `UpsertAdvisory`, `UpsertScore`, `GetAdvisoryByCVE`

### 6) finding
- Ý nghĩa: hợp nhất phát hiện từ SAST/DAST/SCA/Container/IaC; triage, SLA, assignment
- Entities: `Finding`, `Evidence`, (tuỳ chọn) `Attachment`
- VOs: `Severity`, `Status`, `DedupeKey`
- Bất biến: `unique (dedupe_key)`; state machine hợp lệ của `status`; `sla_due` theo severity/env
- Ports: `FindingRepository`
- Use cases: `IngestSARIF/Trivy`, `ListFindings`, `UpdateFindingStatus`

### 7) scanners (integration layer)
- Ý nghĩa: quản lý lượt ingest/chạy tool, đảm bảo idempotent
- Entities: `IngestJob`, `Run`
- Bất biến: at-least-once ingest, idempotent trên `dedupe_key`
- Ports: adaptors cho file/API (không expose domain entity trực tiếp)
- Use cases: `EnqueueIngest`, `RecordRun`

### 8) pipeline (sau)
- Ý nghĩa: policy gate trong CI/CD; artifact/provenance
- Entities: `PipelineRun`, `GatePolicy`, `GateResult`
- Bất biến: quyết định gate tái lập được; policy versioned
- Ports: `PolicyRepository` (nếu cần), không ràng buộc scanner
- Use cases: `EvaluateGate`

### 9) ticketing/notify (sau)
- Ý nghĩa: liên kết issue ngoài (Jira/GitLab), thông báo
- Entities: `TicketLink`, `Subscription`
- Bất biến: `unique (provider, external_id)`; lọc thông báo theo subscription
- Ports: `TicketClient`, `NotifyClient`
- Use cases: `CreateTicketLink`, `SendNotification`

## Data & Migrations (MVP)
- `source_controls`: `(id, url, normalized_url, type)` unique `(normalized_url, type)`
- `repos`: provider repo metadata, unique `(source_control, sourcecontrol_id)`
- `assets`: loại, fqdn/ip/port, owner/env/criticality/lifecycle, `first_seen/last_seen`, `internet_exposed`, `tags`, `meta`
- `components`: `(id, asset_id, purl, name, version, license, source, checksum)` unique `(asset_id, purl)`
- `advisories` + `advisory_scores`
- `component_vulns` (N-N)
- `findings`: `(asset_id|repo_id, tool, rule, location, component_id?, advisory_id?, severity, status, assignee?, sla_due?, evidence, raw, dedupe_key)` unique `(dedupe_key)`

## Biên giới tầng & DTO
- HTTP DTO chỉ sống ở `interfaces/http`; không rò rỉ vào domain.
- Mapping DTO↔VO/Entity diễn ra ở handlers/usecases.
- Đưa validate (format/field) vào VO khi có thể (email, purl, normalized url…)

## Chính sách import & phụ thuộc
- `interfaces/http` → `application` → `domain`
- `infras` → `domain` (implement Ports)
- Chỉ `cmd/api` thấy tất cả để wire

## Kiểm thử
- Unit: VO (edge cases), UseCases (fake ports), mappers
- Handler: table-driven + `httptest`
- Integration: sqlc repos + Postgres, Redis state/limiter, GitLab fake client

## Observability & Security
- Metrics: http, db pool, ingest latency, rate-limit
- Tracing: HTTP + pgx
- JWT: RS256/EdDSA + `kid`; JWKS (public keys)
- PII minimization: băm email cho limiter; không log secrets

## Lộ trình tạo domain (khuyến nghị)
1) `sourcecontrol`, `asset` (migrations + repos + list/sync)
2) `sbom` (CycloneDX import) + `vulnkb` (NVD tối thiểu)
3) `finding` (SARIF/Trivy ingest + triage)
4) `pipeline` + `ticketing/notify`

> Ghi chú: Bắt đầu ít domain, mở rộng khi có use case rõ và bất biến riêng. Tránh gộp domain có vòng đời khác nhau.

## Asset Inventory: Best Practical (Chuẩn thực dụng)
- Phân loại nhất quán: `AssetType = domain | host | ip | service | application` (enum), `Environment = prod|stg|dev|test`, `Criticality = tier1|tier2|tier3`.
- Định danh duy nhất theo loại + normalize mạnh, đặt unique constraint:
  - domain: `normalized_fqdn` (lowercase, punycode, bỏ dấu chấm cuối)
  - ip: `ip`, `version` (4/6)
  - host: `cloud_instance_id` hoặc `machine_id`; fallback `hostname@domain`
  - service: `(host_id, port, protocol)` hoặc `normalized_endpoint`
  - application: `slug` + `environment` (unique `(slug, environment)`)
- Lifecycle: `active` → `stale` (không thấy X ngày) → `retired` (xác nhận); tự động hoá stale dựa `last_seen`.
- Ownership là trung tâm: yêu cầu `owner_team`/`service_owner`; map từ cloud tags, repo/subdir catalog; SLA/risk dựa `Criticality` + `Environment`.
- Ingestion idempotent + dedupe: mọi collector thực hiện `upsert` theo khoá tự nhiên + `TouchSeen` cập nhật `last_seen`; lưu `source`/`evidence`/`payload_hash` để audit.
- Quan hệ then chốt: `application(asset)` ↔ N..N `repo` (monorepo bằng `subdir`), `application(asset)` ↔ N `components` (SBOM) unique `(asset_id, purl)`; `asset` ↔ N `findings`.

## Application trong Asset (mở rộng schema DDD)
- Application là chuyên biệt của `Asset` với `type=application`. Thuộc tính chuyên biệt nên tách về bảng/profile 1–1 để query tốt và không làm nặng bảng `assets`.

### Entities/VOs
- Asset: giữ như hiện tại (`type`, `environment`, `criticality`, `owner_team`, `first_seen`, `last_seen`, `internet_exposed`, `tags`, `meta`).
- ApplicationProfile (thuộc context `asset`): `application_id(FK assets.id)`, `slug(unique)`, `display_name`, `project_name` (tên dự án phát triển), `owner_team`, `language`, `framework`, `runtime`, `tags[]`, `meta jsonb`.
- ApplicationRepo (liên kết với context `sourcecontrol`): `application_id(FK)`, `repo_id(FK)`, `subdir?`, `build_file?`, `default_branch?` với unique `(application_id, repo_id, subdir)`.
- Components (context `sbom`): giữ như docs hiện tại, gắn theo `asset_id` của application.

### Bất biến (Invariants)
- ApplicationProfile: 1–1 với `Asset(type=application)`; `slug` unique trong toàn tổ chức.
- ApplicationRepo: unique `(application_id, repo_id, subdir)` để hỗ trợ monorepo; `subdir` rỗng nghĩa là root.
- Components: unique `(asset_id, purl)`; component immutable theo `checksum`.

### Data & Migrations (bổ sung)
- Bổ sung 2 bảng song song với phần "Data & Migrations (MVP)":
  - `applications`: `(id PK, asset_id unique FK->assets(id), slug unique, display_name, project_name, owner_team, language, framework, runtime, tags text[], meta jsonb, created_at, updated_at)`
  - `application_repos`: `(application_id FK, repo_id FK, subdir text null, build_file text null, default_branch text null, created_at, updated_at, unique(application_id, repo_id, subdir))`

### Ports & Use Cases
- Ports (`internal/application/ports/asset`): `ApplicationRepository` quản lý profile + liên kết repo.
- Use cases (`internal/application/usecase/asset`):
  - `UpsertApplication` (tạo/cập nhật profile 1–1 cho `Asset(type=application)`).
  - `LinkApplicationRepo`/`UnlinkApplicationRepo`.
  - `ListApplications` (filter theo `owner_team`, `language`, `framework`, `repo`, `tag`).
  - `GetApplicationDetail` (profile + repos + SBOM summary + findings gần nhất).
- Dòng dữ liệu: `SyncGitLabRepos` (context `sourcecontrol`) để đồng bộ repo, sau đó `LinkApplicationRepo` theo mapping; `ImportCycloneDX` (context `sbom`) ghi `components(asset_id=application_asset_id)`.

### API gợi ý (interfaces/http)
- `POST /applications` (upsert profile từ `asset_id`), `GET /applications?filters=...`.
- `POST /applications/:id/repos` (link repo + subdir), `DELETE /applications/:id/repos/:repoId`.
- `GET /applications/:id` (gộp profile + repos + components summary + findings mới nhất).

## Clean Architecture + DDD Checklist (áp dụng cho dự án)
- Biên giới import: `interfaces/http` → `application` → `domain`; `infras` → `domain`; chỉ `cmd/api` được wire toàn cục.
- Domain tối giản: chỉ Entities/VOs/Errors, không dùng framework; validate quy tắc vào VO (email, normalized url, purl, slug).
- Use cases tinh gọn: stateless, orchestrate ports, không chứa chi tiết adapter.
- Ports đặt tại `internal/application/ports/<context>`; adapters ở `internal/infras/...`.
- DTO chỉ sống ở `interfaces/http` và map sang VO/Entity ở handler/usecase.
- Idempotency: mọi use case write-side là upsert theo khoá tự nhiên; `TouchSeen` cập nhật `last_seen`.
- Testing: unit VO, unit use case (fake ports), handler tests, integration sqlc+Postgres.

## Roadmap triển khai Clean Architecture + DDD
1) Khởi tạo kiến trúc
   - Cấu trúc thư mục như phần "Quy ước kiến trúc"; wire DI tại `cmd/api`.
   - Thiết lập config, logger, metrics, tracing; chuẩn hoá lỗi domain ↔ HTTP.
2) Migrations & Repos (MVP)
   - Tạo tables: `source_controls`, `repos`, `assets`, `applications`, `application_repos`, `components`, `advisories`, `advisory_scores`, `component_vulns`, `findings`.
   - Sinh code `sqlc` và implement repositories cho `sourcecontrol`, `asset`, `sbom`, `finding`.
3) Use cases cốt lõi
   - `sourcecontrol`: `UpsertSourceControl`, `SyncGitLabRepos`.
   - `asset`: `UpsertAsset`, `ListAssets`, `TouchSeen`, `UpsertApplication`, `LinkApplicationRepo`.
   - `sbom`: `ImportCycloneDX`, `ListComponents`.
   - `finding`: `IngestSARIF/Trivy` (giai đoạn sau), `ListFindings`.
4) Interfaces/HTTP
   - Handlers + DTO cho list/upsert assets, applications, link repos, list components.
   - AuthN/AuthZ dựa `identity` (JWT + policies) cho các route.
5) Observability & Security
   - Metrics: HTTP, DB pool, ingest latency, stale rate; Tracing pgx.
   - PII minimization, audit-log append-only cho thay đổi owner/criticality/exposure.
6) Collectors & Đồng bộ dữ liệu
   - Collector GitLab/GitHub (repos) → `SyncGitLabRepos`.
   - SBOM ingest (CycloneDX) gắn `asset_id` application.
   - DNS/Cert scan để xác định `internet_exposed` cho domain/service.
7) QA & Triển khai tăng dần
   - Unit/Integration tests xanh; dashboard sức khỏe inventory (coverage owner/env/stale/exposed).
   - Rollout theo miền: sourcecontrol → asset/application → sbom → finding.

### Mốc thành công (KPIs)
- ≥95% `application` có `owner_team` và `environment` hợp lệ.
- ≤1% bản ghi trùng (theo unique keys); 100% collectors idempotent.
- Báo cáo stale/retire tự động hoạt động (không false positive quá 5%).