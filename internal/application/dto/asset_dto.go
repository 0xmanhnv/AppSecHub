package dto

import (
	domain "appsechub/internal/domain/asset"
	"time"
)

type UpsertAssetRequest struct {
	ID              string             `json:"id"`
	Type            domain.AssetType   `json:"type"`
	Environment     domain.Environment `json:"environment"`
	Criticality     domain.Criticality `json:"criticality"`
	OwnerTeam       string             `json:"owner_team"`
	InternetExposed bool               `json:"internet_exposed"`
	Tags            []string           `json:"tags"`
	Meta            map[string]any     `json:"meta"`
}

type AssetResponse struct {
	ID              string             `json:"id"`
	Type            domain.AssetType   `json:"type"`
	Environment     domain.Environment `json:"environment"`
	Criticality     domain.Criticality `json:"criticality"`
	OwnerTeam       string             `json:"owner_team"`
	FirstSeenAt     time.Time          `json:"first_seen_at"`
	LastSeenAt      time.Time          `json:"last_seen_at"`
	InternetExposed bool               `json:"internet_exposed"`
	Tags            []string           `json:"tags"`
	Meta            map[string]any     `json:"meta"`
}

type UpsertApplicationRequest struct {
	ID          string         `json:"id"`
	AssetID     string         `json:"asset_id"`
	Slug        string         `json:"slug"`
	DisplayName string         `json:"display_name"`
	ProjectName string         `json:"project_name"`
	OwnerTeam   string         `json:"owner_team"`
	Language    string         `json:"language"`
	Framework   string         `json:"framework"`
	Runtime     string         `json:"runtime"`
	Tags        []string       `json:"tags"`
	Meta        map[string]any `json:"meta"`
}

type ApplicationResponse struct {
	ID          string         `json:"id"`
	AssetID     string         `json:"asset_id"`
	Slug        string         `json:"slug"`
	DisplayName string         `json:"display_name"`
	ProjectName string         `json:"project_name"`
	OwnerTeam   string         `json:"owner_team"`
	Language    string         `json:"language"`
	Framework   string         `json:"framework"`
	Runtime     string         `json:"runtime"`
	Tags        []string       `json:"tags"`
	Meta        map[string]any `json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type LinkApplicationRepoRequest struct {
	RepoID        string `json:"repo_id"`
	Subdir        string `json:"subdir"`
	BuildFile     string `json:"build_file"`
	DefaultBranch string `json:"default_branch"`
}
