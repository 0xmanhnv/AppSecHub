package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appports "appsechub/internal/application/ports/asset"
	domain "appsechub/internal/domain/asset"
	pstore "appsechub/internal/infras/storage/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AssetRepository struct {
	q *pstore.Queries
}

func NewAssetRepository(q *pstore.Queries) *AssetRepository {
	return &AssetRepository{q: q}
}

func (r *AssetRepository) UpsertAsset(ctx context.Context, entity domain.Asset) (domain.Asset, error) {
	uid, err := uuid.Parse(string(entity.ID))
	if err != nil {
		return domain.Asset{}, fmt.Errorf("invalid asset id: %w", err)
	}
	metaBytes, err := json.Marshal(entity.Meta)
	if err != nil {
		return domain.Asset{}, fmt.Errorf("marshal meta: %w", err)
	}
	ownerText := pgtype.Text{String: entity.OwnerTeam, Valid: entity.OwnerTeam != ""}
	err = r.q.UpsertAsset(ctx, pstore.UpsertAssetParams{
		ID:              uid,
		Type:            string(entity.Type),
		Environment:     string(entity.Environment),
		Criticality:     string(entity.Criticality),
		OwnerTeam:       ownerText,
		FirstSeen:       entity.FirstSeenAt,
		LastSeen:        entity.LastSeenAt,
		InternetExposed: entity.InternetExposed,
		Tags:            entity.Tags,
		Meta:            metaBytes,
	})
	if err != nil {
		return domain.Asset{}, err
	}
	// Read back for created/updated timestamps
	row, err := r.q.GetAssetByID(ctx, uid)
	if err != nil {
		return domain.Asset{}, err
	}
	var meta map[string]any
	_ = json.Unmarshal(row.Meta, &meta)
	return domain.Asset{
		ID:              domain.AssetID(row.ID.String()),
		Type:            domain.AssetType(row.Type),
		Environment:     domain.Environment(row.Environment),
		Criticality:     domain.Criticality(row.Criticality),
		OwnerTeam:       row.OwnerTeam.String,
		FirstSeenAt:     row.FirstSeen,
		LastSeenAt:      row.LastSeen,
		InternetExposed: row.InternetExposed,
		Tags:            row.Tags,
		Meta:            meta,
	}, nil
}

func (r *AssetRepository) GetAssetByID(ctx context.Context, id domain.AssetID) (domain.Asset, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return domain.Asset{}, fmt.Errorf("invalid asset id: %w", err)
	}
	row, err := r.q.GetAssetByID(ctx, uid)
	if err != nil {
		return domain.Asset{}, err
	}
	var meta map[string]any
	_ = json.Unmarshal(row.Meta, &meta)
	return domain.Asset{
		ID:              domain.AssetID(row.ID.String()),
		Type:            domain.AssetType(row.Type),
		Environment:     domain.Environment(row.Environment),
		Criticality:     domain.Criticality(row.Criticality),
		OwnerTeam:       row.OwnerTeam.String,
		FirstSeenAt:     row.FirstSeen,
		LastSeenAt:      row.LastSeen,
		InternetExposed: row.InternetExposed,
		Tags:            row.Tags,
		Meta:            meta,
	}, nil
}

func (r *AssetRepository) ListAssets(ctx context.Context, _ appports.ListAssetsFilter) ([]domain.Asset, error) {
	rows, err := r.q.ListAssets(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Asset, 0, len(rows))
	for _, row := range rows {
		var meta map[string]any
		_ = json.Unmarshal(row.Meta, &meta)
		out = append(out, domain.Asset{
			ID:              domain.AssetID(row.ID.String()),
			Type:            domain.AssetType(row.Type),
			Environment:     domain.Environment(row.Environment),
			Criticality:     domain.Criticality(row.Criticality),
			OwnerTeam:       row.OwnerTeam.String,
			FirstSeenAt:     row.FirstSeen,
			LastSeenAt:      row.LastSeen,
			InternetExposed: row.InternetExposed,
			Tags:            row.Tags,
			Meta:            meta,
		})
	}
	return out, nil
}

func (r *AssetRepository) TouchSeen(ctx context.Context, id domain.AssetID, seenAt time.Time) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return fmt.Errorf("invalid asset id: %w", err)
	}
	return r.q.TouchSeen(ctx, pstore.TouchSeenParams{ID: uid, LastSeen: seenAt})
}

