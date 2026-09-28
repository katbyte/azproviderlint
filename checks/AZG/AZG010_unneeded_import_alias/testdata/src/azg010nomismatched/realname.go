package azg010nomismatched

// devices is the package's real name, which its path does not show; goimports writes it
import devices "example.com/iothub"

func realName() {
	devices.New()
}
