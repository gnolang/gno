package commondao

var defaultOptions = Options{
	AllowRender:    true,
	AllowChildren:  true,
	AllowVoting:    true,
	AllowExecution: true,
}

// Option configures DAO options within CommonDAO realm.
type Option func(*Options)

// Options configures DAO behavior within CommonDAO realm.
// These options are only valid for the CommonDAO realm; dApps that
// own a DAO can implement these features by themselves.
type Options struct {
	// AllowListing adds the DAO to the public list of realm DAOs.
	AllowListing bool

	// AllowRender enables support for rendering the DAO within the CommonDAO realm.
	AllowRender bool

	// AllowChildren enables hierarchical DAOs by allowing the creation of SubDAOs.
	AllowChildren bool

	// AllowVoting enables proposal voting support.
	AllowVoting bool

	// AllowExecution allows executing early-passed proposals before their
	// voting deadline. Execution of passed proposals after the deadline is
	// always permitted.
	AllowExecution bool

	// AllowCouncilUpdate allows proposals that add and/or remove council members.
	AllowCouncilUpdate bool

	// AllowTextProposals allows text-only, non-executable proposals.
	AllowTextProposals bool

	// AllowSubDAOProposals allows creating SubDAOs through proposals.
	AllowSubDAOProposals bool

	// AllowTreasuryProposals allows treasury spend proposals, and the
	// ancestor clawback/freeze controls over descendant treasuries.
	AllowTreasuryProposals bool

	// AllowDissolutionProposals allows dissolving DAOs through proposals.
	AllowDissolutionProposals bool

	// MaxActiveProposals caps the number of simultaneously active
	// proposals. Values below one keep the package default.
	MaxActiveProposals int
}

// AllowListing configures DAO listing within CommonDAO realm.
func AllowListing(allow bool) Option {
	return func(o *Options) {
		o.AllowListing = allow
	}
}

// AllowRender configures DAO, proposals and votes rendering within CommonDAO.
func AllowRender(allow bool) Option {
	return func(o *Options) {
		o.AllowRender = allow
	}
}

// AllowChildren configures DAO support for SubDAOs.
func AllowChildren(allow bool) Option {
	return func(o *Options) {
		o.AllowChildren = allow
	}
}

// AllowVoting configures DAO support for voting through CommonDAO realm.
func AllowVoting(allow bool) Option {
	return func(o *Options) {
		o.AllowVoting = allow
	}
}

// AllowExecution configures DAO support for executing early-passed proposals before their voting deadline.
func AllowExecution(allow bool) Option {
	return func(o *Options) {
		o.AllowExecution = allow
	}
}

// AllowCouncilUpdate configures support for council update proposals.
func AllowCouncilUpdate(allow bool) Option {
	return func(o *Options) {
		o.AllowCouncilUpdate = allow
	}
}

// AllowTextProposals configures support for generic text proposals.
func AllowTextProposals(allow bool) Option {
	return func(o *Options) {
		o.AllowTextProposals = allow
	}
}

// AllowSubDAOProposals configures support for SubDAO creation proposals.
func AllowSubDAOProposals(allow bool) Option {
	return func(o *Options) {
		o.AllowSubDAOProposals = allow
	}
}

// AllowTreasuryProposals configures support for treasury proposals:
// spends from the DAO's own treasury, and the ancestor clawback/freeze
// controls over descendant treasuries.
func AllowTreasuryProposals(allow bool) Option {
	return func(o *Options) {
		o.AllowTreasuryProposals = allow
	}
}

// AllowDissolutionProposals configures support for DAO dissolution proposals.
func AllowDissolutionProposals(allow bool) Option {
	return func(o *Options) {
		o.AllowDissolutionProposals = allow
	}
}

// MaxActiveProposals configures the cap for simultaneously active proposals.
// Values below one keep the default cap.
func MaxActiveProposals(max int) Option {
	return func(o *Options) {
		o.MaxActiveProposals = max
	}
}
