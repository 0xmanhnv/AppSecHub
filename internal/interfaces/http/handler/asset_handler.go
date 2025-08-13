package handler

import (
	"context"
	"net/http"
	"time"

	"appsechub/internal/application/dto"
	ports "appsechub/internal/application/ports/asset"
	domain "appsechub/internal/domain/asset"
	"appsechub/internal/interfaces/http/middleware"
	"appsechub/internal/interfaces/http/response"
	"appsechub/internal/interfaces/http/validation"

	"github.com/gin-gonic/gin"
)

// AssetService defines the minimal use cases that the handler needs.
type AssetService interface {
	UpsertAsset(ctx context.Context, entity domain.Asset) (domain.Asset, error)
	ListAssets(ctx context.Context, filter ports.ListAssetsFilter) ([]domain.Asset, error)
	UpsertApplication(ctx context.Context, profile domain.ApplicationProfile) (domain.ApplicationProfile, error)
	LinkApplicationRepo(ctx context.Context, link domain.ApplicationRepoLink) error
}

// PortsFilter mirrors ports.ListAssetsFilter without importing ports into handler signature.
type PortsFilter struct {
	Types        []domain.AssetType
	Environments []domain.Environment
	OwnerTeams   []string
	Tags         []string
	InternetOnly *bool
}

type AssetHandler struct {
	svc AssetService
}

func NewAssetHandler(svc AssetService) *AssetHandler { return &AssetHandler{svc: svc} }

func (h *AssetHandler) UpsertAsset(c *gin.Context) {
	var req dto.UpsertAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := validation.MapBindJSONErrorWithLocale(middleware.GetLocale(c), err)
		response.BadRequest(c, code, msg)
		return
	}
	now := time.Now()
	a, err := domain.NewAsset(domain.AssetID(req.ID), req.Type, req.Environment, req.Criticality, req.OwnerTeam, now)
	if err != nil {
		response.BadRequest(c, "invalid_request", err.Error())
		return
	}
	a.InternetExposed = req.InternetExposed
	a.Tags = req.Tags
	a.Meta = req.Meta
	out, err := h.svc.UpsertAsset(c.Request.Context(), a)
	if err != nil {
		response.InternalError(c, "server_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.AssetResponse{
		ID:              string(out.ID),
		Type:            out.Type,
		Environment:     out.Environment,
		Criticality:     out.Criticality,
		OwnerTeam:       out.OwnerTeam,
		FirstSeenAt:     out.FirstSeenAt,
		LastSeenAt:      out.LastSeenAt,
		InternetExposed: out.InternetExposed,
		Tags:            out.Tags,
		Meta:            out.Meta,
	})
}

func (h *AssetHandler) ListAssets(c *gin.Context) {
	// Parse basic filters from query params
	q := c.Request.URL.Query()
	var f ports.ListAssetsFilter
	if v := q["type"]; len(v) > 0 {
		for _, s := range v {
			f.Types = append(f.Types, domain.AssetType(s))
		}
	}
	if v := q["env"]; len(v) > 0 {
		for _, s := range v {
			f.Environments = append(f.Environments, domain.Environment(s))
		}
	}
	if v := q["owner"]; len(v) > 0 {
		f.OwnerTeams = append(f.OwnerTeams, v...)
	}
	if v := q["tag"]; len(v) > 0 {
		f.Tags = append(f.Tags, v...)
	}
	if v := q.Get("internet_only"); v == "true" {
		b := true
		f.InternetOnly = &b
	}
	items, err := h.svc.ListAssets(c.Request.Context(), f)
	if err != nil {
		response.InternalError(c, "server_error", err.Error())
		return
	}
	out := make([]dto.AssetResponse, 0, len(items))
	for _, a := range items {
		out = append(out, dto.AssetResponse{
			ID:              string(a.ID),
			Type:            a.Type,
			Environment:     a.Environment,
			Criticality:     a.Criticality,
			OwnerTeam:       a.OwnerTeam,
			FirstSeenAt:     a.FirstSeenAt,
			LastSeenAt:      a.LastSeenAt,
			InternetExposed: a.InternetExposed,
			Tags:            a.Tags,
			Meta:            a.Meta,
		})
	}
	response.OK(c, out)
}

func (h *AssetHandler) UpsertApplication(c *gin.Context) {
	var req dto.UpsertApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := validation.MapBindJSONErrorWithLocale(middleware.GetLocale(c), err)
		response.BadRequest(c, code, msg)
		return
	}
	ap := domain.ApplicationProfile{
		ID:          domain.ApplicationID(req.ID),
		AssetID:     domain.AssetID(req.AssetID),
		Slug:        req.Slug,
		DisplayName: req.DisplayName,
		ProjectName: req.ProjectName,
		OwnerTeam:   req.OwnerTeam,
		Language:    req.Language,
		Framework:   req.Framework,
		Runtime:     req.Runtime,
		Tags:        req.Tags,
		Meta:        req.Meta,
	}
	out, err := h.svc.UpsertApplication(c.Request.Context(), ap)
	if err != nil {
		response.InternalError(c, "server_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.ApplicationResponse{
		ID:          string(out.ID),
		AssetID:     string(out.AssetID),
		Slug:        out.Slug,
		DisplayName: out.DisplayName,
		ProjectName: out.ProjectName,
		OwnerTeam:   out.OwnerTeam,
		Language:    out.Language,
		Framework:   out.Framework,
		Runtime:     out.Runtime,
		Tags:        out.Tags,
		Meta:        out.Meta,
		CreatedAt:   out.CreatedAt,
		UpdatedAt:   out.UpdatedAt,
	})
}

func (h *AssetHandler) LinkApplicationRepo(c *gin.Context) {
	applicationID := c.Param("id")
	var req dto.LinkApplicationRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := validation.MapBindJSONErrorWithLocale(middleware.GetLocale(c), err)
		response.BadRequest(c, code, msg)
		return
	}
	link := domain.ApplicationRepoLink{
		ApplicationID: domain.ApplicationID(applicationID),
		RepoID:        req.RepoID,
		Subdir:        req.Subdir,
		BuildFile:     req.BuildFile,
		DefaultBranch: req.DefaultBranch,
	}
	if err := h.svc.LinkApplicationRepo(c.Request.Context(), link); err != nil {
		response.InternalError(c, "server_error", err.Error())
		return
	}
	response.OK(c, gin.H{"linked": true})
}
