package commondao

var defaultOptions = Options{
	AllowRender:    true,
	AllowChildren:  true,
	AllowExecution: true,
}

// Options configures hosting concerns of a DAO within the CommonDAO
// realm: listing, rendering, sub-DAO support and early execution.
// Governance powers themselves (creating and deciding proposals) follow
// the Common DAO Spec defaults and cannot be switched off.
type Options struct {
	// AllowListing adds the DAO to the public list of realm DAOs.
	AllowListing bool

	// AllowRender enables support for rendering the DAO within the CommonDAO realm.
	AllowRender bool

	// AllowChildren enables hierarchical DAOs by allowing the creation
	// of SubDAOs through proposals.
	AllowChildren bool

	// AllowExecution allows executing early-passed proposals before their
	// voting deadline. Execution of passed proposals after the deadline is
	// always permitted.
	AllowExecution bool
}
