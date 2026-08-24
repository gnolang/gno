package treasury

import (
	"chain"
	"chain/banker"
	"errors"

	"gno.land/p/aeddi/panictoerr"
)

var ErrNonCanonicalBanker = errors.New("inner banker is not the canonical chain/banker.Banker")

// CoinsBanker is a Banker that sends banker.Coins.
type CoinsBanker struct {
	owner  address       // The address of this coins banker owner.
	banker banker.Banker // The underlying std banker, must be a BankerTypeRealmSend.
}

var _ Banker = (*CoinsBanker)(nil)

// ID implements Banker.
func (CoinsBanker) ID() string {
	return "Coins"
}

// Send implements Banker.
//
// rlm must be the caller's own captured cur (i.e. the cur of the
// immediate crossing-function caller). Sending with rlm = cur.Previous()
// or any other realm value is rejected: rlm.IsCurrent() asserts pointer
// identity against the topmost crossing frame. Combined with the
// rlm.Address() == cb.owner check, this restricts Send to the owning
// realm acting in its own frame.
func (cb *CoinsBanker) Send(_ int, rlm realm, p Payment) error {
	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}
	if rlm.Address() != cb.owner {
		return ErrCurrentRealmIsNotOwner
	}
	// Check if payment is of type coinsPayment.
	payment, ok := p.(coinsPayment)
	if !ok {
		return ErrInvalidPaymentType
	}

	// Send the coins.
	return panictoerr.PanicToError(func() {
		cb.banker.SendCoins(cb.owner, payment.toAddress, payment.coins)
	})
}

// Balances implements Banker.
func (cb *CoinsBanker) Balances() []Balance {
	// Get the coins from the banker.
	coins := cb.banker.GetCoins(cb.owner)

	// Convert banker.Coins to []Balance.
	balances := make([]Balance, len(coins))
	for i := range coins {
		balances[i] = Balance{
			Denom:  coins[i].Denom,
			Amount: coins[i].Amount,
		}
	}

	return balances
}

// Address implements Banker.
func (cb *CoinsBanker) Address() string {
	return cb.owner.String()
}

// NewCoinsBankerWithOwner creates a new CoinsBanker with the given address.
//
// banker_ must be the canonical Banker produced by banker.NewBanker;
// hand-rolled Banker implementations (no-op fakes, decorators) are
// rejected via banker.IsCanonical. Without this check, a callee
// receiving a *CoinsBanker constructed from a fake banker would not
// be able to tell that Send is a no-op (no real coins move). The
// pkgAddr-vs-owner mismatch is still surfaced lazily by the inner
// banker's own SendCoins check.
func NewCoinsBankerWithOwner(owner address, banker_ banker.Banker) (*CoinsBanker, error) {
	if owner == "" {
		return nil, ErrNoOwnerProvided
	}

	if !banker.IsCanonical(banker_) {
		return nil, ErrNonCanonicalBanker
	}

	return &CoinsBanker{
		owner:  owner,
		banker: banker_,
	}, nil
}

// coinsPayment represents a payment that is issued by a CoinsBanker.
type coinsPayment struct {
	coins     chain.Coins // The coins being sent.
	toAddress address     // The recipient of the payment.
}

var _ Payment = (*coinsPayment)(nil)

// BankerID implements Payment.
func (coinsPayment) BankerID() string {
	return CoinsBanker{}.ID()
}

// String implements Payment.
func (cp coinsPayment) String() string {
	return cp.coins.String() + " to " + cp.toAddress.String()
}

// NewCoinsPayment creates a new coinsPayment.
func NewCoinsPayment(coins chain.Coins, toAddress address) Payment {
	return coinsPayment{
		coins:     coins,
		toAddress: toAddress,
	}
}
