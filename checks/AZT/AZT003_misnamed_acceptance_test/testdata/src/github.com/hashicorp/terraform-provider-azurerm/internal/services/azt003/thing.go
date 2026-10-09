package azt003

import (
	"os"
	"testing"
)

func BuildName(prefix string) string {
	return prefix + "-thing"
}

// Should NOT be flagged: go test only runs functions from _test.go files
func TestAccNotInATestFile(t *testing.T) {
	_ = BuildName("a")
}

// Configured is production code that reads TF_ACC to pick a default; calling it does not make
// a test an acceptance test
func Configured() bool {
	return os.Getenv("TF_ACC") != ""
}
