# AZS003 - TypeList blocks must not allow empty blocks

AZS003 reports a `TypeList` block whose nested schema is all optional, with no `Default`, `DefaultFunc`, `AtLeastOneOf`, or `ExactlyOneOf` on any of its fields. A block like that accepts:

```hcl
resource "azurerm_example" "example" {
  settings {}
}
```

An empty block comes through as a `nil` list element. Expand functions usually do `raw[0].(map[string]interface{})` on it and panic (see [azurerm #11426](https://github.com/hashicorp/terraform-provider-azurerm/issues/11426)), or the user gets a diff that never goes away. Making any one field `Required`, giving it a `Default`, or adding an `AtLeastOneOf` or `ExactlyOneOf` constraint fixes it.

This is a heuristic. A block whose expand function handles the nil element is safe in practice but is still reported. Suppress those with an `//azignore:AZS003` comment.

Ported from [tfproviderlint PR #236 (XS003)](https://github.com/bflad/tfproviderlint/pull/236). The `pluginsdk` aliases used in azurerm are recognised.

## Flagged Code

```go
"settings": {
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"foo": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	},
},
```

## Passing Code

```go
"settings": {
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"foo": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"settings.0.foo"},
			},
		},
	},
},
```

## Ignoring Reports

Put `//azignore:AZS003 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
"settings": { //azignore:AZS003 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
