package thing_test

import (
	"fmt"
	"path/filepath"
)

func present() string {
	return `
resource "azurerm_thing" "test" {
  first  = file("testdata/present.xml")
  second = file("./testdata/present.xml")
  third  = file("${path.module}/testdata/present.xml")
  above  = filebase64("../../testdata/above.pfx")
  folder = "../../testdata/certs"
}
`
}

// want +5 `testdata/missing\.xml does not exist: a test reads test data from its own folder$`
// want +5 `testdata/above\.pfx does not exist: a test reads test data from its own folder, and the nearest copy is \.\./\.\./testdata/above\.pfx`
func missing() string {
	return `
resource "azurerm_thing" "test" {
  first  = file("testdata/missing.xml")
  second = filebase64("testdata/above.pfx")
  third  = filebase64("testdata/above.pfx")
}
`
}

func beside() string {
	return fmt.Sprintf("%s", "testdata/beside.ps1") // want `testdata/beside\.ps1 does not exist: a test reads test data from its own folder, and the nearest copy is \.\./other/testdata/beside\.ps1`
}

func nested() string {
	return "../../testdata/certs/cert_2.pem" // want `\.\./\.\./testdata/certs/cert_2\.pem does not exist: a test reads test data from its own folder$`
}

// a name finished when the test runs, the folder on its own and the end of a longer path are not looked up
func unknown(name string) []string {
	return []string{
		fmt.Sprintf("testdata/cert_%d.pem", 1),
		fmt.Sprintf("testdata/%s.json", name),
		filepath.Join("testdata", name),
		"https://example.com/repo/testdata/missing.xml",
		"mytestdata/missing.xml",
	}
}

func ignored() string {
	return "testdata/ignored.xml" //azignore:AZT009 - written by the test before it is read
}

// inside a string that runs over many lines, the comment goes where the string starts
func ignoredAbove() string {
	//azignore:AZT009 - written by the test before it is read
	return `
resource "azurerm_thing" "test" {
  first = file("testdata/ignored_above.xml")
}
`
}
