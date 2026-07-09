package treasury

import (
	"errors"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/bptree/v0/pager"
	"gno.land/p/nt/ufmt/v0"
)

var (
	ErrNoBankerProvided       = errors.New("no banker provided")
	ErrDuplicateBanker        = errors.New("duplicate banker")
	ErrBankerNotFound         = errors.New("banker not found")
	ErrSendPaymentFailed      = errors.New("failed to send payment")
	ErrNonCanonicalBankerImpl = errors.New("non-canonical Banker impl: only *CoinsBanker / *GRC20Banker accepted")
)

// New creates a new Treasury instance with the given bankers.
//
// pkgPath should be the package path of the owning realm. It is
// captured on the Treasury and used to build the "See full history"
// link in RenderBanker. Pass `cur.PkgPath()` from the owning realm's
// init (or another live-cur context). Passing "" disables the link.
//
// The path is stored as plain data — no IsCurrent guard inside this
// /p/ constructor. The caller is the trust boundary; it must derive
// the value from a real `cur realm` rather than accept it from
// untrusted input. The captured path is consumed only for render
// output (Class-2 designation-forgery shape, accepted as
// display-only — see docs/resources/gno-security.md).
//
// Each banker must be one of treasury's canonical concrete impls
// (*CoinsBanker, *GRC20Banker) per IsCanonicalBanker. Foreign-realm impls
// (including embedded wrappers around canonical types) are rejected: a
// malicious Send impl would receive a capability token via its rlm
// parameter when treasury.Send dispatches into it.
//
// The allowlist validates type only, not captured state. Treasury operators
// must construct their own bankers; never accept a pre-built *Banker value
// from an external realm.
func New(bankers []Banker, pkgPath string) (*Treasury, error) {
	if len(bankers) == 0 {
		return nil, ErrNoBankerProvided
	}

	// Canonical-impl allowlist: reject foreign types (embedding-based
	// bypasses fail here because type assertions are nominal).
	for _, b := range bankers {
		if !IsCanonicalBanker(b) {
			return nil, ErrNonCanonicalBankerImpl
		}
	}

	// Create a new Treasury instance.
	treasury := &Treasury{bankers: bptree.NewBPTree32(), realmPath: pkgPath}

	// Register the bankers.
	for _, banker := range bankers {
		if treasury.bankers.Has(banker.ID()) {
			return nil, ufmt.Errorf("%v: %s", ErrDuplicateBanker, banker.ID())
		}

		treasury.bankers.Set(
			banker.ID(),
			&bankerRecord{banker: banker},
		)
	}

	// Register the Render routes.
	treasury.initRenderRouter()

	return treasury, nil
}

// Send sends a payment using the corresponding banker. rlm is threaded
// to the banker's Send for IsCurrent + owner validation.
func (t *Treasury) Send(_ int, rlm realm, p Payment) error {
	// Get the banker record corresponding to this Payment.
	br := t.bankers.Get(p.BankerID())
	if br == nil {
		return ufmt.Errorf("%v: %s", ErrBankerNotFound, p.BankerID())
	}
	record := br.(*bankerRecord)

	// Send the payment using the corresponding banker.
	if err := record.banker.Send(0, rlm, p); err != nil {
		return ufmt.Errorf("%v: %s", ErrSendPaymentFailed, err)
	}

	// Add the payment to the history of the banker.
	record.history.Append(p)

	return nil
}

// History returns the payment history sent by the banker with the given ID.
// Payments are paginated, with the most recent payments first.
func (t *Treasury) History(
	bankerID string,
	pageNumber int,
	pageSize int,
) ([]Payment, error) {
	// Get the banker record corresponding to this ID.
	br := t.bankers.Get(bankerID)
	if br == nil {
		return nil, ufmt.Errorf("%v: %s", ErrBankerNotFound, bankerID)
	}
	history := br.(*bankerRecord).history

	// Get the page of payments from the history.
	p := pager.NewPager(history.Tree(), pageSize, true)
	page := p.GetPage(pageNumber)

	// Convert the items in the page to a slice of Payments.
	payments := make([]Payment, len(page.Items))
	for i := range page.Items {
		payments[i] = page.Items[i].Value.(Payment)
	}

	return payments, nil
}

// Balances returns the balances of the banker with the given ID.
func (t *Treasury) Balances(bankerID string) ([]Balance, error) {
	// Get the banker record corresponding to this ID.
	br := t.bankers.Get(bankerID)
	if br == nil {
		return nil, ufmt.Errorf("%v: %s", ErrBankerNotFound, bankerID)
	}

	// Get the balances from the banker.
	return br.(*bankerRecord).banker.Balances(), nil
}

// Address returns the address of the banker with the given ID.
func (t *Treasury) Address(bankerID string) (string, error) {
	// Get the banker record corresponding to this ID.
	br := t.bankers.Get(bankerID)
	if br == nil {
		return "", ufmt.Errorf("%v: %s", ErrBankerNotFound, bankerID)
	}

	// Get the address from the banker.
	return br.(*bankerRecord).banker.Address(), nil
}

// HasBanker checks if a banker with the given ID is registered.
func (t *Treasury) HasBanker(bankerID string) bool {
	return t.bankers.Has(bankerID)
}

// ListBankerIDs returns a list of all registered banker IDs.
func (t *Treasury) ListBankerIDs() []string {
	var bankerIDs []string

	t.bankers.Iterate("", "", func(bankerID string, _ any) bool {
		bankerIDs = append(bankerIDs, bankerID)
		return false
	})

	return bankerIDs
}
