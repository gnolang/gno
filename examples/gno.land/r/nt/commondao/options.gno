package commondao

// TODO: List default DAO options somewhere to make devs aware of them
var defaultOptions = Options{
	AllowRender:    true,
	AllowChildren:  true,
	AllowVoting:    true,
	AllowExecution: true,
}

// Options configures the behavior of DAOs created within the CommonDAO realm.
// These options are only valid for the CommonDAO realm. Other dApps that
// own a *commondao.DAO object can implement these features by themselves.
type Options struct {
	// AllowListing allows the DAO to be listed on the CommonDAO realm home page.
	AllowListing bool

	// AllowRender enables support for rendering the DAO within the CommonDAO realm.
	AllowRender bool

	// AllowChildren enables hierarchical DAOs by allowing the creation of SubDAOs.
	AllowChildren bool

	// AllowVoting enables proposal voting.
	AllowVoting bool

	// AllowExecution enables proposal execution.
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
