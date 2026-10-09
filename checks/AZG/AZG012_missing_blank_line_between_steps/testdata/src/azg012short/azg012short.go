package azg012short

// With statements=2 and lines=4 even a short body is checked
func short(n int) int {
	if n < 0 {
		n = 0
	}
	n++ // want `a blank line should separate this from the if above, which ends a step`
	return n
}
