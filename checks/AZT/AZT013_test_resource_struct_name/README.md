# AZT013 - an acceptance test's resource struct must be named after the resource

AZT013 reports the struct a resource's acceptance tests hang off when it is not named after the resource.

Generated tests, such as the identity ones, work the struct's name out from the resource's. A struct named any other way is one they cannot find, and the generation check fails.

## Flagged Code

```go
type StorageAcctTestResource struct{}

func TestAccStorageAccount_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_account", "test")
	r := StorageAcctTestResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
```

```text
test resource StorageAcctTestResource is not named after azurerm_storage_account, which it tests - name it StorageAccountResource
```

## Passing Code

```go
type StorageAccountResource struct{}
```

## What is checked

The struct a test hands to `data.ResourceTest` or another of the `Resource...Test` runners, against the resource the test names in `acceptance.BuildTestData`.

The name wanted is the resource's without the provider, each word given one capital, then `Resource`: `azurerm_cdn_frontdoor_rule` is `CdnFrontdoorRuleResource`, not `CdnFrontDoorRuleResource`.

A struct is reported once, where it is declared. One used for several resources is fine when it is named after any of them.

These are not checked:

- a data source's struct, which is not handed to what runs the test
- a test that does not call `BuildTestData` with the resource's name written out

## Fix

None is offered.

## Ignoring Reports

Put `//azignore:AZT013 - <reason>` at the end of the line that declares the struct, or on the line above it. The reason is required.

```go
type StorageAcctTestResource struct{} //azignore:AZT013 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
