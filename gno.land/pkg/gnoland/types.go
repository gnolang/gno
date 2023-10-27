package gnoland

import (
	"fmt"
	"strings"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
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
	Value   std.Coins
}

func (b *Balance) Parse(line string) error {
	parts := strings.Split(strings.TrimSpace(line), "=") // <address>=<coins>
	if len(parts) != 2 {
		return fmt.Errorf("invalid balance line: %q", line)
	}

	var err error

	b.Address, err = crypto.AddressFromBech32(parts[0])
	if err != nil {
		return fmt.Errorf("invalid balance addr %s: %w", parts[0], err)
	}

	b.Value, err = std.ParseCoins(parts[1])
	if err != nil {
		return fmt.Errorf("invalid balance coins %s: %w", parts[1], err)
	}

	return nil
}

func (b *Balance) UnmarshalJSON(data []byte) error {
	return b.Parse(string(data))
}

func (b *Balance) Marshaljson() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b Balance) String() string {
	return fmt.Sprintf("%s=%s", b.Address.String(), b.Value.String())
}
