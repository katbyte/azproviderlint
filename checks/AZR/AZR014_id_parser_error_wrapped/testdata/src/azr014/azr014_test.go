package azr014

import (
	"fmt"

	"azr014/things"
)

// an error made in a test is not read by a user
func parse(input string) error {
	if _, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("parsing %q: %+v", input, err)
	}

	return nil
}
