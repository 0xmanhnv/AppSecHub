package userusecase

import (
	"appsechub/internal/application/ports"
	domid "appsechub/internal/domain/identity"
)

func NewCreateUserUseCase(repo domid.Repository, hasher PasswordHasher) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo, hasher: hasher}
}

func NewLoginUserUseCase(repo domid.Repository, hasher PasswordHasher, jwt ports.TokenIssuer, store ports.RefreshTokenStore) *LoginUserUseCase {
	return &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt, store: store}
}
