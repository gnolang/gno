package treasury

import (
	"std"

	"gno.land/p/demo/grc/grc20"
	t "gno.land/p/nt/treasury"

	"gno.land/r/demo/grc20reg"
	"gno.land/r/gov/dao"
)

var (
	treasury  *t.Treasury
	tokenKeys = []string{
		// TODO: Add the default GRC20 tokens we want to support here.
	}
)

func init() {
	// Define a token lister for the GRC20Banker.
	// For now, GovDAO uses a static list of tokens.
	grc20Lister := func() map[string]*grc20.Token {
		// Get the GRC20 tokens from the registry.
		tokens := map[string]*grc20.Token{}
		for _, key := range tokenKeys {
			// Get the token by its key.
			token := grc20reg.Get(key)
			if token != nil {
				tokens[key] = token
			}
		}

		return tokens
	}

	// Init the treasury bankers.
	coinsBanker, err := t.NewCoinsBanker(std.NewBanker(std.BankerTypeRealmSend))
	if err != nil {
		panic("failed to create CoinsBanker: " + err.Error())
	}
	grc20Banker, err := t.NewGRC20Banker(grc20Lister)
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
func Send(_ realm, payment t.Payment) {
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

// TestingSetTokenKeys sets the token registry keys for testing purposes.
func TestingSetTokenKeys(_ realm, keys []string) {
	// Ensure that this function can only be called from the treasury/test realm.
	caller := std.PreviousRealm().PkgPath()
	if caller != "gno.land/r/gov/dao/v3/treasury/test" {
		panic("only the treasury/test realm can set token keys, current:" + caller)
	}

	tokenKeys = keys
}
