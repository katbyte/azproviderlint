# azproviderlint

[![GitHub release](https://img.shields.io/github/v/release/katbyte/azproviderlint?color=blueviolet)](https://github.com/katbyte/azproviderlint/releases/latest)
![build](https://github.com/katbyte/azproviderlint/actions/workflows/build.yaml/badge.svg)
![test](https://github.com/katbyte/azproviderlint/actions/workflows/test.yaml/badge.svg)
![lint](https://github.com/katbyte/azproviderlint/actions/workflows/lint.yaml/badge.svg)
![govulncheck](https://github.com/katbyte/azproviderlint/actions/workflows/govulncheck.yaml/badge.svg)
![CodeQL](https://github.com/katbyte/azproviderlint/actions/workflows/codeql-analysis.yml/badge.svg)
[![Go Version](https://img.shields.io/github/go-mod/go-version/katbyte/azproviderlint?color=00ADD8)](https://github.com/katbyte/azproviderlint/blob/main/go.mod)
[![License](https://img.shields.io/github/license/katbyte/azproviderlint?color=blue)](https://github.com/katbyte/azproviderlint/blob/main/LICENSE)

A custom [golangci-lint](https://golangci-lint.run/) module plugin providing Azure provider-specific linting rules built on Go's `analysis` framework.

## Installation

### Standalone

```bash
go install github.com/katbyte/azproviderlint@latest
```

Then run directly (all rules run by default):
```bash
azproviderlint ./...
```

Each rule is also a flag, and setting any rule flag switches to running only the named rules; a category on its own (`-AZG`) runs every rule in that category:
```bash
azproviderlint -AZG001 ./...
azproviderlint -AZR001 -AZR003 ./...
azproviderlint -AZG ./...
azproviderlint -AZG -AZR001 ./...
```

### As a golangci-lint Plugin

Add to your `.custom-gcl.yml`:
```yaml
version: v2.12.2
plugins:
  - module: "github.com/katbyte/azproviderlint"
    import: "github.com/katbyte/azproviderlint/plugin"
    version: v0.1.0
```

Build the custom binary:
```bash
golangci-lint custom
```

Then enable in `.golangci.yml`:
```yaml
linters:
  enable:
    - azproviderlint
  settings:
    custom:
      azproviderlint:
        type: module
```

Individual rules can be enabled/disabled via plugin settings (an empty `enable` list means all rules); an entry names a rule or a whole category (`AZG` matches every AZG rule), and `disable` is applied after `enable`:
```yaml
linters:
  settings:
    custom:
      azproviderlint:
        type: module
        settings:
          enable: [AZG]         # only the AZG rules...
          disable: [AZG005]     # ...except AZG005
```

Some rules take options, set under a rule-name key in the same settings block (or `-<RULE>.<option>` on the standalone binary) — each rule's README documents its options:
```yaml
        settings:
          AZS004: {allow-extra-values: true}
          AZG005: {max-gap: 50}
```

To run just azproviderlint through the custom binary, skipping every other linter:
```bash
custom-gcl run --enable-only azproviderlint ./...
```

There is no CLI flag for a single rule — combine `--enable-only` with an `enable: [AZG001]` list in the plugin settings above, or use the standalone binary's per-rule flags.

### Why the plugin over the standalone binary?

The plugin requires every consumer to build a custom golangci-lint binary (`golangci-lint custom`), but that one-time cost buys a lot on a codebase the size of a provider:

- **One package-load instead of two.** On azurerm this is the dominant cost — loading and type-checking the provider codebase (with its enormous vendor tree) takes minutes, and every separate analysis binary pays it again from scratch. Folding checks into golangci-lint amortizes it, and golangci-lint's result cache makes warm local re-runs dramatically faster; a standalone multichecker reloads the world every single time.
- **Unified config and reporting.** `.golangci.yml` path exclusions (generated files, `/sdk/`, `third_party` — already curated in the provider repo) apply to these checks for free; one output stream, one CI job, SARIF/annotations, and `--new-from-rev` — the killer feature for a codebase with 22+ pre-existing findings per service, since checks can be enforced on new code only instead of azignoring a decade of history.
- **`//nolint` works uniformly** alongside `//azignore` (see [Ignoring Reports](#ignoring-reports)).

The standalone binary remains the right tool for one-off or single-rule runs (`azproviderlint -AZG001 ./...`), editor integrations that expect a plain `analysis`-style vet tool, and quick iteration while developing new checks.

## Rules

Rules are named `AZ<category letter><number>`, aligned with [tfproviderlint](https://github.com/bflad/tfproviderlint)'s category letters (`R`, `S`, `V`, `AT`) where they overlap.

### AZG — General Go Style / Readability

| Rule | Description |
|------|-------------|
| [AZG000](checks/AZG/AZG000_azignore_missing_reason) | azignore directives must give a reason |
| [AZG001](checks/AZG/AZG001_combine_err_assignment_and_check) | combine err assignment and check into one if |
| [AZG002](checks/AZG/AZG002_address_of_single_use_temporary) | use new() instead of a single-use temporary's address |
| [AZG003](checks/AZG/AZG003_pointer_to_enum_conversion) | use pointer.ToEnum for enum conversions |
| [AZG004](checks/AZG/AZG004_zero_value_init_pointer_from) | use pointer.From instead of nil-check dereference |
| [AZG005](checks/AZG/AZG005_single_use_temporary) | inline single-use variable only used in a later assignment or return |
| [AZG006](checks/AZG/AZG006_single_use_call_argument) | inline single-use variable only used in a later function call |
| [AZG007](checks/AZG/AZG007_redundant_zero_value_field) | omit struct literal fields explicitly set to their zero value |
| [AZG008](checks/AZG/AZG008_unchecked_nil_dereference) | pointer dereferences must have a nil guard or use pointer.From |
| [AZG009](checks/AZG/AZG009_nil_dereference_requires_guard) | implicit and write-target nil dereferences need a hand-written guard |

### AZR — Resource Implementation

| Rule | Description |
|------|-------------|
| [AZR001](checks/AZR/AZR001_set_id_dereferenced_pointer) | SetId must use a resource id formatter/parser |
| [AZR002](checks/AZR/AZR002_combined_create_update_method) | separate Create and Update methods |
| [AZR003](checks/AZR/AZR003_resource_data_get_in_delete) | no d.Get in Delete functions |
| [AZR004](checks/AZR/AZR004_resource_id_equality_comparison) | compare resource id types with resourceids.Match |
| [AZR005](checks/AZR/AZR005_case_insensitive_segments_feature_flag) | do not set the case-insensitive segments feature flag |
| [AZR006](checks/AZR/AZR006_stop_context_without_timeouts) | use timeouts wrappers, not StopContext |
| [AZR007](checks/AZR/AZR007_state_change_conf_custom_poller) | use custom pollers instead of StateChangeConf |
| [AZR008](checks/AZR/AZR008_flatten_returns_nil_slice) | flatten functions must return empty slices/maps, not nil |

### AZD — Data Sources

| Rule | Description |
|------|-------------|
| [AZD001](checks/AZD/AZD001_data_source_empty_set_id) | data sources must error when not found, not SetId("") |
| [AZD002](checks/AZD/AZD002_data_source_mark_as_gone) | data sources must error when not found, not MarkAsGone |

### AZS — Schema & Typed SDK Models

| Rule | Description |
|------|-------------|
| [AZS001](checks/AZS/AZS001_typed_sdk_model_64bit_types) | typed SDK model numeric fields must be 64-bit (int64/float64) |
| [AZS002](checks/AZS/AZS002_schema_default_type_mismatch) | schema Default values must match the declared Type |
| [AZS003](checks/AZS/AZS003_schema_allows_empty_block) | TypeList blocks must not allow empty blocks |
| [AZS004](checks/AZS/AZS004_enum_validation_possible_values) | enum validation must use the SDK's possible-values helper |
| [AZS005](checks/AZS/AZS005_resource_missing_data_source) | registered resources must have a same-named data source |
| [AZS006](checks/AZS/AZS006_data_source_missing_properties) | data sources must expose their same-named resource's properties |
| [AZS007](checks/AZS/AZS007_optional_computed_missing_comment) | optional+computed fields must have a Note: O+C comment |
| [AZS008](checks/AZS/AZS008_registration_entries_sorted) | registration entries must be sorted alphabetically |

### AZC — Clients & SDK Usage

| Rule | Description |
|------|-------------|
| [AZC001](checks/AZC/AZC001_client_missing_resource_manager_endpoint) | clients must set an explicit resource manager endpoint |

### AZT — Acceptance Testing

| Rule | Description |
|------|-------------|
| [AZT001](checks/AZT/AZT001_acceptance_test_external_package) | acceptance tests must use a _test package |
| [AZT002](checks/AZT/AZT002_credentials_from_environment) | acctests must not read credentials from the environment |

### AZN — Naming Conventions

_No rules yet — reserved for property naming convention rules (e.g. percentage properties using a `_percentage` suffix rather than `_in_percent`)._

### AZV — Validation

| Rule | Description |
|------|-------------|
| [AZV001](checks/AZV/AZV001_error_should_describe_expected_format) | 'invalid format' error messages must describe the expected format |

## Ignoring Reports

When run via golangci-lint, all azproviderlint reports on a line can be ignored with a `//nolint:azproviderlint` comment at the end of the offending line or on the line immediately preceding it.

To ignore a specific check — leaving the others active, and working under any driver including the standalone binary — use a `//azignore:<Rule> - <reason>` comment in the same positions. Multiple rules can be listed separated by commas, and the reason is free text after the rule list — the `-` separator (`–`/`—` also work) is optional:

```go
d.SetId(*read.ID) //azignore:AZR001 - legacy resource, ID formatter tracked in #1234

//azignore:AZG001,AZR003 combined form obscures the retry loop here
err := client.Delete(ctx, id)
```

The reason is required: directives without one still suppress their target checks, but are themselves reported by [AZG000](checks/AZG/AZG000_azignore_missing_reason) (disable that check to drop the requirement).
