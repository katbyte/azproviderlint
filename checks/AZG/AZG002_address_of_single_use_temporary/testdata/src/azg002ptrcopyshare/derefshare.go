package azg002ptrcopyshare

type props struct{ Name *string }

// fix-pointer-copy=share substitutes the original pointer, dropping the copy.
func derefShare(p props) *string {
	out := *p.Name // want `"out" is only used as an address by the following statement and should be inlined with new\(\) - or, if sharing the original pointer is acceptable, use p\.Name directly`
	return &out
}
