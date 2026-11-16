# AZR011 - typed resources must assert the sdk interface they implement

AZR011 reports a typed resource, data source, list resource, ephemeral resource or action with no `var _ sdk.<Interface> = <Type>{}` line.

The line has the compiler check the type against what the sdk expects, in the file that declares it, and tells a reader what the type is. It matters most for the optional parts. The sdk looks for an `Update` when the provider runs, so one whose signature has drifted is quietly never called; with `var _ sdk.ResourceWithUpdate` the build fails instead.

## Flagged Code

```go
type ArcMachineResource struct{}

func (r ArcMachineResource) Create() sdk.ResourceFunc {
	// ...
}

func (r ArcMachineResource) Update() sdk.ResourceFunc {
	// ...
}
```

```text
ArcMachineResource has no "var _ sdk.ResourceWithUpdate = ArcMachineResource{}" line to have the compiler check it where it is declared
```

## Passing Code

```go
type ArcMachineResource struct{}

var _ sdk.ResourceWithUpdate = ArcMachineResource{}
```

## What counts

A type declared outside a test file that is one of these, by its methods, on the type or on a pointer to it:

| Interface | Asked for |
|---|---|
| `sdk.Resource` | `sdk.ResourceWithUpdate` when it has an `Update`, else `sdk.Resource` |
| `sdk.DataSource` | `sdk.DataSource` |
| `sdk.FrameworkWrappedResource` | `sdk.FrameworkWrappedResourceWithUpdate` when it has an `Update`, else `sdk.FrameworkWrappedResource` |
| `sdk.FrameworkWrappedDataSource` | `sdk.FrameworkWrappedDataSource` |
| `sdk.FrameworkListWrappedResource` | `sdk.FrameworkListWrappedResource` |
| `sdk.EphemeralResource` | `sdk.EphemeralResource` |
| `sdk.Action` | `sdk.Action` |

The sdk is the package whose import path ends in `/internal/sdk`.

A type passes when a `var _` line anywhere in its package, outside the tests, asserts it against any interface the sdk declares. `FooResource{}`, `&FooResource{}` and `(*FooResource)(nil)` all count, alone or in a `var (...)` group. Which interface to name is the author's call: a resource with an update that asserts plain `sdk.Resource` passes.

An interface with the same methods under another name counts too. `sdk.Action` is the framework's `action.ActionWithConfigure`, so asserting that passes.

Not reported:

- types in test files
- generic types and interfaces
- untyped resources, which are functions and have nothing to assert

## Fix

The fix adds the line after the type, using `&` when the methods are on the pointer:

```go
type ArcMachineResource struct{}

var _ sdk.ResourceWithUpdate = ArcMachineResource{}
```

It is offered when the file already imports the sdk. Where it does not, as in a `resource.go` whose methods are in other files, add the import and the line by hand.

## Ignoring Reports

Put `//azignore:AZR011 - <reason>` at the end of the reported type's line, or on the line above it. The reason is required.

```go
type ArcMachineResource struct{} //azignore:AZR011 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
