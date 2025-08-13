package asset

import (
	"context"
	"errors"
	"time"

	domainasset "appsechub/internal/domain/asset"
)

// Repository abstracts persistence for assets and application profiles.
type Repository interface {
	UpsertAsset(ctx context.Context, entity domainasset.Asset) (domainasset.Asset, error)
	GetAssetByID(ctx context.Context, id domainasset.AssetID) (domainasset.Asset, error)
	ListAssets(ctx context.Context, filter ListAssetsFilter) ([]domainasset.Asset, error)
	TouchSeen(ctx context.Context, id domainasset.AssetID, seenAt time.Time) error

	UpsertApplicationProfile(ctx context.Context, profile domainasset.ApplicationProfile) (domainasset.ApplicationProfile, error)
	GetApplicationProfileByAssetID(ctx context.Context, id domainasset.AssetID) (domainasset.ApplicationProfile, error)
	LinkApplicationRepo(ctx context.Context, link domainasset.ApplicationRepoLink) error
	UnlinkApplicationRepo(ctx context.Context, applicationID domainasset.ApplicationID, repoID string, subdir string) error
}

// ListAssetsFilter supports simple filtering without leaking infrastructure details.
type ListAssetsFilter struct {
	Types        []domainasset.AssetType
	Environments []domainasset.Environment
	OwnerTeams   []string
	Tags         []string
	InternetOnly *bool
}

var (
	// ErrNotFound signals missing entities in repository.
	ErrNotFound = errors.New("not found")
)
