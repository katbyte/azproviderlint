package azg010ignore

import (
	s "strings" // want `import alias "s" is not needed, the package name "strings" does not clash`

	// the package names v2021_03_01 and devices each match one of the patterns
	datadog "example.com/datadog/2021-03-01"
	iothub "example.com/iothub"
)

func use() {
	_ = s.ToUpper("x")
	datadog.NewClient()
	iothub.New()
}
