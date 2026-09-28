package azg010testfile

// the package's own test file declares tags at package level, which a pass over the package
// without its tests does not see
import helperTags "example.com/tags"

func use() {
	helperTags.Expand()
}
