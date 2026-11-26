# AZS022 - a string that only takes Enabled or Disabled must be a bool

AZS022 reports a string property whose validation allows only two values that mean on and off.

That is a bool written the long way. The user has to look up two words where `true` and `false` would do.

## Flagged Code

```go
"public_network_access": {
	Type:         pluginsdk.TypeString,
	Optional:     true,
	ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
},
```

```text
string property "public_network_access" only takes "Enabled" and "Disabled" - make it a bool
```

```go
"public_network_access": schema.StringAttribute{
	Optional:   true,
	Validators: []validator.String{stringvalidator.OneOf("Enabled", "Disabled")},
},
```

## Passing Code

```go
"public_network_access_enabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},
```

The bool is named for what it turns on, ending in `_enabled`, and the two words are swapped for it when the request is built and the response is read.

## What is checked

The pairs are `Enabled` and `Disabled`, `Enable` and `Disable`, and `On` and `Off`, in any case and either order.

The values are read from:

- a list written out in `validation.StringInSlice`, as text or as an SDK enum's constants
- an SDK enum's helper, such as `PossibleValuesForPublicNetworkAccess()`
- the framework's `stringvalidator.OneOf` and `OneOfCaseInsensitive`

These are not reported:

- a third value, such as `None`, which is more than a bool can say
- a list that is not written out, or has a value that is not
- a property with `Deprecated` or `DeprecationMessage` set, which is on its way out
- anything in a package named `migration` or `migrations`, which is what old state looked like

Changing the type of a property that has shipped is a breaking change. Until the next major version, ignore the report with a reason.

## Fix

None is offered. The create, update and read code changes with the type.

## Ignoring Reports

Put `//azignore:AZS022 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false), //azignore:AZS022 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
