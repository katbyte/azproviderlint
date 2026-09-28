package azg010

import (
	s "strings" // want `import alias "s" is not needed, the package name "strings" does not clash`

	datadog "example.com/datadog/2021-03-01"       // want `import alias "datadog" is not needed, the package name "v2021_03_01" does not clash`
	set "example.com/go-set/v3"                    // want `import alias "set" repeats the package name`
	iothub "example.com/iothub"                    // want `import alias "iothub" is not needed, the package name "devices" does not clash`
	networkValidate "example.com/network/validate" // want `import alias "networkValidate" is not needed, the package name "validate" does not clash`
)

// fields and methods live in their own namespace, so sharing the package name is fine
type widget struct {
	strings []string
}

func (widget) validate() {}

// and so do labels
func labelled() {
validate:
	for {
		break validate
	}
}

func use() {
	_ = s.ToUpper("x")
	_ = set.Set{}
	datadog.NewClient()
	iothub.New()
	networkValidate.SubnetID()
}

// a local that only shares the alias's name is left as it is
func other() {
	s := 1
	_ = s
}
