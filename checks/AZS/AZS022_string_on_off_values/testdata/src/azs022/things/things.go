// Package things is a minimal stub of a generated SDK package, with its enums and the helpers that list their values.
package things

type PublicAccess string

const (
	PublicAccessDisabled PublicAccess = "Disabled"
	PublicAccessEnabled  PublicAccess = "Enabled"
)

func PossibleValuesForPublicAccess() []string {
	return []string{string(PublicAccessDisabled), string(PublicAccessEnabled)}
}

type Tier string

const (
	TierBasic    Tier = "Basic"
	TierPremium  Tier = "Premium"
	TierStandard Tier = "Standard"
)

func PossibleValuesForTier() []string {
	return []string{string(TierBasic), string(TierPremium), string(TierStandard)}
}

// the older SDKs name the helper the other way round, and return the enum's own type
type Logging string

const (
	LoggingOff Logging = "Off"
	LoggingOn  Logging = "On"
)

func PossibleLoggingValues() []Logging { return []Logging{LoggingOff, LoggingOn} }

// a function that is not a helper for an enum
func PossibleRegions() []string { return []string{"Enabled", "Disabled"} }
