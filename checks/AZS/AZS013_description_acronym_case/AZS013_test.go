package AZS013

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS013(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs013")

	for _, text := range []string{
		"The ID, the IDs, the URL's host and the IPv4 or IPv6 address.",
		"Uses FTPS, LDAPS and an IDS.",
		"The `id` of the `vm`.",
		"See https://example.com/api/id or example.com/sql for the id_token.",
		"A key in ssh-rsa format, sent as api-version.",
		`Run "az vm list --query id" first.`,
		"The clientId, the resource.id and the key=id pair.",
		"A valid idea, a zip file and an idle timeout.",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	for text, want := range map[string][]string{
		"The id.":                           {`"id" should be "ID"`},
		"The Id (or ids) of the vm's nic":   {`"Id" should be "ID"`, `"ids" should be "IDs"`, `"vm" should be "VM"`, `"nic" should be "NIC"`},
		"An Ip-based rule over HTTP/Https.": {`"Ip" should be "IP"`, `"Https" should be "HTTPS"`},
		"A *Dns* zone with a CName record.": {`"Dns" should be "DNS"`, `"CName" should be "CNAME"`},
		"Uses IPV4, Oauth, VCPUs and GRPC.": {`"IPV4" should be "IPv4"`, `"Oauth" should be "OAuth"`, `"VCPUs" should be "vCPUs"`, `"GRPC" should be "gRPC"`},
	} {
		findings := check(text)
		got := make([]string, 0, len(findings))
		for _, finding := range findings {
			got = append(got, finding.Message)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	// the options add an acronym and drop one; they are package state, so this cannot run beside anything else
	for option, acronym := range map[string]string{"acronym": "IoT", "ignore": "id"} {
		if err := Analyzer.Flags.Set(option, acronym); err != nil {
			t.Fatal(err)
		}
	}

	findings := check("The id of the iot hub.")
	got := make([]string, 0, len(findings))
	for _, finding := range findings {
		got = append(got, finding.Message)
	}

	if want := []string{`"iot" should be "IoT"`}; !slices.Equal(got, want) {
		t.Errorf("with IoT added and ID dropped: got %q, want %q", got, want)
	}

	delete(acronyms, "iot")
	acronyms["id"] = "ID"
}
