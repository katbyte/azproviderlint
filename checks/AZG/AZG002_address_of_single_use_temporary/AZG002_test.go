package AZG002

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG002(t *testing.T) {
	t.Parallel()

	dir := analysistest.TestData()

	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg002")

	// use/allow are package state read during run, so the mode fixtures must run sequentially
	// within the same test rather than as parallel siblings
	use = usePointerTo
	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg002pointerto")
	use = useNew
	allow = usePointerTo
	analysistest.Run(t, dir, Analyzer, "azg002allow")
	allow = ""

	fixPointerCopy = fixPointerCopyCopy
	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg002ptrcopycopy")
	fixPointerCopy = fixPointerCopyShare
	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg002ptrcopyshare")
	fixPointerCopy = fixPointerCopyNone
}
