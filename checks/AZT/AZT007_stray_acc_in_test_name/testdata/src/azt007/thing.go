package azt007

import "testing"

// Should NOT be flagged: go test only runs functions from _test.go files
func TestAccThing_storageAccNotATestFile(t *testing.T) {}
