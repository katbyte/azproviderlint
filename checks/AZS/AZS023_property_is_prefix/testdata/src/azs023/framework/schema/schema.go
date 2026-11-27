// Package schema is a minimal stub of the plugin framework's resource schema for the analyzer tests.
package schema

type (
	Attribute interface{}
	Block     interface{}
)

type Schema struct {
	Attributes map[string]Attribute
	Blocks     map[string]Block
}

type BoolAttribute struct {
	Optional bool
	Computed bool
}

type ListNestedBlock struct {
	NestedObject NestedBlockObject
}

type NestedBlockObject struct {
	Attributes map[string]Attribute
}
