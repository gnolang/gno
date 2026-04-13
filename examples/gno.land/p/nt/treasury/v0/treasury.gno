package treasury

import (
	"errors"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/ufmt/v0"
)

var (
	ErrNoBankerProvided  = errors.New("no banker provided")
	ErrDuplicateBanker   = errors.New("duplicate banker")
	ErrBankerNotFound    = errors.New("banker not found")
	ErrSendPaymentFailed = errors.New("failed to send payment")
)

// New creates a new Treasury instance with the given bankers.
func New(bankers []Banker) (*Treasury, error) {
	if len(bankers) == 0 {
		return nil, ErrNoBankerProvided
	}

	// Create a new Treasury instance.
	treasury := &Treasury{bankers: avl.NewTree()}

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

// Send sends a payment using the corresponding banker.
func (t *Treasury) Send(p Payment) error {
	// Get the banker record corresponding to this Payment.
	br, ok := t.bankers.Get(p.BankerID())
	if !ok {
		return ufmt.Errorf("%v: %s", ErrBankerNotFound, p.BankerID())
	}
	record := br.(*bankerRecord)

	// Send the payment using the corresponding banker.
	if err := record.banker.Send(p); err != nil {
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
	br, ok := t.bankers.Get(bankerID)
	if !ok {
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
	br, ok := t.bankers.Get(bankerID)
	if !ok {
		return nil, ufmt.Errorf("%v: %s", ErrBankerNotFound, bankerID)
	}

	// Get the balances from the banker.
	return br.(*bankerRecord).banker.Balances(), nil
}

// Address returns the address of the banker with the given ID.
func (t *Treasury) Address(bankerID string) (string, error) {
	// Get the banker record corresponding to this ID.
	br, ok := t.bankers.Get(bankerID)
	if !ok {
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
