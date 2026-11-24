// Package backups is a minimal stub of an Azure SDK package, for values the provider takes from outside its own code.
package backups

type Action string

const (
	ActionStopProtection Action = "stop_protection"
	ActionSuspendBackups Action = "suspend_backups"
)
