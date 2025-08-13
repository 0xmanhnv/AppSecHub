package userusecase

import (
	"context"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/ports"
	domid "appsechub/internal/domain/identity"
)

// LoginOIDCUseCase issues local tokens for an existing user identified by email (no password).
type LoginOIDCUseCase struct {
	repo              domid.Repository
	jwt               ports.TokenIssuer
	store             ports.RefreshTokenStore
	refreshTTLSeconds int
}

func NewLoginOIDCUseCase(repo domid.Repository, jwt ports.TokenIssuer, store ports.RefreshTokenStore, refreshTTLSeconds int) *LoginOIDCUseCase {
	return &LoginOIDCUseCase{repo: repo, jwt: jwt, store: store, refreshTTLSeconds: refreshTTLSeconds}
}

func (uc *LoginOIDCUseCase) Execute(ctx context.Context, email string) (*dto.LoginResponse, error) {
	emailVO, err := domid.NewEmail(email)
	if err != nil {
		return nil, err
	}
	u, err := uc.repo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	token, err := uc.jwt.GenerateToken(u.ID.String(), string(u.Role))
	if err != nil {
		return nil, err
	}
	var refresh string
	if uc.store != nil {
		ttl := uc.refreshTTLSeconds
		if ttl <= 0 {
			ttl = 3600 * 24 * 7
		}
		refresh, _ = uc.store.Issue(ctx, u.ID.String(), ttl)
	}
	return &dto.LoginResponse{
		AccessToken:  token,
		RefreshToken: refresh,
		User:         dto.UserResponse{ID: u.ID, Email: u.Email.String(), FirstName: u.FirstName, LastName: u.LastName, CreatedAt: u.CreatedAt},
	}, nil
}
