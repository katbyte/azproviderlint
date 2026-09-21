// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeList
	TypeMap
	TypeSet
)

type SchemaConfigMode int

const (
	SchemaConfigModeAuto SchemaConfigMode = iota
	SchemaConfigModeAttr
	SchemaConfigModeBlock
)

type (
	SchemaDefaultFunc      func() (any, error)
	SchemaValidateFunc     func(any, string) ([]string, []error)
	SchemaValidateDiagFunc func(any, string) error
	SchemaDiffSuppressFunc func(string, string, string, any) bool
	SchemaStateFunc        func(any) string
)

type Schema struct {
	Type                  ValueType
	ConfigMode            SchemaConfigMode
	Optional              bool
	Required              bool
	Computed              bool
	ForceNew              bool
	Sensitive             bool
	WriteOnly             bool
	Description           string
	Default               any
	DefaultFunc           SchemaDefaultFunc
	InputDefault          string
	ValidateFunc          SchemaValidateFunc
	ValidateDiagFunc      SchemaValidateDiagFunc
	DiffSuppressFunc      SchemaDiffSuppressFunc
	DiffSuppressOnRefresh bool
	StateFunc             SchemaStateFunc
	MaxItems              int
	MinItems              int
	AtLeastOneOf          []string
	ExactlyOneOf          []string
	ConflictsWith         []string
	RequiredWith          []string
	Elem                  any
}

type Resource struct {
	Schema map[string]*Schema
}
