# AZP004 - registration entries must be sorted alphabetically

AZP004 reports `Registration` methods (`SupportedResources`, `SupportedDataSources`, `Resources`, `DataSources`, and friends) whose map or slice entries are out of alphabetical order, ignoring case.

These lists are where a service says which Terraform types it provides. Keeping them sorted makes them easy to scan, keeps diffs small, and avoids merge conflicts when several PRs add entries at the same time.

Each unsorted section is reported once, at the first entry that is out of place, naming both keys: "`azurerm_disk_encryption_set` should come before `azurerm_managed_disk`".

## Flagged Code

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_managed_disk":     nil,
		"azurerm_availability_set": nil, // should come first alphabetically
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		WorkspaceResource{},
		ApiManagementResource{}, // should come first alphabetically
	}
}
```

## Passing Code

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_availability_set": nil,
		"azurerm_managed_disk":     nil,
	}
}

func (r Registration) SectionedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		// CDN
		"azurerm_cdn_profile": nil,

		// FrontDoor
		"azurerm_cdn_frontdoor_custom_domain": nil,
		"azurerm_cdn_frontdoor_profile":       nil,
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		ApiManagementResource{},
		WorkspaceResource{},
	}
}
```

## What counts

Map and slice literals matching the method's result type are checked whether they are returned directly or assigned to a local first. Literals of other types are ignored.

Sections are sorted independently. A blank line, or a heading comment with a blank line before it, starts a new section, so grouped registrations like the CDN example above are fine. Any other comment belongs to the entry below it.

Both `registration.go` and the generated `registration_gen.go` are checked by default. An unsorted generated file means the generator's input or template needs fixing, not the output.

## Options

| Option | Default | Effect |
|---|---|---|
| `generated` | true | also check generated `registration_gen.go` files |

Set with `-AZP004.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## The fix

`-fix` reorders the entries. Each entry moves as whole lines, so comments attached to it (trailing, directly above, or directly under the opening brace) move with it and section headings stay put. Sections with two entries on one line, or crossed by a multi-line comment, are reported but not rewritten.

## Ignoring Reports

The report is on the method. Put `//azignore:AZP004 - <reason>` at the end of the method's declaration line, or on the line above it. The reason is required. One suppressed method leaves the others checked.

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource { //azignore:AZP004 - intentional ordering
	return map[string]*pluginsdk.Resource{
		"azurerm_managed_disk":     nil,
		"azurerm_availability_set": nil,
	}
}
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
