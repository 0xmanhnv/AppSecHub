package response

import (
	"errors"

	"appsechub/internal/application/apperr"
	domid "appsechub/internal/domain/identity"
)

// FromError maps a domain/application error to HTTP status, code and safe message.
func FromError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, apperr.ErrInvalidCredentials):
		return 401, CodeInvalidCredentials, MsgInvalidCredentials
	case errors.Is(err, apperr.ErrInvalidRefreshToken):
		return 401, CodeInvalidRefreshToken, MsgInvalidRefreshToken
	case errors.Is(err, domid.ErrUserNotFound):
		return 404, CodeNotFound, MsgNotFound
	case errors.Is(err, domid.ErrEmailAlreadyExists):
		return 409, CodeConflict, "email already exists"
	case errors.Is(err, domid.ErrInvalidEmail), errors.Is(err, domid.ErrInvalidRole):
		return 400, CodeInvalidRequest, "invalid request"
	default:
		return 500, CodeServerError, MsgServerError
	}
}
