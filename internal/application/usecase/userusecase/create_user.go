package userusecase

import (
	"context"

	"appsechub/internal/application/dto"
	domid "appsechub/internal/domain/identity"
)

type CreateUserUseCase struct {
	repo   domid.Repository
	hasher PasswordHasher
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
	emailVO, err := domid.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}
	role := domid.Role(input.Role)
	if !role.IsValid() {
		return nil, domid.ErrInvalidRole
	}
	// Enforce public registration only for non-admin users
	if role == domid.RoleAdmin {
		return nil, domid.ErrInvalidRole
	}
	hashed, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	newUser := domid.NewUser(input.FirstName, input.LastName, emailVO, hashed, role)
	if err := domid.ValidateUser(newUser); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, newUser); err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:        newUser.ID,
		Email:     newUser.Email.String(),
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		CreatedAt: newUser.CreatedAt,
	}, nil
}
