package azc001

type Options struct {
	SubscriptionId          string
	ResourceManagerEndpoint string
}

type FoosClient struct{}

func NewFoosClient(subscriptionId string) FoosClient { return FoosClient{} }

func NewFoosClientWithBaseURI(endpoint string, subscriptionId string) FoosClient {
	return FoosClient{}
}

// Should be flagged: client created without an explicit base URI
func badClient(o Options) FoosClient {
	return NewFoosClient(o.SubscriptionId) // want `Azure SDK clients should be created with NewFoosClientWithBaseURI and the resource manager endpoint explicitly specified`
}

// Should NOT be flagged: client created with the resource manager endpoint
func goodClient(o Options) FoosClient {
	return NewFoosClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
}

// Should NOT be flagged: not an options struct subscription id
func goodOtherCall(subscriptionId string) FoosClient {
	return NewFoosClient(subscriptionId)
}

var sdk = struct {
	NewFoosClient func(subscriptionId string) FoosClient
}{}

var ctors = []func(subscriptionId string) FoosClient{NewFoosClient}

// Should be flagged: the constructor is reached through a package-style selector
func badSelectorClient(o Options) FoosClient {
	return sdk.NewFoosClient(o.SubscriptionId) // want `Azure SDK clients should be created with NewFoosClientWithBaseURI and the resource manager endpoint explicitly specified`
}

// Should NOT be flagged: the callee does not name a client
func goodNotAClient(o Options) string {
	return describe(o.SubscriptionId)
}

func describe(subscriptionId string) string { return subscriptionId }

// Should NOT be flagged: the subscription id is read from something other than o
func goodOtherReceiver(opts Options) FoosClient {
	return NewFoosClient(opts.SubscriptionId)
}

// Should NOT be flagged: the callee is an indexed value, not a named constructor
func goodIndexedCallee(o Options) FoosClient {
	return ctors[0](o.SubscriptionId)
}
