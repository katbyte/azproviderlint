package azt003

import "testing"

// Should be flagged: in-package test files are checked too
func TestAccBuildName(t *testing.T) { // want `TestAccBuildName does not run an acceptance test, name it TestBuildName so it is not picked up as one`
	if BuildName("a") != "a-thing" {
		t.Fatal("wrong name")
	}
}
