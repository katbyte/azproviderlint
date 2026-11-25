# AZS010 - resources, properties and blocks must have a description

AZS010 reports a resource, data source, property or block that has no description.

The docs are made from the descriptions, so one left out is a gap on the page.

## Flagged Code

```go
func resourceThing() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: resourceThingRead,

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"sku": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"tier": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}
```

```text
resource has no description
property "name" has no description
block "sku" has no description
property "tier" has no description
```

```go
type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }
```

```text
resource ThingResource has no Description() method
```

## Passing Code

```go
func resourceThing() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Description: "Manages a Thing.",
		Read:        resourceThingRead,

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:        pluginsdk.TypeString,
				Required:    true,
				Description: "The name of the Thing.",
			},
		},
	}
}
```

```go
func (ThingResource) Description() string { return "Manages a Thing." }
```

## What needs a description

- a plugin SDK resource or data source: a `pluginsdk.Resource` that sets `Read`
- a plugin SDK property or block: a `pluginsdk.Schema`, in a schema map or returned by a helper
- a typed resource or data source: a type with a `ResourceType()` method needs a `Description() string` method
- a framework schema, and each attribute and block in it: `Description` or `MarkdownDescription`
- a provider function and each of its parameters: `Description` or `MarkdownDescription`

Provider and `features` block properties count too.

These need none:

- the element of a list or set: `Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}`
- the `pluginsdk.Resource` that holds a block's properties; the block itself is described, and so is each property inside
- a schema handed to a function, such as the one that hashes a set
- anything in a package named `migration` or `migrations`, which is what old state looked like
- test files

A description that is there but empty is [AZS011](../AZS011_description_blank)'s to report. A property whose whole schema comes from a helper in another module, such as `commonschema.Location()`, is not checked here.

## Fix

None is offered. Write the description.

## Ignoring Reports

Put `//azignore:AZS010 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"name": { //azignore:AZS010 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.

A provider that has not written its descriptions yet gets a report for nearly every property. Switch the rule off with `disable: [AZS010]` until it has; see the [root README](../../../README.md).
