# AZG014 - error messages must start in lower case, print the error underneath in full, and have no contractions

AZG014 reports an error message that is not worded the way the provider's others are.

A user reads these when something has gone wrong, after what it happened in: `creating Thing: expected a name`. Printed in full, the error underneath keeps what the API said.

## Flagged Code

```go
return errors.New("Expected a name")

return fmt.Errorf("creating %s: %s", id, err)

return errors.New("`zones` can't be set with `availability_set_id`")
```

```text
error message starts with a capital - write "expected"
error is printed with %s - use %+v, which keeps everything the error says
error message has "can't" - write "cannot"
```

## Passing Code

```go
return errors.New("expected a name")

return fmt.Errorf("creating %s: %+v", id, err)

return errors.New("`zones` cannot be set with `availability_set_id`")
```

## What is checked

The message of a `fmt.Errorf` or `errors.New` call, outside test files.

- `capital`: a message whose first word is a capital and then only small letters. A short form such as `ID`, a name from the code such as `DecodeDiff`, and `Azure`, `Microsoft` and `Terraform` are left alone.
- `error-verb`: an error printed with `%s` or `%v`, and `+%v`, which has the plus on the wrong side. `%w` is left alone: it is how a caller is let ask what the error is.
- `contraction`: `can't`, `won't`, `don't`, `doesn't`, `didn't`, `isn't`, `aren't`, `wasn't`, `weren't`, `couldn't`, `shouldn't`, `wouldn't`, `hasn't`, `haven't` and `mustn't`. `it's` is reported with no fix, as it can be `it is`, `it has` or a misspelt `its`.
- `quoted-value`, which is off unless asked for: a string printed with `%q`, where the docs use backticks. A list or a number printed with `%q` is left alone.

A message that says which value each place prints, such as `%[2]s`, is not checked.

## Fix

Each is fixed where it is: the first letter in lower case, `%+v`, the words written out, and `` `%s` ``.

There is no fix where the message is not one string, or has an escape such as `\"` in it.

## Options

`capital`, `error-verb` and `contraction` can each be switched off, and `quoted-value` switched on. `name` adds a name a message may start with a capital for.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZG014:
            contraction: false
            quoted-value: true
            name: [Databricks, Kubernetes]
```

```sh
azproviderlint -AZG014 -AZG014.quoted-value ./...
```

## Ignoring Reports

Put `//azignore:AZG014 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return fmt.Errorf("creating %s: %s", id, err) //azignore:AZG014 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
