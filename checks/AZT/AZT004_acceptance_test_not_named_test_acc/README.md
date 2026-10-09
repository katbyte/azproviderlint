# AZT004 - acceptance tests must be named TestAcc

AZT004 reports test functions that run an acceptance test but are not named `TestAcc...`.

Acceptance runs pick tests by that prefix. A test named `TestThing_basic` that drives the acceptance framework is skipped by every acceptance run, so the resource it covers is never actually tested against Azure, however complete the test looks.

A function counts as an acceptance test when it does one of these, itself or through another function in the same package:

- hands `t` to the provider's `internal/acceptance` package (`acceptance.BuildTestData(t, ...)`, `data.ResourceTest(t, ...)`, `acceptance.RunTestsInSequence(t, ...)`)
- hands `t` to the Terraform test harness (`resource.Test`, `resource.ParallelTest`)
- looks at `TF_ACC`

Only functions go test runs are checked: a top-level `func TestXxx(t *testing.T)` in a `_test.go` file. Lowercase helpers and methods are left alone, since no runner selects them by name.

## Flagged Code

```go
func TestThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
		data.ImportStep(),
	})
}
```

## Passing Code

```go
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
		data.ImportStep(),
	})
}
```

## No fix

AZT004 only reports. Rename the function so it starts with `TestAcc`. When `Acc` already sits elsewhere in the name (`TestWebAppAccActiveSlot_basic`), move it to the front rather than adding a second one.

This is the mirror of [AZT003](../AZT003_misnamed_acceptance_test), which reports the opposite: a `TestAcc` name on a test that never runs an acceptance test.

## Ignoring Reports

Put `//azignore:AZT004 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
func TestThing_basic(t *testing.T) { //azignore:AZT004 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
