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
}

func (o *Options) SetPrivate(private bool) {
	o.Private = private
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
