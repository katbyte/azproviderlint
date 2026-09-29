package azp004nogen

type autoRegistration struct{}

// Not flagged: generated=false skips registration_gen.go files.
func (autoRegistration) WebsiteCategories() []string {
	return []string{
		"Compute",
		"Chaos Studio",
	}
}
