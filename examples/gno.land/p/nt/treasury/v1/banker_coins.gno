package treasury

import (
	"chain"
	"chain/banker"
	"errors"
	"math"
	"math/overflow"
	"strconv"

	"gno.land/p/aeddi/panictoerr/v0"
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

	// Convert banker.Coins to []Balance. A held coin amount is never
	// negative, so widening it to uint256 is exact.
	balances := make([]Balance, len(coins))
	for i := range coins {
		balances[i] = Balance{
			Denom:  coins[i].Denom,
			Amount: uint256(coins[i].Amount),
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

// NewCoinsPayment creates a new coinsPayment from uint256 amounts.
//
// Chain coins are int64, so each amount is range-checked down here: an amount
// of zero or above math.MaxInt64 is refused (it panics) when the payment is
// built, rather than by the bank keeper at execution, after a vote. Amounts
// of the same denom are summed with a checked add first, so the check applies
// to what will actually be paid.
//
// The payment holds its own sorted, joined chain.Coins, so the caller cannot
// change the amount after the payment exists, and String() renders the set
// that will actually be paid. A malformed denom is still only caught by the
// keeper.
func NewCoinsPayment(amounts []Balance, toAddress address) Payment {
	// Join amounts by denom, in uint256 so that the sum cannot wrap.
	totals := map[string]uint256{}
	denoms := []string{}
	for _, a := range amounts {
		prev, seen := totals[a.Denom]
		if !seen {
			denoms = append(denoms, a.Denom)
		}
		sum, ok := overflow.Addu256(prev, a.Amount)
		if !ok {
			panic("coins payment amount overflows uint256 for denom " + a.Denom)
		}
		totals[a.Denom] = sum
	}

	cz := chain.Coins{}
	for _, denom := range denoms {
		amount, err := coinAmount(totals[denom])
		if err != nil {
			panic(err.Error() + ", got " + strconv.FormatUint256(totals[denom], 10) + denom)
		}
		cz = cz.Add(chain.Coins{chain.NewCoin(denom, amount)})
	}

	return coinsPayment{
		coins:     cz,
		toAddress: toAddress,
	}
}

// ErrCoinAmountOutOfRange is returned when a uint256 amount cannot be paid in
// chain coins, which are positive int64 values.
var ErrCoinAmountOutOfRange = errors.New("coins payment amounts must be in [1, MaxInt64]")

// coinAmount range-checks a uint256 amount down to a positive int64 chain
// coin amount.
func coinAmount(amount uint256) (int64, error) {
	if amount == 0 || amount > math.MaxInt64 {
		return 0, ErrCoinAmountOutOfRange
	}
	return int64(amount), nil
}
