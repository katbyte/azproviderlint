## Unreleased

- add rule `AZS009`: computed-only schema fields must not set input-only attributes (`ValidateFunc`, `MaxItems`, `Default`, `ConflictsWith`, ...) and fields nested in a computed-only block or a typed resource's `Attributes()` must not be `Optional`/`Required`; the plugin SDK misses the nested cases; fixable with `-fix`
- `AZG008`: pointer parameters of nested closures are trusted like a function's own (`func(ctx context.Context, state *pluginsdk.InstanceState) { *state.ID }` no longer reports `state`); `include-parameters` covers them too ([#45](https://github.com/katbyte/azproviderlint/pull/45))
- `AZG008`: an `err`/`ok` companion checked on the left of `&&`/`||` in the same condition as the dereference (`if x, ok := f(); ok && *x != ""`) now counts as a guard ([#45](https://github.com/katbyte/azproviderlint/pull/45))
- add rule `AZR009`: no lifecycle narration logging (`log.Printf("[DEBUG] Creating %s", id)`, `metadata.Logger.Info("Decoding state..")`) inside Create/Read/Update/Delete functions — hashicorp/terraform-provider-azurerm#32423 removed the pattern; ID-rewrite, not-found, and waiting messages are kept; fixable with `-fix`, which also drops a `log` import left unused; `lib/lifecycle` finds lifecycle functions for AZR003 and AZR009
- ci: dev tools pinned in `.tools/go.mod` and built into `.tools/bin` by make (golangci-lint, actionlint, gofumpt; shellcheck and yamllint pinned in the makefile); `make lint` runs a custom golangci-lint with this repo's own AZG rules compiled in; revive enabled with justified opt-outs; new `actionlint`, `yamllint`, `shellcheck` targets and workflows; `check-all` runs everything
- ci: GitHub-owned actions pinned by commit hash; CodeQL scoped to `contents: read`; release `contents: write` scoped to the goreleaser job; `pr-*` workflow names; depscheck also verifies `.tools` tidy and the golangci version match with `.custom-gcl.yml`
- release: cosign keyless-signs `checksums.txt` into a sigstore bundle (`checksums.txt.sigstore.json`), GitHub artifact attestations (`gh attestation verify <file> -R katbyte/azproviderlint`), and SLSA Build L3 provenance from `slsa-github-generator`
- ci: `make typos` spell-checks every file with [typos](https://github.com/crate-ci/typos) (config in `.typos.toml`), pinned in the makefile like shellcheck, with a workflow on every PR; misspell only sees Go files
- docs: every README rewritten in plain language for people new to the provider, with a short glossary in the root README; badge row and lowercase workflow names match katbyte/tctest#138
- SECURITY.md
- `AZG008`: a guard is cancelled by a write to the pointer nested in a block or loop between the check and the dereference — unless the write is judged safe where it happens (a non-nil source, a checked `err`, a fresh nil check after a re-fetch); the same rule keeps `nilguard_returns` from exporting "never nil" for a function with a nested `out = nil`; alias staleness follows the derived path; an if-init reassignment settles the value for the condition too; a default-init counts only when it is the last write ([#37](https://github.com/katbyte/azproviderlint/pull/37))
- `AZG008`: the request-body fix withholding follows the value through any call (`pointer.To(*x)` into a PUT), and a serialising callee counts when the call site supplies the write method; an `err`/`ok` companion overwritten anywhere between the call and its check no longer vouches for the call's other results ([#37](https://github.com/katbyte/azproviderlint/pull/37))
- `AZG008`: map and slice index writes through a dereference (`(*m)[k] = v`) are no longer offered a fix, since `pointer.From` there still panics; left to AZG009 ([#37](https://github.com/katbyte/azproviderlint/pull/37))
- `lib/facts`: shared plumbing for per-function fact analyzers; `nilguard_returns` and `requestbody` rebuilt on it, no behaviour change ([#37](https://github.com/katbyte/azproviderlint/pull/37))
- `AZG008`: more guard shapes recognised — a nil check on an alias of the dereferenced chain (`if v := x.F; v != nil { *x.F }`, `v := x.F` earlier in the block, or in an enclosing if's init), `if x == nil { x = &T{} }` default-init, err/ok companions checked in the call's own if-init or via `if err == nil` / else of `if err != nil`, `commonids` composite-ID `First`/`Second`, `pointer.From(x) != ""` / `len(pointer.From(x)) > 0` conditions, `pointer.ToEnum[T]` and an immediately-invoked func literal whose every return is non-nil as non-nil sources, fields set from a non-nil source (or aliased) inside the composite literal that built the struct, and code after `if x == nil { return } else { ... }`; aliases and companions go stale once reassigned, and a type assertion's ok is not a nil guard ([#36](https://github.com/katbyte/azproviderlint/pull/36))
- `AZG008`: calls to functions that never return nil in a result position now count as non-nil sources — proven per function from its return statements (non-nil sources, proven locals, delegating calls) and shared across packages via analysis facts, so `expand*` helpers returning `&x` and `helpers.ExpandStringSlice` no longer trigger reports at their call sites; facts make the CLI load dependencies from source (≈3× memory on azurerm, same wall time) ([#36](https://github.com/katbyte/azproviderlint/pull/36))
- `AZG008`: dereferences whose value is sent as a PUT/PATCH/POST body are reported without a fix (`requestbody: false` skips them) — the callee serialises that parameter while performing a write, derived by the new `requestbody` fact analyzer from `net/http` method constants and `encoding/json`/`xml` marshalling alone (closures and delegating wrappers included, no SDK names) — directly, via a local, a struct field, a literal, or by address; `pointer.From` there would send an empty write request, so the report names the reason instead ([#36](https://github.com/katbyte/azproviderlint/pull/36))
- `AZG008`: no longer reports (or rewrites into non-compiling code) dereferences whose pointee must be addressable: `(*x).F = v`, `(*x)[i] = v` on an array, `(*x)[:]`, `&(*x).F`, pointer-receiver method calls `(*x).M()` ([#36](https://github.com/katbyte/azproviderlint/pull/36))
- `AZG002`: `out := *p; &out` names both forms in the report; `fix-pointer-copy` picks the fix — `none` (default, person decides), `copy` (`pointer.To(*p)`/`new(*p)`), or `share` (`p` directly, an aliasing change) ([#35](https://github.com/katbyte/azproviderlint/pull/35))

## v0.7.1 (2026-09-04)

- `AZG007`: an `//azignore:AZG007` on a composite literal's opening line suppresses the whole literal, nested literals included ([#33](https://github.com/katbyte/azproviderlint/pull/33))
- `AZG002`/`AZG005`/`AZG006`: don't inline past statements that could observe or change the initializer's call side effects (pointer-like arguments/receivers, both directions); the pure `pointer` package is exempt ([#34](https://github.com/katbyte/azproviderlint/pull/34))

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
