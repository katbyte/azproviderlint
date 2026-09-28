package azg010renames

import (
	s "strings"

	set "example.com/go-set/v3" // want `import alias "set" repeats the package name`
	keyVaultClient "example.com/keyvault/client"
	validate2 "example.com/network/validate" // want `import alias "validate2" is not needed, the package name "validate" does not clash`
)

func use() {
	_ = s.ToUpper("x")
	_ = set.Set{}
	_ = keyVaultClient.Client{}
	validate2.SubnetID()
}
