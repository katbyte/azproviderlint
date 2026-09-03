package azg009notests

type properties struct {
	Count *int
}

type model struct {
	Properties *properties
}

// Not flagged: tests=false skips _test.go files.
func derefInSkippedTest(m model) {
	*m.Properties.Count = 1
}
