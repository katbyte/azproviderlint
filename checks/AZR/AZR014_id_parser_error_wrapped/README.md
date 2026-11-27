# AZR014 - an ID parser's error must be returned as it is

AZR014 reports an ID parser's error that is wrapped in another message.

The parser's error already names the ID, says what is wrong with it and shows what a right one looks like. Wrapping it in "parsing ..." only makes the message the user reads longer.

## Flagged Code

```go
id, err := things.ParseThingID(d.Id())
if err != nil {
	return fmt.Errorf("parsing %q: %+v", d.Id(), err)
}
```

## Passing Code

```go
id, err := things.ParseThingID(d.Id())
if err != nil {
	return err
}
```

## What is checked

A call to a function named `Parse<X>ID` or `Parse<X>IDInsensitively` that gives back an ID with an `ID()` method, followed by an `if err != nil` that puts that error inside `fmt.Errorf`, whether it is returned or added to a list of errors.

These are not checked:

- what reads any ID, such as `resourceids.ParseAzureResourceID`, where saying which ID was wanted adds to the error
- a provider's own parsers that are not named `Parse<X>ID`
- test files, and state migrations, which read IDs as they used to be and say so

## Fix

The `fmt.Errorf` call is replaced with the error.

An ID that was read only to be printed in the message is dropped. The read moves into the check, as in `if _, err := ...; err != nil`. Where the error is used after the check, or something sits between the two, it becomes `_, err :=` instead.

The last fix in a file that then has no other use of `fmt` removes the import.

There is no fix where the ID was declared before it was read, which would leave that variable unused.

## Ignoring Reports

Put `//azignore:AZR014 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return fmt.Errorf("parsing %q: %+v", d.Id(), err) //azignore:AZR014 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
