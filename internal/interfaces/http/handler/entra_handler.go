package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"appsechub/internal/application/usecase/userusecase"
	"appsechub/internal/config"
	"appsechub/internal/infras/oauth"
	resp "appsechub/internal/interfaces/http/response"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type EntraHandler struct {
	cfg        *config.Config
	uc         userusecase.UserUsecases
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	oauth2conf *oauth2.Config
	stateStore oauth.StateStore
}

func NewEntraHandler(ctx context.Context, cfg *config.Config, uc userusecase.UserUsecases, stateStore oauth.StateStore) (*EntraHandler, error) {
	tenant := cfg.Entra.TenantID
	issuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenant)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	oauth2conf := &oauth2.Config{
		ClientID:     cfg.Entra.ClientID,
		ClientSecret: cfg.Entra.ClientSecret,
		RedirectURL:  cfg.Entra.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Entra.Scopes,
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.Entra.ClientID})
	if stateStore == nil {
		stateStore = oauth.NewInMemoryStateStore()
	}
	return &EntraHandler{cfg: cfg, uc: uc, provider: provider, verifier: verifier, oauth2conf: oauth2conf, stateStore: stateStore}, nil
}

func (h *EntraHandler) Login(c *gin.Context) {
	state := randomB64(24)
	codeVerifier := randomB64(48)
	nonce := randomB64(24)
	h.stateStore.Save(state, oauth.StateData{CodeVerifier: codeVerifier, Nonce: nonce, ExpiresAt: time.Now().Add(10 * time.Minute)})

	// PKCE (S256) parameters
	// For simplicity, send code_challenge as plain (not S256) for now; can upgrade to S256 if we compute challenge
	authCodeURL := h.oauth2conf.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce), oauth2.SetAuthURLParam("code_challenge_method", "plain"), oauth2.SetAuthURLParam("code_challenge", codeVerifier))
	c.Redirect(http.StatusFound, authCodeURL)
}

func (h *EntraHandler) Callback(c *gin.Context) {
	ctx := c.Request.Context()
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		resp.BadRequest(c, resp.CodeInvalidRequest, "missing state or code")
		return
	}
	sd, ok := h.stateStore.GetAndDelete(state)
	if !ok {
		resp.BadRequest(c, resp.CodeInvalidRequest, "invalid state")
		return
	}
	token, err := h.oauth2conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", sd.CodeVerifier))
	if err != nil {
		resp.Unauthorized(c, resp.CodeUnauthorized, "oauth exchange failed")
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		resp.Unauthorized(c, resp.CodeUnauthorized, "missing id_token")
		return
	}
	idt, err := h.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		resp.Unauthorized(c, resp.CodeUnauthorized, "invalid id_token")
		return
	}
	// Validate nonce
	var claims struct {
		Email      string `json:"email"`
		Nonce      string `json:"nonce"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		OID        string `json:"oid"`
		TID        string `json:"tid"`
	}
	if err := idt.Claims(&claims); err != nil {
		resp.Unauthorized(c, resp.CodeUnauthorized, "invalid claims")
		return
	}
	if claims.Nonce == "" || claims.Nonce != sd.Nonce {
		resp.Unauthorized(c, resp.CodeUnauthorized, "nonce mismatch")
		return
	}

	// Map to local user (simplified): require email claim
	email := claims.Email
	if email == "" {
		resp.BadRequest(c, resp.CodeInvalidRequest, "email not provided by provider")
		return
	}
	// Issue local JWT via usecase login-like path (no password). Require user pre-provisioned.
	lr, err := h.uc.LoginOIDC(ctx, email)
	if err != nil {
		resp.Unauthorized(c, resp.CodeUnauthorized, "token issue failed")
		return
	}
	resp.OK(c, lr)
}

func randomB64(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
