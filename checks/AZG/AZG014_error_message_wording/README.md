# AZG014 - error messages must print the error underneath in full, have no contractions, and put values in backticks

AZG014 reports an error message that is not worded the way the provider's others are.

A user reads these when something has gone wrong. Printed in full, the error underneath keeps what the API said. Words written out and values in backticks match the docs.

## Flagged Code

```go
return fmt.Errorf("creating %s: %s", id, err)

return fmt.Errorf("retrieving %q: %+v", name, err)

return errors.New("`zones` can't be set with `availability_set_id`")
```

```text
error is printed with %s - use %+v, which keeps everything the error says
error message puts a value in quotes with %q - put it in backticks, as the docs do
error message has "can't" - write "cannot"
```

## Passing Code

```go
return fmt.Errorf("creating %s: %+v", id, err)

return fmt.Errorf("retrieving `%s`: %+v", name, err)

return errors.New("`zones` cannot be set with `availability_set_id`")
```

## What is checked

The message of a `fmt.Errorf` or `errors.New` call, outside test files.

- `error-verb`: an error printed with `%s` or `%v`. `%w` is left alone: it is how a caller is let ask what the error is.
- `contraction`: `can't`, `won't`, `don't`, `doesn't`, `didn't`, `isn't`, `aren't`, `wasn't`, `weren't`, `couldn't`, `shouldn't`, `wouldn't`, `hasn't`, `haven't` and `mustn't`.
- `quoted-value`: a string printed with `%q`. A list or a number printed with `%q` is left alone.

A message that says which value each place prints, such as `%[2]s`, is not checked.

## Fix

Each is fixed where it is: `%+v`, the words written out, and `` `%s` ``.

There is no fix where the message is not one string, or has an escape such as `\"` in it.

## Options

Each kind can be switched off on its own.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZG014:
            quoted-value: false
```

```sh
azproviderlint -AZG014 -AZG014.quoted-value=false ./...
```

## Ignoring Reports

Put `//azignore:AZG014 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return fmt.Errorf("creating %s: %s", id, err) //azignore:AZG014 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
