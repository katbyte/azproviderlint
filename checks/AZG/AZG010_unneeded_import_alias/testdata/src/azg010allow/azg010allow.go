package azg010allow

import (
	s "strings"

	iothub "example.com/iothub"
	networkValidate "example.com/network/validate"
	helperTags "example.com/tags" // want `import alias "helperTags" is not needed, the package name "tags" does not clash`
)

func use() {
	_ = s.ToUpper("x")
	iothub.New()
	networkValidate.SubnetID()
	helperTags.Expand()
}
