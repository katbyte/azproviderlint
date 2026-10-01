// Package resource is a minimal stub of the terraform-plugin-testing harness.
package resource

const EnvTfAcc = "TF_ACC"

// T mirrors the harness's own testing interface, which *testing.T satisfies.
type T interface {
	Fatal(args ...any)
}

type TestStep struct {
	Config string
}

type TestCase struct {
	Steps []TestStep
}

func Test(t T, c TestCase) {}

func ParallelTest(t T, c TestCase) {}
