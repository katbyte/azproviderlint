package azg009

// _test.go files are checked by default (tests=true).
func derefInTest(m model) {
	use(m.Properties.Count) // want "selecting `Count` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
}
