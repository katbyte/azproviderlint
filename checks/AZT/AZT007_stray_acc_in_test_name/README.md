# AZT007 - acceptance test names must not carry a second Acc

AZT007 reports a `TestAcc` or `testAcc` function whose name has `Acc` in it again as a word of its own, as in `TestAccThing_storageAccBehindFireWall`.

`Acc` is the marker that makes a function an acceptance test. A second one reads as a mistake, and as an abbreviation of account it is unclear. Write the word out.

## Flagged Code

```go
func TestAccMsSqlServerExtendedAuditingPolicy_storageAccBehindFireWall(t *testing.T) {
```

## Passing Code

```go
func TestAccMsSqlServerExtendedAuditingPolicy_storageAccountBehindFirewall(t *testing.T) {
```

## What counts

`Acc` counts when it is a whole camel-case word: followed by an upper-case letter, an underscore, or the end of the name. `Account`, `Access` and `accelerated` are left alone, and so is the `Acc` of the prefix itself.

Only `TestAcc` and lowercase `testAcc` functions in `_test.go` files are checked. A plain `Test` name with `Acc` in the wrong place, like `TestWebAppAccActiveSlot_basic`, is [AZT004](../AZT004_acceptance_test_not_named_test_acc)'s report when it runs an acceptance test.

## No fix

AZT007 only reports: whether `Acc` meant account, access or something else is the author's call.

## Ignoring Reports

Put `//azignore:AZT007 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
func TestAccThing_storageAccBehindFireWall(t *testing.T) { //azignore:AZT007 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
