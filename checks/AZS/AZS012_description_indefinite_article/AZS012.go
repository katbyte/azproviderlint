// Package AZS012 defines an analyzer that reports "a" where "an" belongs, and the other way round, in resource and property descriptions.
package AZS012

import (
	"flag"
	"fmt"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks each "a" and "an" in a description against how the next word is said: "an integer", "a user", "an SSL certificate", "a URL".
// It only reports where it is sure, so a word it cannot place is left alone.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS012",
	Doc:      `check that resource and property descriptions use the right one of "a" and "an"`,
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS012_description_indefinite_article/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// said holds the words whose article cannot be worked out from their letters, in lowercase, with the article they take. An empty article means
// both are in use, so neither is reported.
var said = map[string]string{}

func init() {
	// acronyms said as a word that starts with a consonant sound, and ones starting with U, which is said "you"
	for word := range strings.FieldsSeq("nat nic nas numa nist sas saml san sim siem scim soap sip scsi sata suse sox scep sarif soar sase scom lan lun luis lamp linq lifo mac mime mars mapi ram raid rest radius rom rhel hana hipaa fifo fhir fips fat eula uuid ui uefi uri url utc udp upn usb utf unc ux usa uat uae uami umi ua ulid") {
		said[word] = "a"
	}

	// acronyms said letter by letter whose first letter's name starts with a vowel sound
	for word := range strings.FieldsSeq("msi mfa msix msal msa mca mtu mta mi ha hdi hmac ldap ldaps lro lte nva nvme nginx npu rbac rpo rto rsa ru ra rdma sso sse sla soa ssis ssas sni sre slo sap sbom sa se fci fpga http https html xml xsd xss xslt ssh ssl sftp smtp snmp smb sms sdk nfs nsg ntp ntfs ntlm npm fqdn ftp ftps rdp rpc rfc hsm hdd hpc mqtt") {
		said[word] = "an"
	}

	// said both ways: "a sequel server" and "an S-Q-L server"
	for word := range strings.FieldsSeq("sql sku snat spa sha sid spi soc sles mau faq noc") {
		said[word] = ""
	}

	Analyzer.Flags.Init("AZS012", flag.ContinueOnError)
	Analyzer.Flags.Func("a", `a word that takes "a", such as an acronym said as a word; repeat for several`, func(word string) error {
		said[strings.ToLower(word)] = "a"
		return nil
	})
	Analyzer.Flags.Func("an", `a word that takes "an", such as an acronym said letter by letter; repeat for several`, func(word string) error {
		said[strings.ToLower(word)] = "an"
		return nil
	})
}

// words that an article never comes before, so an "a" in front of one is a letter or a name
var notNouns = strings.Fields("or and is are was be to of in on at for with as by from nor but if then vs")

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

// check returns each "a" or "an" in the description that is the wrong one for the word after it.
func check(text string) []description.Finding {
	var findings []description.Finding
	words := description.Words(description.Mask(text))
	for i, word := range words {
		article := word.Trim("")
		if i+1 == len(words) || !slices.Contains([]string{"a", "an", "A", "An"}, article.Text) {
			continue
		}

		// a capital A in the middle of a sentence is a letter or a name: "a type A record"
		if before := strings.TrimRight(words[max(i-1, 0)].Text, ")]}\"'*_"); article.Text == "A" && i > 0 && (before == "" || !strings.ContainsRune(".?!:", rune(before[len(before)-1]))) {
			continue
		}

		// the word after it, less what closes it and where a link on it points
		next := words[i+1].Trim("")
		next.Text, _, _ = strings.Cut(next.Text, "]")
		next.Text = strings.TrimRight(next.Text, ".,;:?!)}\"'*_"+string(description.Blank))
		if next.Text == "" || slices.Contains(notNouns, strings.ToLower(next.Text)) {
			continue
		}

		want := articleFor(next.Text)
		if want == "" || strings.EqualFold(want, article.Text) {
			continue
		}

		if article.Text[0] == 'A' {
			want = "A" + want[1:]
		}
		findings = append(findings, description.Finding{Start: article.Start, End: article.End, Message: fmt.Sprintf("%q should be %q", article.Text+" "+next.Text, want+" "+next.Text), Fix: "Write " + want, New: want})
	}

	return findings
}

// articleFor returns the article the word takes, "a" or "an", going by how it is said and not how it is spelt. It returns nothing for a word
// that is said both ways, or that it cannot place.
func articleFor(word string) string {
	// ".NET" is said "dot net"
	if len(word) > 1 && word[0] == '.' {
		return "a"
	}

	// a number goes by how it is read out: "an 8", "an 11", "an 18", "an 80", but "a 1" and "a 110"
	if digits := len(word) - len(strings.TrimLeft(word, "0123456789")); digits > 0 {
		switch number := word[:digits]; {
		case number[0] == '8':
			return "an"
		case (strings.HasPrefix(number, "11") || strings.HasPrefix(number, "18")) && digits%3 == 2:
			return "an"
		case (strings.HasPrefix(number, "11") || strings.HasPrefix(number, "18")) && digits == 4:
			return ""
		default:
			return "a"
		}
	}

	// only the letters the word opens with matter: "user" in "user-assigned", "S" in "S3", "x" in "x509"
	head := word[:len(word)-len(strings.TrimLeft(word, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"))]
	if head == "" {
		return ""
	}

	lower := strings.ToLower(head)
	letter := "a"
	if strings.ContainsRune("aefhilmnorsx", rune(lower[0])) {
		letter = "an"
	}

	// a letter on its own is said by its name, and so is one in front of a capital: "an x509", "an mTLS", "a vCore"
	capitals := len(head) - len(strings.TrimLeft(head, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	if len(head) == 1 || (capitals == 0 && head[1] < 'a') {
		return letter
	}

	// the capitals a word opens with are an acronym, less the last one when it starts a word of its own: "HD" in "HDInsight"
	acronym := lower
	if capitals >= 2 {
		acronym = lower[:capitals]
		if len(head)-capitals >= 2 {
			acronym = acronym[:capitals-1]
		}
	}
	if len(acronym) == 1 {
		return letter
	}

	for _, known := range []string{acronym, strings.TrimSuffix(acronym, "s")} {
		if want, ok := said[known]; ok && (len(known) >= 3 || capitals >= 2 || want == "a") {
			return want
		}
	}

	// read as a word, it goes by the sound it starts with
	spoken := "a"
	switch {
	case strings.ContainsRune("aeiou", rune(acronym[0])):
		spoken = "an"

		// "you" and "won" sounds: "a user", "a unique", "a European", "a one-time"; but "an unintended", "an unimportant"
		if hasPrefix(acronym, "eu", "ew", "uni", "use", "usa", "usu", "usi", "uti", "uto", "ute", "ura", "ure", "uri", "uro", "uru", "ubiq", "uga", "ukr", "uku", "ufo") && (!hasPrefix(acronym, "unin", "unim", "uniden", "unill", "uniss") || hasPrefix(acronym, "unimod")) {
			spoken = "a"
		}
		if slices.Contains([]string{"one", "once", "ones", "oneself", "oneness", "onetime"}, acronym) {
			spoken = "a"
		}

	// an x before a consonant is said as the letter: "an Xbox", "an xterm", but "a Xen host"
	case acronym[0] == 'x' && !strings.ContainsRune("aeiouy", rune(acronym[1])):
		spoken = "an"

	case hasPrefix(acronym, "hour", "honest", "honor", "honour", "heir"): //nolint:misspell // both spellings of honour are words it has to know
		spoken = "an"

	case hasPrefix(acronym, "herb", "historic"):
		return ""
	}

	// an acronym is said as a word or letter by letter; where the two differ, one with no vowel after its first letter can only be spelt out
	switch {
	case capitals < 2 && (len(acronym) < 3 || strings.ContainsAny(acronym[1:], "aeiouy")):
		return spoken
	case spoken == letter:
		return spoken
	case capitals >= 2 && !strings.ContainsAny(acronym[1:], "aeiouy"):
		return letter
	default:
		return ""
	}
}

// hasPrefix reports whether the word starts with any of the prefixes.
func hasPrefix(word string, prefixes ...string) bool {
	return slices.ContainsFunc(prefixes, func(prefix string) bool { return strings.HasPrefix(word, prefix) })
}
