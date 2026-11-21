# AZS014 - resource and property descriptions must not repeat what the schema says

AZS014 reports text in a resource or property description that the schema already states: a tag such as `(Optional)` or `(Boolean)`, the default, the values it can take, that a change forces a new resource.

The docs take those from the schema. A description that repeats them says it twice, and the copy goes stale when the schema changes.

## Flagged Code

```go
"mode": {
	Type:         pluginsdk.TypeString,
	Optional:     true,
	ForceNew:     true,
	Default:      "Fast",
	ValidateFunc: validation.StringInSlice([]string{"Fast", "Slow"}, false),
	Description:  "(Optional) The mode of the Thing. Possible values are `Fast` and `Slow`. Defaults to `Fast`. Changing this forces a new Thing to be created.",
},
```

```text
description has the tag "(Optional)" - the schema already says so
description lists the values it can take - the schema's validation already does
description gives the default - the schema's Default already does
description says a change forces a new resource - the schema's ForceNew already does
```

## Passing Code

```go
"mode": {
	Type:         pluginsdk.TypeString,
	Optional:     true,
	ForceNew:     true,
	Default:      "Fast",
	ValidateFunc: validation.StringInSlice([]string{"Fast", "Slow"}, false),
	Description:  "The mode of the Thing.",
},
```

## What is checked

Each kind of repeat has a name, and can be switched off by it.

| Kind | Reports | When | Example |
|---|---|---|---|
| `tag` | a tag copied from the docs | always | `(Optional)`, `(Boolean)`, `(Optional, Sensitive)` |
| `possible-values` | a sentence listing the values a property takes | its validation lists or bounds them, or it is a bool | `Possible values are...`, `Must be between...` |
| `default` | a sentence giving the default | the property sets `Default` | `Defaults to...`, `The default is...` |
| `force-new` | a sentence saying a change forces a new resource | the property is `ForceNew` | `Changing this forces a new Thing to be created.` |
| `conflicts-with` | a sentence saying what the property conflicts with | the property sets `ConflictsWith` | `Conflicts with...` |
| `defined-below` | words that point at the docs page | always | `as defined below`, `described above` |

The tags are `Required`, `Optional`, `Computed`, `Sensitive`, `Deprecated`, `ForceNew`, `Read-Only`, `Write-Only`, `Boolean`, `Bool`, `String`, `Number`, `Integer`, `Int`, `Float`, `List`, `Set`, `Map`, `Block` and `Object`, in brackets and capitalised, alone or several together.

The validation that counts is a list or a range: `StringInSlice`, `IntInSlice`, `FloatInSlice`, `IntBetween`, `IntAtLeast`, `IntAtMost` and the `Float` ones, on the property or on the element of its list, and the framework's `OneOf`, `Between`, `AtLeast` and `AtMost`. In the framework a property is `ForceNew` when a plan modifier is `RequiresReplace`.

These pass, because the description is the only place it is written down:

- `Defaults to...` on a property with no `Default`, such as one the API fills in
- `Possible values are...` on a property with no validation, or with validation that is not a list or a range
- `Changing this forces...` on a property that is not `ForceNew`
- the same said inside a longer sentence: `The mode, which defaults to Fast.`

## What counts

A description is text written out in the code, outside test files, in one of these places:

- `Description` on a plugin SDK schema property, block or resource (`pluginsdk.Schema`, `pluginsdk.Resource`)
- `Description` or `MarkdownDescription` on a framework schema, attribute or block
- `Description` or `MarkdownDescription` on a provider function or one of its parameters
- what a typed resource or data source returns from `Description() string`

Only the first two have a schema beside them to repeat. The others are checked for tags and for words that point at the docs page.

## Fix

The tag, the words or the whole sentence is removed. There is no fix when:

- every sentence of the description repeats the schema, so nothing would be left
- the description is a constant or is joined from parts, which is reported as a whole

## Options

| Option | Default | Effect |
|---|---|---|
| each kind in the table above | `true` | `false` stops reporting that kind |

```yaml
          AZS014:
            conflicts-with: false
            defined-below: false
```

```bash
azproviderlint -AZS014 -AZS014.conflicts-with=false -AZS014.defined-below=false ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS014 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "The mode of the Thing. Defaults to `Fast`.", //azignore:AZS014 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
