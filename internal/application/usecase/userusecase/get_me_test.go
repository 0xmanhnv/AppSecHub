package userusecase

import (
	"context"
	"testing"
	"time"

	domid "appsechub/internal/domain/identity"

	"github.com/google/uuid"
)

type repoOne struct{ u *domid.User }

func (r repoOne) Save(ctx context.Context, u *domid.User) error                  { return nil }
func (r repoOne) GetByID(ctx context.Context, id uuid.UUID) (*domid.User, error) { return r.u, nil }
func (r repoOne) GetByEmail(ctx context.Context, email domid.Email) (*domid.User, error) {
	return r.u, nil
}
func (r repoOne) GetAll(ctx context.Context) ([]*domid.User, error) {
	return []*domid.User{r.u}, nil
}
func (r repoOne) Update(ctx context.Context, u *domid.User) error { return nil }
func (r repoOne) Delete(ctx context.Context, id uuid.UUID) error  { return nil }

func TestGetMe_Success(t *testing.T) {
	uid := uuid.New()
	u := &domid.User{ID: uid, FirstName: "A", LastName: "B", Email: domid.Email("a@b.com"), CreatedAt: time.Now()}
	usecase := NewGetMeUseCase(repoOne{u: u})
	got, err := usecase.Execute(context.Background(), uid.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != uid {
		t.Fatalf("expected %s, got %s", uid, got.ID)
	}
}
