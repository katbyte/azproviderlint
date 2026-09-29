package azg010ownpackage

import sdk "example.com/sdk/azg010ownpackage" // want `import alias "sdk" is not needed, the package name "azg010ownpackage" does not clash`

func use() {
	sdk.New()
}
