# AZT005 - resources and data sources must have basic, requiresImport, complete and update acceptance tests

AZT005 reports a registered data source with no `basic` acceptance test, and a registered resource missing any of `basic`, `requiresImport`, and, when the resource can be updated in place, `complete` and `update`.

These are the tests every resource is expected to ship with. `basic` proves it can be created and read back, `requiresImport` proves an existing object is not silently adopted, `complete` proves every field round-trips, and `update` proves changes apply without recreating the object.

## Flagged Code

```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_example": resourceExample(), // example_resource_test.go has only TestAccExample_basic
	}
}

func resourceExample() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceExampleCreate,
		Read:   resourceExampleRead,
		Update: resourceExampleUpdate,
		Delete: resourceExampleDelete,
	}
}
```

```text
resource "azurerm_example" is missing acceptance tests: TestAccExample_requiresImport, TestAccExample_complete, TestAccExample_update
```

## Passing Code

```go
// example_resource_test.go
func TestAccExample_basic(t *testing.T)          { /* ... */ }
func TestAccExample_requiresImport(t *testing.T) { /* ... */ }
func TestAccExample_complete(t *testing.T)       { /* ... */ }
func TestAccExample_update(t *testing.T)         { /* ... */ }

// example_data_source_test.go
func TestAccExampleDataSource_basic(t *testing.T) { /* ... */ }
```

## What counts

Every registration style is read: untyped `SupportedResources()` / `SupportedDataSources()` map keys, typed `Resources()` / `DataSources()` elements, and `FrameworkResources()` elements, named by their `ResourceType()` methods. Entries registered behind a feature flag count too.

A resource can be updated when it declares an Update handler: an `Update`, `UpdateContext` or `UpdateWithoutTimeout` key on an untyped resource, or an `Update` method on a typed or framework one. The plugin SDK only allows an Update handler when some field is not ForceNew, so the two are the same thing. A resource without one needs only `basic` and `requiresImport`.

Test files are a separate package, so they are read from disk. Every `TestAcc` or lowercase `testAcc` function counts when it is in a test file named after the file declaring the resource (`example_resource.go` covers `example_resource_test.go` and `example_resource_other_test.go`), or when its name spells the type name without `azurerm_`, ignoring case and underscores (`TestAccMsSqlServer` for `azurerm_mssql_server`, with an older `AzureRM` prefix, a `Resource` suffix, or `DataSource` at either end of a data source name also accepted).

The test name's segments after the first underscore name its steps. A segment that starts with the step name counts, so `TestAccExample_completeWithTags`, `TestAccExample_updateSku` and `TestAccExample_list_basic` all cover their steps.

Not reported:

- entries whose value is declared in another package, or cannot be resolved
- a resource whose Update handler is set to `nil`, which is treated as having none

## The fix

AZT005 only reports. Add the missing tests, named for the steps they cover.

## Ignoring Reports

Put `//azignore:AZT005 - <reason>` at the end of the registration line, or on the line above it. The reason is required.

```go
"azurerm_example": resourceExample(), //azignore:AZT005 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
