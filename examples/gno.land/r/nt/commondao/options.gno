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
}

func (o *Options) SetAllowListing(allow bool) {
	o.AllowListing = allow
}

func (o *Options) SetAllowRender(allow bool) {
	o.AllowRender = allow
}

func (o *Options) SetAllowChildren(allow bool) {
	o.AllowChildren = allow
}

func (o *Options) SetAllowVoting(allow bool) {
	o.AllowVoting = allow
}

func (o *Options) SetAllowExecution(allow bool) {
	o.AllowExecution = allow
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
