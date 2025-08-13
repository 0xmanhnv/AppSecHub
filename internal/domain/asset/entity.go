package asset

import (
	"errors"
	"time"
)

// AssetID is a strong-typed identifier for an Asset.
type AssetID string

// Asset represents any managed item in the inventory.
// For application-specific attributes, see ApplicationProfile.
type Asset struct {
	ID              AssetID
	Type            AssetType
	Environment     Environment
	Criticality     Criticality
	OwnerTeam       string
	FirstSeenAt     time.Time
	LastSeenAt      time.Time
	InternetExposed bool
	Tags            []string
	Meta            map[string]any
}

// NewAsset constructs an Asset and validates invariants.
func NewAsset(id AssetID, assetType AssetType, env Environment, crit Criticality, ownerTeam string, now time.Time) (Asset, error) {
	if err := validateAssetFields(assetType, env, crit); err != nil {
		return Asset{}, err
	}
	return Asset{
		ID:              id,
		Type:            assetType,
		Environment:     env,
		Criticality:     crit,
		OwnerTeam:       ownerTeam,
		FirstSeenAt:     now,
		LastSeenAt:      now,
		InternetExposed: false,
		Tags:            nil,
		Meta:            map[string]any{},
	}, nil
}

// TouchSeen updates the LastSeenAt timestamp, ensuring it never goes backwards.
func (a *Asset) TouchSeen(seenAt time.Time) {
	if seenAt.After(a.LastSeenAt) {
		a.LastSeenAt = seenAt
	}
}

func validateAssetFields(assetType AssetType, env Environment, crit Criticality) error {
	if !assetType.IsValid() {
		return ErrInvalidAssetType
	}
	if !env.IsValid() {
		return ErrInvalidEnvironment
	}
	if !crit.IsValid() {
		return ErrInvalidCriticality
	}
	return nil
}

var (
	// ErrInvalidAssetType indicates the provided AssetType is not supported.
	ErrInvalidAssetType = errors.New("invalid asset type")
	// ErrInvalidEnvironment indicates the provided Environment value is invalid.
	ErrInvalidEnvironment = errors.New("invalid environment")
	// ErrInvalidCriticality indicates the provided Criticality value is invalid.
	ErrInvalidCriticality = errors.New("invalid criticality")
)
