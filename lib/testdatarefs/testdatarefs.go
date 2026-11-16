// Package testdatarefs finds the testdata paths that Go source names, for the checks that compare what tests read against what is on disk.
package testdatarefs

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Ref is one mention of a testdata path.
type Ref struct {
	Pos     token.Pos // where the mention starts, when it came from a file
	Lit     token.Pos // where the string or comment holding it starts, which for a string over many lines is the nearest place a comment can go
	Offset  int       // where the mention starts in the string it was found in
	Text    string    // the path as written, such as ../../testdata/cert.pfx
	Name    string    // the part after testdata/, empty when the folder itself is meant
	Path    string    // what it names on disk, taken from the folder of the file it was found in
	Pattern bool      // Path is only the fixed start of a name built at run time, as in testdata/cert_%d.pem
	Rooted  bool      // starts where a path starts, so it is safe to read it as relative to that folder
}

// Covers reports whether the mention is of path, of a folder holding it, or of a pattern it fits.
func (r Ref) Covers(path string) bool {
	if r.Pattern {
		return strings.HasPrefix(path, r.Path)
	}

	return path == r.Path || strings.HasPrefix(path, r.Path+string(filepath.Separator))
}

// any run of ./ and ../ in front counts as part of the path, and so does what follows up to the first character a file name would not hold
var mention = regexp.MustCompile(`((?:\.\.?/)*)testdata((?:/[\w.\-/]*)?)`)

// terraform's name for the folder of the config being run, which for a test is the folder of the test
const pathModule = "${path.module}/"

// InString returns the testdata paths s mentions, read as relative to dir.
func InString(dir, s string) []Ref {
	var refs []Ref
	for _, m := range mention.FindAllStringSubmatchIndex(s, -1) {
		start, end := m[0], m[1]
		prefix, tail := s[m[2]:m[3]], s[m[4]:m[5]]

		// testdata as part of a longer word is some other name
		if start > 0 && isNameByte(s[start-1]) {
			continue
		}
		if tail == "" && end < len(s) && isNameByte(s[end]) {
			continue
		}

		// after a slash it is the end of a longer path whose start is unknown, a URL say
		rooted := start == 0 || s[start-1] != '/' || strings.HasSuffix(s[:start], pathModule)

		// a name that runs into a verb or a wildcard is finished at run time: a folder still means everything in it, a partial name everything starting with it
		pattern := end < len(s) && strings.IndexByte("%*{$[", s[end]) >= 0 && tail != "" && !strings.HasSuffix(tail, "/")
		if !pattern {
			tail = strings.TrimRight(tail, "./")
		}

		refs = append(refs, Ref{
			Offset:  start,
			Text:    prefix + "testdata" + tail,
			Name:    strings.TrimPrefix(tail, "/"),
			Path:    filepath.Join(dir, filepath.FromSlash(prefix+"testdata"+tail)),
			Pattern: pattern,
			Rooted:  rooted,
		})
	}

	return refs
}

// InFile returns the testdata paths that the string literals and go:embed lines of file mention, read as relative to the folder the file is in.
func InFile(fset *token.FileSet, file *ast.File) []Ref {
	dir := filepath.Dir(fset.Position(file.Pos()).Filename)

	var refs []Ref
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		for _, ref := range InString(dir, s) {
			// a raw string holds its text as written, so the mention can be pointed at; an escaped one may not line up
			ref.Pos, ref.Lit = lit.Pos(), lit.Pos()
			if strings.HasPrefix(lit.Value, "`") {
				ref.Pos += token.Pos(1 + ref.Offset)
			}

			refs = append(refs, ref)
		}

		return true
	})

	for _, group := range file.Comments {
		for _, comment := range group.List {
			patterns, ok := strings.CutPrefix(comment.Text, "//go:embed ")
			if !ok {
				continue
			}

			for _, ref := range InString(dir, patterns) {
				ref.Pos, ref.Lit = comment.Pos(), comment.Pos()
				refs = append(refs, ref)
			}
		}
	}

	return refs
}

func isNameByte(b byte) bool {
	return b == '_' || b == '-' || b == '.' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
