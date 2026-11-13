package azg013

import "fmt"

type resource struct{}

// the shape that started this: a test config helper copied without the template it was written against
func (r resource) template(n int) string {
	return fmt.Sprintf("%s %d", jobStepBasic(n), n) // want `resource\.template always calls jobStepBasic, which always calls resource\.template again, so neither returns`
}

func jobStepBasic(n int) string {
	r := resource{}

	return fmt.Sprintf("%s basic", r.template(n))
}

func first() {
	second() // want `first always calls second, which always calls third, which always calls first again, so none of them return`
}

func second() {
	fmt.Println("second")
	third()
}

func third() {
	x := 1
	{
		x++
	}

	defer fmt.Println(x)

	first()
}

// before the branch is taken, what the branch asks is always worked out
func asks(n int) bool {
	if answers(n) { // want `asks always calls answers, which always calls asks again, so neither returns`
		return true
	}

	return false
}

func answers(n int) bool {
	var ok = asks(n) || n > 0

	return ok
}

type pointer struct{}

func (p *pointer) one() int {
	return p.two() + 1 // want `pointer\.one always calls pointer\.two, which always calls pointer\.one again, so neither returns`
}

func (p *pointer) two() int {
	return p.one() + 1
}

func generic[T any](v T) T {
	return other(v) // want `generic always calls other, which always calls generic again, so neither returns`
}

func other[T any](v T) T {
	return generic[T](v)
}
