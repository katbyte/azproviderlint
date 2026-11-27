# AZS026 - a bool property must end in _enabled, or a word for its state such as _required

AZS026 reports a bool property whose name does not end in `_enabled` or a word for its state.

A user can then tell from the name alone that the property is on or off: `compression_enabled`, or `mtls_required` where `_enabled` would read oddly. azurerm's naming guide asks for this.

## Flagged Code

```go
"zone_redundant": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},

"public_access_disabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},
```

```text
bool property "zone_redundant" should end in _enabled, _required, _accepted, _allowed, _supported, _verified, _configured or _only
bool property "public_access_disabled" should end in _enabled, _required, _accepted, _allowed, _supported, _verified, _configured or _only
```

## Passing Code

```go
"zone_redundancy_enabled": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},

"mtls_required": {
	Type:     pluginsdk.TypeBool,
	Optional: true,
},
```

## What is checked

A bool, in a plugin SDK or framework schema, whose name does not end in one of `_enabled`, `_required`, `_accepted`, `_allowed`, `_supported`, `_verified`, `_configured` or `_only`, and is not one of those words on its own, such as `enabled`.

`_disabled` is not allowed: the guide asks for a bool the API calls disable to be turned around and named for what it turns on.

A property whose schema a function hands back, such as `"zone_redundant": helpers.ZoneRedundant()`, is checked where it is named, in whichever package the function is.

These are not checked:

- a property that is `Deprecated`
- state migrations

Renaming a property that has shipped is a breaking change, so this is for new schema.

## Fix

None is offered.

## Options

`ending` is what a bool's name may end in. Each one given replaces the default list.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS026:
            ending: [_enabled, _required, _accepted]
```

```sh
azproviderlint -AZS026 -AZS026.ending=_enabled -AZS026.ending=_required ./...
```

## Ignoring Reports

Put `//azignore:AZS026 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
"zone_redundant": { //azignore:AZS026 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
