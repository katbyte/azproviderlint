package azg000

// Should NOT be flagged: directive with a dash-separated reason
func withReason() {
	err := doSomething() //azignore:AZG001 - legacy pattern, refactor tracked in #42
	_ = err
}

// Should NOT be flagged: reason without a dash separator
func withReasonNoDash() {
	err := doSomething() //azignore:AZG001 the combined form obscures the retry loop
	_ = err
}

// Should NOT be flagged: multiple rules with a reason
func multipleRulesWithReason() {
	err := doSomething() //azignore:AZG001,AZR001 - deliberate
	_ = err
}

// Should NOT be flagged: em dash separator
func emDashReason() {
	err := doSomething() //azignore:AZP003 — deliberately not exposed in the data source
	_ = err
}

// Should NOT be flagged: directive with a reason on the line preceding the code
func reasonOnLineAbove() {
	//azignore:AZG001 - the combined form obscures the retry loop here
	err := doSomething()
	_ = err
}

// Should NOT be flagged: prose mentioning the directive syntax, not a directive itself
// (suppress reports with //azignore:AZG001 when needed)
func proseMention() {}

// Should be flagged: bare directive at the end of the line
func bareSameLine() {
	err := doSomething() /* want `azignore directive must include a reason` */ //azignore:AZG001
	_ = err
}

// Should be flagged: bare directive on its own line
func bareLineAbove() {
	/* want `azignore directive must include a reason` */ //azignore:AZG001
	err := doSomething()
	_ = err
}

// Should be flagged: bare directive listing multiple rules
func bareList() {
	err := doSomething() /* want `azignore directive must include a reason` */ //azignore:AZG001, AZR001
	_ = err
}

// Should be flagged: a dash with nothing after it is not a reason
func emptyReasonAfterDash() {
	err := doSomething() /* want `azignore directive must include a reason` */ //azignore:AZG001 -
	_ = err
}

// Should be flagged: the gofmt-spaced form is still a directive
func bareGofmtSpaced() {
	err := doSomething() /* want `azignore directive must include a reason` */ // azignore:AZG001
	_ = err
}

func doSomething() error {
	return nil
}
