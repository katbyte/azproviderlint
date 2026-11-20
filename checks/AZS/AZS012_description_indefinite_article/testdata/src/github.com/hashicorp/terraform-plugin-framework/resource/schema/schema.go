// Package schema is a minimal stub of the plugin framework's resource/schema for the analyzer tests.
package schema

type Attribute interface{ isAttribute() }

type Schema struct {
	Description         string
	MarkdownDescription string
	Attributes          map[string]Attribute
}

type StringAttribute struct {
	Optional            bool
	Description         string
	MarkdownDescription string
}

func (StringAttribute) isAttribute() {}
