package azg010

import (
	// a parameter, a named result and a receiver in this file are named like the packages
	helperValidate "example.com/compute/validate"
	helperSet "example.com/go-set/v3"
	helperAccounts "example.com/storage/accounts"

	// tags is declared at package level in another file of this package
	helperTags "example.com/tags"
)

type holder struct{}

func parameter(set helperSet.Set) helperSet.Set {
	return set
}

func namedResult() (accounts helperAccounts.Account) {
	return accounts
}

func (validate holder) receiver() {
	helperValidate.VirtualMachineName()
	helperTags.Expand()
}
