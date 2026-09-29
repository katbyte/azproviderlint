package azg010

import (
	"example.com/compute/validate"
	networkValidate "example.com/network/validate" // the unaliased import already takes the name validate
)

func unaliased() {
	validate.VirtualMachineName()
	networkValidate.SubnetID()
}
