## v0.17.1 (2026-11-26)

- `AZS005`: in a framework attribute, `NestedObject` now goes after `DeprecationMessage`, with the other fields that say what it holds ([#124](https://github.com/katbyte/azproviderlint/pull/124))

## v0.17.0 (2026-11-25)

- add rule `AZS005`: schema properties whose fields are not written in the agreed order, such as `Computed` before `Optional` ([#116](https://github.com/katbyte/azproviderlint/pull/116))
- `AZS015`: line breaks can be reported in resource, property, list or set descriptions only ([#117](https://github.com/katbyte/azproviderlint/pull/117))
- `AZS003` and the description rules: state migration schemas are no longer reported, in a package named `migration` or `migrations` ([#118](https://github.com/katbyte/azproviderlint/pull/118))
- `AZG010`: aliases can be allowed by pattern, such as every alias ending in `Beta` ([#120](https://github.com/katbyte/azproviderlint/pull/120))
- `AZS010`: a type with a field named `ResourceType`, such as the acceptance tests' `TestData`, is no longer reported as a resource ([#118](https://github.com/katbyte/azproviderlint/pull/118))
- fixing with every rule on no longer skips a file, or stops it building, when two rules fix the same code ([#122](https://github.com/katbyte/azproviderlint/pull/122))
- `AZG008`: the fix no longer leaves brackets behind, and reading an element of a slice or string, such as `(*items)[0]`, is reported without one ([#123](https://github.com/katbyte/azproviderlint/pull/123))

## v0.16.0 (2026-11-24)

- add rule `AZS018`: resource and property descriptions with a note in them, such as `**Note:** ...` ([#105](https://github.com/katbyte/azproviderlint/pull/105))
- add rule `AZS019`: resource and property descriptions written as a question, such as `Should the Thing be enabled?` ([#108](https://github.com/katbyte/azproviderlint/pull/108))
- add rule `AZS020`: resource and property descriptions naming a resource or property the provider does not have, such as a property since renamed ([#112](https://github.com/katbyte/azproviderlint/pull/112))
- `AZS015`: a word with no space before a bracket, such as `size(in GB)`, is now reported ([#106](https://github.com/katbyte/azproviderlint/pull/106))
- `AZS017`: wordy phrases such as `or not` and `in order to` are now reported and shortened ([#109](https://github.com/katbyte/azproviderlint/pull/109))
- `AZS014`: a range or a length the schema does not check, such as `must be between 5 and 300 seconds`, is now reported, and more ways of listing values are recognised, such as `Acceptable values include ...` ([#110](https://github.com/katbyte/azproviderlint/pull/110))
- descriptions handed to a shared schema helper, such as `helpers.NameSchema("The name.")`, are now checked by every description rule ([#111](https://github.com/katbyte/azproviderlint/pull/111))
- `AZS014`: a read-only property described as something to set, such as `Whether to ...`, is now reported ([#111](https://github.com/katbyte/azproviderlint/pull/111))
- `AZS014`: value lists in the middle of a sentence, such as `..., possible values are a and b`, and more ways of saying what else must be set are recognised ([#111](https://github.com/katbyte/azproviderlint/pull/111))
- `AZS015`: a value in quotes where the docs use backticks, such as `"TemporaryDisk"`, is now reported ([#111](https://github.com/katbyte/azproviderlint/pull/111))
- `AZS014`: a list that says what each value means, such as `` `Regional` - all nodes in one region``, is no longer reported ([#113](https://github.com/katbyte/azproviderlint/pull/113))
- `AZS014`: removing `as defined below` no longer leaves words such as `which are` behind ([#113](https://github.com/katbyte/azproviderlint/pull/113))
- two rules fixing the same words no longer lose every fix in the file ([#113](https://github.com/katbyte/azproviderlint/pull/113))

## v0.15.1 (2026-11-23)

- re-release of v0.15.0, whose binaries were published without their build provenance ([#104](https://github.com/katbyte/azproviderlint/pull/104))

## v0.15.0 (2026-11-23)

- add rule `AZS017`: resource and property descriptions that open with `Specifies`, such as `Specifies the name` for `The name` ([#98](https://github.com/katbyte/azproviderlint/pull/98))
- `AZS014`: possible values, defaults and the like in the description of a read-only property are now reported and removed ([#102](https://github.com/katbyte/azproviderlint/pull/102))

## v0.14.0 (2026-11-22)

- add rule `AZP007`: links to pages that have moved, such as `docs.microsoft.com` to `learn.microsoft.com` ([#96](https://github.com/katbyte/azproviderlint/pull/96))
- add rule `AZS010`: resources, properties and blocks with no description ([#93](https://github.com/katbyte/azproviderlint/pull/93))
- add rule `AZS011`: resource and property descriptions left blank ([#87](https://github.com/katbyte/azproviderlint/pull/87))
- add rule `AZS012`: resource and property descriptions with `a` where `an` belongs, or the other way round ([#88](https://github.com/katbyte/azproviderlint/pull/88))
- add rule `AZS013`: resource and property descriptions with an acronym in the wrong case, such as `id` for `ID` ([#89](https://github.com/katbyte/azproviderlint/pull/89))
- add rule `AZS014`: resource and property descriptions that repeat the schema, such as `(Optional)`, the default or the possible values ([#92](https://github.com/katbyte/azproviderlint/pull/92))
- add rule `AZS015`: resource and property descriptions with punctuation mistakes, such as no capital, no full stop or a stray space ([#91](https://github.com/katbyte/azproviderlint/pull/91))
- add rule `AZS016`: resource and property descriptions with an Azure product name in lowercase, such as `function app` for `Function App` ([#97](https://github.com/katbyte/azproviderlint/pull/97))

## v0.13.0 (2026-11-15)

- add rule `AZP006`: a service's shared code in a package named `helper`, where the provider uses `helpers` ([#83](https://github.com/katbyte/azproviderlint/pull/83))
- add rule `AZR011`: typed resources and data sources with no `var _ sdk.Resource = ...` line ([#82](https://github.com/katbyte/azproviderlint/pull/82))
- add rule `AZT009`: tests naming a testdata file that is not there ([#79](https://github.com/katbyte/azproviderlint/pull/79))
- add rule `AZT010`: testdata files that no test names ([#79](https://github.com/katbyte/azproviderlint/pull/79))
- built with Go 1.26.11, which fixes CVE-2026-111943

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
