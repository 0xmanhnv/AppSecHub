package response

import (
	"errors"
	"testing"

	"appsechub/internal/application/apperr"
	domid "appsechub/internal/domain/identity"
)

func TestFromError_Mapping(t *testing.T) {
	cases := []struct {
		in         error
		wantStatus int
	}{
		{apperr.ErrInvalidCredentials, 401},
		{apperr.ErrInvalidRefreshToken, 401},
		{domid.ErrUserNotFound, 404},
		{domid.ErrEmailAlreadyExists, 409},
		{errors.New("x"), 500},
	}
	for _, c := range cases {
		got, _, _ := FromError(c.in)
		if got != c.wantStatus {
			t.Fatalf("status = %d, want %d", got, c.wantStatus)
		}
	}
}
