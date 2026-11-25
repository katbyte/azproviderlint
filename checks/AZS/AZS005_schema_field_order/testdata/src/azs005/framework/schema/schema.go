// Package schema is a minimal stub of the plugin framework's resource schema for the analyzer tests.
package schema

type Attribute interface{}

type Schema struct {
	Attributes  map[string]Attribute
	Description string
}

type StringAttribute struct {
	Required            bool
	Optional            bool
	Computed            bool
	Sensitive           bool
	Description         string
	MarkdownDescription string
	Validators          []any
	PlanModifiers       []any
}

type ListNestedAttribute struct {
	NestedObject NestedAttributeObject
	Optional     bool
	Description  string
	Validators   []any
}

type NestedAttributeObject struct {
	Attributes map[string]Attribute
}
