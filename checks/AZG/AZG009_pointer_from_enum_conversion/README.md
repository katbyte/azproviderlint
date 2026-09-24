# AZG009 - use pointer.FromEnum for enum conversions

AZG009 reports `string(pointer.From(input))`, which should be `pointer.FromEnum(input)`.

Both helpers come from [go-azure-helpers](https://github.com/hashicorp/go-azure-helpers). `FromEnum` expresses the conversion directly and still returns an empty string for a nil pointer.

Only [go-azure-sdk](https://github.com/hashicorp/go-azure-sdk) enums are reported: a named string type from a go-azure-sdk package that has a generated `PossibleValuesFor<Name>()` helper. `pointer.To` on strings, numbers, and other types is left alone.

## Flagged Code

```go
output.StartTLSPolicy = string(pointer.From(input.StartTLSPolicy))
d.Set("allocation_policy", string(pointer.From(props.AllocationPolicy)))
strings.ToLower(string(pointer.From(input.EndpointType)))
```

## Passing Code

```go
output.StartTLSPolicy = pointer.FromEnum(input.StartTLSPolicy)
d.Set("allocation_policy", pointer.FromEnum(props.AllocationPolicy))
strings.ToLower(pointer.FromEnum(input.EndpointType))
```

## The fix

`-fix` removes the outer string conversion and changes `From` to `FromEnum`.

## Ignoring Reports

Put `//azignore:AZG009 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
return string(pointer.From(input)) //azignore:AZG009 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
