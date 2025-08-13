package asset

// AssetType enumerates supported asset categories.
type AssetType string

const (
	AssetTypeDomain      AssetType = "domain"
	AssetTypeHost        AssetType = "host"
	AssetTypeIP          AssetType = "ip"
	AssetTypeService     AssetType = "service"
	AssetTypeApplication AssetType = "application"
)

// IsValid returns true if the AssetType is recognized.
func (t AssetType) IsValid() bool {
	switch t {
	case AssetTypeDomain, AssetTypeHost, AssetTypeIP, AssetTypeService, AssetTypeApplication:
		return true
	default:
		return false
	}
}

// Environment enumerates deployment environments.
type Environment string

const (
	EnvironmentProd Environment = "prod"
	EnvironmentStg  Environment = "stg"
	EnvironmentDev  Environment = "dev"
	EnvironmentTest Environment = "test"
)

// IsValid returns true if the Environment is recognized.
func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentProd, EnvironmentStg, EnvironmentDev, EnvironmentTest:
		return true
	default:
		return false
	}
}

// Criticality enumerates business impact tiers.
type Criticality string

const (
	CriticalityTier1 Criticality = "tier1"
	CriticalityTier2 Criticality = "tier2"
	CriticalityTier3 Criticality = "tier3"
)

// IsValid returns true if the Criticality is recognized.
func (c Criticality) IsValid() bool {
	switch c {
	case CriticalityTier1, CriticalityTier2, CriticalityTier3:
		return true
	default:
		return false
	}
}
