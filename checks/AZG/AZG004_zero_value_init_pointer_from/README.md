# AZG004 - use pointer.From instead of nil-check dereference

AZG004 reports the pattern of setting a variable to its zero value and then overwriting it from a pointer if the pointer is not nil:

```go
enabled := false
if props.Enabled != nil {
	enabled = *props.Enabled
}
```

`pointer.From(x)` from [go-azure-helpers](https://github.com/hashicorp/go-azure-helpers) does exactly that in one line: it returns `*x`, or the zero value when `x` is nil.

## Flagged Code

```go
enabled := false
if props.Enabled != nil {
	enabled = *props.Enabled
}

// the var form counts too
var enabled bool
if props.Enabled != nil {
	enabled = *props.Enabled
}
```

## Passing Code

```go
enabled := pointer.From(props.Enabled)

// not reported: a non-zero starting value changes the meaning
enabled := true
if props.Enabled != nil {
	enabled = *props.Enabled
}
```

## What counts

The two statements must be next to each other. The variable must start as a zero value (`false`, `0`, `""`, `nil`), whether written as `y := 0`, `var y T`, `var y T = 0`, or `var y = 0`. The `if` must have exactly one condition, `x != nil`, no `else`, and a body that is just `y = *x` where `x` is the same expression that was checked.

## The fix

`-fix` collapses both statements into `y := pointer.From(x)`. It uses whatever name the file already imports the pointer package under, or adds the import in sorted position among the non-standard-library imports so gci grouping is kept.

## Ignoring Reports

Put `//azignore:AZG004 - <reason>` at the end of the declaration line, or on the line above it. The reason is required.

```go
enabled := false //azignore:AZG004 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
