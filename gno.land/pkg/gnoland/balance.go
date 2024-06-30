package gnoland

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type Balance struct {
	Address bft.Address
	Amount  std.Coins
}

func (b *Balance) Verify() error {
	if b.Address.IsZero() {
		return ErrBalanceEmptyAddress
	}

	if b.Amount.Len() == 0 {
		return ErrBalanceEmptyAmount
	}

	return nil
}

func (b *Balance) Parse(entry string) error {
	parts := strings.Split(strings.TrimSpace(entry), "=") // <address>=<coins>
	if len(parts) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	var err error

	b.Address, err = crypto.AddressFromBech32(parts[0])
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", parts[0], err)
	}

	b.Amount, err = std.ParseCoins(parts[1])
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", parts[1], err)
	}

	return nil
}

func (b *Balance) UnmarshalAmino(rep string) error {
	return b.Parse(rep)
}

func (b Balance) MarshalAmino() (string, error) {
	return b.String(), nil
}

func (b Balance) String() string {
	return fmt.Sprintf("%s=%s", b.Address.String(), b.Amount.String())
}

type Balances map[crypto.Address]Balance

func NewBalances() Balances {
	return make(Balances)
}

func (bs Balances) Set(address crypto.Address, amount std.Coins) {
	bs[address] = Balance{
		Address: address,
		Amount:  amount,
	}
}

func (bs Balances) Get(address crypto.Address) (balance Balance, ok bool) {
	balance, ok = bs[address]
	return
}

func (bs Balances) List() []Balance {
	list := make([]Balance, 0, len(bs))
	for _, balance := range bs {
		list = append(list, balance)
	}
	return list
}

// LeftMerge left-merges the two maps
func (bs Balances) LeftMerge(from Balances) {
	for key, bVal := range from {
		if _, present := (bs)[key]; !present {
			(bs)[key] = bVal
		}
	}
}

func GetBalancesFromEntries(entries ...string) (Balances, error) {
	balances := NewBalances()
	return balances, balances.LoadFromEntries(entries...)
}

// LoadFromEntries extracts the balance entries in the form of <address>=<amount>
func (bs Balances) LoadFromEntries(entries ...string) error {
	for _, entry := range entries {
		var balance Balance
		if err := balance.Parse(entry); err != nil {
			return fmt.Errorf("unable to parse balance entry: %w", err)
		}
		bs[balance.Address] = balance
	}

	return nil
}

func GetBalancesFromSheet(sheet io.Reader) (Balances, error) {
	balances := NewBalances()
	return balances, balances.LoadFromSheet(sheet)
}

// LoadFromSheet extracts the balance sheet from the passed in
// balance sheet file, that has the format of <address>=<amount>ugnot
func (bs Balances) LoadFromSheet(sheet io.Reader) error {
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

		if err := bs.LoadFromEntries(entry); err != nil {
			return fmt.Errorf("unable to load entries: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error encountered while scanning, %w", err)
	}

	return nil
}
