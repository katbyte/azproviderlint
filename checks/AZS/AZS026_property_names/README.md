# AZS026 - a bool must end in _enabled, a block must have a singular name, and a list of values a plural one

AZS026 reports a property whose name does not say what it holds the way the provider's others do.

A user guesses how to write a property from its name: `compression_enabled` is on or off, `rule` is written once per rule, `zones` takes several values.

## Flagged Code

```go
"zone_redundant": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},

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
bool property "zone_redundant" should end in _enabled
block "rules" has a plural name - a block is one thing each time it is written, so name it in the singular
list "zone" has a singular name - it holds many values, so name it in the plural
```

## Passing Code

```go
"zone_redundancy_enabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},

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

- `bool-enabled`: a bool, in a plugin SDK or framework schema, whose name is not `enabled` and does not end in `_enabled`.
- `block-singular`: a plugin SDK list or set of blocks that the user sets, with a plural name. One that is only read back is a list of them, and is left alone.
- `list-plural`: a plugin SDK list or set of plain values with a singular name.

A name is plural when its last word ends in `s`, but not in `ss`, `us` or `is`. The rule knows these words are not plurals: `alias`, `aws`, `cors`, `dns`, `https`, `kms`, `kubernetes`, `nfs`, `os`, `redis`, `sas`, `series`, `tls`. And that these are: `criteria`, `data`, `media`, `metadata`.

These are not checked:

- a property that is `Deprecated`
- a list whose `Elem` comes from a function
- state migrations

Renaming a property that has shipped is a breaking change, so this is for new schema.

## Fix

None is offered.

## Options

Each kind can be switched off on its own. `bool-ending` is what a bool may end in, and replaces `_enabled`. `singular-word` and `plural-word` add to the words the rule knows.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS026:
            block-singular: false
            bool-ending: [_enabled, _required, _accepted]
            singular-word: [settings]
            plural-word: [sheep]
```

```sh
azproviderlint -AZS026 -AZS026.bool-ending=_enabled -AZS026.bool-ending=_required ./...
```

## Ignoring Reports

Put `//azignore:AZS026 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"zone_redundant": { //azignore:AZS026 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
