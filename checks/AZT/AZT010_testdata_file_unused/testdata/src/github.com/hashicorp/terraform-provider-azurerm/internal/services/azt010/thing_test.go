package azt010_test

import "fmt"

func config() string {
	return fmt.Sprintf(`
resource "azurerm_thing" "test" {
  first  = file("testdata/used.xml")
  second = file("testdata/certs/cert_%d.pem")
  third  = "./testdata/scripts"
}
`, 1)
}
