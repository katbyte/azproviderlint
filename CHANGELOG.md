## v0.12.0 (2026-10-09)

- add rule `AZT003`: tests named `TestAcc` that never run an acceptance test ([#64](https://github.com/katbyte/azproviderlint/pull/64))
- add rule `AZT004`: acceptance tests not named `TestAcc`, which acceptance runs skip ([#67](https://github.com/katbyte/azproviderlint/pull/67))
- add rule `AZT006`: update tests that never update anything ([#69](https://github.com/katbyte/azproviderlint/pull/69))
- add rule `AZT007`: a second `Acc` in an acceptance test's name ([#72](https://github.com/katbyte/azproviderlint/pull/72))
- add rule `AZT008`: acceptance tests in a file not named after their resource's file ([#73](https://github.com/katbyte/azproviderlint/pull/73))
- built with Go 1.26.9, which fixes CVE-2026-78667, CVE-2026-78669 and CVE-2026-97031

## v0.11.0 (2026-09-28)

- **breaking**: provider-wide rules move to a new `AZP` category: `AZS005` to `AZP002`, `AZS006` to `AZP003`, `AZS008` to `AZP004`, `AZR005` to `AZP005`; update ignore comments and settings ([#48](https://github.com/katbyte/azproviderlint/pull/48))
- add rule `AZG010`: import aliases that are not needed ([#62](https://github.com/katbyte/azproviderlint/pull/62))
- golangci settings: an option can take a list of values ([#62](https://github.com/katbyte/azproviderlint/pull/62))

## v0.10.0 (2026-09-24)

- add rule `AZG009`: enum pointers read with `pointer.From` and a cast, where `pointer.FromEnum` does both ([#60](https://github.com/katbyte/azproviderlint/pull/60))

## v0.9.0 (2026-09-22)

- update a dependency for CVE-2026-56864 and CVE-2026-56865; the linter was not affected, but the released binary now scans clean ([#50](https://github.com/katbyte/azproviderlint/pull/50))
- `AZR010`: an argument wrapped in extra parentheses is no longer missed
- `AZR003`: reads made through a renamed copy of the resource data are now reported
- `AZR003`/`AZR009`: a same-named function in another package is no longer mistaken for the resource's own
- `AZG008`: request bodies are recognised in more cases, so fewer unsafe fixes are offered
- `AZS009`: values set through a constant or a cast are now reported and fixed
- add rule `AZP001`: Microsoft docs links that force a language, such as `/en-us/`; starts the `AZP` category for provider-wide conventions ([#47](https://github.com/katbyte/azproviderlint/pull/47))

## v0.8.0 (2026-09-09)

- add rule `AZR010`: nil checks made before calling a flatten function, which should handle nil itself ([#46](https://github.com/katbyte/azproviderlint/pull/46))
- add rule `AZS009`: computed-only fields carrying settings that only apply to input ([#44](https://github.com/katbyte/azproviderlint/pull/44))
- add rule `AZR009`: log lines that only narrate a resource being created, read, updated or deleted ([#41](https://github.com/katbyte/azproviderlint/pull/41))
- `AZG008`: many more kinds of nil check are recognised, so fewer safe reads are reported ([#36](https://github.com/katbyte/azproviderlint/pull/36), [#37](https://github.com/katbyte/azproviderlint/pull/37), [#45](https://github.com/katbyte/azproviderlint/pull/45))
- `AZG008`: a nil check no longer counts once the pointer has been reassigned ([#37](https://github.com/katbyte/azproviderlint/pull/37))
- `AZG008`: values sent in a request body are reported without a fix, since the fix would send an empty value ([#36](https://github.com/katbyte/azproviderlint/pull/36))
- `AZG008`: no report where the code needs the original value, not a copy ([#36](https://github.com/katbyte/azproviderlint/pull/36), [#37](https://github.com/katbyte/azproviderlint/pull/37))
- `AZG002`: new `fix-pointer-copy` option chooses the fix for a copied pointer ([#35](https://github.com/katbyte/azproviderlint/pull/35))
- release: binaries are signed and carry build provenance ([#39](https://github.com/katbyte/azproviderlint/pull/39))

## v0.7.1 (2026-09-04)

- `AZG007`: an ignore comment on a literal's opening line covers the whole literal ([#33](https://github.com/katbyte/azproviderlint/pull/33))
- `AZG002`/`AZG005`/`AZG006`: no longer report where moving a call could change what the code does ([#34](https://github.com/katbyte/azproviderlint/pull/34))

## v0.7.0 (2026-09-03)

- add rule `AZG008`: pointers read without a nil check ([#31](https://github.com/katbyte/azproviderlint/pull/31))
- add rule `AZG007`: struct fields set to the value they would have anyway ([#24](https://github.com/katbyte/azproviderlint/pull/24))
- add rule `AZS008`: registration entries out of alphabetical order ([#23](https://github.com/katbyte/azproviderlint/pull/23))

## v0.6.0 (2026-09-02)

- add rule `AZG002`: a variable created only to take its address ([#29](https://github.com/katbyte/azproviderlint/pull/29))
- **breaking**: the old `AZG002` is renamed `AZV001`, as it checks validation messages; update ignore comments and settings ([#28](https://github.com/katbyte/azproviderlint/pull/28))
- plugin settings: `enable` and `disable` accept a whole category, such as `AZG`
- built with Go 1.26.8

## v0.5.1 (2026-09-02)

- `AZG005`/`AZG006`: no longer report where the fix could change the value used

## v0.5.0 (2026-09-02)

- `AZG005`: also reports a variable used once further down the same block ([#25](https://github.com/katbyte/azproviderlint/pull/25))
- `AZR008`: catches more ways of returning nil, maps included ([#27](https://github.com/katbyte/azproviderlint/pull/27))
- add rule `AZG006`: a variable used once, as an argument to a later call ([#26](https://github.com/katbyte/azproviderlint/pull/26))

## v0.4.0 (2026-09-01)

- add rule `AZR008`: flatten functions that return nil where an empty list is expected ([#22](https://github.com/katbyte/azproviderlint/pull/22))
- add rule `AZS007`: optional and computed fields without a comment saying why ([#20](https://github.com/katbyte/azproviderlint/pull/20))

## v0.3.2 (2026-08-28)

- `AZS004`: suggests the newer validation helper where the provider has it ([azurerm#33246](https://github.com/hashicorp/terraform-provider-azurerm/pull/33246)) ([#21](https://github.com/katbyte/azproviderlint/pull/21))

## v0.3.1 (2026-08-28)

- plugin settings: rule names match in any letter case ([#19](https://github.com/katbyte/azproviderlint/pull/19))
- `AZS004`: the suggested replacement for older SDK enums now compiles ([#19](https://github.com/katbyte/azproviderlint/pull/19))

## v0.3.0 (2026-08-27)

- ignore comments can carry a reason: `//azignore:AZR001 - why` ([#18](https://github.com/katbyte/azproviderlint/pull/18))
- add rule `AZG000`: ignore comments without a reason ([#18](https://github.com/katbyte/azproviderlint/pull/18))
- `AZS004`: also reports listed values the enum does not have; new `allow-missing-values` and `allow-extra-values` options ([#17](https://github.com/katbyte/azproviderlint/pull/17))
- `AZS006`: new `ignore-sensitive` option; ignore comments work on a single property ([#15](https://github.com/katbyte/azproviderlint/pull/15))
- add rule `AZR007`: `StateChangeConf` used where a custom poller should be ([#6](https://github.com/katbyte/azproviderlint/pull/6))

## v0.2.0 (2026-08-18)

- `AZT002`: only checks test files ([#13](https://github.com/katbyte/azproviderlint/pull/13))
- add rule `AZG003`: enum pointers made with a cast and `pointer.To`, where `pointer.ToEnum` does both ([#5](https://github.com/katbyte/azproviderlint/pull/5), [#10](https://github.com/katbyte/azproviderlint/pull/10))
- add rule `AZG004`: a nil check and read written out where `pointer.From` would do ([#5](https://github.com/katbyte/azproviderlint/pull/5), [#11](https://github.com/katbyte/azproviderlint/pull/11))
- add rule `AZG005`: a variable used once, on the very next line ([#12](https://github.com/katbyte/azproviderlint/pull/12))
- add rule `AZS002`: schema defaults of the wrong type, from tfproviderlint [#329](https://github.com/bflad/tfproviderlint/pull/329) ([#4](https://github.com/katbyte/azproviderlint/pull/4))
- add rule `AZS003`: blocks that can be left empty, from tfproviderlint [#236](https://github.com/bflad/tfproviderlint/pull/236) ([#4](https://github.com/katbyte/azproviderlint/pull/4))
- add rule `AZS004`: enum values listed by hand where the SDK already provides the list ([#7](https://github.com/katbyte/azproviderlint/pull/7))
- add rule `AZS005`: resources without a data source of the same name ([#8](https://github.com/katbyte/azproviderlint/pull/8))
- add rule `AZS006`: data sources missing properties their resource has ([#9](https://github.com/katbyte/azproviderlint/pull/9))
- built with Go 1.25.13, which fixes GO-2026-6218

## v0.1.0 (2026-08-07)

Initial release!

- add rule `AZG001`: an error assigned on one line and checked on the next, where one `if` would do
- add rule `AZS001`: model number fields that are not 64-bit
- the provider's script-based checks, rewritten as rules:
  - `AZG002`: error messages that call a format invalid without saying what is expected
  - `AZR001`: resource IDs set from a raw pointer, not a parsed ID
  - `AZR002`: one combined create and update, where two functions are expected
  - `AZR003`: config read inside a delete function
  - `AZC001`: clients created without an explicit endpoint
  - `AZR004`: resource IDs compared directly, not with `resourceids.Match`
  - `AZR005`: use of the unreleased case-insensitive segments flag
  - `AZD001`: data sources that clear their ID when nothing is found, where an error is expected
  - `AZD002`: data sources that mark themselves gone, where an error is expected
  - `AZR006`: a context taken without a timeout
  - `AZT001`: acceptance test files not in a `_test` package
  - `AZT002`: tests reading credentials from the environment
- release binaries for linux, macOS, windows, freebsd, openbsd and solaris
- a `version` command
- rules can be turned on and off in the golangci-lint plugin settings
