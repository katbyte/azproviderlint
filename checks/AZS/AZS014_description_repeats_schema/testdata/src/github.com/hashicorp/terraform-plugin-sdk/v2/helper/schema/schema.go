// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeInt
	TypeString
	TypeList
)

type SchemaValidateFunc func(any, string) ([]string, []error)

type Schema struct {
	Type          ValueType
	Required      bool
	Optional      bool
	Computed      bool
	Sensitive     bool
	ForceNew      bool
	Default       any
	ValidateFunc  SchemaValidateFunc
	ConflictsWith []string
	ExactlyOneOf  []string
	AtLeastOneOf  []string
	RequiredWith  []string
	MaxItems      int
	MinItems      int
	Description   string
	Elem          any
}

type Resource struct {
	Description string
	Schema      map[string]*Schema
}
