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
- a property that is part of the resource's ID may sit among them. The rule takes one ending in `_name` or `_id`, such as `lab_name` or `workspace_id`, to be, and one that is `Required` and `ForceNew`. It cannot read the ID itself, so a property that only looks the part passes.
- a `name` that is only read back is not the one the user gives, and goes where the read-only properties do.
- `tags` comes after every other property the user sets. Properties that are only read back may follow it.

The order of the properties in between is not checked.

A property whose schema a function hands back, such as `commonschema.LocationComputed()`, is read for whether it is only read back, in whichever package the function is.

These are left alone:

- a block inside the resource, and what a block holds when a function builds it: a `Resource` with no `Read`
- a typed resource's `Attributes`, which are only read back
- data source files, and a function with `DataSource` in its name
- state migrations

## Fix

None is offered.

## Options

`first` is the properties that go first, in order, and `last` the ones that go last. `parent-ending` is what a property that may sit among the first ends in. Each replaces its default.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS027:
            first: [name, resource_group_name, location]
            last: [tags]
            parent-ending: [_name, _id]
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
