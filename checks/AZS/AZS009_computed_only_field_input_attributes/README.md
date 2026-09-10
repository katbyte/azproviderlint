# AZS009 - computed-only fields must not set input-only schema attributes

AZS009 reports a computed-only schema field that sets something only user input can use, such as `ValidateFunc`, `MaxItems`, or `Default`. It also reports `Optional` or `Required` on a field nested inside a computed-only block.

A computed-only field (`Computed: true` with no `Optional` or `Required`) is filled in by the provider from the API. The user never writes it, so there is nothing to validate, default, count, or check for conflicts. Those attributes are dead weight at best. The plugin SDK rejects most of them when the provider starts, but only on the field that says `Computed`. It does not notice that everything inside a computed-only block is computed-only too, and it does not look at an element schema's `ValidateFunc`. This check does.

## Flagged Code

```go
"status": {
	Type:         pluginsdk.TypeString,
	Computed:     true,
	ValidateFunc: validation.StringIsNotEmpty,
},
```

```go
"network": {
	Type:     pluginsdk.TypeList,
	Computed: true,
	MaxItems: 1,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"subnet_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true, // nothing in this block can be configured
				ValidateFunc: commonids.ValidateSubnetID,
			},
		},
	},
},
```

```go
func (r ExampleResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"endpoint": {
			Type:         pluginsdk.TypeString,
			ValidateFunc: validation.IsURLWithHTTPS, // Attributes() are all computed
		},
	}
}
```

## Passing Code

```go
"status": {
	Type:     pluginsdk.TypeString,
	Computed: true,
},
```

```go
"network": {
	Type:     pluginsdk.TypeList,
	Computed: true,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"subnet_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	},
},
```

```go
// optional+computed is configurable, so validation belongs
"name": {
	Type:         pluginsdk.TypeString,
	Optional:     true,
	Computed:     true,
	ValidateFunc: validation.StringIsNotEmpty,
},
```

## What counts

A field is computed-only when `Computed: true` is a constant and neither `Optional` nor `Required` is. Everything reached through its `Elem`, at any depth, is computed-only as well. In a typed resource, every field in `Attributes()` is computed-only because azurerm's SDK wrapper sets `Computed` on all of them at start-up.

Reported on any such field:

- `Default`, `DefaultFunc`, `InputDefault`
- `ValidateFunc`, `ValidateDiagFunc`
- `DiffSuppressFunc`, `DiffSuppressOnRefresh`, `StateFunc`
- `MaxItems`, `MinItems`
- `AtLeastOneOf`, `ExactlyOneOf`, `ConflictsWith`, `RequiredWith`
- `WriteOnly`
- `ConfigMode: SchemaConfigModeBlock`

Reported on a field nested in a computed-only block, or listed in `Attributes()`: `Optional: true` and `Required: true`.

Not reported:

- attributes set to a zero value (`MaxItems: 0`, `Default: nil`)
- `Computed` set from a variable rather than a constant, since the field may not be computed-only
- an `Elem` that comes from a function call rather than a literal
- `ForceNew`, `Sensitive`, `Description`, `Deprecated`, and `Set`, which describe state and are fine on computed fields

## The fix

`-fix` deletes the attribute's line. For a nested `Optional` or `Required` it renames the key to `Computed`, or deletes the line when the field already says `Computed`. An `Optional` inside `Attributes()` gets no fix, because the right move is usually to the `Arguments()` map, and that is a person's call.

## Ignoring Reports

Put `//azignore:AZS009 - <reason>` at the end of the attribute's line, or on the line above it. The reason is required.

```go
ValidateFunc: validation.StringIsNotEmpty, //azignore:AZS009 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
