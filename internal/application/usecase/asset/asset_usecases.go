package asset

import (
	"context"
	"time"

	ports "appsechub/internal/application/ports/asset"
	domain "appsechub/internal/domain/asset"
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service { return &Service{repo: repo} }

// UpsertAsset creates or updates an asset while enforcing domain rules.
func (s *Service) UpsertAsset(ctx context.Context, entity domain.Asset) (domain.Asset, error) {
	return s.repo.UpsertAsset(ctx, entity)
}

// ListAssets retrieves assets matching filter.
func (s *Service) ListAssets(ctx context.Context, filter ports.ListAssetsFilter) ([]domain.Asset, error) {
	return s.repo.ListAssets(ctx, filter)
}

// TouchSeen updates last-seen timestamp idempotently.
func (s *Service) TouchSeen(ctx context.Context, id domain.AssetID, seenAt time.Time) error {
	return s.repo.TouchSeen(ctx, id, seenAt)
}

// UpsertApplication creates/updates application profile for an application asset.
func (s *Service) UpsertApplication(ctx context.Context, profile domain.ApplicationProfile) (domain.ApplicationProfile, error) {
	return s.repo.UpsertApplicationProfile(ctx, profile)
}

// LinkApplicationRepo links an application to a repository (monorepo supported via subdir).
func (s *Service) LinkApplicationRepo(ctx context.Context, link domain.ApplicationRepoLink) error {
	return s.repo.LinkApplicationRepo(ctx, link)
}

// UnlinkApplicationRepo removes the link between an application and a repository.
func (s *Service) UnlinkApplicationRepo(ctx context.Context, appID domain.ApplicationID, repoID string, subdir string) error {
	return s.repo.UnlinkApplicationRepo(ctx, appID, repoID, subdir)
}
