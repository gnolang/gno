package authz

/*
package TBD

import (
	"errors"
	"std"
)

// TODO: add a sync/async example with Do(action, callback) -- consider another
// package.

type Doer interface {
	Do(action string, fn func() error) error
}

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
