package azg010ignore

import (
	s "strings" // want `import alias "s" is not needed, the package name "strings" does not clash`

	datadog "example.com/datadog/2021-03-01"
)

func use() {
	_ = s.ToUpper("x")
	datadog.NewClient()
}
