// Package off is checked with every kind of wording switched off.
package off

import (
	"errors"
	"fmt"
)

func wording(name string, err error) error {
	_ = fmt.Errorf("creating %q: %s", name, err)
	_ = errors.New("`zones` can't be set")

	return nil
}
