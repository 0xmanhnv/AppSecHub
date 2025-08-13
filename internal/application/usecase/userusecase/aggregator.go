package userusecase

import (
	"context"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/ports"
	domid "appsechub/internal/domain/identity"
)

// UserUsecases defines the application-facing interface for user-related operations.
type UserUsecases interface {
	Register(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error)
	GetMe(ctx context.Context, userID string) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordRequest) error
	Refresh(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	// LoginOIDC logs in an existing provisioned user by email (no password), issuing local tokens.
	LoginOIDC(ctx context.Context, email string) (*dto.LoginResponse, error)
}

// userUsecasesAggregator is a thin wrapper delegating to concrete use cases.
type userUsecasesAggregator struct {
	create   *CreateUserUseCase
	login    *LoginUserUseCase
	loginSSO *LoginOIDCUseCase
	getMe    *GetMeUseCase
	change   *ChangePasswordUseCase
	refresh  *RefreshUseCase
}

func NewUserUsecases(repo domid.Repository, hasher PasswordHasher, jwt ports.TokenIssuer) UserUsecases {
	return &userUsecasesAggregator{
		create:   NewCreateUserUseCase(repo, hasher),
		login:    NewLoginUserUseCase(repo, hasher, jwt, nil),
		loginSSO: NewLoginOIDCUseCase(repo, jwt, nil, 0),
		getMe:    NewGetMeUseCase(repo),
		change:   NewChangePasswordUseCase(repo, hasher),
		refresh:  NewRefreshUseCase(repo, jwt),
	}
}

func NewUserUsecasesWithStore(repo domid.Repository, hasher PasswordHasher, jwt ports.TokenIssuer, store ports.RefreshTokenStore, refreshTTLSeconds int) UserUsecases {
	return &userUsecasesAggregator{
		create:   NewCreateUserUseCase(repo, hasher),
		login:    &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt, store: store, refreshTTLSeconds: refreshTTLSeconds},
		loginSSO: NewLoginOIDCUseCase(repo, jwt, store, refreshTTLSeconds),
		getMe:    NewGetMeUseCase(repo),
		change:   NewChangePasswordUseCase(repo, hasher),
		refresh:  NewRefreshUseCaseWithStore(repo, jwt, store, refreshTTLSeconds),
	}
}

func (u *userUsecasesAggregator) Register(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
	return u.create.Execute(ctx, input)
}

func (u *userUsecasesAggregator) Login(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
	return u.login.Execute(ctx, input)
}

func (u *userUsecasesAggregator) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
	return u.getMe.Execute(ctx, userID)
}

func (u *userUsecasesAggregator) ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordRequest) error {
	return u.change.Execute(ctx, userID, input)
}

func (u *userUsecasesAggregator) Refresh(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	return u.refresh.Execute(ctx, refreshToken)
}

func (u *userUsecasesAggregator) Logout(ctx context.Context, refreshToken string) error {
	return u.refresh.Revoke(ctx, refreshToken)
}

func (u *userUsecasesAggregator) LoginOIDC(ctx context.Context, email string) (*dto.LoginResponse, error) {
	return u.loginSSO.Execute(ctx, email)
}
