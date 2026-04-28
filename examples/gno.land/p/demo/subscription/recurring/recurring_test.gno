package recurring

import (
	"chain"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice    = testutils.TestAddress("alice")
	bob      = testutils.TestAddress("bob")
	charlie  = testutils.TestAddress("charlie")
	payee    = testutils.TestAddress("payee")
)

// mockSender is a test fixture implementing subscription.Sender.
// It records each Send call and can be capped to fail on the next
// invocation (simulating "cap exceeded" or other Send-side failures).
type mockSender struct {
	sent      []sendRecord
	failNext  bool
	failMsg   string
}

type sendRecord struct {
	to  address
	amt chain.Coins
}

func (m *mockSender) Send(to address, amt chain.Coins) {
	if m.failNext {
		panic(m.failMsg)
	}
	m.sent = append(m.sent, sendRecord{to, amt})
}

// totalUgnot returns the total ugnot moved by all Send calls so far.
func (m *mockSender) totalUgnot() int64 {
	var total int64
	for _, r := range m.sent {
		total += r.amt.AmountOf("ugnot")
	}
	return total
}

func TestRecurringSubscription(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)
	ms := &mockSender{}

	err := rs.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected Subscribe to succeed for Alice")
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected 1000 ugnot pulled")

	err = rs.HasValidSubscription(alice)
	uassert.NoError(t, err, "Expected Alice to have access")

	_, err = rs.GetExpiration(alice)
	uassert.NoError(t, err, "Expected to get expiration for Alice")
}

func TestRecurringSubscriptionGift(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)
	ms := &mockSender{}

	// Alice's mockSender pays for Bob's subscription.
	err := rs.Subscribe(ms, bob)
	uassert.NoError(t, err, "Expected gift Subscribe to succeed for Bob")

	err = rs.HasValidSubscription(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = rs.HasValidSubscription(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestRecurringSubscriptionExpiration(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	ms := &mockSender{}

	err := rs.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected Subscribe to succeed for Alice")

	err = rs.HasValidSubscription(alice)
	uassert.NoError(t, err, "Expected Alice to have access")

	expiration := time.Now().Add(-time.Hour * 2)
	rs.subs.Set(alice.String(), expiration)

	err = rs.HasValidSubscription(alice)
	uassert.Error(t, err, "Expected Alice's subscription to be expired")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)

	err := rs.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	testing.SetRealm(testing.NewUserRealm(bob))
	err = rs.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}

func TestGetAmount(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)

	uassert.Equal(t, rs.GetAmount(), int64(1000), "Expected initial amount 1000")

	err := rs.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	uassert.Equal(t, rs.GetAmount(), int64(2000), "Expected updated amount 2000")
}

// TestSenderFailure replaces the old TestIncorrectPaymentAmount: in the
// new model, "wrong amount" is enforced by the sender (cap-token), not
// the lib. A failing Send panics out of Subscribe; the lib leaves no
// subscription record.
func TestSenderFailure(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)
	ms := &mockSender{failNext: true, failMsg: "cap exceeded"}

	defer func() {
		r := recover()
		uassert.NotEqual(t, r, nil, "Expected Subscribe to panic when Send fails")
	}()
	rs.Subscribe(ms, alice)
	t.Fatal("unreachable; Subscribe should have panicked")
}

func TestMultiplePaymentsForSameUser(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour*24, 1000, payee)
	ms := &mockSender{}

	err := rs.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected first Subscribe to succeed for Alice")
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected 1000 ugnot pulled on first sub")

	err = rs.Subscribe(ms, alice)
	uassert.Error(t, err, "Expected second Subscribe to fail (already-sub)")
	// No additional payment should have been pulled.
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected no additional ugnot on duplicate sub")
}

func TestRecurringSubscriptionWithMultiplePayments(t *testing.T) {
	rs := NewRecurringSubscription(time.Hour, 1000, payee)
	ms := &mockSender{}

	err := rs.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected first Subscribe to succeed for Alice")

	err = rs.HasValidSubscription(alice)
	uassert.NoError(t, err, "Expected Alice to have access after first payment")

	// Force expiry, then re-subscribe.
	rs.subs.Set(alice.String(), time.Now().Add(-time.Hour*2))

	err = rs.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected second Subscribe to succeed for Alice (post-expiry)")
	uassert.Equal(t, ms.totalUgnot(), int64(2000), "Expected 2000 ugnot total across two periods")

	err = rs.HasValidSubscription(alice)
	uassert.NoError(t, err, "Expected Alice to have access after second payment")
}
