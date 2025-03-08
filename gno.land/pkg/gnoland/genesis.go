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
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/pelletier/go-toml"
)

const initGasPrice = "1ugnot/1000gas"

// LoadGenesisBalancesFile loads genesis balances from the provided file path.
func LoadGenesisBalancesFile(path string) (Balances, error) {
	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
	content, err := osm.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")

	balances := make(Balances, len(lines))
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

		balances.Set(addr, coins)
	}

	return balances, nil
}

func splitTypedName(typedName string) (name string, type_ string) {
	parts := strings.Split(typedName, ".")
	if len(parts) == 1 {
		return typedName, ""
	} else if len(parts) == 2 {
		return parts[0], parts[1]
	} else {
		panic("malforumed typed name: expected <name> or <name>.<type> but got " + typedName)
	}
}

// LoadGenesisParamsFile loads genesis params from the provided file path.
func LoadGenesisParamsFile(path string, ggs *GnoGenesisState) error {
	content, err := osm.ReadFile(path)
	if err != nil {
		return err
	}

	// Parameters are grouped by modules (or more specifically module:submodule).
	// The vm module uses the submodule for realm package paths.
	// If only the module is specified, the submodule is assumed to be "p"
	// for keeper param structs.
	m := map[string] /* <module>(:<submodule>)? */ map[string] /* <name> */ interface{} /* <value> */ {}
	err = toml.Unmarshal(content, &m)
	if err != nil {
		return err
	}

	// XXX Write onto ggs for other keeper params.

	// Write onto ggs.VM.Params.
	if vmparams, ok := m["vm"]; ok {
		for name, value := range vmparams {
			name, _ := splitTypedName(name)
			switch name {
			case "chain_domain":
				ggs.VM.Params.ChainDomain = value.(string)
			case "sysusers_pkgpath":
				ggs.VM.Params.SysUsersPkgPath = value.(string)
			default:
				return errors.New("unexpected vm parameter " + name)
			}
		}
	}

	// Write onto ggs.VM.RealmParams.
	for modrlm, values := range m {
		if !strings.HasPrefix(modrlm, "vm:") {
			continue
		}
		parts := strings.Split(modrlm, ":")
		numparts := len(parts)
		if numparts == 1 {
			// keeper param struct (sys param). skip
		} else if numparts == 2 {
			realm := parts[1]
			// XXX validate realm part.
			for name, value := range values {
				name, type_ := splitTypedName(name)
				if type_ == "strings" {
					vz := value.([]interface{})
					sz := make([]string, len(vz))
					for i, v := range vz {
						sz[i] = v.(string)
					}
					value = sz
				}
				param := params.NewParam(realm+":"+name, value)
				ggs.VM.RealmParams = append(ggs.VM.RealmParams, param)
			}
		} else {
			return errors.New("invalid key " + modrlm + ", expected format <module>:<realm>:<name>")
		}
	}
	return nil
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
func LoadPackagesFromDir(dir string, creator bft.Address, fee std.Fee) ([]TxWithMetadata, error) {
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

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()
	txs := make([]TxWithMetadata, 0, len(nonDraftPkgs))
	for _, pkg := range nonDraftPkgs {
		tx, err := LoadPackage(pkg, creator, fee, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to load package %q: %w", pkg.Dir, err)
		}

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
	memPkg := gno.MustReadMemPackage(pkg.Dir, pkg.Name)
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

func DefaultGenState() GnoGenesisState {
	authGen := auth.DefaultGenesisState()
	gp, err := std.ParseGasPrice(initGasPrice)
	if err != nil {
		panic(err)
	}
	authGen.Params.InitialGasPrice = gp

	gs := GnoGenesisState{
		Balances: []Balance{},
		Txs:      []TxWithMetadata{},
		Auth:     authGen,
		Bank:     bank.DefaultGenesisState(),
		VM:       vmm.DefaultGenesisState(),
	}
	return gs
}
