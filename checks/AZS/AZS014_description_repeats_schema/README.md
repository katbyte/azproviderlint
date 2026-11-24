# AZS014 - resource and property descriptions must not repeat what the schema says

AZS014 reports text in a resource or property description that the schema already states: a tag such as `(Optional)` or `(Boolean)`, the default, the values it can take, how long or how many it can be, what it conflicts with or needs, that a change forces a new resource.

The docs take those from the schema. A description that repeats them says it twice, and the copy goes stale when the schema changes.

On a read-only property nothing can be set, so the values it can take, a default and the rest are reported whether or not the schema says them.

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

```go
"state": {
	Type:        pluginsdk.TypeString,
	Computed:    true,
	Description: "The state of the Thing. Possible values include `Running` and `Stopped`. Defaults to `Running`.",
},
```

```text
description lists the values it can take - a read-only property cannot be set
description gives the default - a read-only property cannot be set
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

"state": {
	Type:        pluginsdk.TypeString,
	Computed:    true,
	Description: "The state of the Thing.",
},
```

## What is checked

Each kind of repeat has a name, and can be switched off by it.

| Kind | Reports | When | Example |
|---|---|---|---|
| `tag` | a tag copied from the docs | always | `(Optional)`, `(Boolean)`, `(Optional, Sensitive)` |
| `possible-values` | a sentence listing the values a property takes, or the list at the end of one | its validation lists or bounds them, or it is a bool | `Possible values are...`, `Acceptable values include...`, `Values are a, b and c.`, `Valid options are...`, `Can be a or b.`, `Only a is supported.`, `Must be between...`, `The range is 1 - 5.`, `... - can be either a or b.`, `..., possible values are a and b.`, `Currently the only possible value is a.`, `..., one of a or b.`, `..., must be one or more from the following: a and b.` |
| `default` | a sentence giving the default, or the default at the end of one or in brackets | always; it is taken out only when the property sets `Default` | `Defaults to...`, `If blank, defaults to...`, `..., which defaults to Fast.`, `(defaults to Fast)` |
| `force-new` | a sentence saying a change forces a new resource | the property is `ForceNew` | `Changing this forces a new resource.`, `Changing the name forces a new Thing to be created.` |
| `conflicts-with` | a sentence saying what the property conflicts with | the property sets `ConflictsWith` or `ExactlyOneOf` | `Conflicts with...`, `Cannot be specified with...`, `Only one of a or b can be specified.` |
| `one-of` | a sentence saying one of several properties must be set | the property sets `ExactlyOneOf` or `AtLeastOneOf` | `Exactly one of a or b must be specified.`, `Either a or both b and c must be specified.` |
| `required-with` | a sentence saying what else must be set | the property sets `RequiredWith` | `Required with a.`, `Requires a to be set.`, `a and b must also be set.`, `The a and b properties must be specified together or not at all.`, `If a is specified, b is required.`, `Only valid in conjunction with a.` |
| `item-count` | a sentence saying how many items or blocks there can be | the property sets `MaxItems` or `MinItems` | `A maximum of 100 rule blocks may be defined.` |
| `length` | how long a value can be, as a sentence or inside one | its validation gives a length | `Must be between 1 and 80 characters.`, `...up to 80 characters in length` |
| `defined-below` | words or a sentence that point at the docs page | always | `as defined below`, `Each rule block supports fields defined below.` |
| `read-only-setting` | a read-only property described as something to set | the property is read-only | `Whether to keep the path.`, `Set to true to...`, `Whether the Thing should be enabled.` |
| `unchecked-range` | a sentence giving a range or a length | nothing in the schema checks it | `This value must be between 5 and 300 seconds.`, `The range is 100 - 65000.`, `Must be between 1 and 80 characters.` |

The tags are `Required`, `Optional`, `Computed`, `Sensitive`, `Deprecated`, `ForceNew`, `Read-Only`, `Write-Only`, `Boolean`, `Bool`, `String`, `Number`, `Integer`, `Int`, `Float`, `List`, `Set`, `Map`, `Block` and `Object`, in brackets and capitalised, alone or several together.

Validation counts when it is a list or a range, such as `StringInSlice`, `IntBetween` or `ISO8601DurationBetween`, on the property or on the element of its list, or a length, such as `StringLenBetween`. The framework's validators and plan modifiers that say the same things count too.

A read-only property is `Computed` and neither `Optional` nor `Required`. Every kind but `tag` and `defined-below` is reported on it and taken out, whatever its schema says.

A read-only property holds what the provider fills in, so its description says what that is: `Whether the path is kept.`, not `Whether to keep the path.`

A description handed to a helper is checked against the schema the helper sets it on, in whichever package the helper is.

A default is always reported. The fix takes it out when the property sets `Default`. When it does not, set `Default`, take the sentence out, or ignore the report with the reason it has to stay.

A range or a length belongs in the schema, where it is enforced. It is reported, and left in place, when the property has no validation, or only one that checks its format, such as `ISO8601Duration` or `StringIsNotEmpty`. The report names the validation to use where it can: `ISO8601DurationBetween`, `IntBetween`, `FloatBetween` or `StringLenBetween`. A property validated by a function of the provider's own is left alone, since that function may check the range.

These pass on a property that can be set, because the description is the only place it is written down:

- `Possible values are...`, `Changing this forces...`, a count or `Exactly one of...` on a property whose schema does not say it
- a list of values that only holds some of the time: `If mode is Fast, possible values are a and b.`
- a sentence that says more than the schema does: `Only one rule block can be specified unless mode is Fast.`
- something needed only for some values, which the schema cannot say: `Required when mode is Fast.`

## What counts

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

Only a property has a schema beside it to repeat. Anything else is checked just for tags and for words that point at the docs page.

## Fix

The tag, the words or the whole sentence is removed. There is no fix when:

- every sentence of the description repeats the schema, so nothing would be left
- a length sits inside a sentence that says more, which has to be reworded by hand
- the property can be set and sets no `Default`
- a range or a length is not checked by the schema, which needs a validation first
- a read-only property is described as something to set, which has to be reworded
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
