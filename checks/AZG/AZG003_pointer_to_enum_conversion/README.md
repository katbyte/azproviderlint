# AZG003 - use pointer.ToEnum for enum conversions

AZG003 reports `pointer.To(sdk.SomeEnum(v))`, which should be `pointer.ToEnum[sdk.SomeEnum](v)`.

Both come from [go-azure-helpers](https://github.com/hashicorp/go-azure-helpers). `ToEnum` says what is happening and names the enum type once instead of wrapping a conversion in a call.

Only [go-azure-sdk](https://github.com/hashicorp/go-azure-sdk) enums are reported: a named string type from a go-azure-sdk package that has a generated `PossibleValuesFor<Name>()` helper. `pointer.To` on strings, numbers, and other types is left alone.

`ToEnum` takes a `string`. When `v` is some other named string type, such as another enum, the rewrite needs a `string(v)` wrap to compile, and the report says so.

## Flagged Code

```go
return pointer.To(virtualmachines.VirtualMachinePriorityTypes(priority))
return pointer.To(managedclusters.ArtifactSource(config["artifact_source"].(string)))
```

## Passing Code

```go
return pointer.ToEnum[virtualmachines.VirtualMachinePriorityTypes](priority)
return pointer.ToEnum[managedclusters.ArtifactSource](config["artifact_source"].(string))

// not enums, not reported
return pointer.To("regular string")
return pointer.To(42)
```

## The fix

`-fix` rewrites the call, adding the `string(...)` wrap when it is needed.

## Ignoring Reports

Put `//azignore:AZG003 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
return pointer.To(virtualmachines.OperatingSystemTypes("Linux")) //azignore:AZG003 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
