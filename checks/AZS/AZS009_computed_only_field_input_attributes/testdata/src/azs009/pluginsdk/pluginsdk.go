// Package pluginsdk mirrors terraform-provider-azurerm's internal/tf/pluginsdk type aliases.
package pluginsdk

import "azs009/schema"

type (
	Schema   = schema.Schema
	Resource = schema.Resource
)

const (
	TypeBool   = schema.TypeBool
	TypeInt    = schema.TypeInt
	TypeString = schema.TypeString
	TypeList   = schema.TypeList
	TypeSet    = schema.TypeSet
)
