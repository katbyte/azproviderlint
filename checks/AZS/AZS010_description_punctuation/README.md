# AZS010 - resource and property descriptions must be punctuated properly

AZS010 reports punctuation mistakes in resource and property descriptions: no capital at the start, no full stop at the end, and stray or missing spaces and marks in between.

The descriptions are what the provider's documentation is made from, so a slip here is a slip on the page.

## Flagged Code

```go
"mode": {
	Type:        pluginsdk.TypeString,
	Description: "the mode of the Thing.  Possible values are `Fast`,`Slow` and `Off` . defaults to `Off`",
},
```

```text
description should start with a capital letter
description has more than one space in a row
description has no space after ','
description has a space before '.'
sentence should start with a capital letter
description should end with a full stop
```

## Passing Code

```go
"mode": {
	Type:        pluginsdk.TypeString,
	Description: "The mode of the Thing. Possible values are `Fast`, `Slow` and `Off`. Defaults to `Off`.",
},
```

## What is checked

Each kind of mistake has a name, and can be switched off by it.

| Kind | Reports | Example |
|---|---|---|
| `capital` | a description that starts with a lowercase word | `the name of the Thing.` |
| `full-stop` | a description that does not end with a full stop or a question mark | `The name of the Thing` |
| `question-mark` | a question that ends with a full stop | `Should the Thing be enabled.` |
| `sentence-capital` | a later sentence that starts with a lowercase word | `The name. changing this...` |
| `leading-mark` | a dash, bullet or other mark before the first word | `- The name of the Thing.` |
| `edge-space` | a space at the start or the end | `"The name. "` |
| `double-space` | more than one space in a row | `The name.  Changing this...` |
| `line-break` | a tab or a line break inside the description | `"The name.\nChanging this..."` |
| `space-before` | a space before `,` `.` `;` `:` `?` or `!` | `The name of the Thing .` |
| `space-after` | no space after `,` or `;`, after the full stop between two sentences, or after a `:` that runs into code | `` `Fast`,`Slow` `` |
| `bracket-space` | a space just inside a bracket | `( in seconds )` |
| `code-space` | code in backticks that touches the word beside it or other code, or has a space just inside the backticks | `` Defaults to`30` ``, `` `30 ` `` |
| `doubled-punctuation` | two marks in a row | `..` `,.` `?.` |
| `unbalanced` | a backtick, double quote or bracket with no partner | `` one of `v1` or `v2 `` |
| `abbreviation` | `e.g.`, `i.e.` and `etc.` without their full stops | `eg`, `i.e`, `etc` |
| `pasted-characters` | curly quotes, non-breaking and invisible spaces, `…`, and en and em dashes | `“Enabled”` |

Code in backticks, the address a link points to, and web addresses are skipped, so their dots, colons and commas are never punctuation.

These pass:

- a description that starts with code or a mixed-case name: `` `sku_name` is... ``, `iSCSI targets...`
- a description ending in `?`, in `.)`, or in `etc.`
- `e.g.`, `i.e.`, `etc.` and `vs.` in the middle of a sentence
- `...`, `::`, `.NET`, `1,000`, `10:30`, and `a)` or `1)` numbering the points of a list
- a plural made from code, `` `subnet`s ``, and a space that is the whole value, `` ` ` ``

A description is a question when it opens with `Should`, `Is`, `Are`, `Does`, `Do`, `Can`, `Must`, `Has`, `Have`, `Will`, `Was` or `Were`. Its first sentence is the question, so `Should the Thing be enabled. Defaults to true.` is reported too, and so is one behind a tag such as `(Boolean)`.

It is a statement when the next word leaves the subject out: `Must be between 1 and 5.`, `Can only be set when...`, `Is required when...`, `Has no effect...`.

A description ending in code or a link is reported as missing its full stop, since the code is not the end of the sentence.

## What counts

A description is text written out in the code, outside test files, in one of these places:

- `Description` on a plugin SDK schema property or resource (`pluginsdk.Schema`, `pluginsdk.Resource`)
- `Description` or `MarkdownDescription` on a framework schema, attribute or block
- `Description` or `MarkdownDescription` on a provider function or one of its parameters
- what a typed resource or data source returns from `Description() string`

Provider and `features` block descriptions count too. Text built when the provider runs, such as with `fmt.Sprintf`, is not checked.

## Fix

Most kinds have a fix that changes only the characters involved. There is none where the right answer is a judgement:

- an unbalanced backtick, quote or bracket
- two pieces of code with nothing between them
- a description ending in `,`, `;`, `:` or `!`
- mixed marks such as `,.`
- a `;`, `:` or `,` on its own at the end, which is usually a value that belongs in backticks: `` separated by `;` ``
- an em dash between two words with no spaces
- a description that is a constant or is joined from parts, which is reported as a whole

## Options

| Option | Default | Effect |
|---|---|---|
| each kind in the table above | `true` | `false` stops reporting that kind |

```yaml
          AZS010:
            question-mark: false
            pasted-characters: false
```

```bash
azproviderlint -AZS010 -AZS010.question-mark=false -AZS010.pasted-characters=false ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS010 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "The paths to skip, separated by ;", //azignore:AZS010 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
