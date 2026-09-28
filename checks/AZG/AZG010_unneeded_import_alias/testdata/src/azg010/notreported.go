package azg010

import (
	_ "example.com/tags"

	// the path suggests "iothub", so goimports writes the package name explicitly
	devices "example.com/iothub"

	// both share the package name validate, so dropping either alias is only safe while the other stays
	computeValidate "example.com/compute/validate"
	networkValidate "example.com/network/validate"

	// a local variable in this file is named client
	keyVaultClient "example.com/keyvault/client"

	// the package declares tags at package level
	helperTags "example.com/tags"

	// max is a builtin
	limits "example.com/limits"

	// a type switch in this file names its variable accounts
	storageAccounts "example.com/storage/accounts"

	// the file's own package is named azg010
	sdk "example.com/sdk/azg010"
)

var tags = []string{}

func notReported() {
	devices.New()
	computeValidate.VirtualMachineName()
	networkValidate.SubnetID()
	helperTags.Expand()
	_ = limits.Items
	sdk.New()

	client := keyVaultClient.Client{}
	_ = client
}

func typeSwitch(v any) {
	switch accounts := v.(type) {
	case storageAccounts.Account:
		_ = accounts
	}
}
