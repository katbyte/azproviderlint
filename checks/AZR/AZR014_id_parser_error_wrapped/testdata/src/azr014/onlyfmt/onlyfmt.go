// Package onlyfmt uses fmt for nothing but the wrapping, so the last fix takes the import with it.
package onlyfmt

import (
	"fmt"

	"azr014/things"
)

func read(input string) error {
	if _, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("parsing %q: %+v", input, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}

func flatten(input string) error {
	if _, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("parsing the thing ID %q: %+v", input, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}
