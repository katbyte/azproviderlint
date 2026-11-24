# AZS017 - resource and property descriptions must not have extra words

AZS017 reports words in a resource or property description that add nothing: a verb it opens with that only says it is a description, such as `Specifies`, and wordy phrases such as `or not` and `in order to`.

## Flagged Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "Specifies the name of the Webhook.",
},

"caching_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Whether to enable caching or not.",
},
```

```text
description opens with "Specifies the" - open with "The"
description has "or not" - it adds nothing
```

## Passing Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "The name of the Webhook.",
},

"caching_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Whether to enable caching.",
},
```

## What counts

**A verb the description opens with**, in any case:

`Specifies` `Specify` `Indicates` `Denotes` `Defines` `Determines`

Only the opening is checked: `The name of the Thing. Specifies the Thing to use.` passes. A verb that says what the property does, such as `Configures`, is not on the list.

**A wordy phrase**, anywhere and in any case:

| Phrase | Becomes |
|---|---|
| `or not`, after `whether` or closing a clause | removed |
| `in order to`, `in order for` | `to`, `for` |
| `due to the fact that` | `because` |
| `in the event that` | `if` |
| `for the purpose of` | `for` |
| `is able to`, `are able to` | `can` |
| `prior to` | `before` |
| `please`, `please note`, `please note that`, `note that`, `it should be noted that` | removed |

`empty or not set` passes, and so does a phrase inside code in backticks.

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Fix

An opening verb is taken out, and the description opens with the word that followed it:

| After the verb | Opens with |
|---|---|
| `the`, `a`, `an` | `The`, `A`, `An` |
| `whether`, `if` | `Whether` |
| `which`, `how`, `where`, `what` | `Which`, `How`, `Where`, `What` |

After any other word there is no fix, since what is left would not open a description: `Specifies interval in minutes…` has to be reworded by hand.

A phrase is swapped for its shorter form or removed, and the next word gets a capital where the phrase opened a sentence. There is no fix where the result would not read: `agree to in order to sign up`, `Please note: ...`.

## Options

| Option | Default | Effect |
|---|---|---|
| `verb` | | more opening verbs to report |
| `phrase` | | more phrases to report, as `phrase=shorter form`; with nothing after the `=` it is removed |
| `ignore` | | verbs and phrases from the lists above to leave alone |

Each takes several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZS017:
            verb: [Configures, Controls]
            phrase: ["at this time=now", "basically="]
            ignore: [Defines, prior to]
```

```bash
azproviderlint -AZS017 -AZS017.verb=Configures -AZS017.phrase="at this time=now" -AZS017.ignore=Defines ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS017 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Specifies the name of the Webhook.", //azignore:AZS017 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
