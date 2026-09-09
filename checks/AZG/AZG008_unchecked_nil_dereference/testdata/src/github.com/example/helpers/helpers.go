// Package helpers is a stand-in for the provider's expand helpers used by the AZG008
// analysistest fixtures: functions whose non-nil results are learned through facts.
package helpers

// ExpandStringSlice always returns a pointer to a (possibly empty) slice.
func ExpandStringSlice(input []interface{}) *[]string {
	result := make([]string, 0)
	for _, item := range input {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return &result
}

// ExpandMaybe returns nil for empty input.
func ExpandMaybe(input []interface{}) *[]string {
	if len(input) == 0 {
		return nil
	}
	return ExpandStringSlice(input)
}

// Parse always returns a value; the error is independent of it.
func Parse(input string) (*int, error) {
	n := len(input)
	return &n, nil
}
