// Package defaults is checked with no options set, when a value in quotes is not reported.
package defaults

import "fmt"

func wording(name string, err error) error {
	_ = fmt.Errorf("retrieving %q: %+v", name, err)
	_ = fmt.Errorf("retrieving %q: %s", name, err) // want `error is printed with %s`

	return nil
}
