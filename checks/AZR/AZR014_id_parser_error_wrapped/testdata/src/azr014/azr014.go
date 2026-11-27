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

// an error that is not the parser's is left alone, and so is one set again before it is checked
func more(input string) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		input = ""
		return fmt.Errorf("parsing %q: %+v", input, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	other, err := things.ParseThingID(input)
	err = check(other)
	if err != nil {
		return fmt.Errorf("checking %q: %+v", input, err)
	}

	if _, err = things.ValidateThingID(input); err != nil {
		return fmt.Errorf("validating %q: %+v", input, err)
	}

	_ = id

	return nil
}

func check(*things.ThingId) error { return nil }

// added to a list of errors, or checked a line or two after the parse
func list(inputs []string) (errs []error) {
	for _, input := range inputs {
		id, err := things.ParseThingID(input)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing %q: %+v", input, err)) // want `the ID parser's error already says what is wrong with the ID`
			continue
		}

		other, err := things.ParseThingID(input)
		_ = id
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing %q: %+v", input, err)) // want `the ID parser's error already says what is wrong with the ID`
		}

		_ = other
	}

	return errs
}

// an error set again in one arm of an if is no longer the parser's
func again(input string, ready bool) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		return err
	}

	if ready {
		err = check(id)
	}

	if err != nil {
		return fmt.Errorf("checking %q: %+v", input, err)
	}

	return nil
}

// an ID read only to be printed in the message is not kept once the message goes
func printed(input string) error {
	id, err := things.ParseThingID(input)
	if err != nil {
		return fmt.Errorf("parsing %s: %+v", id, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}

func printedAgain(input string) error {
	var err error
	if input == "" {
		return err
	}

	id, err := things.ParseThingID(input)
	if err != nil {
		return fmt.Errorf("parsing %s: %+v", id, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}

func printedInCheck(input string) error {
	if id, err := things.ParseThingID(input); err != nil {
		return fmt.Errorf("parsing %s: %+v", id, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}

// an ID used again after the check is kept
func used(input string) (string, error) {
	id, err := things.ParseThingID(input)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %+v", id, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return id.ID(), nil
}

// an ID declared before it is read would be left unused there, so there is no fix
func declared(input string) error {
	var id *things.ThingId
	var err error

	id, err = things.ParseThingID(input)
	if err != nil {
		return fmt.Errorf("parsing %s: %+v", id, err) // want `the ID parser's error already says what is wrong with the ID`
	}

	return nil
}
