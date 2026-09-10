package returns

type T struct{ N int }

func newT() *T { return &T{} } // want newT:"nonnil\\(0\\)"

func viaNew() *T { return new(T) } // want viaNew:"nonnil\\(0\\)"

func viaLocal() *T { // want viaLocal:"nonnil\\(0\\)"
	out := &T{}
	out.N = 1
	return out
}

func viaNamed() (out *T) { // want viaNamed:"nonnil\\(0\\)"
	out = &T{}
	return
}

func viaCallee() *T { return newT() } // want viaCallee:"nonnil\\(0\\)"

func viaLater() *T { return declaredLater() } // want viaLater:"nonnil\\(0\\)"

func declaredLater() *T { return &T{} } // want declaredLater:"nonnil\\(0\\)"

func withErr(fail bool) (*T, error) { // want withErr:"nonnil\\(0\\)"
	if fail {
		return &T{}, errFail
	}
	return &T{N: 1}, nil
}

func passthrough() (*T, error) { return withErr(false) } // want passthrough:"nonnil\\(0\\)"

func second() (int, *T) { return 0, &T{} } // want second:"nonnil\\(1\\)"

func (t *T) clone() *T { return &T{N: t.N} } // want clone:"nonnil\\(0\\)"

// no fact: one path returns nil
func maybe(fail bool) *T {
	if fail {
		return nil
	}
	return &T{}
}

// no fact: a parameter passes through
func echo(t *T) *T { return t }

// no fact: an unknown callee
func unknown() *T { return external() }

// no fact: the local is reassigned from an unknown source before the return
func reassigned(t *T) *T {
	out := &T{}
	out = t
	return out
}

// no fact: the naked return's named result was never assigned
func namedZero() (out *T) { return }

// no fact: a nested literal's return does not count for the enclosing function
func nested() *T {
	f := func() *T { return &T{} }
	return f()
}

// no fact: non-pointer results are not tracked
func value() T { return T{} }

// no fact: the guard is on another path than the returned one
func guardedOther(a, b *T) *T {
	if a == nil {
		return &T{}
	}
	return b
}

// self-recursion converges to no fact
func recursive(n int) *T {
	if n == 0 {
		return recursive(1)
	}
	return recursive(n - 1)
}

var errFail error

func external() *T

// no fact: a write nested in a block after the non-nil assignment
func nestedNil(fail bool) *T {
	out := &T{}
	if fail {
		out = nil
	}
	return out
}

// no fact: a loop body rewrites the local before a later iteration's return
func loopNil(items []int) (out *T) {
	out = &T{}
	for _, i := range items {
		if i == 0 {
			return
		}
		out = nil
	}
	return out
}
