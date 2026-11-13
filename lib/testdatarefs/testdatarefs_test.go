package testdatarefs

import (
	"path/filepath"
	"testing"
)

func TestInString(t *testing.T) {
	t.Parallel()

	type want struct {
		text    string
		path    string
		pattern bool
		rooted  bool
	}

	tests := []struct {
		name string
		in   string
		want []want
	}{
		{
			name: "a file in the folder's own testdata",
			in:   `contents = filebase64("testdata/cert.pfx")`,
			want: []want{{text: "testdata/cert.pfx", path: "svc/res/testdata/cert.pfx", rooted: true}},
		},
		{
			name: "a leading dot slash",
			in:   `zip_deploy_file = "./testdata/app.zip"`,
			want: []want{{text: "./testdata/app.zip", path: "svc/res/testdata/app.zip", rooted: true}},
		},
		{
			name: "two folders up",
			in:   `zip_deploy_file = "../../testdata/app.zip"`,
			want: []want{{text: "../../testdata/app.zip", path: "testdata/app.zip", rooted: true}},
		},
		{
			name: "terraform's path.module is the test's folder",
			in:   `file("${path.module}/testdata/a.xml")`,
			want: []want{{text: "testdata/a.xml", path: "svc/res/testdata/a.xml", rooted: true}},
		},
		{
			name: "two in one string",
			in:   `file("testdata/a.xml") file("testdata/sub/b.ps1")`,
			want: []want{
				{text: "testdata/a.xml", path: "svc/res/testdata/a.xml", rooted: true},
				{text: "testdata/sub/b.ps1", path: "svc/res/testdata/sub/b.ps1", rooted: true},
			},
		},
		{
			name: "the folder on its own, as handed to filepath.Join",
			in:   `testdata`,
			want: []want{{text: "testdata", path: "svc/res/testdata", rooted: true}},
		},
		{
			name: "the folder with a name added at run time",
			in:   `testdata/%s.json`,
			want: []want{{text: "testdata", path: "svc/res/testdata", rooted: true}},
		},
		{
			name: "the start of a name finished at run time",
			in:   `testdata/cert_%d.pem`,
			want: []want{{text: "testdata/cert_", path: "svc/res/testdata/cert_", pattern: true, rooted: true}},
		},
		{
			name: "a go:embed wildcard",
			in:   `testdata/*.json`,
			want: []want{{text: "testdata", path: "svc/res/testdata", rooted: true}},
		},
		{
			name: "the full stop that ends a sentence",
			in:   `could not read testdata/cert.pfx.`,
			want: []want{{text: "testdata/cert.pfx", path: "svc/res/testdata/cert.pfx", rooted: true}},
		},
		{
			name: "the end of a longer path is not rooted",
			in:   `https://example.com/repo/testdata/cert.pfx`,
			want: []want{{text: "testdata/cert.pfx", path: "svc/res/testdata/cert.pfx"}},
		},
		{
			name: "part of a longer word",
			in:   `mytestdata/cert.pfx and testdata_dir and my.testdata`,
		},
		{
			name: "no mention",
			in:   `resource "azurerm_resource_group" "test" {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := InString(filepath.FromSlash("svc/res"), tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("InString(%q) found %d mentions, want %d: %+v", tt.in, len(got), len(tt.want), got)
			}

			for i, w := range tt.want {
				g := got[i]
				if g.Text != w.text || g.Path != filepath.FromSlash(w.path) || g.Pattern != w.pattern || g.Rooted != w.rooted {
					t.Errorf("InString(%q)[%d] = {Text: %q, Path: %q, Pattern: %t, Rooted: %t}, want %+v", tt.in, i, g.Text, g.Path, g.Pattern, g.Rooted, w)
				}
			}
		})
	}
}

func TestCovers(t *testing.T) {
	t.Parallel()

	file := filepath.FromSlash("svc/testdata/certs/cert_1.pem")
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "the file itself", in: "testdata/certs/cert_1.pem", want: true},
		{name: "a folder holding it", in: "testdata/certs", want: true},
		{name: "the testdata folder", in: "testdata", want: true},
		{name: "a pattern it fits", in: "testdata/certs/cert_%d.pem", want: true},
		{name: "another file", in: "testdata/certs/cert_2.pem", want: false},
		{name: "a file whose name it starts with", in: "testdata/certs/cert_1", want: false},
		{name: "a pattern it does not fit", in: "testdata/certs/key_%d.pem", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			refs := InString("svc", tt.in)
			if len(refs) != 1 {
				t.Fatalf("InString(%q) found %d mentions, want 1", tt.in, len(refs))
			}

			if got := refs[0].Covers(file); got != tt.want {
				t.Errorf("%q covers %s = %t, want %t", tt.in, file, got, tt.want)
			}
		})
	}
}
