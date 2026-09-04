## Unreleased

- `AZG007`: a `//azignore:AZG007` directive on a composite literal's opening line suppresses every finding within that literal, including nested literals ([#33](https://github.com/katbyte/azproviderlint/pull/33))
- `AZG002`/`AZG005`/`AZG006`: no longer inline past an intervening statement that reads a variable the initializer's calls may mutate through a pointer-like argument or receiver (`result := astutil.Apply(f, ...)` then a read of `f`), nor when an intervening call receives an operand pointer-like — moving the call re-orders observable side effects; the pure go-azure-helpers pointer package is exempt ([#34](https://github.com/katbyte/azproviderlint/pull/34))

## v0.7.0 (2026-09-03)

- add rule `AZG008`: pointer dereferences (`*props.Status`) must have a reachable nil guard; fixes to `pointer.From`/`pointer.FromEnum`; options `include-parameters`, `tests`, `fix-with` ([#31](https://github.com/katbyte/azproviderlint/pull/31))
- add rule `AZG007`: struct literal fields set to their zero value should be omitted (named-constant zeros are kept); fixable with `-fix`; `tests` opts into test files ([#24](https://github.com/katbyte/azproviderlint/pull/24))
- add rule `AZS008`: `registration.go` entries must be sorted alphabetically, checked per section; fixable with `-fix`; `generated: false` skips `registration_gen.go` ([#23](https://github.com/katbyte/azproviderlint/pull/23))

## v0.6.0 (2026-09-02)

- add rule `AZG002`: single-use temporaries only used as `&v` should be `new(<expr>)` (go1.26; or `pointer.To` via `use`); new mode also rewrites existing `pointer.To(x)` calls unless `allow: pointer.To`; fixable with `-fix` ([#29](https://github.com/katbyte/azproviderlint/pull/29))
- **breaking**: rename `AZG002` to `AZV001` — it polices validation error messages, so it moves to the reserved AZV category; update `//azignore:AZG002` comments and settings references ([#28](https://github.com/katbyte/azproviderlint/pull/28))
- plugin settings: `enable`/`disable` entries can name a whole category (`enable: [AZG]`); `disable` applies after `enable`
- build and scan with Go 1.26.8

## v0.5.1 (2026-09-02)

- `AZG005`/`AZG006`: no longer report when an intervening statement writes to, takes the address of, or shadows anything the initializer reads — the fix could change the value read

## v0.5.0 (2026-09-02)

- `AZG005`: also flag temporaries consumed by a later statement in the same block, within `max-gap` lines (default 100) ([#25](https://github.com/katbyte/azproviderlint/pull/25))
- `AZR008`: also cover map results, naked returns with an unassigned named result, and provably-nil variables; a provably-nil error no longer masks a finding ([#27](https://github.com/katbyte/azproviderlint/pull/27))
- add rule `AZG006`: single-use variables only used as an argument of a later call should be inlined; tuned by `max-gap` (default 100), `only-when-literals`, and `maximum-arguments`; fixable with `-fix` ([#26](https://github.com/katbyte/azproviderlint/pull/26))

## v0.4.0 (2026-09-01)

- add rule `AZR008`: `flatten*` functions should return `[]T{}` instead of `nil`; error paths exempt, naked returns out of scope; fixable with `-fix` ([#22](https://github.com/katbyte/azproviderlint/pull/22))
- add rule `AZS007`: schema fields with both `Optional: true` and `Computed: true` need a `// Note: O+C because ...` comment; `exclude-packages` skips listed packages ([#20](https://github.com/katbyte/azproviderlint/pull/20))

## v0.3.2 (2026-08-28)

- `AZS004`: track-1 advice suggests `validation.StringInEnumSlice(cdn.PossibleTransformValues(), false)` when the validation package exports a generic wrapper ([azurerm#33246](https://github.com/hashicorp/terraform-provider-azurerm/pull/33246)); `pointer.FromEnumSlice(pointer.To(...))` remains the fallback ([#21](https://github.com/katbyte/azproviderlint/pull/21))

## v0.3.1 (2026-08-28)

- plugin settings: rule names matched case-insensitively (golangci's YAML decoding lowercases keys) ([#19](https://github.com/katbyte/azproviderlint/pull/19))
- `AZS004`: track-1 enum advice now compiles, via `pointer.FromEnumSlice(pointer.To(...))` ([#19](https://github.com/katbyte/azproviderlint/pull/19))

## v0.3.0 (2026-08-27)

- `//azignore` directives take an optional reason after the rule list (`//azignore:AZR001 - deliberate subset`) ([#18](https://github.com/katbyte/azproviderlint/pull/18))
- add rule `AZG000`: report `//azignore` directives without a reason ([#18](https://github.com/katbyte/azproviderlint/pull/18))
- `AZS004`: also report list values not in the enum; new `allow-missing-values`/`allow-extra-values` flags ([#17](https://github.com/katbyte/azproviderlint/pull/17))
- `AZS006`: new `ignore-sensitive` flag; `//azignore:AZS006` works on individual properties ([#15](https://github.com/katbyte/azproviderlint/pull/15))
- add rule `AZR007`: `StateChangeConf` from `helper/retry` should be a custom poller implementing `pollers.PollerType` ([#6](https://github.com/katbyte/azproviderlint/pull/6))

## v0.2.0 (2026-08-18)

- `AZT002`: only check `_test.go` files ([#13](https://github.com/katbyte/azproviderlint/pull/13))
- add rule `AZG003`: `pointer.To(sdk.SomeEnum(v))` should be `pointer.ToEnum[sdk.SomeEnum](v)`; fixable with `-fix` ([#5](https://github.com/katbyte/azproviderlint/pull/5), [#10](https://github.com/katbyte/azproviderlint/pull/10))
- add rule `AZG004`: zero-value declaration plus nil-check dereference should be `pointer.From(x)`; fixable with `-fix` ([#5](https://github.com/katbyte/azproviderlint/pull/5), [#11](https://github.com/katbyte/azproviderlint/pull/11))
- add rule `AZG005`: single-use temporaries immediately consumed by the next statement should be inlined; fixable with `-fix` ([#12](https://github.com/katbyte/azproviderlint/pull/12))
- add rule `AZS002`: schema `Default` values must match the declared `Type` (ports tfproviderlint [#329](https://github.com/bflad/tfproviderlint/pull/329) S038) ([#4](https://github.com/katbyte/azproviderlint/pull/4))
- add rule `AZS003`: optional/required `TypeList` blocks must not allow empty blocks (ports tfproviderlint [#236](https://github.com/bflad/tfproviderlint/pull/236) XS003) ([#4](https://github.com/katbyte/azproviderlint/pull/4))
- add rule `AZS004`: enum validation should use the SDK's `PossibleValuesFor<Enum>()` helper ([#7](https://github.com/katbyte/azproviderlint/pull/7))
- add rule `AZS005`: registered resources should have a data source of the same name ([#8](https://github.com/katbyte/azproviderlint/pull/8))
- add rule `AZS006`: data sources should not be missing properties of the same-named resource ([#9](https://github.com/katbyte/azproviderlint/pull/9))
- build and scan with Go 1.25.13 (fixes GO-2026-6218); govulncheck honours `.go-version`

## v0.1.0 (2026-08-07)

Initial release!

- add rule `AZG001`: `_, err := SomeFunc()` followed by `if err != nil` should be a single `if` init statement
- add rule `AZS001`: typed SDK model numeric fields (tagged `tfschema`) must be `int64`/`float64`
- port the grep/sed based checks from terraform-provider-azurerm's `scripts/checks/` to AST-based rules:
  - `AZG002`: unclear `invalid format of ...` error messages
  - `AZR001`: `d.SetId(*ptr)` instead of a Resource ID Formatter/Parser's `id.ID()`
  - `AZR002`: combined `CreateUpdate` methods instead of separate Create and Update
  - `AZR003`: `d.Get`/`metadata.ResourceData.Get` inside Delete functions
  - `AZC001`: Azure SDK clients created without an explicit resource manager endpoint
  - `AZR004`: Resource IDs compared with `==`/`!=` instead of `resourceids.Match`
  - `AZR005`: assignments to the unreleased `TreatUserSpecifiedSegmentsAsCaseInsensitive` feature flag
  - `AZD001`: data sources calling `d.SetId("")` instead of returning an error
  - `AZD002`: data sources calling `metadata.MarkAsGone` instead of returning an error
  - `AZR006`: `ctx` assigned from `meta.(*clients.Client).StopContext` without a timeouts wrapper
  - `AZT001`: resource/data source acceptance test files not using a `_test` package
  - `AZT002`: tests reading `ARM_CLIENT_ID`/`ARM_CLIENT_SECRET` credentials from the environment
- release binaries with goreleaser (linux/darwin/windows/freebsd/openbsd/solaris) on tagged releases
- add a `version` subcommand printing the version and git commit
- support per-rule `enable`/`disable` lists via golangci-lint plugin settings
