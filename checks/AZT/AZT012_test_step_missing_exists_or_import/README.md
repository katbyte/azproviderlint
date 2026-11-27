# AZT012 - each step of a resource's test must check the resource exists and be followed by an import step

AZT012 reports an acceptance test step that applies a config but never checks the resource exists, or is not followed by an import step.

Applying a config proves only that Terraform did not fail. The exists check proves the resource was made. The import step proves what was saved matches what the API gives back. azurerm's testing guide asks for an import step after every step.

## Flagged Code

```go
func TestAccThing_update(t *testing.T) {
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
		{
			Config: r.complete(data),
		},
	})
}
```

```text
step 3 of TestAccThing_update never checks the resource exists - add check.That(data.ResourceName).ExistsInAzure(r) to its Check
step 3 of TestAccThing_update is not followed by an import step - add data.ImportStep() after it
```

## Passing Code

```go
func TestAccThing_update(t *testing.T) {
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
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}
```

## What is checked

Every step of a `TestAcc...` function that hands a list of steps, written out in place, to `data.ResourceTest` or another of the `Resource...Test` runners. A step is held to this when it sets `Config`, or is made with `data.ApplyStep`.

- the step must check the resource exists: `ExistsInAzure`, or another check whose name has `Exists` in it. `data.ApplyStep` does this itself.
- the next step must import: `data.ImportStep()` or another helper whose name starts with `Import`, or a step with `ImportState: true`. Trying to create the resource again, with `data.RequiresImportErrorStep` or a step expecting `acceptance.RequiresImportError`, also counts.

These are not checked:

- a step that expects an error, only plans, destroys, or checks the resource is gone with `DoesNotExistInAzure`
- a data source's test, which has nothing to exist or import: one with `DataSource` in its name, or whose data is built for `data.<type>`
- steps built by a function

## Fix

None is offered.

## Ignoring Reports

Put `//azignore:AZT012 - <reason>` at the end of the step's first line, or on the line above it. The reason is required.

```go
{ //azignore:AZT012 - <reason>
	Config: r.template(data),
},
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
