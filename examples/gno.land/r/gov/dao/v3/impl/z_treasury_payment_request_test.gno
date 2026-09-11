package impl

import (
	"testing"

	trs_pkg "gno.land/p/nt/treasury/v0"
	"gno.land/p/nt/urequire/v0"
)

// Named z_* for the same reason as z_upgrade_request_test.gno: govdao_test.gno
// asserts hard-coded proposal ids, and this test creates no proposal.

// foreignPayment satisfies trs_pkg.Payment without being one of treasury's own
// impls. BankerID names a banker the treasury really has registered, and
// String says whatever this realm likes.
type foreignPayment struct{}

func (foreignPayment) BankerID() string { return "Coins" }

func (foreignPayment) String() string { return "42ugnot to g1nobody" }

// The immutability NewCoinsPayment provides is a property of the concrete
// type, not of the Payment interface: a foreign impl renders a payment line
// members vote on and then fails CoinsBanker.Send's type assertion at
// execution, spending a whole vote cycle on a proposal that can never
// execute. Rejected at construction so the proposer sees it, not the voters.
func TestTreasuryPaymentRequestRejectsForeignPayment(cur realm, t *testing.T) {
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))

	// PanicsWithMessage, not AbortsWithMessage: same-realm call, so the guard
	// raises a plain panic rather than a cross-realm abort.
	urequire.PanicsWithMessage(t, cur,
		"payment must be built by treasury.NewCoinsPayment or treasury.NewGRC20Payment",
		func() {
			NewTreasuryPaymentRequest(cur, foreignPayment{}, "a reason")
		})
}

// The check must not reject the real thing. GRC20 is used here because the
// Coins banker is the one govdao_test.gno's shared state leaves funded; both
// are canonical, and only the canonicality gate is under test.
func TestTreasuryPaymentRequestAcceptsCanonicalPayment(cur realm, t *testing.T) {
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))

	urequire.NotPanics(t, cur, func() {
		NewTreasuryPaymentRequest(cur, trs_pkg.NewGRC20Payment("TOK", 1, m1), "a reason")
	})
}
