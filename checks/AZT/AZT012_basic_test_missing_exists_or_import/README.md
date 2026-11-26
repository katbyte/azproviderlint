# AZT012 - a resource's basic test must check the resource exists and import it

AZT012 reports a basic acceptance test that never checks the resource exists, or never imports it.

Applying a config proves only that Terraform did not fail. The exists check proves the resource was made. The import step proves what was saved matches what the API gives back.

## Flagged Code

```go
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
	})
}
```

```text
TestAccThing_basic never checks the resource exists - add check.That(data.ResourceName).ExistsInAzure(r) to a step
TestAccThing_basic never imports the resource - add data.ImportStep() after the config is applied
```

## Passing Code

```go
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}
```

## What is checked

A function named `TestAcc<X>_basic` that hands a list of steps, written out in place, to `data.ResourceTest` or another of the `Resource...Test` runners.

- the steps must call `ExistsInAzure`
- the steps must import: `data.ImportStep()` or another helper whose name starts with `Import`, or a step with `ImportState: true`

These are not checked:

- a data source's test, which has nothing to exist or import
- steps built by a function
- any test not named `_basic`

## Fix

None is offered.

## Ignoring Reports

Put `//azignore:AZT012 - <reason>` at the end of the line with the test's name, or on the line above it. The reason is required.

```go
func TestAccThing_basic(t *testing.T) { //azignore:AZT012 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
