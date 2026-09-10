// Command azproviderlint runs the analyzers in checks/ over Go packages, as a standalone
// multichecker; the same analyzers ship as a golangci-lint module plugin (see plugin/).
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/katbyte/azproviderlint/checks"
	"github.com/katbyte/azproviderlint/version"
)

func main() {
	// multichecker owns the CLI, so handle `azproviderlint version` before it takes over
	if len(os.Args) > 1 && os.Args[1] == "version" {
		v := version.Version
		if v == "" {
			v = "dev"
		}
		if version.GitCommit != "" {
			v += " (" + version.GitCommit + ")"
		}
		fmt.Println("azproviderlint " + v)
		return
	}

	os.Args = expandCategoryFlags(os.Args)
	multichecker.Main(checks.All...)
}

// expandCategoryFlags rewrites category flags like -AZG into the individual rule flags of
// that category (-AZG001 -AZG003 ...) before multichecker parses the command line.
func expandCategoryFlags(args []string) []string {
	byCategory := map[string][]string{}
	for _, a := range checks.All {
		if i := strings.IndexFunc(a.Name, unicode.IsDigit); i > 0 {
			cat := a.Name[:i]
			byCategory[cat] = append(byCategory[cat], a.Name)
		}
	}

	out := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			cat := strings.ToUpper(strings.TrimLeft(arg, "-"))
			if names, ok := byCategory[cat]; ok {
				for _, name := range names {
					out = append(out, "-"+name)
				}
				continue
			}
		}
		out = append(out, arg)
	}
	return out
}
