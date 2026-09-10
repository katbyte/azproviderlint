package nilguard

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestReturnsAnalyzer(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), ReturnsAnalyzer, "returns")
}
