# AZS028 - a list or set of blocks must have a singular name, and a list or set of values a plural one

AZS028 reports a list or set property whose name does not say how it is written.

A user guesses how to write a property from its name: `rule` is a block written once per rule, `zones` takes several values.

## Flagged Code

```go
"rules": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem:     &pluginsdk.Resource{...},
},

"zone": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
},
```

```text
block "rules" has a plural name - a block is one thing each time it is written, so name it in the singular
list "zone" has a singular name - it holds many values, so name it in the plural
```

## Passing Code

```go
"rule": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem:     &pluginsdk.Resource{...},
},

"zones": {
	Type:     pluginsdk.TypeList,
	Optional: true,
	Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
},
```

## What is checked

- `block-singular`: a plugin SDK list or set of blocks that the user sets, with a plural name. One that is only read back is a list of them, and is left alone.
- `list-plural`: a plugin SDK list or set of plain values with a singular name.

A word is plural when it ends in `s`, but not in `ss`, `us` or `is`. The rule knows these words are not plurals: `alias`, `analytics`, `aws`, `contains`, `cors`, `diagnostics`, `dns`, `equals`, `https`, `insights`, `kms`, `kubernetes`, `nfs`, `os`, `redis`, `sas`, `series`, `tls`, `windows`. And that these are: `apis`, `cpus`, `criteria`, `gpus`, `uris`, `vcpus`. `data`, `list`, `media` and `metadata` have no other form, and are fine as either.

A block is named by its last word. A list is plural when its last word is, or the word before one of these: `after`, `allowed`, `base64`, `for`, `from`, `in`, `not`, `of`, `only`, `to`, `with`. So `days_of_week` and `tables_to_include` are plural.

These are not checked:

- a property that is `Deprecated`
- a list whose `Elem` comes from a function
- state migrations

Renaming a property that has shipped is a breaking change, so this is for new schema.

## Fix

None is offered.

## Options

Each kind can be switched off on its own. `singular-word` and `plural-word` add to the words the rule knows, and a word given as both is fine as either. `trailing-word` adds a word that may come after the plural in a list's name.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS028:
            block-singular: false
            singular-word: [settings]
            plural-word: [sheep]
            trailing-word: [dark]
```

```sh
azproviderlint -AZS028 -AZS028.singular-word=settings -AZS028.plural-word=sheep ./...
```

## Ignoring Reports

Put `//azignore:AZS028 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"zone": { //azignore:AZS028 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
