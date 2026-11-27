# AZS025 - number properties must be validated

AZS025 reports a number property the user sets that has no validation.

Without it a value the API will refuse, such as one out of range, is only found part way through an apply. With it Terraform says so when the plan is made.

## Flagged Code

```go
"count": {
	Type:     pluginsdk.TypeInt,
	Required: true,
},
```

```text
number property "count" has no validation - add a ValidateFunc that checks what the API accepts
```

```go
"ports": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem: &pluginsdk.Schema{
		Type: pluginsdk.TypeInt,
	},
},
```

```text
the numbers in "ports" have no validation - add a ValidateFunc that checks what the API accepts
```

## Passing Code

```go
"count": {
	Type:         pluginsdk.TypeInt,
	Required:     true,
	ValidateFunc: validation.IntBetween(1, 10),
},
```

## What is checked

- a plugin SDK `TypeInt` or `TypeFloat` that is `Required` or `Optional`: it needs a `ValidateFunc` or `ValidateDiagFunc`
- the numbers in a list or set that is `Required` or `Optional`: the `Elem` needs one
- a framework number attribute, such as `Int64Attribute`, that is `Required` or `Optional`: it needs `Validators`

A schema handed back by a function is followed into whichever package the function is in. One in the provider's own code is reported where it is written. One from another module, such as go-azure-helpers, is reported where it is used, as the provider cannot change it.

The rule only looks for validation being there. Whether it checks enough is for a reviewer.

These are not reported:

- a property that is only `Computed`, which the user does not set
- the numbers in a map, which cannot be checked one by one
- a property with `Deprecated` or `DeprecationMessage` set, which is on its way out
- anything in a package named `migration` or `migrations`, which is what old state looked like
- a schema in a test file, which is never shipped

Adding validation to a property that has shipped can refuse a value that used to work. Take care that what is added is no stricter than the API.

## Fix

None is offered. What a valid value looks like is in the API's spec or its docs.

## Ignoring Reports

Put `//azignore:AZS025 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"count": { //azignore:AZS025 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
