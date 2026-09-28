package azg010mismatched

// devices is the package's real name, which its path does not show; goimports writes it, and
// an azignore on the import is how a consumer calls out the mismatch
import devices "example.com/iothub" // want `import alias "devices" repeats the package name`

func realName() {
	devices.New()
}
