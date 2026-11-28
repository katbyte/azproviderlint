# AZS029 - a write-only property must be tied to its version and to the property it stands in for

AZS029 reports a write-only property that is not tied to its version, or to the property it stands in for.

Terraform keeps no write-only value, so it cannot tell when one has changed. The version beside it is what says so. Unless the schema ties them together, a user can set a value that is never sent, or set the same secret two ways at once.

## Flagged Code

```go
"password": {
	Type:      pluginsdk.TypeString,
	Optional:  true,
	Sensitive: true,
},

"password_wo": {
	Type:      pluginsdk.TypeString,
	Optional:  true,
	WriteOnly: true,
},

"password_wo_version": {
	Type:     pluginsdk.TypeInt,
	Optional: true,
},
```

```text
"password" and "password_wo" can both be set - add ConflictsWith: []string{"password_wo"}
"password_wo" can be set without "password_wo_version" - add RequiredWith: []string{"password_wo_version"}
"password_wo" and "password" can both be set - add ConflictsWith: []string{"password"}
"password_wo_version" can be set without "password_wo" - add RequiredWith: []string{"password_wo"}
"password_wo_version" can be set to 0, which is what an unset version is stored as - add ValidateFunc: validation.IntAtLeast(1)
```

## Passing Code

```go
"password": {
	Type:          pluginsdk.TypeString,
	Optional:      true,
	Sensitive:     true,
	ConflictsWith: []string{"password_wo"},
},

"password_wo": {
	Type:          pluginsdk.TypeString,
	Optional:      true,
	WriteOnly:     true,
	RequiredWith:  []string{"password_wo_version"},
	ConflictsWith: []string{"password"},
},

"password_wo_version": {
	Type:         pluginsdk.TypeInt,
	Optional:     true,
	RequiredWith: []string{"password_wo"},
	ValidateFunc: validation.IntAtLeast(1),
},
```

## What is checked

For each plugin SDK property with `WriteOnly: true` in a map of properties, written out or handed back by a function:

- a `<name>_version` property sits beside it
- each names the other in `RequiredWith`, unless the other is `Required` and so always set
- the version is checked with `validation.IntAtLeast(1)`, or `IntBetween` starting at 1
- when the name ends in `_wo` and the property without that ending can be set, each names the other in `ConflictsWith` or `ExactlyOneOf`

A name in those lists can be the whole path to a property in a block, such as `secret.0.value_wo`. A list that is not written out in place, or is in a function, is taken to name what it should.

These are not checked:

- framework schemas
- state migrations

## Fix

None is offered.

## Ignoring Reports

Put `//azignore:AZS029 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"password_wo": { //azignore:AZS029 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
