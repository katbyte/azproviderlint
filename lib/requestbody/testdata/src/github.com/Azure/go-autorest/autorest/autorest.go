// Package autorest is a minimal stand-in for go-autorest's preparers, shaped like the real
// ones: AsPut carries the method as a string literal, WithJSON marshals inside a closure.
package autorest

import "encoding/json"

type PrepareDecorator func()

func WithMethod(method string) PrepareDecorator { return func() { _ = method } }

func AsPut() PrepareDecorator   { return WithMethod("PUT") }   // want AsPut:"writes"
func AsPatch() PrepareDecorator { return WithMethod("PATCH") } // want AsPatch:"writes"
func AsGet() PrepareDecorator   { return WithMethod("GET") }

func WithJSON(v interface{}) PrepareDecorator { // want WithJSON:"serialises\\(0\\)"
	return func() {
		_, _ = json.Marshal(v)
	}
}

func CreatePreparer(d ...PrepareDecorator) int { return 0 }
