// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeString
	TypeList
	TypeSet
)

type Schema struct {
	Type        ValueType
	Optional    bool
	Computed    bool
	Description string
	Elem        any
	Set         any
}

type Resource struct {
	Description string
	Read        func() error
	ReadContext func() error
	Schema      map[string]*Schema
}

func HashSchema(*Schema) any { return nil }
