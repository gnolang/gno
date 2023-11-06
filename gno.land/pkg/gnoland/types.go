package gnoland

import (
	"errors"
	"fmt"
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
