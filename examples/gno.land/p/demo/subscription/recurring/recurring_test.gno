package recurring

import (
	"chain"
	"testing"
	"time"

	"gno.land/p/jaekwon/allowancesender/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	payee = testutils.TestAddress("payee")
)

// stubSender is a Send-only stand-in for *AllowanceSender used by these
// tests. Subscribe accepts a concrete *allowancesender.AllowanceSender,
// and outside the allowancesender package we cannot construct a working
// one (the IsCanonical gate rejects synthetic bankers). Tests therefore
// inject behavior through stubSender by wrapping it as a fake
// AllowanceSender via the allowancesender package's same-package helpers.
//
// Since allowancesender's struct fields are unexported, we can't build
// one here either. Instead each test uses the empty-struct shortcut
// `&allowancesender.AllowanceSender{}` which compiles cross-package
// (no field-name reference) but produces an unusable instance whose
// Send always panics. That panic propagates out of Subscribe, the tx
// reverts, and tests assert the panic shape. For success-path tests,
// we sidestep the AllowanceSender plumbing by constructing the
// RecurringSubscription, then directly setting subs entries via the
// in-package field — testing Subscribe's bookkeeping without the cap.
//
// This is sufficient for unit-level coverage of the subscription state
// machine (already-sub gate, expiration, ownership, idempotency).
// Full payment-path coverage requires integration txtars, which are
// future work.

func TestNewRecurringSubscription_RejectsInvalidPayee(t *testing.T) {
	uassert.PanicsWithMessage(t, ErrInvalidPayee.Error(), func() {
		NewRecurringSubscription(time.Hour, 1000, address(""))
	})
}

func TestSubscribe_RejectsInvalidSubscriber(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	err := rs.Subscribe(&allowancesender.AllowanceSender{}, address(""))
	uassert.ErrorIs(t, err, ErrInvalidSubscriber)
}

func TestHasValidSubscription_NoSub(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	err := rs.HasValidSubscription(alice)
	uassert.ErrorIs(t, err, ErrNoSub)
}

func TestHasValidSubscription_Active(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	rs.subs.Set(alice.String(), time.Now().Add(time.Hour))
	uassert.NoError(t, rs.HasValidSubscription(alice))
}

func TestHasValidSubscription_Expired(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	rs.subs.Set(alice.String(), time.Now().Add(-time.Hour))
	uassert.ErrorIs(t, rs.HasValidSubscription(alice), ErrSubExpired)
}

// TestSubscribe_AlreadyActiveReturnsErrBeforeSend exercises the
// duplicate-detection short-circuit: when the subscriber is already
// active, Subscribe returns ErrAlreadySub WITHOUT calling al.Send. We
// pass an unusable empty AllowanceSender; if Send were called it would
// panic, but the early return prevents that.
func TestSubscribe_AlreadyActiveReturnsErrBeforeSend(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	rs.subs.Set(alice.String(), time.Now().Add(time.Hour))

	err := rs.Subscribe(&allowancesender.AllowanceSender{}, alice)
	uassert.ErrorIs(t, err, ErrAlreadySub)
}

// TestSubscribe_SendPanicLeavesNoRecord exercises the send-side failure:
// junk AllowanceSender → Send panics → tx revert. Since unit tests don't
// have tx semantics, the panic propagates. Recover and assert no
// subscription was recorded.
func TestSubscribe_SendPanicLeavesNoRecord(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)

	func() {
		defer func() { _ = recover() }()
		rs.Subscribe(&allowancesender.AllowanceSender{}, bob)
	}()

	uassert.ErrorIs(t, rs.HasValidSubscription(bob), ErrNoSub)
}

func TestUpdateAmount_OwnerSucceeds(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour, 1000, payee)

	uassert.NoError(t, rs.UpdateAmount(2000))
	uassert.Equal(t, int64(2000), rs.GetAmount())
}

func TestUpdateAmount_NonOwnerFails(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour, 1000, payee)

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.ErrorIs(t, rs.UpdateAmount(3000), ErrNotAuthorized)
}

func TestGetAmount(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	uassert.Equal(t, int64(1000), rs.GetAmount())
}

func TestGetPayee(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	uassert.Equal(t, payee.String(), rs.GetPayee().String())
}

func TestGetExpiration_NoSub(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	_, err := rs.GetExpiration(alice)
	uassert.ErrorIs(t, err, ErrNoSub)
}

func TestGetExpiration_Active(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	expected := time.Now().Add(time.Hour)
	rs.subs.Set(alice.String(), expected)

	got, err := rs.GetExpiration(alice)
	uassert.NoError(t, err)
	uassert.Equal(t, expected.String(), got.String())
}

// silence unused-import warning for chain (used in production code paths).
var _ = chain.Coins{}
