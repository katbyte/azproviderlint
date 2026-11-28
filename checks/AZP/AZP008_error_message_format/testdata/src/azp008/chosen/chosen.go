// Package chosen is checked with errors printed with %w and values put in quotes with %q.
package chosen

import (
	"errors"
	"fmt"
)

func wording(name string, err error) error {
	_ = fmt.Errorf("creating %s: %+v", name, err)  // want `error is printed with %\+v - use %w`
	_ = fmt.Errorf("creating %s: %s", name, err)   // want `error is printed with %s - use %w`
	_ = fmt.Errorf("creating `%s`: %w", name, err) // want "error message puts a value in backticks - use %q"
	_ = fmt.Errorf("expanding: +%v", err)          // want `error message has \+%v, with the plus on the wrong side - use %w`
	_ = fmt.Errorf("expanding: +%v", name)         // want `error message has \+%v, with the plus on the wrong side - use %\+v`

	return nil
}

func fine(name string, count int, err error) error {
	_ = fmt.Errorf("creating %q: %w", name, err)

	// what is not text is left in its backticks, and so is a name written in them
	_ = fmt.Errorf("expected `%d`", count)
	_ = errors.New("`zones` cannot be set")

	return nil
}
