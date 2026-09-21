# AZT002 - acceptance tests must not read credentials from the environment

AZT002 reports tests that read the provider's own credentials with `os.Getenv("ARM_CLIENT_ID")`, `os.Getenv("ARM_CLIENT_SECRET")`, or `os.Getenv("ARM_CLIENT_SECRET_ALT")`.

Those are the credentials the test framework itself runs with, and they have broad permissions. A test config should not reuse them. Instead, create a User Assigned Identity in the test config with the smallest role that works. It gets cleaned up with the rest of the test resources.

Only `_test.go` files are checked. The provider runtime and the test framework have to read these variables.

## Flagged Code

```go
clientId := os.Getenv("ARM_CLIENT_ID")
clientSecret := os.Getenv("ARM_CLIENT_SECRET")
```

## Passing Code

```hcl
resource "azurerm_user_assigned_identity" "test" {
  name                = "acctest-uai-${var.random_integer}"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_role_assignment" "test" {
  scope                = azurerm_resource_group.test.id
  role_definition_name = "Reader"
  principal_id         = azurerm_user_assigned_identity.test.principal_id
}
```

## Ignoring Reports

Put `//azignore:AZT002 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
clientId := os.Getenv("ARM_CLIENT_ID") //azignore:AZT002 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
