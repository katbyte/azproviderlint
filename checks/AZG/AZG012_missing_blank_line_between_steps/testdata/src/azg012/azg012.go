package azg012

import "strings"

type item struct {
	name string
	ok   bool
}

func lookup(name string) (item, bool) {
	return item{name: name}, name != ""
}

func report(s string) {}

// Should be flagged: a long body whose steps run together. The function body itself is three
// statements, so the final return needs no blank line.
func long(names []string) int {
	count := 0
	for _, name := range names {
		if name == "" {
			continue
		}
		it, ok := lookup(name) // want `a blank line should separate this from the if above, which ends a step`
		if !ok || !it.ok {
			continue
		} // a comment on the next line belongs to the next step, so the blank goes before it
		// want `a blank line should separate this from the if above, which ends a step`
		rest := strings.TrimPrefix(it.name, "x")
		if rest == "" {
			report(rest)
			continue
		}
		report(rest) // want `a blank line should separate this from the if above, which ends a step`
		count++
	}
	return count
}

// A run of one-line guards stays together, but the statement after the last one starts a
// step. A literal assignment does not end a step, and a blank line already there is enough.
func guards(names []string) []string {
	var out []string
	for _, name := range names {
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "_") {
			continue
		}
		if len(name) > 64 {
			continue
		}
		it := item{ // want `a blank line should separate this from the if above, which ends a step`
			name: name,
		}
		if it.name == "skip" {
			continue
		}

		out = append(out, it.name)
		switch it.name {
		case "a":
			if it.ok {
				report("a")
			}
		case "b":
			report("b")
		}
		count := len(out) // want `a blank line should separate this from the switch above, which ends a step`
		report(string(rune(count)))
	}

	return out
}

// Should NOT be flagged: a short body is one group
func short(allow string) map[string]bool {
	allowed := map[string]bool{}
	for form := range strings.SplitSeq(allow, ",") {
		form = strings.TrimSpace(form)
		if form == "" {
			continue
		}
		if form != "x" {
			return nil
		}
		allowed[form] = true
	}
	return allowed
}

// Should be flagged: a closure body is checked like any other
func closure(names []string) func() int {
	return func() int {
		n := 0
		for _, name := range names {
			if name != "" {
				n++
			}
		}
		if n > 10 { // want `a blank line should separate this from the for above, which ends a step`
			n = 10
		}
		report("n") // want `a blank line should separate this from the if above, which ends a step`
		report("m")
		report("o")
		return n
	}
}
