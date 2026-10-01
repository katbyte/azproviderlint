# AZT003 - TestAcc functions must run an acceptance test

AZT003 reports test functions named `TestAcc...` that never run an acceptance test.

Acceptance runs pick tests by that prefix. A unit test that carries it is scheduled and reported with the acceptance tests, although it never touches Azure.

A function counts as an acceptance test when it does one of these, itself or through another function in the same package:

- hands `t` to the provider's `internal/acceptance` package (`acceptance.BuildTestData(t, ...)`, `data.ResourceTest(t, ...)`, `acceptance.RunTestsInSequence(t, ...)`)
- hands `t` to the Terraform test harness (`resource.Test`, `resource.ParallelTest`)
- looks at `TF_ACC`

Other functions in the package are followed whether they are called or passed as a value, so sequential tests that list `testAccThing_basic` in a map pass. A helper in a different package is not followed.

Only `_test.go` files are checked. Names like `TestAccountName` are left alone, since `Acc` there is the start of a word.

## Flagged Code

```go
func TestAccContainerRegistryName_validation(t *testing.T) {
	for _, tc := range cases {
		_, errors := validate.ContainerRegistryName(tc.Value, "name")
		if len(errors) != tc.ErrCount {
			t.Fatalf("expected %d errors, got %d", tc.ErrCount, len(errors))
		}
	}
}
```

## Passing Code

```go
func TestContainerRegistryName_validation(t *testing.T) {
	for _, tc := range cases {
		_, errors := validate.ContainerRegistryName(tc.Value, "name")
		if len(errors) != tc.ErrCount {
			t.Fatalf("expected %d errors, got %d", tc.ErrCount, len(errors))
		}
	}
}
```

```go
func TestAccContainerRegistry_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry", "test")
	r := ContainerRegistryResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
		data.ImportStep(),
	})
}
```

## No fix

AZT003 only reports. Usually the answer is to drop `Acc` from the name. If the test was meant to be an acceptance test and lost its body, restore or delete it instead.

## Ignoring Reports

Put `//azignore:AZT003 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
func TestAccThing_validation(t *testing.T) { //azignore:AZT003 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
