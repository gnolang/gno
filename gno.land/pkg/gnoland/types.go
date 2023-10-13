package gnoland

import (
	"errors"
	"fmt"
	"strings"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
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

type PackagePath struct {
	Creator bft.Address
	Deposit std.Coins
	Fee     std.Fee
	Path    string
}

func (p PackagePath) Load() ([]std.Tx, error) {
	if p.Creator.IsZero() {
		return nil, errors.New("empty creator address")
	}

	if p.Path == "" {
		return nil, errors.New("empty package path")
	}

	// list all packages from target path
	pkgs, err := gnomod.ListPkgs(p.Path)
	if err != nil {
		return nil, fmt.Errorf("listing gno packages: %w", err)
	}

	// Sort packages by dependencies.
	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("sorting packages: %w", err)
	}

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()
	txs := []std.Tx{}
	for _, pkg := range nonDraftPkgs {
		// Open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

		// Create transaction
		tx := std.Tx{
			Fee: p.Fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: p.Creator,
					Package: memPkg,
					Deposit: p.Deposit,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}
