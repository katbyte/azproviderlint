# AZP008 - error messages must be formatted correctly

AZP008 reports an error message that is not formatted correctly. Each way it checks can be switched off on its own.

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
- `error-verb`: an error printed with `%s` or `%v` where it should be `%+v`, and `+%v`, which has the plus on the wrong side. `%w` is left alone: it prints the same text and keeps the error underneath, so the code that gets the message can still tell what went wrong.
- `contraction`: `can't`, `won't`, `don't`, `doesn't`, `didn't`, `isn't`, `aren't`, `wasn't`, `weren't`, `couldn't`, `shouldn't`, `wouldn't`, `hasn't`, `haven't` and `mustn't`. `it's` is reported with no fix, as it can be `it is`, `it has` or a misspelt `its`.
- `quoted-error`: a string printed with `%q`, where the docs use backticks. A list or a number printed with `%q` is left alone.

A message that says which value each place prints, such as `%[2]s`, is not checked.

## Fix

Each is fixed where it is: the first letter in lower case, `%+v`, the words written out, and `` `%s` `` in place of `%q`.

There is no fix where the message is not one string, or has an escape such as `\"` in it.

## Options

`capital` and `contraction` can each be switched off with `false`. `name` adds a name a message may start with a capital for.

`error-verb` is how an error is printed: `%+v` by default, or `%w`, `%v` or `%s`. `%w` is always allowed. `none` switches it off.

`quoted-error` is how a value is put in quotes: `backticks` by default, or `%q`, which reports a string printed as `` `%s` `` instead. `none` switches it off.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZP008:
            error-verb: "%w"
            quoted-error: none
            name: [Databricks, Kubernetes]
```

```sh
azproviderlint -AZP008 -AZP008.error-verb=%w -AZP008.quoted-error=none ./...
```

## Ignoring Reports

Put `//azignore:AZP008 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return fmt.Errorf("creating %s: %s", id, err) //azignore:AZP008 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
