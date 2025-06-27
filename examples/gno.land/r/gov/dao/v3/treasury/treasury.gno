package treasury

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	t "gno.land/p/nt/treasury"

	"gno.land/r/gov/dao"
)

var treasury *t.Treasury

func init() {
	// Define a lister for the GRC20Banker that returns a static list
	// of GRC20 tokens.
	grc20Lister := func() []*grc20.Token {
		tokens := []*grc20.Token{
			// TODO: Add the GRC20 tokens we want to support.
		}

		return tokens
	}

	// Init the treasury bankers.
	coinsBanker, err := t.NewCoinsBanker(std.NewBanker(std.BankerTypeRealmSend))
	if err != nil {
		panic("failed to create CoinsBanker: " + err.Error())
	}
	grc20Banker, err := t.NewGRC20Banker(grc20regLister)
	if err != nil {
		panic("failed to create GRC20Banker: " + err.Error())
	}
	bankers := []t.Banker{
		coinsBanker,
		grc20Banker,
	}

	// Create the treasury instance with the bankers.
	treasury, err = t.New(bankers)
	if err != nil {
		panic("failed to create treasury: " + err.Error())
	}
}

// Send sends a payment using the treasury instance.
func Send(cur realm, payment t.Payment) {
	// NOTE: Replace this by cur.Previous().PkgPath() when it will be implemented.
	pkgPath := std.PreviousRealm().PkgPath()

	// Check if the current realm is allowed to send payments.
	if !dao.InAllowedDAOs(pkgPath) {
		panic("this Realm is not allowed to send payment: " + pkgPath)
	}

	// Send the payment using the treasury instance.
	if err := treasury.Send(payment); err != nil {
		panic(err)
	}
}

// History returns the n last payments sent by the banker with the given ID.
// Returns the full history if n is 0 or greater than the number of payments.
func History(bankerID string, n uint) []t.Payment {
	history, err := treasury.History(bankerID, n)
	if err != nil {
		panic("failed to get history: " + err.Error())
	}

	return history
}

// Balances returns the balances of the banker with the given ID.
func Balances(bankerID string) []t.Balance {
	balances, err := treasury.Balances(bankerID)
	if err != nil {
		panic("failed to get balances: " + err.Error())
	}

	return balances
}

// Address returns the address of the banker with the given ID.
func Address(bankerID string) string {
	addr, err := treasury.Address(bankerID)
	if err != nil {
		panic("failed to get address: " + err.Error())
	}

	return addr
}

// HasBanker checks if a banker with the given ID is registered.
func HasBanker(bankerID string) bool {
	return treasury.HasBanker(bankerID)
}

// ListBankerIDs returns a list of all registered banker IDs.
func ListBankerIDs() []string {
	return treasury.ListBankerIDs()
}

func Render(path string) string {
	return treasury.Render(path)
}
