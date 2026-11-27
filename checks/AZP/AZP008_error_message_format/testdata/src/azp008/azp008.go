package azp008

import (
	"errors"
	"fmt"
)

type thingId struct{}

func (thingId) String() string { return "" }

const prefix = "creating"

func wording(id thingId, name string, count int, err error) error {
	_ = fmt.Errorf("creating %s: %s", id, err)         // want `error is printed with %s - use %\+v`
	_ = fmt.Errorf("creating %s: %v", id, err)         // want `error is printed with %v - use %\+v`
	_ = fmt.Errorf("creating %q: %+v", name, err)      // want "error message puts a value in quotes with %q - put it in backticks, as the docs do"
	_ = fmt.Errorf("%d%% of %q: %s", count, name, err) // want `error message puts a value in quotes with %q` `error is printed with %s`

	_ = errors.New("`zones` can't be set with `availability_set_id`") // want `error message has "can't" - write "cannot"`
	_ = fmt.Errorf("Doesn't exist: %+v", err)                         // want `error message has "Doesn't" - write "does not"`
	_ = errors.New("Expected a name")                                 // want `error message starts with a capital - write "expected"`
	_ = fmt.Errorf("Creating %s: %+v", id, err)                       // want `error message starts with a capital - write "creating"`
	_ = errors.New("it won't work and shouldn't be tried")            // want `error message has "won't" - write "will not"` `error message has "shouldn't" - write "should not"`
	_ = errors.New("once it's enabled it cannot be turned off") // want `error message has "it's" - write "it is", "it has" or "its"`
	_ = fmt.Errorf("expanding `pool`: +%v", err)   // want `error message has \+%v, with the plus on the wrong side - use %\+v`
	_ = fmt.Errorf("expanding `pool`: +%v", name)  // want `error message has \+%v, with the plus on the wrong side - use %\+v`

	// a message that is not one plain string is reported as a whole, with no fix
	_ = fmt.Errorf(prefix+" %s: %s", id, err)  // want `error is printed with %s`
	_ = fmt.Errorf("creating \"it\": %v", err) // want `error is printed with %v`

	return nil
}

func fine(id thingId, name string, names []string, width int, err error) error {
	_ = fmt.Errorf("creating %s: %+v", id, err)
	_ = fmt.Errorf("creating `%s`: %+v", name, err)
	_ = errors.New("`zones` cannot be set with `availability_set_id`")

	// %w prints the same, and keeps the error for the code that gets the message
	_ = fmt.Errorf("creating %s: %w", id, err)

	// what is not text is not a value in the docs' sense, and its quotes are not the same as backticks
	_ = fmt.Errorf("expected one of %q", names)
	_ = fmt.Errorf("unexpected character %q", 'a')

	// the text of an error, not the error
	_ = fmt.Errorf("creating %s: %s", id, err.Error())

	// a place that says which value it prints is not followed
	_ = fmt.Errorf("creating %[2]s: %[1]s", err, id)
	_ = fmt.Errorf("creating %*s: %+v", width, name, err)

	// a short form or a name may start a message, and so may a value
	_ = errors.New("ID was not found")
	_ = errors.New("IPv4 is not supported")
	_ = errors.New("Azure refused the request")
	_ = errors.New("DecodeDiff was handed nothing")
	_ = errors.New("IotHub was not found")
	_ = fmt.Errorf("`%s` is not a name", name)

	// a word this rule does not know is left alone
	_ = errors.New("the thing ain't there")

	return nil
}
