// Package acceptance is a minimal stub of the provider's acceptance test framework.
package acceptance

import "testing"

// TestStep mirrors the harness step; azurerm aliases it from terraform-plugin-testing.
type TestStep struct {
	Config      string
	ImportState bool
}

type TestData struct {
	ResourceName string
}

func BuildTestData(t *testing.T, resourceType, resourceLabel string) TestData {
	return TestData{ResourceName: resourceType + "." + resourceLabel}
}

func (td TestData) ResourceTest(t *testing.T, r any, steps []TestStep) {}

func (td TestData) ImportStep() TestStep {
	return TestStep{ImportState: true}
}
