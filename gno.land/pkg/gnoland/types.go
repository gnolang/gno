package gnoland

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []Balance `json:"balances"`
	Txs      []std.Tx  `json:"txs"`
}

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

func New() Balances {
	return make(Balances)
}

func (balances Balances) Set(address crypto.Address, amount std.Coins) {
	balances[address] = Balance{
		Address: address,
		Amount:  amount,
	}
}

func (balances Balances) Get(address crypto.Address) (balance Balance, ok bool) {
	balance, ok = balances[address]
	return
}

func (balances Balances) List() []Balance {
	list := make([]Balance, 0, len(balances))
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
		var balance Balance
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
