// Package pluginsdk is a minimal stand-in for azurerm's plugin SDK wrapper used only by the
// AZP003 analysistest fixtures.
package pluginsdk

// Resource is a minimal stand-in for the plugin SDK resource type.
type Resource struct {
	Schema map[string]*Schema
}

// Schema is a minimal stand-in for the plugin SDK schema type.
type Schema struct {
	Type      int
	Required  bool
	Optional  bool
	Computed  bool
	WriteOnly bool
	Sensitive bool
	Elem      interface{}
}

const (
	TypeString = iota
	TypeList
)
