package lifetime

import (
	"chain"
	"testing"

	"gno.land/p/jaekwon/allowancesender/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	payee = testutils.TestAddress("payee")
)

func TestNewLifetimeSubscription_RejectsInvalidPayee(cur realm, t *testing.T) {
	uassert.PanicsWithMessage(t, cur, ErrInvalidPayee.Error(), func() {
		NewLifetimeSubscription(alice, 1000, address(""))
	})
}

func TestSubscribe_RejectsInvalidSubscriber(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)
	err := ls.Subscribe(&allowancesender.AllowanceSender{}, address(""))
	uassert.ErrorIs(t, err, ErrInvalidSubscriber)
}

func TestHasValidSubscription_NoSub(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)
	uassert.ErrorIs(t, ls.HasValidSubscription(alice), ErrNoSub)
}

func TestHasValidSubscription_Active(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)
	ls.subs.Set(alice.String(), true)
	uassert.NoError(t, ls.HasValidSubscription(alice))
}

func TestSubscribe_AlreadySubscribedReturnsErrBeforeSend(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)
	ls.subs.Set(alice.String(), true)

	err := ls.Subscribe(&allowancesender.AllowanceSender{}, alice)
	uassert.ErrorIs(t, err, ErrAlreadySub)
}

func TestSubscribe_SendPanicLeavesNoRecord(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)

	func() {
		defer func() { _ = recover() }()
		ls.Subscribe(&allowancesender.AllowanceSender{}, bob)
	}()

	uassert.ErrorIs(t, ls.HasValidSubscription(bob), ErrNoSub)
}

func TestUpdateAmount_OwnerSucceeds(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = ls.UpdateAmount(0, cur, 2000) }(cross(cur))
	uassert.NoError(t, err)
	uassert.Equal(t, int64(2000), ls.GetAmount())
}

func TestUpdateAmount_NonOwnerFails(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)

	testing.SetRealm(testing.NewUserRealm(bob))
	var err error
	func(cur realm) { err = ls.UpdateAmount(0, cur, 3000) }(cross(cur))
	uassert.ErrorIs(t, err, ErrNotAuthorized)
}

func TestGetPayee(cur realm, t *testing.T) {
	ls := NewLifetimeSubscription(alice, 1000, payee)
	uassert.Equal(t, payee.String(), ls.GetPayee().String())
}

var _ = chain.Coins{}
