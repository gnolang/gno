package commondao

// TODO: List default DAO options somewhere to make devs aware of them
var defaultOptions = Options{
	AllowRender:    true,
	AllowChildren:  true,
	AllowVoting:    true,
	AllowExecution: true,
}

// Option configures DAO options within CommonDAO realm.
type Option func(*Options)

// Options configures DAO behavior within CommonDAO realm.
// These options are only valid for the CommonDAO ream, dApps that
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

	// AllowExecution enables proposal execution support.
	AllowExecution bool

	// AllowCouncilUpdate allows creating a proposal to update DAO members.
	AllowCouncilUpdate bool

	// AllowTextProposals allows text-only, non-executable proposals.
	AllowTextProposals bool

	// AllowSubDAOProposals allows creating SubDAOs though proposals.
	AllowSubDAOProposals bool

	// AllowDissolutionProposals allows dissolving DAOs though proposals.
	AllowDissolutionProposals bool

	// MaxActiveProposals caps the number of simultaneously active
	// proposals. Values below one keep the package default.
	MaxActiveProposals int
}

// SetAllowListing toggles DAO public listing support.
func (o *Options) SetAllowListing(allow bool) {
	o.AllowListing = allow
}

// SetAllowRender toggles DAO render support within the CommonDAO realm.
func (o *Options) SetAllowRender(allow bool) {
	o.AllowRender = allow
}

// SetAllowChildren toggles DAO support for SubDAOs.
// Enabling child DAOs is necessary for hierarchical tree based DAOs.
func (o *Options) SetAllowChildren(allow bool) {
	o.AllowChildren = allow
}

// SetAllowVoting toggles proposal voting support though the CommonDAO realm.
func (o *Options) SetAllowVoting(allow bool) {
	o.AllowVoting = allow
}

// SetAllowExecution toggles proposal execution support though the CommonDAO realm.
func (o *Options) SetAllowExecution(allow bool) {
	o.AllowExecution = allow
}

// SetAllowCouncilUpdate toggles support for council update proposals.
// This type of proposal allows adding or removing council members.
func (o *Options) SetAllowCouncilUpdate(allow bool) {
	o.AllowCouncilUpdate = allow
}

// SetAllowTextProposals toggles support generic text proposals.
func (o *Options) SetAllowTextProposals(allow bool) {
	o.AllowTextProposals = allow
}

// SetAllowSubDAOProposals toggles support for SubDAO creation proposals.
// SubDAOs can only be created when DAO children support is enabled.
func (o *Options) SetAllowSubDAOProposals(allow bool) {
	o.AllowSubDAOProposals = allow
}

// SetAllowDissolutionProposals toggles support for DAO dissolution proposals.
func (o *Options) SetAllowDissolutionProposals(allow bool) {
	o.AllowDissolutionProposals = allow
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

// AllowExecution configures DAO support for executing proposals through CommonDAO realm.
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

// AllowTextProposals configures support generic text proposals.
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
