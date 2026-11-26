# AZS024 - string properties must be validated

AZS024 reports a string property the user sets that has no validation.

Without it a value the API will refuse is only found part way through an apply. With it Terraform says so when the plan is made.

## Flagged Code

```go
"name": {
	Type:     pluginsdk.TypeString,
	Required: true,
},
```

```text
string property "name" has no validation - add a ValidateFunc that checks what the API accepts
```

```go
"zones": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem: &pluginsdk.Schema{
		Type: pluginsdk.TypeString,
	},
},
```

```text
the strings in "zones" have no validation - add a ValidateFunc that checks what the API accepts
```

```go
"name": schema.StringAttribute{
	Required: true,
},
```

## Passing Code

```go
"name": {
	Type:         pluginsdk.TypeString,
	Required:     true,
	ValidateFunc: validate.ThingName,
},
```

```go
"zones": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem: &pluginsdk.Schema{
		Type:         pluginsdk.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"1", "2", "3"}, false),
	},
},
```

```go
"name": schema.StringAttribute{
	Required:   true,
	Validators: []validator.String{stringvalidator.LengthBetween(3, 24)},
},
```

## What is checked

- a plugin SDK string that is `Required` or `Optional`: it needs a `ValidateFunc` or `ValidateDiagFunc`
- the strings in a list or set that is `Required` or `Optional`: the `Elem` needs one
- a framework `StringAttribute` that is `Required` or `Optional`: it needs `Validators`

The rule only looks for validation being there. Whether it checks enough, such as `validation.StringIsNotEmpty` on a name the API has rules for, is for a reviewer.

These are not reported:

- a property that is only `Computed`, which the user does not set
- the strings in a map, which cannot be checked one by one
- a property with `Deprecated` or `DeprecationMessage` set, which is on its way out
- a property whose whole schema comes from a function, such as `commonschema.Location()`
- anything in a package named `migration` or `migrations`, which is what old state looked like

Adding validation to a property that has shipped can refuse a value that used to work. Take care that what is added is no stricter than the API.

## Fix

None is offered. What a valid value looks like is in the API's spec or its docs.

## Ignoring Reports

Put `//azignore:AZS024 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"name": { //azignore:AZS024 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
