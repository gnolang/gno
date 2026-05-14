package lifetime

import (
	"chain"
	"testing"

	"gno.land/p/jaekwon/allowancesender/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	payee = testutils.TestAddress("payee")
)

func TestNewLifetimeSubscription_RejectsInvalidPayee(t *testing.T) {
	uassert.PanicsWithMessage(t, ErrInvalidPayee.Error(), func() {
		NewLifetimeSubscription(1000, address(""))
	})
}

func TestSubscribe_RejectsInvalidSubscriber(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	err := ls.Subscribe(&allowancesender.AllowanceSender{}, address(""))
	uassert.ErrorIs(t, err, ErrInvalidSubscriber)
}

func TestHasValidSubscription_NoSub(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	uassert.ErrorIs(t, ls.HasValidSubscription(alice), ErrNoSub)
}

func TestHasValidSubscription_Active(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ls.subs.Set(alice.String(), true)
	uassert.NoError(t, ls.HasValidSubscription(alice))
}

func TestSubscribe_AlreadySubscribedReturnsErrBeforeSend(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ls.subs.Set(alice.String(), true)

	err := ls.Subscribe(&allowancesender.AllowanceSender{}, alice)
	uassert.ErrorIs(t, err, ErrAlreadySub)
}

func TestSubscribe_SendPanicLeavesNoRecord(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)

	func() {
		defer func() { _ = recover() }()
		ls.Subscribe(&allowancesender.AllowanceSender{}, bob)
	}()

	uassert.ErrorIs(t, ls.HasValidSubscription(bob), ErrNoSub)
}

func TestUpdateAmount_OwnerSucceeds(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000, payee)

	uassert.NoError(t, ls.UpdateAmount(2000))
	uassert.Equal(t, int64(2000), ls.GetAmount())
}

func TestUpdateAmount_NonOwnerFails(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000, payee)

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.ErrorIs(t, ls.UpdateAmount(3000), ErrNotAuthorized)
}

func TestGetPayee(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	uassert.Equal(t, payee.String(), ls.GetPayee().String())
}

var _ = chain.Coins{}
