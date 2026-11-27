// Package schema is a minimal stub of the plugin framework's resource schema for the analyzer tests.
package schema

type Attribute interface{}

type Schema struct {
	Attributes map[string]Attribute
}

type StringAttribute struct {
	Required           bool
	Optional           bool
	Validators         []any
	DeprecationMessage string
}
