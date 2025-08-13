# Domains: identity, project, policy/rbac (AppSecHub)

## Mục tiêu
- Chuẩn hóa tách domain cho mô hình: global role (identity) + project-scoped roles (project) + policy (role→permissions).
- Bám Clean Architecture/DDD đang dùng trong AppSecHub.

## Bounded contexts

### identity
- Ý nghĩa: Danh tính toàn cục, xác thực, vai trò toàn cục.
- Entities
  - User { id, email, password_hash, global_role: 'admin'|'user', created_at, updated_at }
- VOs
  - Email, GlobalRole
- Invariants
  - Email unique
  - GlobalRole ∈ {'admin','user'}
- Ports
  - UserRepository

### project
- Ý nghĩa: Quản lý dự án và membership theo vai trò cục bộ (manager|validator|dev).
- Entities
  - Project { id, name, ... }
  - ProjectMembership { project_id, user_id, project_role }
- VOs
  - ProjectID, ProjectRole(enum: manager|validator|dev)
- Invariants
  - (project_id, user_id) unique
  - ProjectRole hợp lệ theo enum
- Ports
  - ProjectRepository
  - MembershipRepository

### policy (authorization)
- Ý nghĩa: Cung cấp bảng ánh xạ role → permissions (chuẩn YAML, hot-reload).
- Entities
  - Policy { role, permissions[] }
- VOs
  - PermissionKey (string, định dạng `resource:action`, ví dụ: `project:write`)
- Invariants
  - Role phải có trong YAML; permissions là tập con keys được hệ thống nhận diện
- Ports
  - PolicyProvider (load/refresh từ YAML; cache in-memory)

## Quan hệ
- identity.User 1-* project.ProjectMembership (mỗi membership nằm trong 1 project)
- project.Project 1-* project.ProjectMembership
- policy.Policy dùng bởi Authorizer để quyết định quyền

## Schema (Postgres)
```sql
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  global_role text NOT NULL CHECK (global_role IN ('admin','user')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE project_memberships (
  project_id uuid NOT NULL,
  user_id uuid NOT NULL,
  project_role text NOT NULL CHECK (project_role IN ('manager','validator','dev')),
  PRIMARY KEY (project_id, user_id),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## RBAC/Authorizer (application service)
- Quy tắc:
  - Nếu user.global_role = 'admin' → allow tất cả
  - Ngược lại: lấy project_id (từ path/query/body), tra membership → project_role
  - policy: role→permissions (YAML) → kiểm tra chứa đủ permissions yêu cầu
- Middleware đề xuất: `RequireProjectPermissions(perms... , extractProjectID)`
  - extractProjectID: hàm lấy project_id từ request
  - Cache membership và policy 60s, bust khi thay đổi
  - Audit quyết định allow/deny kèm `request_id`, `user_id`, `project_id`, `perms`

## Permissions mẫu (YAML)
```yaml
roles:
  manager:
    - project:read
    - project:write
    - member:manage
    - repo:sync
    - finding:triage
    - sbom:import
  validator:
    - project:read
    - finding:triage
    - approve:gate
  dev:
    - project:read
    - finding:view
```

## Mapping router
- `PUT /v1/projects/:id` → RequireProjectPermissions(project:write)
- `POST /v1/projects/:id/members` → RequireProjectPermissions(member:manage)
- `POST /v1/projects/:id/repos/sync` → RequireProjectPermissions(repo:sync)
- `POST /v1/projects/:id/findings/triage` → RequireProjectPermissions(finding:triage)

## Value Objects
- Email (identity) — dùng `net/mail`, không nhận display name
- GlobalRole — enum admin|user
- ProjectRole — enum manager|validator|dev
- PermissionKey — chuỗi `resource:action`
- ProjectID/UserID — uuid

## Biên giới & phụ thuộc
- `interfaces/http` → `application` → `domain`
- `infras` → `domain` (implement Ports)
- Chỉ `cmd/api` wire mọi thứ
- DTO chỉ sống ở HTTP; domain VO/Entity không import framework

## Kiểm thử
- Authorizer: bảng quyết định (role→perms), trường hợp admin bypass
- MembershipRepository: CRUD + unique constraint
- Middleware RequireProjectPermissions: allow/deny theo các case
- End-to-end: route project update với các vai trò khác nhau

## Lộ trình triển khai
1) Thêm migrations users/projects/project_memberships (nếu chưa đủ)
2) Tạo domain `project` + Ports + Repos (sqlc)
3) PolicyProvider từ YAML (cache) + Authorizer service
4) Middleware `RequireProjectPermissions` + áp dụng lên routes dự án
5) Viết tests (unit + handler)