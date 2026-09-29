# AZS007 - optional+computed fields must have a Note: O+C comment

AZS007 reports a schema field that is both `Optional: true` and `Computed: true` with no `// Note: O+C because ...` comment between the two lines.

Optional plus Computed (O+C) means "the user may set this, and if they do not, Azure picks a value". It is the right choice sometimes, but it also hides drift: if Azure changes the value behind the user's back, Terraform will not notice. So each use needs a written reason explaining the API behaviour that made it necessary. The comment makes that easy to review and easy to revisit.

The comment must match `// Note: O+C` (any casing) and sit on a line between `Optional:` and `Computed:`. Only the first line of a multi-line comment has to match. Both the untyped schema and the typed SDK's `Arguments()` / `Attributes()` methods are checked, and the `pluginsdk` aliases used in azurerm are recognised.

## Flagged Code

```go
"max_message_size_in_kilobytes": {
	Type:     schema.TypeInt,
	Optional: true,
	Computed: true,
},
```

```go
"max_message_size_in_kilobytes": {
	Type:     schema.TypeInt,
	Optional: true,
	// this needs to be computed
	Computed: true,
},
```

## Passing Code

```go
"max_message_size_in_kilobytes": {
	Type:     schema.TypeInt,
	Optional: true,
	// NOTE: O+C this gets a variable default based on the sku and can be updated without issues
	Computed: true,
},
```

```go
"administrator_login": {
	Type:     schema.TypeString,
	Optional: true,
	// Note: O+C because Azure returns a generated value if
	// azure_active_directory_administrator.azuread_authentication_only_enabled is true
	Computed: true,
	ForceNew: true,
},
```

## Options

| Option | Default | Effect |
|---|---|---|
| `exclude-packages` | (empty) | comma-separated package names to skip, such as state-migration snapshot packages |

Set with `-AZS007.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Put `//azignore:AZS007 - <reason>` at the end of the `Computed:` line, or on the line above it. The reason is required.

```go
Computed: true, //azignore:AZS007 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
