// Package things is a minimal stub of an SDK package with a resource ID and its parsers, for the analyzer tests.
package things

type ThingId struct{}

func (ThingId) ID() string { return "" }

func ParseThingID(input string) (*ThingId, error) { return &ThingId{}, nil }

func ParseThingIDInsensitively(input string) (*ThingId, error) { return &ThingId{}, nil }

// Parsed has no ID to write out again, so what reads it is not an ID parser.
type Parsed struct{}

func ParseAzureResourceID(input string) (*Parsed, error) { return &Parsed{}, nil }

func ValidateThingID(input string) (bool, error) { return true, nil }
