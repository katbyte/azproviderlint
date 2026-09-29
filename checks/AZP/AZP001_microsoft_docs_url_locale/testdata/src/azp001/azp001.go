package azp001

// Should be flagged: a locale segment in a comment.
// See https://learn.microsoft.com/en-us/azure/aks/egress-outboundtype for details. // want "Microsoft docs URL has a locale segment `en-us/` - drop it so readers get their own language"
const outboundDoc = "https://docs.microsoft.com/en-gb/azure/batch/batch-api-basics#pool" // want "Microsoft docs URL has a locale segment `en-gb/` - drop it so readers get their own language"

// Should be flagged twice: two URLs in one literal, both fixed.
var both = `https://learn.microsoft.com/en-us/rest/api/keyvault/ and https://msdn.microsoft.com/en-us/library/x` // want "locale segment `en-us/`" "locale segment `en-us/`"

// Should NOT be flagged: already locale-neutral.
// https://learn.microsoft.com/azure/azure-arc/kubernetes/conceptual-agent-overview
const neutral = "https://docs.microsoft.com/azure/cosmos-db/how-to-configure-firewall"

// Should NOT be flagged: other Microsoft hosts and non-locale segments.
const portal = "https://portal.azure.com/en-us/#home"
const notLocale = "https://learn.microsoft.com/enus/azure/"

/* Should be flagged inside a block comment too: https://docs.microsoft.com/en-us/azure/aks/ */ // want "locale segment `en-us/`"

func use() string { return outboundDoc + both + neutral + portal + notLocale }
