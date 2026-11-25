package azg010allowpattern

import (
	iothubBeta "example.com/iothub"
	helperTags "example.com/tags" // want `import alias "helperTags" is not needed, the package name "tags" does not clash`
)

func use() {
	iothubBeta.New()
	helperTags.Expand()
}
