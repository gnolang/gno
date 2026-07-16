package treasury

import (
	"errors"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/bptree/v0/list"
	"gno.land/p/nt/mux/v0"
)

// Treasury is the main structure that holds all bankers and their payment
// history. It also provides a router for rendering the treasury pages.
type Treasury struct {
	bankers   *bptree.BPTree // string -> *bankerRecord
	router    *mux.Router
	realmPath string // owning realm's PkgPath, captured at New() (used for render links)
}

// bankerRecord holds a Banker and its payment history.
type bankerRecord struct {
	banker  Banker
	history list.List // List of Payment.
}

// Banker is an interface that allows for banking operations.
//
// SECURITY: Send takes (int, realm, Payment), so handing a Banker value to
// untrusted code yields a capability token to whatever Send impl that code
// dispatches into. The set of canonical impls is closed (*CoinsBanker,
// *GRC20Banker); any public function that accepts a Banker as a parameter
// from external callers MUST verify it via IsCanonicalBanker and reject
// otherwise. treasury.New enforces this for its own intake; future
// Banker-accepting APIs must do the same. An unexported-marker "seal" does
// NOT defend against this — see
// p/test/seal/filetests/z_seal_iface_embedding_filetest.gno.
//
// Note that IsCanonicalBanker validates dynamic TYPE only, not captured
// STATE: a canonical *CoinsBanker constructed via NewCoinsBankerWithOwner
// with a hostile owner argument passes the allowlist but its read methods
// (Balances, Address) report data tied to that hostile address. Treasury
// operators must construct their own bankers and NEVER accept pre-built
// *Banker values from external realms.
type Banker interface {
	ID() string                     // Get the ID of the banker.
	Send(int, realm, Payment) error // Send a payment to a recipient.
	Balances() []Balance            // Get the balances of the banker.
	Address() string                // Get the address of the banker to receive payments.
}

// IsCanonicalBanker reports whether b is one of treasury's canonical
// concrete Banker impls. Use this at any public entry point in /p/ or /r/
// that accepts a Banker from an external caller before invoking its methods.
//
// Foreign types — including embedding-based wrappers like
// `type Evil struct { *CoinsBanker }` — are rejected because type assertions
// are nominal: *Evil is not *CoinsBanker, regardless of method promotion.
//
// To add a new canonical type: extend the switch below AND add a regression
// test (under filetests/ in this package) that an embedded-impl bypass is
// rejected.
//
// Mirrors the precedent of chain/banker.IsCanonical and
// p/jaekwon/allowancesender's canonical-impl check.
func IsCanonicalBanker(b Banker) bool {
	switch b.(type) {
	case *CoinsBanker, *GRC20Banker:
		return true
	default:
		return false
	}
}

// Payment is an interface that allows getting details about a payment.
type Payment interface {
	BankerID() string // Get the ID of the banker that can process this payment.
	String() string   // Get a string representation of the payment.
}

// Balance represents the balance of an asset held by a Banker.
type Balance struct {
	Denom  string // The denomination of the asset
	Amount int64  // The amount of the asset
}

// Common Banker errors.
var (
	ErrCurrentRealmIsNotOwner = errors.New("current realm is not the owner of the banker")
	ErrNoOwnerProvided        = errors.New("no owner provided")
	ErrInvalidPaymentType     = errors.New("invalid payment type")
	ErrSpoofedRealm           = errors.New("rlm does not match the current crossing frame")
)
