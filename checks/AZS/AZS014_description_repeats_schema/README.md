# AZS014 - resource and property descriptions must not repeat what the schema says

AZS014 reports text in a resource or property description that the schema already states: a tag such as `(Optional)` or `(Boolean)`, the default, the values it can take, how long or how many it can be, what it conflicts with or needs, that a change forces a new resource.

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
| `possible-values` | a sentence listing the values a property takes, or the list at the end of one | its validation lists or bounds them, or it is a bool | `Possible values are...`, `Valid options are...`, `Can be a or b.`, `Only a is supported.`, `Must be between...`, `... - can be either a or b.` |
| `default` | a sentence giving the default, or the default at the end of one or in brackets | always; it is taken out only when the property sets `Default` | `Defaults to...`, `If blank, defaults to...`, `..., which defaults to Fast.`, `(defaults to Fast)` |
| `force-new` | a sentence saying a change forces a new resource | the property is `ForceNew` | `Changing this forces a new resource.`, `Changing the name forces a new Thing to be created.` |
| `conflicts-with` | a sentence saying what the property conflicts with | the property sets `ConflictsWith` or `ExactlyOneOf` | `Conflicts with...`, `Cannot be specified with...`, `Only one of a or b can be specified.` |
| `one-of` | a sentence saying one of several properties must be set | the property sets `ExactlyOneOf` or `AtLeastOneOf` | `Exactly one of a or b must be specified.` |
| `required-with` | a sentence saying what else must be set | the property sets `RequiredWith` | `Required with a.`, `Requires a to be set.`, `a and b must also be set.`, `Only valid in conjunction with a.` |
| `item-count` | a sentence saying how many items or blocks there can be | the property sets `MaxItems` or `MinItems` | `A maximum of 100 rule blocks may be defined.` |
| `length` | how long a value can be, as a sentence or inside one | its validation gives a length | `Must be between 1 and 80 characters.`, `...up to 80 characters in length` |
| `defined-below` | words or a sentence that point at the docs page | always | `as defined below`, `Each rule block supports fields defined below.` |

The tags are `Required`, `Optional`, `Computed`, `Sensitive`, `Deprecated`, `ForceNew`, `Read-Only`, `Write-Only`, `Boolean`, `Bool`, `String`, `Number`, `Integer`, `Int`, `Float`, `List`, `Set`, `Map`, `Block` and `Object`, in brackets and capitalised, alone or several together.

The validation that counts is a list or a range: `StringInSlice`, `IntInSlice`, `FloatInSlice`, `IntBetween`, `IntAtLeast`, `IntAtMost` and the `Float` ones, on the property or on the element of its list, and the framework's `OneOf`, `Between`, `AtLeast` and `AtMost`. A length comes from `StringLenBetween`, or the framework's `LengthBetween`, `LengthAtLeast` and `LengthAtMost`.

In the framework a property is `ForceNew` when a plan modifier is `RequiresReplace`, and its item count, one-of and required-with rules are the `SizeAtMost`, `SizeAtLeast`, `SizeBetween`, `ExactlyOneOf`, `AtLeastOneOf` and `AlsoRequires` validators.

A default is the schema's to give, so one written in a description is always reported. When the property sets `Default` the fix takes it out. When it does not, the report has no fix: set `Default` in the schema, or drop the sentence.

These pass, because the description is the only place it is written down:

- `Possible values are...` on a property with no validation, or with validation that is not a list or a range
- `Changing this forces...` on a property that is not `ForceNew`
- `A maximum of 100 rule blocks...`, `Exactly one of...` and a length on a property whose schema does not set them
- a sentence that says more than the schema does: `Only one rule block can be specified unless mode is Fast.`
- something needed only for some values, which the schema cannot say: `Required when mode is Fast.`

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
- a length sits inside a sentence that says more, which has to be reworded by hand
- the property sets no `Default`, so the description is the only place the default is written down
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
