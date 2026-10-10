# AZT008 - acceptance tests must be in a file named after the resource's file

AZT008 reports an acceptance test that sits in a test file not named after the file declaring the resource or data source it tests.

Tests for the resource declared in `lb_probe_resource.go` belong in `lb_probe_resource_test.go`. When they live in `loadbalancer_probe_resource_test.go` nobody reading the resource finds them, tooling that pairs a resource with its tests by file name misses them, and a test that drifted into another resource's file goes unnoticed.

## Flagged Code

```go
// lb_probe_resource.go declares the resource registered as "azurerm_lb_probe"

// loadbalancer_probe_resource_test.go
func TestAccAzureRMLoadBalancerProbe_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	// ...
}
```

```text
TestAccAzureRMLoadBalancerProbe_basic and 5 more in this file build test data for resource "azurerm_lb_probe", declared in lb_probe_resource.go, so they belong in lb_probe_resource_test.go or lb_probe_resource_<part>_test.go
```

## Passing Code

```go
// lb_probe_resource_test.go, or lb_probe_resource_identity_gen_test.go
func TestAccLbProbe_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	// ...
}
```

## What counts

What a test tests is what it builds test data for: the second argument of `acceptance.BuildTestData`. Every function in a `_test.go` file that makes that call is checked, lowercase helpers included.

- `"azurerm_foo"` is the resource and `"data.azurerm_foo"` the data source; a constant works as well as a literal
- `FooResource{}.ResourceType()` is whatever that type is registered as, also inside a `fmt.Sprintf("data.%s", ...)`

The declaring file is the one holding the function or type that the service's registration registers under that name: `resourceFoo()` in `SupportedResources()`, `FooResource{}` in `Resources()`, and the data source and framework equivalents. Test files are a separate package, so the service package's own files are read from disk.

A test file is right when its name is the declaring file's name with `_test` added, or with a part in between: `foo_resource_test.go`, `foo_resource_network_test.go`, `foo_resource_identity_gen_test.go`. The part has to come after the whole name, so `foo_network_resource_test.go` is reported.

One report is made per test file and resource, on the first such test, with a count of the rest.

Not reported:

- a test that builds test data for several things, when its file is right for any one of them
- a test run through `DataSourceTest` under the bare resource name, in either the resource's or the data source's test file
- a name nothing in the test's own package registers, ephemeral resources included
- a type that is not known until the test runs, such as a loop variable

## No fix

AZT008 only reports. Rename the test file, or move the test to the file it belongs in. If the test is in the right file and builds test data for the wrong thing, correct the `BuildTestData` argument instead.

## Options

| Option | Default | Effect |
|---|---|---|
| `suffix` | true | also accept a part between the declaring file's name and `_test.go`, so one resource's tests can be split across files; false requires exactly `<file>_test.go` |

Set with `-AZT008.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Put `//azignore:AZT008 - <reason>` at the end of the reported test's line, or on the line above it. The reason is required. One directive covers the whole file's report for that resource.

```go
func TestAccAzureRMLoadBalancerProbe_basic(t *testing.T) { //azignore:AZT008 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
