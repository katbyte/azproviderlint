# AZS018 - resource and property descriptions must not have a note in them

AZS018 reports a note written into a resource or property description, such as `**Note:** only valid with a subnet`, and removes it.

A description says what the property is. A note is a callout of its own on the docs page, and belongs there.

## Flagged Code

```go
"internal_load_balancer_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Should the Container Environment operate in Internal Load Balancing Mode? **Note:** can only be set to `true` if `infrastructure_subnet_id` is specified.",
},
```

```text
description has a note in it - a note belongs in the docs, not the description
```

## Passing Code

```go
"internal_load_balancer_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Should the Container Environment operate in Internal Load Balancing Mode?",
},
```

## What counts

A note opens with `Note`, in either of these ways:

- in bold, anywhere and in any case: `**Note:**`, `**NOTE**:`, `__Note:__`
- plain with a colon, where a sentence starts: `Note:`, `NOTE:`

An arrow from the docs in front, as in `~> **Note:**`, goes with it. Everything from there to the end of the description is the note.

These pass: `A note about this user.`, `Please note these values are only shown once.`, and a note inside code in backticks.

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Fix

The note is removed. Move it to the docs first if it is not there. There is no fix when the note is the whole description.

## Options

| Option | Default | Effect |
|---|---|---|
| `label` | | more words a note opens with |
| `ignore` | | `Note`, to leave it alone |

Both take several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZS018:
            label: [Important, Warning]
```

```bash
azproviderlint -AZS018 -AZS018.label=Important -AZS018.label=Warning ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS018 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "The name. **Note:** it cannot be changed.", //azignore:AZS018 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
