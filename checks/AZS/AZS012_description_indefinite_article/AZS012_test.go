package AZS012

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS012(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs012")

	for want, words := range map[string][]string{
		// said with a vowel sound, whatever the first letter
		"an": {
			"integer", "Automation", "API", "ID", "IP", "IPv4", "IDs", "optional", "unused", "unintended", "uninitialized", "unimportant", "unidentified", "update", "upper", "Ubuntu", "umbrella", "only", "onboarding", "iSCSI", "eDTU", "OAuth", "ETag",
			"hour", "hourly", "hour-long", "honest", "honorary", "heir",
			"SSL", "SSH", "HTTP", "HTTPS", "HSM", "FQDN", "FTP", "NSG", "NFS", "NVMe", "XML", "MSI", "MFA", "RBAC", "RSA", "LDAP", "SLA", "SSO", "LTR", "SLB", "HDInsight", "HBase", "HL7", "MX", "SR-IOV",
			"http", "xml", "ssh", "nginx", "Xbox", "Xcode", "xterm", "XPath", "yttrium", "Yvonne", "onerous", "unable", "unassigned",
			"S3", "x509", "X-Forwarded-For", "e-mail", "mTLS", "sFTP", "F5", "R-squared",
			"8", "8-bit", "80", "800", "8080", "11", "18", "11th", "18-month", "11,000", "18000",
		},

		// said with a consonant sound, whatever the first letter
		"a": {
			"Certificate", "resource", "hotel", "history", "hybrid", "Hyper-V", "year", "YAML", "JSON",
			"user", "User", "user's", "username", "user-assigned", "unique", "unit", "union", "uniform", "universal", "Unix", "unicode", "unidirectional", "unimodal", "usage", "usable", "useful", "utility", "URL", "URI", "UUID", "UI", "UDP", "UTC", "UTF-8", "url", "uri", "utf-8", "European", "euro",
			"one", "one-time", "One", "once",
			"NAT", "NIC", "SAS", "SAML", "SCSI", "LUN", "LAN", "MAC", "RAM", "REST", "RAID", "RHEL", "FHIR", "FIPS", "EULA", "NAS",
			"VM", "VNet", "vCore", "DNS", "TLS", "TCP", "CIDR", "GUID", "JWT", "PEM", "dns", "vm", "Kubernetes", "PostgreSQL", "MySQL", "macOS",
			".NET", "B-tree", "t-shirt", "Y-axis", "U-turn", "ui", "ux", "UX", "Xen", "Xamarin", "xylophone", "unanimous", "Utah", "sap", "oneself", "YML", "year",
			"1", "10", "100", "110", "180", "2-node", "0", "12", "17", "1,000",
		},

		// said both ways, or not something it can place
		"": {
			"SQL", "SQLServer", "sql", "SKU", "SKUs", "SNAT", "SHA", "sha256", "herb", "historic", "historical",
			"NULL", "NOT", "SELECT", "MAX", "msg", "src", "svc", "mgmt", "std", "yml",
			"1100", "1800", "$web", "<name>", "#1", "",
		},
	} {
		for _, word := range words {
			if got := articleFor(word); got != want {
				t.Errorf("%q takes %q, want %q", word, got, want)
			}
		}
	}

	for _, text := range []string{
		"An integer, a user, an SSL certificate and a URL.",
		"Either a `map` or an `identity` block.",
		"A type A or AAAA record.",
		"Rated from a to z, or grade a or b.",
		"The vitamin a is not an article.",
		"Ends with a",
		"In version 1.2 A or AAAA records are allowed, and in v1.3 A IN records.",
		"It has a ha effect and a mi effect.",
		"A [link](https://example.com) and an [other](https://example.com/a).",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	for text, want := range map[string][]string{
		"a integer.":                  {`"a integer" should be "an integer"`},
		"Takes (a hour).":             {`"a hour" should be "an hour"`},
		"It is done. A API is made.":  {`"A API" should be "An API"`},
		"An user and a *API* key.":    {`"An user" should be "A user"`, `"a API" should be "an API"`},
		"An [user](https://x.com/y).": {`"An user" should be "A user"`},
		"Uses an 1 GB disk.":          {`"an 1" should be "a 1"`},
		"Uses a 8 GB disk.":           {`"a 8" should be "an 8"`},
		"Has an ui and an Xen host.":  {`"an ui" should be "a ui"`, `"an Xen" should be "a Xen"`},
		"Runs on a Xbox? A API call.": {`"a Xbox" should be "an Xbox"`, `"A API" should be "An API"`},
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

	// the options teach it a word; they are package state, so this cannot run beside anything else
	if got := check("An NIMBUS and a SPROCKET."); len(got) > 0 {
		t.Errorf("an acronym it cannot place should be left alone, got %q", got[0].Message)
	}

	for option, word := range map[string]string{"a": "NIMBUS", "an": "Sprocket"} {
		if err := Analyzer.Flags.Set(option, word); err != nil {
			t.Fatal(err)
		}
	}

	findings := check("An NIMBUS and a SPROCKET.")
	got := make([]string, 0, len(findings))
	for _, finding := range findings {
		got = append(got, finding.Message)
	}

	if want := []string{`"An NIMBUS" should be "A NIMBUS"`, `"a SPROCKET" should be "an SPROCKET"`}; !slices.Equal(got, want) {
		t.Errorf("with both words taught: got %q, want %q", got, want)
	}

	delete(said, "nimbus")
	delete(said, "sprocket")
}
