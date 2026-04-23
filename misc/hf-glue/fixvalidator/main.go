// fixvalidator rewrites the validator set in a gnoland genesis.json to a
// single validator. Input can be either a priv_validator_key.json file, or
// a pair of bech32 strings (address + pubkey) for key-less environments.
//
// Usage:
//
//	fixvalidator --priv-key <path>      --genesis <path> [--name NAME] [--power N]
//	fixvalidator --address g1... --pubkey gpub1... --genesis <path> [--name NAME] [--power N]
//
// This is testbed glue (misc/hf-glue). Not intended to be installed.
package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/gnolang/gno/gno.land/pkg/gnoland" // register GnoGenesisState amino type
	"github.com/gnolang/gno/tm2/pkg/amino"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func main() {
	var (
		privPath    string
		addrStr     string
		pubkeyStr   string
		genesisPath string
		name        string
		power       int64
	)

	flag.StringVar(&privPath, "priv-key", "", "path to priv_validator_key.json (alternative to --address/--pubkey)")
	flag.StringVar(&addrStr, "address", "", "validator bech32 address g1... (alternative to --priv-key; requires --pubkey)")
	flag.StringVar(&pubkeyStr, "pubkey", "", "validator bech32 pubkey gpub1... (required with --address)")
	flag.StringVar(&genesisPath, "genesis", "", "path to genesis.json to rewrite in place")
	flag.StringVar(&name, "name", "hf-glue-local", "validator name")
	flag.Int64Var(&power, "power", 10, "validator voting power")
	flag.Parse()

	if genesisPath == "" {
		fmt.Fprintln(os.Stderr, "--genesis is required")
		os.Exit(2)
	}

	address, pubkey, err := resolveValidator(privPath, addrStr, pubkeyStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	if err := run(genesisPath, address, pubkey, name, power); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// resolveValidator returns (address, pubkey) from either a priv_validator_key.json
// path or an explicit (bech32 address, bech32 pubkey) pair. When both are
// supplied, --priv-key wins. When using --address/--pubkey, the address is
// verified to match the pubkey's derived address.
func resolveValidator(privPath, addrStr, pubkeyStr string) (crypto.Address, crypto.PubKey, error) {
	if privPath != "" {
		pv, err := signer.LoadFileKey(privPath)
		if err != nil {
			return crypto.Address{}, nil, fmt.Errorf("load priv key: %w", err)
		}
		return pv.Address, pv.PubKey, nil
	}
	if addrStr == "" || pubkeyStr == "" {
		return crypto.Address{}, nil, fmt.Errorf("either --priv-key OR (--address AND --pubkey) is required")
	}
	address, err := crypto.AddressFromBech32(addrStr)
	if err != nil {
		return crypto.Address{}, nil, fmt.Errorf("parse address %q: %w", addrStr, err)
	}
	pubkey, err := crypto.PubKeyFromBech32(pubkeyStr)
	if err != nil {
		return crypto.Address{}, nil, fmt.Errorf("parse pubkey %q: %w", pubkeyStr, err)
	}
	if derived := pubkey.Address(); address != derived {
		return crypto.Address{}, nil, fmt.Errorf("--address %s does not match --pubkey (derives %s)", address, derived)
	}
	return address, pubkey, nil
}

func run(genesisPath string, address crypto.Address, pubkey crypto.PubKey, name string, power int64) error {
	genDoc, err := bft.GenesisDocFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}

	oldCount := len(genDoc.Validators)
	genDoc.Validators = []bft.GenesisValidator{{
		Address: address,
		PubKey:  pubkey,
		Power:   power,
		Name:    name,
	}}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return fmt.Errorf("validate genesis after rewrite: %w", err)
	}

	data, err := amino.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(genesisPath, data, 0o644); err != nil {
		return fmt.Errorf("write genesis: %w", err)
	}

	fmt.Printf("replaced %d validator(s) with single validator:\n", oldCount)
	fmt.Printf("  address: %s\n", address.String())
	fmt.Printf("  pubkey:  %s\n", pubkey.String())
	fmt.Printf("  name:    %s\n", name)
	fmt.Printf("  power:   %d\n", power)
	return nil
}
