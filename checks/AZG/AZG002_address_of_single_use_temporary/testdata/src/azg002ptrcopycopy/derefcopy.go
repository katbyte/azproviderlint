package azg002ptrcopycopy

type props struct{ Name *string }

// fix-pointer-copy=copy keeps the copy-preserving inline.
func derefCopy(p props) *string {
	out := *p.Name // want `"out" is only used as an address by the following statement and should be inlined with new\(\) - or, if sharing the original pointer is acceptable, use p\.Name directly`
	return &out
}
