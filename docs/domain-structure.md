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