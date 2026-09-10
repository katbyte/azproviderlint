// Package commonids is a minimal stand-in for go-azure-helpers' commonids used only by the
// AZG008 analysistest fixtures.
package commonids

// CompositeResourceID pairs two resource IDs; both fields are always populated.
type CompositeResourceID[T1 any, T2 any] struct {
	First  *T1
	Second *T2
}

// NewCompositeResourceID builds a composite ID from two IDs.
func NewCompositeResourceID[T1 any, T2 any](first *T1, second *T2) CompositeResourceID[T1, T2] {
	return CompositeResourceID[T1, T2]{First: first, Second: second}
}

// ParseCompositeResourceID parses input into the two IDs.
func ParseCompositeResourceID[T1 any, T2 any](input string, first *T1, second *T2) (*CompositeResourceID[T1, T2], error) {
	return &CompositeResourceID[T1, T2]{First: first, Second: second}, nil
}
