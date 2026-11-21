// Package schema is a minimal stub of the plugin framework's resource/schema for the analyzer tests.
package schema

type Attribute interface{ isAttribute() }

type Block interface{ isBlock() }

type Schema struct {
	Description         string
	MarkdownDescription string
	Attributes          map[string]Attribute
	Blocks              map[string]Block
}

type StringAttribute struct {
	Optional            bool
	Description         string
	MarkdownDescription string
}

func (StringAttribute) isAttribute() {}

type NestedBlockObject struct {
	Attributes map[string]Attribute
}

type ListNestedBlock struct {
	Description         string
	MarkdownDescription string
	NestedObject        NestedBlockObject
}

func (ListNestedBlock) isBlock() {}
