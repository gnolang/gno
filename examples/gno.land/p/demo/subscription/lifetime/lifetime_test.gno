package lifetime

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
	payee   = testutils.TestAddress("payee")
)

// mockSender implements subscription.Sender for unit tests.
type mockSender struct {
	sent     []sendRecord
	failNext bool
	failMsg  string
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

func (m *mockSender) totalUgnot() int64 {
	var total int64
	for _, r := range m.sent {
		total += r.amt.AmountOf("ugnot")
	}
	return total
}

func TestLifetimeSubscription(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ms := &mockSender{}

	err := ls.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected Subscribe to succeed for Alice")
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected 1000 ugnot pulled")

	err = ls.HasValidSubscription(alice)
	uassert.NoError(t, err, "Expected Alice to have access")
}

func TestLifetimeSubscriptionGift(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ms := &mockSender{}

	err := ls.Subscribe(ms, bob)
	uassert.NoError(t, err, "Expected gift Subscribe to succeed for Bob")

	err = ls.HasValidSubscription(bob)
	uassert.NoError(t, err, "Expected Bob to have access")

	err = ls.HasValidSubscription(charlie)
	uassert.Error(t, err, "Expected Charlie to fail access check")
}

func TestUpdateAmountAuthorization(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000, payee)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	testing.SetRealm(testing.NewUserRealm(bob))
	err = ls.UpdateAmount(3000)
	uassert.Error(t, err, "Expected Bob to fail when updating amount")
}

// TestSenderFailure replaces TestIncorrectPaymentAmount and
// TestGiftSubscriptionWithIncorrectAmount: the lib trusts the sender
// to refuse over-cap or under-funded sends. A failing Send panics out
// of Subscribe; the lib leaves no subscription record.
func TestSenderFailure(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ms := &mockSender{failNext: true, failMsg: "cap exceeded"}

	defer func() {
		r := recover()
		uassert.NotEqual(t, r, nil, "Expected Subscribe to panic when Send fails")
	}()
	ls.Subscribe(ms, alice)
	t.Fatal("unreachable; Subscribe should have panicked")
}

func TestSenderFailureLeavesNoRecord(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ms := &mockSender{failNext: true, failMsg: "cap exceeded"}

	func() {
		defer func() { recover() }()
		ls.Subscribe(ms, bob)
	}()

	err := ls.HasValidSubscription(bob)
	uassert.Error(t, err, "Expected Bob to not have access after failed Subscribe")
}

func TestMultipleSubscriptionAttempts(t *testing.T) {
	ls := NewLifetimeSubscription(1000, payee)
	ms := &mockSender{}

	err := ls.Subscribe(ms, alice)
	uassert.NoError(t, err, "Expected first subscription to succeed")
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected 1000 ugnot pulled")

	err = ls.Subscribe(ms, alice)
	uassert.Error(t, err, "Expected second subscription to fail (already-sub)")
	// No additional payment should have been pulled.
	uassert.Equal(t, ms.totalUgnot(), int64(1000), "Expected no additional ugnot on duplicate sub")
}

// TestUpdateAmountEffectiveness verifies that UpdateAmount actually
// changes what the lib pulls on subsequent Subscribe calls. The mock
// sender records each Send so we can verify the new amount lands.
func TestUpdateAmountEffectiveness(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	ls := NewLifetimeSubscription(1000, payee)

	err := ls.UpdateAmount(2000)
	uassert.NoError(t, err, "Expected Alice to succeed in updating amount")

	ms := &mockSender{}
	err = ls.Subscribe(ms, bob)
	uassert.NoError(t, err, "Expected Subscribe to succeed at updated amount")
	uassert.Equal(t, ms.totalUgnot(), int64(2000), "Expected 2000 ugnot pulled (post-update)")
}
