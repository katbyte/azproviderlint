# AZS005 - registered resources must have a same-named data source

AZS005 reports a registered resource with no data source of the same Terraform type name in the service package.

Most resources should have a data source so other configs can look them up. Not every resource needs one, and suppressions are expected. The point of the check is that skipping the data source is a decision someone made and wrote down, not something that was forgotten.

## Flagged Code

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_example": resourceExample(), // no data source named azurerm_example registered
	}
}
```

## Passing Code

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_example": resourceExample(),
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_example": dataSourceExample(),
	}
}
```

## What counts

All three registration styles are read, and a resource registered one way is matched by a data source registered any other way:

- untyped plugin SDK: keys of the `SupportedResources()` / `SupportedDataSources()` maps
- typed SDK: elements of `Resources()` / `DataSources()`, named by their `ResourceType()` methods
- framework: elements of `FrameworkResources()` / `FrameworkDataSources()`, named the same way

Also handled:

- entries added conditionally (`resources["azurerm_x"] = ...` or `append(out, FooResource{})` behind a feature flag)
- generated auto-registration (`append(out, r.autoRegistration.Resources()...)`), whose entries are picked up from the generated methods
- type names held in package variables (`var FooResourceName = "azurerm_foo"`) and named constants

Never reported: action-style resources that run an operation or mint a credential rather than manage something durable, recognised today by the `_run_command` and `_sas_token` suffixes.

If any data source name in the package cannot be resolved, the whole package is skipped, so an unreadable data source can never cause a false "missing" report. An unresolvable resource entry is skipped on its own.

## Ignoring Reports

Put `//azignore:AZS005 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
"azurerm_example": resourceExample(), //azignore:AZS005 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
