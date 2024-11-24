package gnoland

import (
	"errors"
	"fmt"
	"strings"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/pelletier/go-toml"
)

// LoadGenesisBalancesFile loads genesis balances from the provided file path.
func LoadGenesisBalancesFile(path string) ([]Balance, error) {
	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
	content, err := osm.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")

	balances := make([]Balance, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// remove comments.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// skip empty lines.
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=") // <address>=<coin>
		if len(parts) != 2 {
			return nil, errors.New("invalid genesis_balance line: " + line)
		}

		addr, err := crypto.AddressFromBech32(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid balance addr %s: %w", parts[0], err)
		}

		coins, err := std.ParseCoins(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid balance coins %s: %w", parts[1], err)
		}

		balances = append(balances, Balance{
			Address: addr,
			Amount:  coins,
		})
	}

	return balances, nil
}

// LoadGenesisParamsFile loads genesis params from the provided file path.
func LoadGenesisParamsFile(path string) ([]Param, error) {
	// each param is in the form: key.kind=value
	content, err := osm.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m := map[string] /*category*/ map[string] /*key*/ map[string] /*kind*/ interface{} /*value*/ {}
	err = toml.Unmarshal(content, &m)
	if err != nil {
		return nil, err
	}

	params := make([]Param, 0)
	for category, keys := range m {
		for key, kinds := range keys {
			for kind, val := range kinds {
				param := Param{
					key:  category + "." + key,
					kind: kind,
				}
				switch kind {
				case "uint64": // toml
					param.value = uint64(val.(int64))
				default:
					param.value = val
				}
				if err := param.Verify(); err != nil {
					return nil, err
				}
				params = append(params, param)
			}
		}
	}

	return params, nil
}

// LoadGenesisTxsFile loads genesis transactions from the provided file path.
// XXX: Improve the way we generate and load this file
func LoadGenesisTxsFile(path string, chainID string, genesisRemote string) ([]TxWithMetadata, error) {
	txs := make([]TxWithMetadata, 0)

	txsBz, err := osm.ReadFile(path)
	if err != nil {
		return nil, err
	}
	txsLines := strings.Split(string(txsBz), "\n")
	for _, txLine := range txsLines {
		if txLine == "" {
			continue // Skip empty line.
		}

		// Patch the TX.
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

		var tx TxWithMetadata
		if err := amino.UnmarshalJSON([]byte(txLine), &tx); err != nil {
			return nil, fmt.Errorf("unable to Unmarshall txs file: %w", err)
		}

		txs = append(txs, tx)
	}

	return txs, nil
}

// LoadPackagesFromDir loads gno packages from a directory.
// It creates and returns a list of transactions based on these packages.
func LoadPackagesFromDir(dir string, creatorMnemonic string, chainID string, fee std.Fee) ([]TxWithMetadata, error) {
	// list all packages from target path
	pkgs, err := gnomod.ListPkgs(dir)
	if err != nil {
		return nil, fmt.Errorf("listing gno packages: %w", err)
	}

	// Sort packages by dependencies.
	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("sorting packages: %w", err)
	}

	kb := keys.NewInMemory()
	// Save the account
	info, err := kb.CreateAccount(
		"deployer",
		creatorMnemonic,
		"",
		"",
		0,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create account, %w", err)
	}

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()
	txs := make([]TxWithMetadata, 0, len(nonDraftPkgs))

	for _, pkg := range nonDraftPkgs {
		tx, err := LoadPackage(pkg, info.GetAddress(), fee, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to load package %q: %w", pkg.Dir, err)
		}

		// Both account number and account sequence are 0 on genesis transactions
		txData, err := tx.GetSignBytes(chainID, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to generate mnemonic, %w", err)
		}
		sig, pub, err := kb.Sign("deployer", "", txData)
		if err != nil {
			return nil, err
		}
		tx.Signatures = []std.Signature{{
			PubKey:    pub,
			Signature: sig,
		}}
		txs = append(txs, TxWithMetadata{
			Tx: tx,
		})
	}

	return txs, nil
}

// LoadPackage loads a single package into a `std.Tx`
func LoadPackage(pkg gnomod.Pkg, creator bft.Address, fee std.Fee, deposit std.Coins) (std.Tx, error) {
	var tx std.Tx

	// Open files in directory as MemPackage.
	memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)
	err := memPkg.Validate()
	if err != nil {
		return tx, fmt.Errorf("invalid package: %w", err)
	}

	// Create transaction
	tx.Fee = fee
	tx.Msgs = []std.Msg{
		vmm.MsgAddPackage{
			Creator: creator,
			Package: memPkg,
			Deposit: deposit,
		},
	}
	tx.Signatures = make([]std.Signature, len(tx.GetSigners()))

	return tx, nil
}
