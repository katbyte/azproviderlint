# AZS002 - schema Default values must match the declared Type

AZS002 reports a schema field whose `Default` does not match its `Type`, such as `Default: true` on a `TypeInt`.

The plugin SDK's own `InternalValidate` does not check this, so the mistake only shows up as an error when someone runs a plan.

Named constants are resolved, so `Default: SkuStandard` is checked against the constant's type. `TypeFloat` accepts both int and float constants. Defaults that are not constants are skipped, as are list, set, and map types, which cannot have a literal default. The `pluginsdk` aliases used in azurerm are recognised.

Ported from [tfproviderlint PR #329 (S038)](https://github.com/bflad/tfproviderlint/pull/329).

## Flagged Code

```go
&schema.Schema{
	Type:     schema.TypeInt,
	Optional: true,
	Default:  true,
}
```

## Passing Code

```go
&schema.Schema{
	Type:     schema.TypeBool,
	Optional: true,
	Default:  true,
}
```

## Ignoring Reports

Put `//azignore:AZS002 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
Default: true, //azignore:AZS002 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
