// Package acceptance is a minimal stub of the provider's internal/acceptance for the analyzer tests.
package acceptance

import "testing"

type TestStep struct{ Config string }

type TestData struct{ ResourceName string }

func BuildTestData(t *testing.T, resourceType, label string) TestData { return TestData{} }

func (TestData) ResourceTest(t *testing.T, r any, steps []TestStep) {}

func (TestData) ResourceSequentialTest(t *testing.T, r any, steps []TestStep) {}

func (TestData) DataSourceTest(t *testing.T, steps []TestStep) {}

// Shared is a test resource kept in another package.
type Shared struct{}
