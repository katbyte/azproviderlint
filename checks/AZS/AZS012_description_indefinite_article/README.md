# AZS012 - resource and property descriptions must use the right one of "a" and "an"

AZS012 reports `a` where `an` belongs, and the other way round, in resource and property descriptions.

Which one is right depends on how the next word is said, not how it is spelt: `an hour`, `a user`, `an SSL certificate`, `a URL`.

## Flagged Code

```go
"count": {
	Type:        pluginsdk.TypeInt,
	Description: "A optional count, as a integer, for an Certificate or a API.",
},
```

```text
"A optional" should be "An optional"
"a integer" should be "an integer"
"an Certificate" should be "a Certificate"
"a API" should be "an API"
```

## Passing Code

```go
"count": {
	Type:        pluginsdk.TypeInt,
	Description: "An optional count, as an integer, for a Certificate or an API.",
},
```

## How the next word is read

| The next word | Takes | Examples |
|---|---|---|
| starts with a vowel | `an` | `an integer`, `an Automation Account`, `an unused` |
| starts with a "you" or "won" sound | `a` | `a user`, `a unique`, `a utility`, `a European`, `a one-time` |
| starts with a consonant | `a` | `a Certificate`, `a hybrid`, `a year` |
| starts with a silent h | `an` | `an hour`, `an honest`, `an heir` |
| is an acronym said letter by letter | by the name of its first letter: `an` for A E F H I L M N O R S X | `an SSL`, `an HTTP`, `an FQDN`, `an MSI`, `a URL`, `a VM`, `a DNS` |
| is an acronym said as a word | by its sound | `a NAT`, `a NIC`, `a SAS`, `a SAML`, `a RAM` |
| is a letter and a number or a capital | by the name of the letter | `an S3`, `an x509`, `an mTLS`, `a vCore` |
| is a number | by how it is read out | `an 8`, `an 11`, `an 18`, `an 80`, `a 1`, `a 110` |
| starts with a dot | `a` | `a .NET` |

Only the letters a word opens with are read, so `user-assigned` is `user` and `HDInsight` is `HD`.

It only reports where it is sure. These are left alone:

- acronyms said both ways: `SQL`, `SKU`, `SNAT`, `SHA`, `SPA`, `FAQ`
- `herb` and `historic`, which differ by dialect
- an acronym with a vowel in it that it has not been taught, since `a NOVA` and `an NVA` cannot be told apart by their letters
- a word in backticks, and anything that does not start with a letter, a number or a dot
- a capital `A` in the middle of a sentence, and an `a` before `or`, `and`, `is` or `to`, which is a letter or a name: `a type A or AAAA record`

## What counts

A description is text written out in the code, outside test files, in one of these places:

- `Description` on a plugin SDK schema property or resource (`pluginsdk.Schema`, `pluginsdk.Resource`)
- `Description` or `MarkdownDescription` on a framework schema, attribute or block
- `Description` or `MarkdownDescription` on a provider function or one of its parameters
- what a typed resource or data source returns from `Description() string`

Provider and `features` block descriptions count too. Text built when the provider runs, such as with `fmt.Sprintf`, is not checked.

## Fix

The fix swaps the article and keeps its capital.

## Options

| Option | Default | Effect |
|---|---|---|
| `a` | | words that take `a` |
| `an` | | words that take `an` |

Both take several values: a list in the golangci settings, or the flag repeated on the command line. They teach it an acronym it leaves alone, or overrule one it gets wrong. Case does not matter.

```yaml
          AZS012:
            a: [NOVA, SQL]
            an: [NVA]
```

```bash
azproviderlint -AZS012 -AZS012.a=NOVA -AZS012.a=SQL -AZS012.an=NVA ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS012 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Either a A or a B record.", //azignore:AZS012 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
