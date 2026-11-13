# AZT010 - testdata files must be named by a test

AZT010 reports files in a `testdata` folder that no Go file names.

Test data outlives the test that used it. A test is moved or deleted, its certificate or script stays behind, and the next reader cannot tell whether it is safe to remove.

## Flagged Code

```text
internal/services/signalr/
  registration.go
  signalr_service_custom_domain_resource_test.go   reads testdata/custom-domain-cert-signalr-acctest.pfx
  testdata/
    custom-domain-cert-signalr-acctest.pfx
    custom-domain-cert-signalr.pfx                  nothing reads this
```

```text
testdata/custom-domain-cert-signalr.pfx is not named by any Go file in this folder or beneath it
```

Several files are listed in one report:

```text
3 files in testdata are not named by any Go file in this folder or beneath it: ad.ps1, certs/old.pem, connect.ps1
```

## Passing Code

```go
// any Go file in the folder holding testdata, or in a folder beneath it
contents = filebase64("testdata/custom-domain-cert-signalr.pfx")
contents = filebase64("../../testdata/custom-domain-cert-signalr.pfx")
```

## What counts

A file is used when a string or `//go:embed` line in a Go file names it, read the same way as [AZT009](../AZT009_testdata_file_missing): `testdata/<name>` with any `./` or `../` in front, taken from the folder of the file that names it. Every Go file in the folder holding `testdata` and in the folders beneath it is read, tests or not.

A file is also used when a Go file names:

- a folder it is in: `testdata/certs`, or `testdata` on its own as in `filepath.Join("testdata", name)`
- the start of its name, finished when the test runs: `testdata/cert_%d.pem` covers `cert_1.pem`
- the end of a longer path that ends in its name, such as a URL

One report is made per folder, on the package line of its first file by name.

Not reported:

- Go files, and files whose name starts with a dot
- a `testdata` folder that holds Go files anywhere in it: that is source for a tool to read, such as a linter's own test cases
- a `testdata` folder with no Go file beside it

A file named only by another testdata file, such as a script that a template runs, is reported. So is a file read only by a test in some other folder; it belongs with the test that reads it.

## No fix

AZT010 only reports. Delete the file, or move it beside the test that reads it.

## Options

| Option | Default | Effect |
|---|---|---|
| `ignore` | | regular expressions; a file is skipped when its path under `testdata` matches any of them |

`ignore` takes several values: a list in the golangci settings, or the flag repeated on the command line. It suits files kept on purpose, such as notes on how a certificate was made.

```yaml
          AZT010:
            ignore:
              - '(^|/)HOWTO\.md$'
              - '\.cnf$'
```

```bash
azproviderlint -AZT010 '-AZT010.ignore=(^|/)HOWTO\.md$' '-AZT010.ignore=\.cnf$' ./...
```

A pattern matches anywhere in the path unless it is anchored with `^` and `$`. See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZT010 - <reason>` at the end of the reported package line, or on the line above it. The reason is required. It covers every file in that folder's `testdata`, so prefer the `ignore` option for single files.

```go
package signalr //azignore:AZT010 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
