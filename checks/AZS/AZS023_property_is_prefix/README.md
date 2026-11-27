# AZS023 - property names must not start with is_

AZS023 reports a schema property whose name starts with `is_`.

The prefix says nothing the rest of the name does not, and the provider's other properties do without it.

## Flagged Code

```go
"is_storage_enabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},
```

```text
property "is_storage_enabled" starts with is_ - name it "storage_enabled"
```

## Passing Code

```go
"storage_enabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},
```

## What is checked

The name a property is given in a schema, whatever its type:

- in a map of plugin SDK properties, at any depth
- where a property is added to such a map later: `out["is_legacy"] = &pluginsdk.Schema{...}`
- in a map of framework attributes or blocks

A name that only starts with the same letters, such as `issuer`, is not reported. Nor is anything in a package named `migration` or `migrations`, which is what old state looked like.

Renaming a property that has shipped is a breaking change. Until the next major version, ignore the report with a reason.

## Fix

None is offered. The name is also used wherever the property is read and set, in the tests and in the docs.

## Ignoring Reports

Put `//azignore:AZS023 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"is_storage_enabled": { //azignore:AZS023 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
