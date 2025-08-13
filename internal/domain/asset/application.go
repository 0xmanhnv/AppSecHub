package asset

import "time"

// ApplicationID is a strong-typed identifier for an application profile.
type ApplicationID string

// ApplicationProfile stores attributes specific to application assets.
// Invariant: one profile per Asset with Type=application (1-1 mapping).
type ApplicationProfile struct {
	ID          ApplicationID
	AssetID     AssetID
	Slug        string
	DisplayName string
	ProjectName string
	OwnerTeam   string
	Language    string
	Framework   string
	Runtime     string
	Tags        []string
	Meta        map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ApplicationRepoLink links an application to a SCM repository.
// Supports monorepo via optional Subdir.
type ApplicationRepoLink struct {
	ApplicationID ApplicationID
	RepoID        string
	Subdir        string
	BuildFile     string
	DefaultBranch string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
