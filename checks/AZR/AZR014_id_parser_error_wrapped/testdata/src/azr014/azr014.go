package azr014

import (
	"fmt"

	"azr014/things"
)

func read(input string) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		return fmt.Errorf("parsing %q: %+v", input, err) // want `the ID parser's error already says what is wrong with the ID and what a right one looks like - return it as it is`
	}

	other, err := things.ParseThingIDInsensitively(input)
	if err != nil {
		return fmt.Errorf("parsing `thing_id`: %+v", err) // want `the ID parser's error already says what is wrong with the ID`
	}

	if _, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("flattening `thing_id`: %s", err) // want `the ID parser's error already says what is wrong with the ID`
	}

	_, _ = id, other

	return nil
}

func expand(input string) (*things.ThingId, error) {
	switch input {
	case "":
		return nil, nil

	default:
		id, err := things.ParseThingID(input)
		if err != nil {
			return nil, fmt.Errorf("parsing %q as a Thing ID: %+v", input, err) // want `the ID parser's error already says what is wrong with the ID`
		}

		return id, nil
	}
}

// returned as it is
func passes(input string) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		return err
	}

	_ = id

	return nil
}

// what reads any ID says less, and saying which ID was wanted adds to it
func general(input string) error {
	id, err := things.ParseAzureResourceID(input)
	if err != nil {
		return fmt.Errorf("parsing %q as a Thing ID: %+v", input, err)
	}

	_ = id

	return nil
}

// a check that does more than return is left alone, and so is an error that is not the parser's
func more(input string) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		input = ""
		return fmt.Errorf("parsing %q: %+v", input, err)
	}

	if _, err = things.ValidateThingID(input); err != nil {
		return fmt.Errorf("validating %q: %+v", input, err)
	}

	_ = id

	return nil
}
