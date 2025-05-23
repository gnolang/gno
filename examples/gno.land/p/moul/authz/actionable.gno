package authz

// TODO: add a sync/async example with Do(action, callback) -- consider another
// package.

/*
type Doer interface {
	Do(action string, fn func() error) error
}

func ActionableProposalSystem(proposers, executors Membership, inboxFn ExecutableInboxFn) ActionableProposalSystem {
	return ActionableProposalSystem{proposers, executors, inboxFn}

type ExecutableFn func() error

type ExecutableInboxFn func(metadata any, fn ExecutableFn) error

type ActionableProposalSystem struct {
	proposers, executors Membership
}

func (aps *ActionableProposalSystem) DoByCurrent(metadata any, fn func()) error {
	if !aps.proposers.Has(std.CurrentRealm().Address()) {
		return errors.New("unauthorized")
	}

	wrappedFn := func() error {
		if !aps.executors.Has(std.CurrentRealm().Address()) {
			return errors.New("unauthorized")
		}
		return fn()
	}

	return aps.inboxFn(metadata, wrappedFn)
}

/*
type AuthorityDoer struct {
	auth Authority
}

func (ad AuthorityDoer) CurrentDo(action string, fn func() error) error {
	caller := std.CurrentRealm().Address()
	if ad.auth.Can(caller, action) {
		return fn()
	}
	return errors.New("unauthorized")
}

func ExampleDoer() {
	auth, _ := NewWithCurrentAdmin()
	doer := AuthorityDoer{auth}
	doer.CurrentDo("propose", func() error {
		println("hello")
		return nil
	})
}
*/
