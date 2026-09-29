package helpers

// FlattenStringSlice handles nil input itself.
func FlattenStringSlice(input *[]string) []interface{} {
	if input == nil {
		return []interface{}{}
	}
	out := make([]interface{}, 0, len(*input))
	for _, v := range *input {
		out = append(out, v)
	}
	return out
}
