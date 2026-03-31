package treasury

import (
	"chain/runtime"
	"errors"
	"strconv"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ufmt/v0"
)

var (
	ErrNoListerProvided   = errors.New("no lister provided")
	ErrGRC20TokenNotFound = errors.New("GRC20 token not found")
)

// GRC20Banker is a Banker that sends GRC20 tokens listed using a getter
// set during initialization.
type GRC20Banker struct {
	owner  address         // The address of this GRC20 banker owner.
	lister TokenListerFunc // Allows to list tokens from methods that require it.
}

// TokenListerFunc is a function type that returns a map of GRC20 tokens.
type TokenListerFunc func() map[string]*grc20.Token

var _ Banker = (*GRC20Banker)(nil)

// ID implements Banker.
func (GRC20Banker) ID() string {
	return "GRC20"
}

// Send implements Banker.
func (gb *GRC20Banker) Send(p Payment) error {
	if runtime.CurrentRealm().Address() != gb.owner {
		return ErrCurrentRealmIsNotOwner
	}

	payment, ok := p.(grc20Payment)
	if !ok {
		return ErrInvalidPaymentType
	}

	// Get the GRC20 tokens using the lister.
	tokens := gb.lister()

	// Look for the token corresponding to the payment tokenKey.
	token, ok := tokens[payment.tokenKey]
	if !ok {
		return ufmt.Errorf("%v: %s", ErrGRC20TokenNotFound, payment.tokenKey)
	}

	// Send the token.
	return token.RealmTeller().Transfer(payment.toAddress, payment.amount)
}

// Balances implements Banker.
func (gb *GRC20Banker) Balances() []Balance {
	// Get the GRC20 tokens from the lister.
	tokens := gb.lister()

	// Convert GRC20 tokens to []Balance.
	var balances []Balance
	for key, token := range tokens {
		balances = append(balances, Balance{
			Denom:  key,
			Amount: token.BalanceOf(gb.owner),
		})
	}
	return balances
}

// Address implements Banker.
func (gb *GRC20Banker) Address() string {
	return gb.owner.String()
}

// NewGRC20Banker creates a new GRC20Banker with the current Realm's address
// as the owner.
func NewGRC20Banker(lister TokenListerFunc) (*GRC20Banker, error) {
	owner := runtime.CurrentRealm().Address()

	return NewGRC20BankerWithOwner(owner, lister)
}

// NewGRC20BankerWithOwner creates a new GRC20Banker with the given address.
func NewGRC20BankerWithOwner(owner address, lister TokenListerFunc) (*GRC20Banker, error) {
	if owner == "" {
		return nil, ErrNoOwnerProvided
	}

	if lister == nil {
		return nil, ErrNoListerProvided
	}

	return &GRC20Banker{
		owner:  owner,
		lister: lister,
	}, nil
}

// grc20Payment represents a payment that is issued by a GRC20Banker.
type grc20Payment struct {
	tokenKey  string  // The key associated with the GRC20 token.
	amount    int64   // The amount of token to send.
	toAddress address // The recipient of the payment.
}

var _ Payment = (*grc20Payment)(nil)

// BankerID implements Payment.
func (grc20Payment) BankerID() string {
	return GRC20Banker{}.ID()
}

// String implements Payment.
func (gp grc20Payment) String() string {
	amount := strconv.Itoa(int(gp.amount))
	return amount + gp.tokenKey + " to " + gp.toAddress.String()
}

// NewGRC20Payment creates a new grc20Payment.
func NewGRC20Payment(tokenKey string, amount int64, toAddress address) Payment {
	return grc20Payment{
		tokenKey:  tokenKey,
		amount:    amount,
		toAddress: toAddress,
	}
}
