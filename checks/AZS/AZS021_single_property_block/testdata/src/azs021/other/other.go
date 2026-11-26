// Package other has types that look like a schema and are not one.
package other

type Schema struct {
	MaxItems int
	Elem     any
}

type Resource struct {
	Schema map[string]*Schema
}
