package thing_test

func config() string {
	return `
resource "azurerm_thing" "test" {
  first  = filebase64("../../testdata/used_from_beneath.pfx")
  second = file("testdata/own.json")
}
`
}
