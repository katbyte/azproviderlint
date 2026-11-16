package azt010used

import "path/filepath"

// the name comes later, so any file in the folder may be the one read
func path(name string) string {
	return filepath.Join("testdata", name)
}
