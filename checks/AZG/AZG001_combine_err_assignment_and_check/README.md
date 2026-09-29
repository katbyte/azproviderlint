# AZG001 - combine err assignment and check into one if

AZG001 reports `err := f()` or `_, err := f()` followed straight away by `if err != nil`. The two should be one `if` with an init statement.

The combined form keeps `err` scoped to its check and keeps the happy path at the outer indent, which is how the rest of the provider reads.

Only safe cases are reported: every other value on the left side must be `_`, and an `err` declared with `:=` must not be used again after the `if`, since combining moves it into the `if`'s scope.

## Flagged Code

```go
_, err := client.Delete(ctx, id)
if err != nil {
	return fmt.Errorf("deleting %s: %+v", id, err)
}
```

```go
err := resourceGroupClient.WaitForDeletion(ctx, id)
if err != nil {
	return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
}
```

## Passing Code

```go
if _, err := client.Delete(ctx, id); err != nil {
	return fmt.Errorf("deleting %s: %+v", id, err)
}
```

```go
if err := resourceGroupClient.WaitForDeletion(ctx, id); err != nil {
	return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
}
```

## Ignoring Reports

Put `//azignore:AZG001 - <reason>` at the end of the assignment line, or on the line above it. The reason is required.

```go
_, err := client.Delete(ctx, id) //azignore:AZG001 - <reason>
if err != nil {
	return err
}
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
