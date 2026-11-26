// Package cty is a minimal stub of github.com/zclconf/go-cty/cty for the analyzer tests.
package cty

type Value struct{}

func (Value) IsNull() bool                 { return false }
func (Value) IsKnown() bool                { return true }
func (Value) IsWhollyKnown() bool          { return true }
func (Value) GetAttr(string) Value         { return Value{} }
func (Value) LengthInt() int               { return 0 }
func (Value) AsValueSlice() []Value        { return nil }
func (Value) AsValueMap() map[string]Value { return nil }
func (Value) AsString() string             { return "" }
func (Value) True() bool                   { return false }
