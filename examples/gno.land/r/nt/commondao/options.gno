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
