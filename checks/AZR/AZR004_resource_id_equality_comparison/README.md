# AZR004 - compare resource id types with resourceids.Match

AZR004 reports resource IDs compared with `==` or `!=`, such as `a.ID() == b.ID()`.

Parts of an Azure resource ID come from the user, and Azure compares those parts without caring about case. So two IDs can name the same resource and still be different strings. Use `resourceids.Match(a, b)` from `github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids`, and `!resourceids.Match(a, b)` in place of `!=`.

## Flagged Code

```go
if subnetId.ID() == other.ID() {
	// ...
}
```

## Passing Code

```go
if resourceids.Match(subnetId, other) {
	// ...
}
```

## Ignoring Reports

Put `//azignore:AZR004 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
if subnetId.ID() == other.ID() { //azignore:AZR004 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
