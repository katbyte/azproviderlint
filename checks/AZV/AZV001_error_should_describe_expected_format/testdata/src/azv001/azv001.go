package azv001

import "fmt"

// Should be flagged: unclear error message
func badError(name string) error {
	return fmt.Errorf("invalid format of %q", name) // want `unclear error message: describe the expected format instead of 'invalid format of'`
}

// Should be flagged: unclear error message in a plain string
func badErrorString(name string) error {
	msg := "invalid format of the value provided" // want `unclear error message: describe the expected format instead of 'invalid format of'`
	return fmt.Errorf("%s: %s", msg, name)
}

// Should NOT be flagged: descriptive error message
func goodError(name string) error {
	return fmt.Errorf("%q must start with a letter, may contain letters and numbers, and must end with a letter", name)
}

// Should NOT be flagged: 'invalid format' without 'of'
func goodErrorShort(name string) error {
	return fmt.Errorf("%q has an invalid format: expected `letter[letter|number]*letter`", name)
}

// Should NOT be flagged: non-string literals are skipped
const maxLength = 64

// Should be flagged: a raw string literal is inspected too
func badRawString(name string) error {
	return fmt.Errorf(`invalid format of %s`, name) // want `unclear error message: describe the expected format instead of 'invalid format of'`
}

// Should NOT be flagged: 'invalid format of' must be followed by a space
func goodSuffix(name string) error {
	return fmt.Errorf("%q: invalid format offset", name)
}
