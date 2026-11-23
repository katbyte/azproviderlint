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
	Computed            bool
	Description         string
	MarkdownDescription string
	Default             any
	Validators          []any
	PlanModifiers       []any
}

func (StringAttribute) isAttribute() {}

type BoolAttribute struct {
	Optional            bool
	Default             any
	Description         string
	MarkdownDescription string
}

func (BoolAttribute) isAttribute() {}

type ListNestedBlock struct {
	Description         string
	MarkdownDescription string
	Validators          []any
}

func (ListNestedBlock) isBlock() {}
