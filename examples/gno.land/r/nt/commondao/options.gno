package commondao

// TODO: List default DAO options somewhere to make devs aware of them
var defaultOptions = Options{
	AllowRender:    true,
	AllowChildren:  true,
	AllowVoting:    true,
	AllowExecution: true,
}

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

	// AllowMembersUpdate allows creating a proposal to update DAO members.
	AllowMembersUpdate bool

	// AllowTextProposals allows creating general text proposals.
	AllowTextProposals bool

	// AllowSubDAOProposals allows creating SubDAOs though proposals.
	AllowSubDAOProposals bool

	// AllowDissolutionProposals allows dissolving DAOs though proposals.
	AllowDissolutionProposals bool
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

// SetAllowMembersUpdate toggles support for DAO members update proposals.
// This type of proposal allows adding or removing DAO members.
func (o *Options) SetAllowMembersUpdate(allow bool) {
	o.AllowMembersUpdate = allow
}

// SetAllowTextProposals toggles support generic text proposals.
func (o *Options) SetAllowTextProposals(allow bool) {
	o.AllowTextProposals = allow
}

// SetAllowSubDAOPorposals toggles support for SubDAO creation proposals.
// SubDAOs can only be created when DAO children support is enabled.
func (o *Options) SetAllowSubDAOPorposals(allow bool) {
	o.AllowSubDAOProposals = allow
}

// SetAllowDissolutionProposals toggles support for DAO dissolution proposals.
func (o *Options) SetAllowDissolutionProposals(allow bool) {
	o.AllowDissolutionProposals = allow
}
