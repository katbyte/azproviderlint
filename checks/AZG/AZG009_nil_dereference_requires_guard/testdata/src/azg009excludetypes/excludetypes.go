package azg009excludetypes

type widgetsClient struct{ Endpoint *string }

func (widgetsClient) Get() string { return "" }

type account struct{ SubscriptionId string }

type properties struct{ Count *int }

type model struct {
	Client     *widgetsClient
	Account    *account
	Properties *properties
}

func use(interface{}) {}

// Not flagged: `*Client` matches widgetsClient, for the method call and the field alike.
func excludedClient(m model) {
	use(m.Client.Get())
	use(m.Client.Endpoint)
}

// Not flagged: a fully qualified pattern matches too.
func excludedQualified(m model) {
	use(m.Account.SubscriptionId)
}

// Should be flagged: properties matches no pattern.
func notExcluded(m model) {
	use(m.Properties.Count) // want "selecting `Count` implicitly dereferences possibly-nil `m.Properties`"
}
