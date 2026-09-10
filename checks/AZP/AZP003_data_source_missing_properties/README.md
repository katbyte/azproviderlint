# AZP003 - data sources must expose their same-named resource's properties

AZP003 pairs each data source with the resource of the same name and reports resource properties the data source does not expose.

A data source is meant to mirror its resource. The usual way they drift apart is someone adding a property to the resource and forgetting the data source. This check catches that.

## Flagged Code

```go
func resourceExample() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
			"zone": {Type: pluginsdk.TypeString, Optional: true},
		},
	}
}

func dataSourceExample() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
			// "zone" is not exposed
		},
	}
}
```

## Passing Code

```go
func dataSourceExample() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
			"zone": {Type: pluginsdk.TypeString, Computed: true},
		},
	}
}
```

## What counts

Pairing works the same way as in AZP002 across untyped, typed, and framework registrations. The schema is read per style: the untyped resource function's body, the typed `Arguments()` and `Attributes()` methods, or the framework `Schema()` method. Property names are collected all the way down, and helper functions in the same package that return schema pieces are followed.

A resource property counts as covered if its name appears anywhere in the data source schema, top level or nested. That means a restructured schema can hide a gap, but it also means restructuring never causes a false report. A missing block is reported once, by dotted path, and its children are not listed separately.

Not reported:

- write-only arguments (`WriteOnly: true`, or names ending in `_wo` / `_wo_version`), which can never be read back
- pairs where either side has no readable properties
- data sources with a non-constant schema key

## Options

| Option | Default | Effect |
|---|---|---|
| `ignore-sensitive` | false | skip resource properties marked `Sensitive: true`, and their children |

For providers whose policy is to keep secrets out of data sources across the board. Set with `-AZP003.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Some properties have no place in a data source, such as create-only inputs, so suppressions are expected.

To exempt one property, put the directive on that property in the **resource** schema. The rest of the schema is still checked:

```go
"customer_managed_key": { //azignore:AZP003 - deliberately not exposed in the data source
	Type:      pluginsdk.TypeString,
	Sensitive: true,
},
```

To silence every report for a data source, put `//azignore:AZP003 - <reason>` on its registration line, or on the line above it. The reason is required.

```go
"azurerm_example": dataSourceExample(), //azignore:AZP003 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
