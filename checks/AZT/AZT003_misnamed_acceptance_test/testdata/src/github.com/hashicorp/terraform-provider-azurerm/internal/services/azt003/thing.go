package azt003

import "testing"

func BuildName(prefix string) string {
	return prefix + "-thing"
}

// Should NOT be flagged: go test only runs functions from _test.go files
func TestAccNotInATestFile(t *testing.T) {
	_ = BuildName("a")
}
