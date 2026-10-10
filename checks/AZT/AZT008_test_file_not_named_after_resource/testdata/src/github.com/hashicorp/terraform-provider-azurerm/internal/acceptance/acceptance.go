// Package acceptance is a minimal stub of the provider's acceptance test framework.
package acceptance

import "testing"

type TestStep struct {
	Config string
}

type TestData struct {
	ResourceName string
}

func BuildTestData(t *testing.T, resourceType, resourceLabel string) TestData {
	return TestData{ResourceName: resourceType + "." + resourceLabel}
}

func (td TestData) ResourceTest(t *testing.T, r any, steps []TestStep) {}

func (td TestData) DataSourceTest(t *testing.T, steps []TestStep) {}
