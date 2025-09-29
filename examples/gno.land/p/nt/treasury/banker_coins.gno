package treasury

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"

	"gno.land/p/aeddi/panictoerr"
)

var ErrNoStdBankerProvided = errors.New("no std banker provided")

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
func (cb *CoinsBanker) Send(p Payment) error {
	if runtime.CurrentRealm().Address() != cb.owner {
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

// NewCoinsBanker creates a new CoinsBanker with the current Realm's address
// as the owner.
func NewCoinsBanker(banker_ banker.Banker) (*CoinsBanker, error) {
	owner := runtime.CurrentRealm().Address()

	return NewCoinsBankerWithOwner(owner, banker_)
}

// NewCoinsBankerWithOwner creates a new CoinsBanker with the given address.
func NewCoinsBankerWithOwner(owner address, banker_ banker.Banker) (*CoinsBanker, error) {
	if owner == "" {
		return nil, ErrNoOwnerProvided
	}

	// NOTE: Should we add methods to std.Banker to check both its type and the
	// associated Realm for this kind of case?
	// For example:
	// if banker.Type() != std.BankerTypeRealmSend { panic("banker must be of type std.BankerTypeRealmSend") }
	// if banker.Realm().Address() != owner { panic("banker must be owned by the given owner address") }
	if banker_ == nil {
		return nil, ErrNoStdBankerProvided
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
