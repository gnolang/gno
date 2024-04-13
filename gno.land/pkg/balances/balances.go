package balances

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type Balances map[crypto.Address]gnoland.Balance

func New() Balances {
	return make(Balances)
}

func (balances Balances) Set(address crypto.Address, amount std.Coins) {
	balances[address] = gnoland.Balance{
		Address: address,
		Amount:  amount,
	}
}

func (balances Balances) Get(address crypto.Address) (balance gnoland.Balance, ok bool) {
	balance, ok = balances[address]
	return
}

func (balances Balances) List() []gnoland.Balance {
	list := make([]gnoland.Balance, 0, len(balances))
	for _, balance := range balances {
		list = append(list, balance)
	}
	return list
}

// leftMerge left-merges the two maps
func (a Balances) LeftMerge(b Balances) {
	for key, bVal := range b {
		if _, present := (a)[key]; !present {
			(a)[key] = bVal
		}
	}
}

func GetBalancesFromEntries(entries ...string) (Balances, error) {
	balances := New()
	return balances, balances.LoadFromEntries(entries...)
}

// LoadFromEntries extracts the balance entries in the form of <address>=<amount>
func (balances Balances) LoadFromEntries(entries ...string) error {
	for _, entry := range entries {
		var balance gnoland.Balance
		if err := balance.Parse(entry); err != nil {
			return fmt.Errorf("unable to parse balance entry: %w", err)
		}
		balances[balance.Address] = balance
	}

	return nil
}

func GetBalancesFromSheet(sheet io.Reader) (Balances, error) {
	balances := New()
	return balances, balances.LoadFromSheet(sheet)
}

// LoadFromSheet extracts the balance sheet from the passed in
// balance sheet file, that has the format of <address>=<amount>ugnot
func (balances Balances) LoadFromSheet(sheet io.Reader) error {
	// Parse the balances
	scanner := bufio.NewScanner(sheet)

	for scanner.Scan() {
		entry := scanner.Text()

		// Remove comments
		entry = strings.Split(entry, "#")[0]
		entry = strings.TrimSpace(entry)

		// Skip empty lines
		if entry == "" {
			continue
		}

		if err := balances.LoadFromEntries(entry); err != nil {
			return fmt.Errorf("unable to load entries: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error encountered while scanning, %w", err)
	}

	return nil
}
