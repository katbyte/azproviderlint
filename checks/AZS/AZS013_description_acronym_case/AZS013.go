// Package AZS013 defines an analyzer that reports acronyms written in the wrong case in resource and property descriptions.
package AZS013

import (
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that an acronym in a description is written the way it is everywhere else: "ID", not "id" or "Id". Code in backticks and web
// addresses are left alone, since there the case is part of the value.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS013",
	Doc:      "check that acronyms in resource and property descriptions are written in their usual case",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS013_description_acronym_case/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// acronyms holds how each acronym is written, by its lowercase form. None of them is also an everyday word.
var acronyms = map[string]string{}

func init() {
	for acronym := range strings.FieldsSeq("ID IP URL URI API DNS HTTP HTTPS JSON XML YAML SQL SSL TLS SSH TCP UDP VM CPU GPU SKU FQDN CIDR GUID UUID UTC HTML FTP FTPS SFTP SMTP NFS SMB VPN IOPS TTL ACL HSM MSI RBAC SAML JWT SDK CLI NSG CDN CORS CSV RSA AES LDAP LDAPS DHCP BGP ICMP SSD HDD VHD MFA SSO NAT NIC SAS LUN WAF KQL PEM PFX PKI CNAME AKS PHP TDE UPN SPN ASCII IPv4 IPv6 OAuth vCPU NVMe DDoS OData gRPC VNet") {
		acronyms[strings.ToLower(acronym)] = acronym
	}

	Analyzer.Flags.Init("AZS013", flag.ContinueOnError)
	Analyzer.Flags.Func("acronym", "another acronym to check, written the way it should be; repeat for several", func(acronym string) error {
		acronyms[strings.ToLower(acronym)] = acronym
		return nil
	})
	Analyzer.Flags.Func("ignore", "an acronym to leave alone; repeat for several", func(acronym string) error {
		delete(acronyms, strings.ToLower(acronym))
		return nil
	})
}

var quoted = regexp.MustCompile(`"[^"]*"`)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for _, found := range description.All(pass, insp) {
		for _, finding := range check(found.Text) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns each acronym in the description that is not written in its usual case.
func check(text string) []description.Finding {
	// what is inside double quotes is a value or a command, the same as code in backticks
	masked := []byte(description.Mask(text))
	for _, at := range quoted.FindAllIndex(masked, -1) {
		for i := at[0]; i < at[1]; i++ {
			if masked[i] != ' ' {
				masked[i] = description.Blank
			}
		}
	}

	var findings []description.Finding
	for _, word := range description.Words(string(masked)) {
		// the text of a link is prose; where it points is not
		word.Text, _, _ = strings.Cut(word.Text, "]")
		word.End = word.Start + len(word.Text)
		word = word.Trim(".,;:?!)}\"'*_")

		// a word with a dot, an underscore or an equals sign in it is a name or a value that is missing its backticks, not prose
		if strings.ContainsAny(word.Text, "._=@\\"+string(description.Blank)) {
			continue
		}

		// "ip-based" and "http/https" are checked a part at a time
		for start := 0; start <= len(word.Text); {
			end := start + strings.IndexAny(word.Text[start:]+"-", "-/")
			part := strings.TrimSuffix(word.Text[start:end], "'s")
			lower := strings.ToLower(part)
			want, ok := acronyms[lower]

			// a lowercase part of a hyphenated word is as likely a name as prose: "ssh-rsa", "api-version"
			if part == lower && strings.Contains(word.Text, "-") {
				ok = false
			}
			// "IDs" and "ids" are more than one ID, where "FTPS" is an acronym of its own
			if plural := strings.TrimSuffix(lower, "s"); !ok && strings.HasSuffix(part, "s") && !strings.HasSuffix(plural, "s") && (part != lower || !strings.Contains(word.Text, "-")) {
				want, ok = acronyms[plural]
				want += "s"
			}

			if ok && part != want {
				findings = append(findings, description.Finding{Start: word.Start + start, End: word.Start + start + len(part), Message: fmt.Sprintf("%q should be %q", part, want), Fix: "Write " + want, New: want})
			}
			start = end + 1
		}
	}

	return findings
}