func (r *AssetRepository) UpsertApplicationProfile(ctx context.Context, profile domain.ApplicationProfile) (domain.ApplicationProfile, error) {
	pid, err := uuid.Parse(string(profile.ID))
	if err != nil {
		return domain.ApplicationProfile{}, fmt.Errorf("invalid application id: %w", err)
	}
	aid, err := uuid.Parse(string(profile.AssetID))
	if err != nil {
		return domain.ApplicationProfile{}, fmt.Errorf("invalid asset id: %w", err)
	}
	metaBytes, err := json.Marshal(profile.Meta)
	if err != nil {
		return domain.ApplicationProfile{}, fmt.Errorf("marshal meta: %w", err)
	}
	err = r.q.UpsertApplicationProfile(ctx, pstore.UpsertApplicationProfileParams{
		ID:          pid,
		AssetID:     aid,
		Slug:        profile.Slug,
		DisplayName: pgtype.Text{String: profile.DisplayName, Valid: profile.DisplayName != ""},
		ProjectName: pgtype.Text{String: profile.ProjectName, Valid: profile.ProjectName != ""},
		OwnerTeam:   pgtype.Text{String: profile.OwnerTeam, Valid: profile.OwnerTeam != ""},
		Language:    pgtype.Text{String: profile.Language, Valid: profile.Language != ""},
		Framework:   pgtype.Text{String: profile.Framework, Valid: profile.Framework != ""},
		Runtime:     pgtype.Text{String: profile.Runtime, Valid: profile.Runtime != ""},
		Tags:        profile.Tags,
		Meta:        metaBytes,
	})
	if err != nil {
		return domain.ApplicationProfile{}, err
	}
	row, err := r.q.GetApplicationProfileByAssetID(ctx, aid)
	if err != nil {
		return domain.ApplicationProfile{}, err
	}
	var meta map[string]any
	_ = json.Unmarshal(row.Meta, &meta)
	return domain.ApplicationProfile{
		ID:          domain.ApplicationID(row.ID.String()),
		AssetID:     domain.AssetID(row.AssetID.String()),
		Slug:        row.Slug,
		DisplayName: row.DisplayName.String,
		ProjectName: row.ProjectName.String,
		OwnerTeam:   row.OwnerTeam.String,
		Language:    row.Language.String,
		Framework:   row.Framework.String,
		Runtime:     row.Runtime.String,
		Tags:        row.Tags,
		Meta:        meta,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *AssetRepository) GetApplicationProfileByAssetID(ctx context.Context, id domain.AssetID) (domain.ApplicationProfile, error) {
	aid, err := uuid.Parse(string(id))
	if err != nil {
		return domain.ApplicationProfile{}, fmt.Errorf("invalid asset id: %w", err)
	}
	row, err := r.q.GetApplicationProfileByAssetID(ctx, aid)
	if err != nil {
		return domain.ApplicationProfile{}, err
	}
	var meta map[string]any
	_ = json.Unmarshal(row.Meta, &meta)
	return domain.ApplicationProfile{
		ID:          domain.ApplicationID(row.ID.String()),
		AssetID:     domain.AssetID(row.AssetID.String()),
		Slug:        row.Slug,
		DisplayName: row.DisplayName.String,
		ProjectName: row.ProjectName.String,
		OwnerTeam:   row.OwnerTeam.String,
		Language:    row.Language.String,
		Framework:   row.Framework.String,
		Runtime:     row.Runtime.String,
		Tags:        row.Tags,
		Meta:        meta,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *AssetRepository) LinkApplicationRepo(ctx context.Context, link domain.ApplicationRepoLink) error {
	aid, err := uuid.Parse(string(link.ApplicationID))
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	return r.q.LinkApplicationRepo(ctx, pstore.LinkApplicationRepoParams{
		ApplicationID: aid,
		RepoID:        link.RepoID,
		Subdir:        pgtype.Text{String: link.Subdir, Valid: link.Subdir != ""},
		BuildFile:     pgtype.Text{String: link.BuildFile, Valid: link.BuildFile != ""},
		DefaultBranch: pgtype.Text{String: link.DefaultBranch, Valid: link.DefaultBranch != ""},
	})
}

func (r *AssetRepository) UnlinkApplicationRepo(ctx context.Context, appID domain.ApplicationID, repoID string, subdir string) error {
	aid, err := uuid.Parse(string(appID))
	if err != nil {
		return fmt.Errorf("invalid application id: %w", err)
	}
	var subdirParam any
	if subdir != "" {
		subdirParam = pgtype.Text{String: subdir, Valid: true}
	} else {
		subdirParam = nil
	}
	return r.q.UnlinkApplicationRepo(ctx, pstore.UnlinkApplicationRepoParams{
		ApplicationID: aid,
		RepoID:        repoID,
		Column3:       subdirParam,
	})
}
