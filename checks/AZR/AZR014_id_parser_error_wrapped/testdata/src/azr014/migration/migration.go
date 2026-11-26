// Package migration reads IDs as they used to be, and says so.
package migration

import (
	"fmt"

	"azr014/things"
)

func upgrade(input string) error {
	if _, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("parsing the old ID %q: %+v", input, err)
	}

	return nil
}
