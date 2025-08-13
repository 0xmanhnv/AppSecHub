-- Asset Inventory schema (assets, applications, application_repos)

CREATE TABLE IF NOT EXISTS assets (
  id UUID PRIMARY KEY,
  type TEXT NOT NULL CHECK (type IN ('domain','host','ip','service','application')),
  environment TEXT NOT NULL CHECK (environment IN ('prod','stg','dev','test')),
  criticality TEXT NOT NULL CHECK (criticality IN ('tier1','tier2','tier3')),
  owner_team TEXT,
  first_seen TIMESTAMPTZ NOT NULL,
  last_seen TIMESTAMPTZ NOT NULL,
  internet_exposed BOOLEAN NOT NULL DEFAULT FALSE,
  tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
CREATE INDEX IF NOT EXISTS idx_assets_env ON assets(environment);
CREATE INDEX IF NOT EXISTS idx_assets_owner ON assets(owner_team);
CREATE INDEX IF NOT EXISTS idx_assets_last_seen ON assets(last_seen);

DROP TRIGGER IF EXISTS trg_assets_set_updated_at ON assets;
CREATE TRIGGER trg_assets_set_updated_at
BEFORE UPDATE ON assets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- Application profile (1-1 with assets where type=application)
CREATE TABLE IF NOT EXISTS applications (
  id UUID PRIMARY KEY,
  asset_id UUID NOT NULL UNIQUE REFERENCES assets(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  display_name TEXT,
  project_name TEXT,
  owner_team TEXT,
  language TEXT,
  framework TEXT,
  runtime TEXT,
  tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_applications_owner ON applications(owner_team);

DROP TRIGGER IF EXISTS trg_applications_set_updated_at ON applications;
CREATE TRIGGER trg_applications_set_updated_at
BEFORE UPDATE ON applications
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- Link applications to repositories (supports monorepo via subdir)
CREATE TABLE IF NOT EXISTS application_repos (
  application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
  repo_id TEXT NOT NULL,
  subdir TEXT,
  build_file TEXT,
  default_branch TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT application_repos_unique UNIQUE (application_id, repo_id, subdir)
);

CREATE INDEX IF NOT EXISTS idx_application_repos_repo ON application_repos(repo_id);

DROP TRIGGER IF EXISTS trg_application_repos_set_updated_at ON application_repos;
CREATE TRIGGER trg_application_repos_set_updated_at
BEFORE UPDATE ON application_repos
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


