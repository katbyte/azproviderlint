# AZS005 - the fields of a schema property must be written in the agreed order

AZS005 reports a schema property whose fields are not in the agreed order.

Every property then reads the same way: what it is, its type, whether it is set, how it is checked, and last what it holds.

## Flagged Code

```go
"sku": {
	Type:         pluginsdk.TypeString,
	ValidateFunc: validation.StringIsNotEmpty,
	Computed:     true,
	Optional:     true,
	Description:  "The SKU of the Thing.",
},
```

```go
"rule": {
	Type: pluginsdk.TypeList,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{},
	},
	Optional: true,
},
```

## Passing Code

```go
"sku": {
	Description:  "The SKU of the Thing.",
	Type:         pluginsdk.TypeString,
	Optional:     true,
	Computed:     true,
	ValidateFunc: validation.StringIsNotEmpty,
},
```

```go
"rule": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem: &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{},
	},
},
```

## The order

For a plugin SDK schema:

`Description`, `Type`, `Required`, `Optional`, `Computed`, `ForceNew`, `Default`, `DefaultFunc`, `InputDefault`, `Sensitive`, `WriteOnly`, `MinItems`, `MaxItems`, `ConfigMode`, `ValidateFunc`, `ValidateDiagFunc`, `StateFunc`, `DiffSuppressFunc`, `DiffSuppressOnRefresh`, `Set`, `RequiredWith`, `ConflictsWith`, `ExactlyOneOf`, `AtLeastOneOf`, `Deprecated`, `Elem`

For a framework attribute or block, the same order with its own field names:

`Description`, `MarkdownDescription`, `CustomType`, `ElementType`, `NestedObject`, `AttributeTypes`, `Required`, `Optional`, `Computed`, `Default`, `Sensitive`, `WriteOnly`, `Validators`, `PlanModifiers`, `DeprecationMessage`, `Attributes`, `Blocks`

The plugin SDK order is the one most properties in the azurerm, azuread and aws providers already use, with what the property holds, such as `Elem`, last. About nine in ten in each of the three already fit.

A property that is both set and read back is `Optional`, then `Computed`.

A field the order does not name is not checked, and stays where it is.

## Fix

The fields are put in order, each with its own line's comment and any comment above it, and the values are lined up again as `gofmt` would. A `Note: O+C` comment stays between `Optional` and `Computed`.

A property inside another is put in order in the same fix.

There is no fix where two fields share a line in a property written over several.

## Options

`order` is the order for a plugin SDK schema, and `framework-order` that for a framework attribute or block. Either replaces its default.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS005:
            order: [Type, Description, Required, Optional, Computed, Elem]
```

```sh
azproviderlint -AZS005 -AZS005.order=Type -AZS005.order=Description -AZS005.order=Required ./...
```

## Ignoring Reports

Put `//azignore:AZS005 - <reason>` at the end of the reported field's line, or on the line above it. The reason is required.

```go
Type: pluginsdk.TypeString, //azignore:AZS005 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
