// Package pluginsdk mirrors terraform-provider-azurerm's internal/tf/pluginsdk type aliases.
package pluginsdk

import "azs025/schema"

type (
	Schema   = schema.Schema
	Resource = schema.Resource
)

const (
	TypeBool   = schema.TypeBool
	TypeInt    = schema.TypeInt
	TypeFloat  = schema.TypeFloat
	TypeString = schema.TypeString
	TypeList   = schema.TypeList
	TypeSet    = schema.TypeSet
	TypeMap    = schema.TypeMap
)
