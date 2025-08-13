package userusecase

import (
	"context"

	"appsechub/internal/application/dto"
	domid "appsechub/internal/domain/identity"

	"github.com/google/uuid"
)

type GetMeUseCase struct{ repo domid.Repository }

func NewGetMeUseCase(repo domid.Repository) *GetMeUseCase { return &GetMeUseCase{repo: repo} }

func (uc *GetMeUseCase) Execute(ctx context.Context, userID string) (*dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domid.ErrInvalidID
	}
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	}, nil
}
