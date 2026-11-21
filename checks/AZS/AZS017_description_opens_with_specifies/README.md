# AZS017 - resource and property descriptions must not open with "Specifies"

AZS017 reports a resource or property description that opens with a verb which only says it is a description, such as `Specifies the name of the Thing.`, where most of the provider says what the property is: `The name of the Thing.`

## Flagged Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "Specifies the name of the Webhook.",
},

"http2_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Specifies if the HTTP2 protocol should be enabled.",
},
```

```text
description opens with "Specifies the" - open with "The"
description opens with "Specifies if" - open with "Whether"
```

## Passing Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "The name of the Webhook.",
},

"http2_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Whether the HTTP2 protocol should be enabled.",
},
```

## What counts

A description whose first word is one of these verbs, in any case:

`Specifies` `Specify` `Indicates` `Denotes` `Defines` `Determines`

Only the opening is checked: `The name of the Thing. Specifies the Thing to use.` passes. A verb that says what the property does, such as `Configures`, is not on the list.

A description is text written out in the code, outside test files: `Description` on a plugin SDK schema property, block or resource, `Description` or `MarkdownDescription` on a framework schema, attribute or block, or on a provider function or one of its parameters, and what a typed resource or data source returns from `Description() string`. Provider and `features` block descriptions count too.

## Fix

The fix takes the verb out and opens with the word that followed it:

| After the verb | Opens with |
|---|---|
| `the`, `a`, `an` | `The`, `A`, `An` |
| `whether`, `if` | `Whether` |
| `which`, `how`, `where`, `what` | `Which`, `How`, `Where`, `What` |

After any other word there is no fix, since what is left would not open a description: `Specifies interval in minutes…` has to be reworded by hand.

## Options

| Option | Default | Effect |
|---|---|---|
| `verb` | | more opening verbs to report |
| `ignore` | | verbs from the list above to leave alone |

Both take several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZS017:
            verb: [Configures, Controls]
            ignore: [Defines]
```

```bash
azproviderlint -AZS017 -AZS017.verb=Configures -AZS017.verb=Controls -AZS017.ignore=Defines ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS017 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Specifies the name of the Webhook.", //azignore:AZS017 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
