package commondao

var defaultOptions = Options{
	AllowChildren:  true,
	AllowVoting:    true,
	AllowExecution: true,
}

// Options configures DAO behavior within commondao realm.
// These options are only valid for the commondao ream, dApps that
// own a DAO can implement these features by themselves.
type Options struct {
	// Private makes DAO private so it's not rendered by the realm.
	Private bool

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

	// AllowDismissalProposals allows dismissing DAOs though proposals.
	AllowDismissalProposals bool
}

// SetPrivate toggles a DAO between private or public.
// Private DAO's are not rendered by the realm.
func (o *Options) SetPrivate(private bool) {
	o.Private = private
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

// SetAllowDismissalProposals toggles support for DAO dismissal proposals.
func (o *Options) SetAllowDismissalProposals(allow bool) {
	o.AllowDismissalProposals = allow
}
