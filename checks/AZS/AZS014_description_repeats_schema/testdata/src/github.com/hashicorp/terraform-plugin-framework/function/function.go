// Package function is a minimal stub of the plugin framework's function package for the analyzer tests.
package function

type Parameter interface{ isParameter() }

type Definition struct {
	Summary             string
	Description         string
	MarkdownDescription string
	Parameters          []Parameter
}

type StringParameter struct {
	Name                string
	Description         string
	MarkdownDescription string
}

func (StringParameter) isParameter() {}
