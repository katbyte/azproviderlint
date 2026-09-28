package azg010

import (
	validate "example.com/keyvault/client" // want `import alias "validate" is not needed, the package name "client" does not clash`

	// the import above goes by the name validate until its alias is removed, so this one is
	// only found by a second run
	networkValidate "example.com/network/validate"
)

func aliasClash() {
	_ = validate.Client{}
	networkValidate.SubnetID()
}
