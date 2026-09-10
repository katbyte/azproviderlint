// Package fwschema is a minimal stand-in for the terraform-plugin-framework schema packages
// used only by the AZP003 analysistest fixtures.
package fwschema

// Schema is a minimal stand-in for the framework schema type.
type Schema struct {
	Attributes map[string]Attribute
	Blocks     map[string]Block
}

// Attribute is a minimal stand-in for the framework attribute interface.
type Attribute interface{}

// Block is a minimal stand-in for the framework block interface.
type Block interface{}

// StringAttribute is a minimal concrete attribute.
type StringAttribute struct {
	Required bool
	Computed bool
}
