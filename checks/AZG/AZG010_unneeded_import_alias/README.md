# AZG010 - drop import aliases that are not needed

AZG010 reports an import alias when the package's own name would work just as well: no other import, package-level declaration, local variable, parameter, or builtin in the file goes by that name. That covers aliases that repeat the package name, and aliases left behind once the clash they worked around has gone.

## Flagged Code

```go
import (
	set "github.com/hashicorp/go-set/v3"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate"
)
```

## Passing Code

```go
import (
	"github.com/hashicorp/go-set/v3"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate"
)
```

```go
// not reported: both packages are named validate
import (
	computeValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/compute/validate"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate"
)
```

## What counts

Not reported:

- the name is taken by another import, a package-level declaration, or a builtin
- the name is declared at package level in the package's own test files, which would break the test build
- the name is declared anywhere in the file, even in another function, so the rename would shadow or be shadowed
- two imports share a package name, even if both are aliased, since dropping either alias only works while the other stays
- the name is the file's own package name, or the package an external test file tests; `keyvault.BaseClient` inside package `keyvault` reads as a reference to itself
- an alias that is the package name when the path suggests something else (`devices "…/iothub"`); goimports writes those
- generated files

## The fix

`-fix` removes the alias and renames every use in the file. If the package's name differs from what its path suggests (`v2021_03_01` from `…/2021-03-01`), the fix swaps the alias for the package name instead of removing it, as goimports would write it.

Removing one alias can free a name another alias was avoiding, so a second `-fix` run can find more.

## Options

| Option | Default | Effect |
|---|---|---|
| `ignore` | | regular expressions; an import is skipped when its package name matches any of them |
| `allow` | | aliases that are never reported |

Both take several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZG010:
            ignore:
              - '^v\d{4}_\d{2}_\d{2}' # versioned SDK packages: v2021_03_01, v2021_04_01_preview
              - '^v\d+_\d+$'          # v7_4
            allow: [log, azValidate]
```

```bash
azproviderlint -AZG010 '-AZG010.ignore=^v\d{4}_\d{2}_\d{2}' '-AZG010.ignore=^v\d+_\d+$' -AZG010.allow=log -AZG010.allow=azValidate ./...
```

`ignore` is matched against the package's name, not the alias or the import path. A pattern matches anywhere in the name unless it is anchored with `^` and `$`. Give each pattern its own entry; they are not split on commas, since a pattern can contain one (`\d{2,4}`).

`allow` entries are whole alias names. They can also be comma-separated in one entry (`log,azValidate`).

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZG010 - <reason>` at the end of the import line, or on the line above it. The reason is required.

```go
networkValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate" //azignore:AZG010 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
