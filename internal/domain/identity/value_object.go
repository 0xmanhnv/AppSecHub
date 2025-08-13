package identity

import (
	appval "appsechub/pkg/validator"
	"strings"
)

// Email VO
type Email string

func (e Email) String() string { return string(e) }
func (e Email) IsValid() bool  { return appval.IsValidEmail(string(e)) }
func NewEmail(s string) (Email, error) {
	email := Email(strings.TrimSpace(s))
	if !email.IsValid() {
		return "", ErrInvalidEmail
	}
	return email, nil
}

// Role VO
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleViewer:
		return true
	default:
		return false
	}
}
