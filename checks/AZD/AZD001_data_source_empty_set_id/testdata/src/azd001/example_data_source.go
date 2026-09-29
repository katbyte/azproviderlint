package azd001

// Should be flagged: SetId("") in a data source file
func exampleDataSourceRead(d *ResourceData) error {
	d.SetId("") // want `data sources should return an error when a resource cannot be found instead of calling SetId with an empty string`
	return nil
}

// Should NOT be flagged: setting a real ID
func exampleDataSourceReadGood(d *ResourceData, id string) error {
	d.SetId(id)
	return nil
}

func (d *ResourceData) Set(key string, value interface{}) {}

func markGone(id string) {}

// Should NOT be flagged: two-argument calls are not SetId
func exampleDataSourceSetEmpty(d *ResourceData) error {
	d.Set("name", "")
	return nil
}

// Should NOT be flagged: a plain function call, not a method on d
func exampleDataSourceHelperCall(d *ResourceData) error {
	markGone("")
	return nil
}

// Should NOT be flagged: a different method taking one string
func exampleDataSourceOtherMethod(d *ResourceData) error {
	d.SetType("")
	return nil
}

func (d *ResourceData) SetType(t string) {}

// Should NOT be flagged: a non-literal argument
func exampleDataSourceNonLiteral(d *ResourceData) error {
	empty := ""
	d.SetId(empty)
	return nil
}
