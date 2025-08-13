-- name: UpsertAsset :exec
INSERT INTO assets (id, type, environment, criticality, owner_team, first_seen, last_seen, internet_exposed, tags, meta)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
  type = EXCLUDED.type,
  environment = EXCLUDED.environment,
  criticality = EXCLUDED.criticality,
  owner_team = EXCLUDED.owner_team,
  last_seen = GREATEST(assets.last_seen, EXCLUDED.last_seen),
  internet_exposed = EXCLUDED.internet_exposed,
  tags = EXCLUDED.tags,
  meta = EXCLUDED.meta;

-- name: GetAssetByID :one
SELECT id, type, environment, criticality, owner_team, first_seen, last_seen, internet_exposed, tags, meta, created_at, updated_at
FROM assets
WHERE id = $1;

-- name: ListAssets :many
SELECT id, type, environment, criticality, owner_team, first_seen, last_seen, internet_exposed, tags, meta, created_at, updated_at
FROM assets
ORDER BY last_seen DESC;

-- name: TouchSeen :exec
UPDATE assets
SET last_seen = GREATEST(last_seen, $2)
WHERE id = $1;

-- Application profile

-- name: UpsertApplicationProfile :exec
INSERT INTO applications (id, asset_id, slug, display_name, project_name, owner_team, language, framework, runtime, tags, meta)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
  asset_id = EXCLUDED.asset_id,
  slug = EXCLUDED.slug,
  display_name = EXCLUDED.display_name,
  project_name = EXCLUDED.project_name,
  owner_team = EXCLUDED.owner_team,
  language = EXCLUDED.language,
  framework = EXCLUDED.framework,
  runtime = EXCLUDED.runtime,
  tags = EXCLUDED.tags,
  meta = EXCLUDED.meta;

-- name: GetApplicationProfileByAssetID :one
SELECT id, asset_id, slug, display_name, project_name, owner_team, language, framework, runtime, tags, meta, created_at, updated_at
FROM applications
WHERE asset_id = $1;

-- name: LinkApplicationRepo :exec
INSERT INTO application_repos (application_id, repo_id, subdir, build_file, default_branch)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (application_id, repo_id, subdir) DO UPDATE SET
  build_file = EXCLUDED.build_file,
  default_branch = EXCLUDED.default_branch;

-- name: UnlinkApplicationRepo :exec
DELETE FROM application_repos
WHERE application_id = $1 AND repo_id = $2 AND ((subdir IS NULL AND $3 IS NULL) OR subdir = $3);

