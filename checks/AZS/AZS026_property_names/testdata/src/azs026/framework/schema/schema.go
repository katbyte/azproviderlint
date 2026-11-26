// Package schema is a minimal stub of the plugin framework's resource schema for the analyzer tests.
package schema

type Attribute interface{}

type Schema struct {
	Attributes map[string]Attribute
}

type BoolAttribute struct {
	Optional           bool
	Computed           bool
	DeprecationMessage string
}
