# AZS027 - name, resource_group_name and location must come first in a resource schema, and tags last

AZS027 reports a resource whose well-known properties are not where a reader looks for them.

Every resource then starts and ends the same way: what names it and where it lives first, `tags` last.

## Flagged Code

```go
return map[string]*pluginsdk.Schema{
	"tags": commonschema.Tags(),

	"sku_name": {
		Type:     pluginsdk.TypeString,
		Required: true,
	},

	"name": {
		Type:     pluginsdk.TypeString,
		Required: true,
	},

	"resource_group_name": commonschema.ResourceGroupName(),
}
```

```text
property "tags" should come after "sku_name" - it goes last
property "name" should come before "tags" - name, resource_group_name, location go first
property "resource_group_name" should come before "tags" - name, resource_group_name, location go first
```

## Passing Code

```go
return map[string]*pluginsdk.Schema{
	"name": {
		Type:     pluginsdk.TypeString,
		Required: true,
	},

	"resource_group_name": commonschema.ResourceGroupName(),

	"sku_name": {
		Type:     pluginsdk.TypeString,
		Required: true,
	},

	"tags": commonschema.Tags(),
}
```

## What is checked

A resource's own properties, written out as a map in a plugin SDK or framework schema.

- `name`, `resource_group_name` and `location` come before every other property, in that order. A resource need not have all three.
- `tags` comes after every other property.

The order of the properties in between is not checked.

These are left alone:

- a block inside the resource
- a typed resource's `Attributes`, which are only read back
- data source files
- state migrations

## Fix

None is offered.

## Options

`first` is the properties that go first, in order, and `last` the ones that go last. Each replaces its default.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS027:
            first: [name, resource_group_name, location]
            last: [tags]
```

```sh
azproviderlint -AZS027 -AZS027.first=display_name -AZS027.first=name ./...
```

## Ignoring Reports

Put `//azignore:AZS027 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"tags": commonschema.Tags(), //azignore:AZS027 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
