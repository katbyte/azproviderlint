package azg010nogenerated

import (
	s "strings" // want `import alias "s" is not needed, the package name "strings" does not clash`
)

func written() {
	_ = s.ToUpper("x")
}
