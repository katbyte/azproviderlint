package azs029

import (
	"azs029/pluginsdk"
	"azs029/validation"
)

// the value, its version and the property it stands in for, each tied to the others
func wired() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"password": {
			Type:          pluginsdk.TypeString,
			Optional:      true,
			Sensitive:     true,
			ConflictsWith: []string{"password_wo"},
		},

		"password_wo": {
			Type:          pluginsdk.TypeString,
			Optional:      true,
			WriteOnly:     true,
			RequiredWith:  []string{"password_wo_version"},
			ConflictsWith: []string{"password"},
		},

		"password_wo_version": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			RequiredWith: []string{"password_wo"},
			ValidateFunc: validation.IntAtLeast(1),
		},
	}
}

func unwired() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"password": { // want `"password" and "password_wo" can both be set - add ConflictsWith: \[\]string\{"password_wo"\}`
			Type:      pluginsdk.TypeString,
			Optional:  true,
			Sensitive: true,
		},

		"password_wo": { // want `"password_wo" can be set without "password_wo_version" - add RequiredWith: \[\]string\{"password_wo_version"\}` `"password_wo" and "password" can both be set - add ConflictsWith: \[\]string\{"password"\}`
			Type:      pluginsdk.TypeString,
			Optional:  true,
			WriteOnly: true,
		},

		"password_wo_version": { // want `"password_wo_version" can be set without "password_wo" - add RequiredWith: \[\]string\{"password_wo"\}` `"password_wo_version" can be set to 0, which is what an unset version is stored as - add ValidateFunc: validation.IntAtLeast\(1\)`
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},

		"key_wo": { // want `write-only property "key_wo" has no "key_wo_version" beside it - nothing says when its value has changed`
			Type:      pluginsdk.TypeString,
			Optional:  true,
			WriteOnly: true,
		},
	}
}

var together = []string{"token", "token_wo"}

// exactly one of the two is as good as one conflicting with the other, a name can be the whole path to a property in a block, and a list that is
// not written out could name anything
func otherWays() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"value": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"value", "value_wo"},
		},

		"value_wo": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			WriteOnly:    true,
			RequiredWith: []string{"secret.0.value_wo_version"},
			ExactlyOneOf: []string{"value", "value_wo"},
		},

		"value_wo_version": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			RequiredWith: []string{"secret.0.value_wo"},
			ValidateFunc: validation.IntBetween(1, 100),
		},

		"token": {
			Type:          pluginsdk.TypeString,
			Optional:      true,
			ConflictsWith: together,
		},

		"token_wo": {
			Type:          pluginsdk.TypeString,
			Optional:      true,
			WriteOnly:     true,
			RequiredWith:  []string{"token_wo_version"},
			ConflictsWith: together,
		},

		// a version from a function that hands back nothing readable is there, but what it says cannot be read
		"token_wo_version": version(),

		// a property that is only read back cannot be set beside the write-only one
		"route_table_id": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"route_table_id_wo": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			WriteOnly:    true,
			RequiredWith: []string{"route_table_id_wo_version"},
		},

		"route_table_id_wo_version": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			RequiredWith: []string{"route_table_id_wo"},
			ValidateFunc: validation.IntAtLeast(1),
		},
	}
}

func version() *pluginsdk.Schema { return nil }

// two properties that are always set cannot be set one without the other
func required() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"connection_string_wo": {
			Type:      pluginsdk.TypeString,
			Required:  true,
			WriteOnly: true,
		},

		"connection_string_wo_version": {
			Type:         pluginsdk.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},

		// a value that is always set can still be set without a version that is not, and that version still has to be at least 1
		"key_wo": { // want `"key_wo" can be set without "key_wo_version"`
			Type:      pluginsdk.TypeString,
			Required:  true,
			WriteOnly: true,
		},

		"key_wo_version": { // want `"key_wo_version" can be set to 0, which is what an unset version is stored as`
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},
	}
}

// a property whose schema a function hands back is read for what it is, but its lists could name anything
func fromFunctions() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"secret_wo": writeOnly(), // want `write-only property "secret_wo" has no "secret_wo_version" beside it`

		"pin_wo": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			WriteOnly:    true,
			RequiredWith: []string{"pin_wo_version"},
		},

		"pin_wo_version": anyVersion(), // want `"pin_wo_version" can be set to 0, which is what an unset version is stored as`

		"phrase": optional(),

		"phrase_wo": { // want `"phrase_wo" and "phrase" can both be set - add ConflictsWith: \[\]string\{"phrase"\}`
			Type:         pluginsdk.TypeString,
			Optional:     true,
			WriteOnly:    true,
			RequiredWith: []string{"phrase_wo_version"},
		},

		"phrase_wo_version": countedVersion(),
	}
}

func writeOnly() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:      pluginsdk.TypeString,
		Optional:  true,
		WriteOnly: true,
	}
}

func anyVersion() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeInt,
		Optional: true,
	}
}

func countedVersion() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:         pluginsdk.TypeInt,
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(1),
	}
}

func optional() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeString,
		Optional: true,
	}
}
