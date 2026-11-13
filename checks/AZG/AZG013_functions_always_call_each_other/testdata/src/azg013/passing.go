package azg013

import (
	"fmt"
	"os"
)

// a way out before the call
func countDown(n int) int {
	if n == 0 {
		return 0
	}

	return countBack(n)
}

func countBack(n int) int {
	return countDown(n - 1)
}

// the call is only made on one branch
func walk(n int) {
	if n > 0 {
		walkBack(n)
	}
}

func walkBack(n int) {
	walk(n - 1)
}

// the right side of && and || may never run
func even(n int) bool {
	return n == 0 || odd(n-1)
}

func odd(n int) bool {
	return n != 0 && even(n-1)
}

// only made in a loop, a switch, or a function written in place
func loops(items []int) {
	for range items {
		loopsBack(items)
	}
}

func loopsBack(items []int) {
	loops(items[1:])
}

func switches(n int) {
	switch n {
	case 1:
		switchesBack(n)
	}
}

func switchesBack(n int) {
	switches(n)
}

func later() func() {
	return func() { laterBack() }
}

func laterBack() {
	later()
}

// deferred and started calls are left alone
func deferred() {
	defer deferredBack()
}

func deferredBack() {
	go deferred()
}

// nothing after a panic or an exit is sure to run
func panics() {
	panic("not implemented")
	panicsBack()
}

func panicsBack() {
	panics()
}

func exits() {
	os.Exit(1)
	exitsBack()
}

func exitsBack() {
	exits()
}

// one function calling itself is staticcheck's to report
func forever() {
	forever()
}

// a chain that does not come back
func top() string {
	return middle()
}

func middle() string {
	return bottom()
}

func bottom() string {
	return fmt.Sprint("bottom")
}

// the same method name on two types is two functions
type outer struct{ in inner }

type inner struct{}

func (o outer) read() string {
	return o.in.read()
}

func (inner) read() string {
	return "read"
}

// a call through an interface or a variable could go anywhere
type reader interface{ read() string }

func throughInterface(r reader) string {
	return r.read()
}

func throughVariable(next func()) {
	next()
}

func throughVariableBack() {
	throughVariable(throughVariableBack)
}
