# AZT006 - update tests must apply two different configs

AZT006 reports a test named for an update step that applies a single config, or the same config more than once.

An update test exists to run the resource's Update path. With one config it only proves the resource can be created, and with the same config twice the second apply is a no-op plan. Either way the step the test is named for never runs, and the name says it did.

## Flagged Code

```go
func TestAccThing_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
		},
		data.ImportStep(),
	})
}
```

## Passing Code

```go
func TestAccThing_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
		},
		data.ImportStep(),
	})
}
```

## What counts

A test is an update test when a segment of its name after the first underscore starts with `update`: `TestAccThing_update`, `TestAccThing_updateSku`, `testAccThing_list_update`. `TestAccThing_autoUpdateEnabled` is a feature name, not the step, and is left alone. Lowercase `testAcc` helpers count, since sequential tests keep their steps there.

The configs are the `Config` values of the `TestStep` literals in the function body, compared as written. The same helper with different arguments is a different config.

Not reported:

- a test with no config steps of its own, which delegates to a helper; the helper is checked instead
- files other than `_test.go`

## No fix

AZT006 only reports. Add a second step with a changed config, or rename the test if it was never meant to update.

## Ignoring Reports

Put `//azignore:AZT006 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
func TestAccThing_update(t *testing.T) { //azignore:AZT006 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
