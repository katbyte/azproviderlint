# AZT009 - testdata files a test names must exist

AZT009 reports a test that names a `testdata` file that is not there.

A test runs in its own folder, so `testdata/winrm.ps1` means the file next to the test. Move a test to another folder without its test data and nothing notices until the test is run.

## Flagged Code

```go
// resources/mssql_virtual_machine/resource_test.go, with no testdata folder beside it
func (r MsSqlVirtualMachineResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurerm_windows_virtual_machine" "test" {
  custom_data = filebase64("testdata/winrm.ps1")
}
`, data.RandomInteger)
}
```

```text
testdata/winrm.ps1 does not exist: a test reads test data from its own folder, and the nearest copy is ../mssql_virtual_machine_availability_group_listener/testdata/winrm.ps1
```

## Passing Code

```go
// resources/mssql_virtual_machine/testdata/winrm.ps1 exists
  custom_data = filebase64("testdata/winrm.ps1")

// or the test names the folder the file is in
  custom_data = filebase64("../../testdata/winrm.ps1")
```

## What counts

Every string in a `_test.go` file is read for `testdata/<name>`, with any `./` or `../` in front, and so is every `//go:embed` line. The path is taken from the folder of the file that names it. `${path.module}/testdata/<name>` means the same folder.

When the file is missing the report says where the nearest file of that name is: in a `testdata` folder above first, then in one beside the test's folder.

Not reported:

- a name finished when the test runs, such as `testdata/cert_%d.pem` or `filepath.Join("testdata", name)`
- the end of a longer path, such as a URL
- a folder: `testdata/certs` passes when the folder exists
- strings in files that are not tests

## No fix

AZT009 only reports. Move or copy the file to the test's `testdata` folder, or correct the path.

## Ignoring Reports

Put `//azignore:AZT009 - <reason>` at the end of the reported line, or on the line above it. The reason is required. No comment can go inside a string that runs over several lines, so there the directive goes on the line the string starts, or the line above that.

```go
	//azignore:AZT009 - <reason>
	return fmt.Sprintf(`
resource "azurerm_windows_virtual_machine" "test" {
  custom_data = filebase64("testdata/winrm.ps1")
}
`, data.RandomInteger)
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
