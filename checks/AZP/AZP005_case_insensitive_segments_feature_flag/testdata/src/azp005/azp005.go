package azp005

type Features struct {
	TreatUserSpecifiedSegmentsAsCaseInsensitive bool
	SomethingElse                               bool
}

// Should be flagged: configuring the case-aware comparisons feature flag
func badAssign(f *Features) {
	f.TreatUserSpecifiedSegmentsAsCaseInsensitive = true // want `TreatUserSpecifiedSegmentsAsCaseInsensitive must not be set, the case-aware comparisons feature is not ready for use`
}

// Should NOT be flagged: other feature flags
func goodAssign(f *Features) {
	f.SomethingElse = true
}

// Should NOT be flagged: reading the flag
func goodRead(f *Features) bool {
	return f.TreatUserSpecifiedSegmentsAsCaseInsensitive
}
