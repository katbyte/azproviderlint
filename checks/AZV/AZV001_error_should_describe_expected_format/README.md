# AZV001 - 'invalid format' error messages must describe the expected format

AZV001 reports validation errors that say `invalid format of ...`.

That message tells the user something is wrong but not how to fix it. Say what a valid value looks like instead.

## Flagged Code

```go
return fmt.Errorf("invalid format of %q", name)
```

## Passing Code

```go
return fmt.Errorf("%q must start with a letter, may contain letters and numbers, and must end with a letter", name)
```

## Ignoring Reports

Put `//azignore:AZV001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
return fmt.Errorf("invalid format of %q", name) //azignore:AZV001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
