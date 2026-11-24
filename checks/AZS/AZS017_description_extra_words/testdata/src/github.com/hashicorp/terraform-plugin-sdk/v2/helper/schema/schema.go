// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeString
	TypeList
)

type Schema struct {
	Type        ValueType
	Optional    bool
	Computed    bool
	Description string
	Elem        any
}

type Resource struct {
	Description string
	Schema      map[string]*Schema
}
