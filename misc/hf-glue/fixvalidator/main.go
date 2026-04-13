// fixvalidator rewrites the validator set in a gnoland genesis.json to a
// single validator loaded from a priv_validator_key.json file.
//
// Usage:
//
//	fixvalidator --priv-key <path> --genesis <path> [--name NAME] [--power N]
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
)

func main() {
	var (
		privPath    string
		genesisPath string
		name        string
		power       int64
	)

	flag.StringVar(&privPath, "priv-key", "", "path to priv_validator_key.json")
	flag.StringVar(&genesisPath, "genesis", "", "path to genesis.json to rewrite in place")
	flag.StringVar(&name, "name", "hf-glue-local", "validator name")
	flag.Int64Var(&power, "power", 10, "validator voting power")
	flag.Parse()

	if privPath == "" || genesisPath == "" {
		fmt.Fprintln(os.Stderr, "both --priv-key and --genesis are required")
		os.Exit(2)
	}

	if err := run(privPath, genesisPath, name, power); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(privPath, genesisPath, name string, power int64) error {
	pv, err := signer.LoadFileKey(privPath)
	if err != nil {
		return fmt.Errorf("load priv key: %w", err)
	}

	genDoc, err := bft.GenesisDocFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}

	oldCount := len(genDoc.Validators)
	genDoc.Validators = []bft.GenesisValidator{{
		Address: pv.Address,
		PubKey:  pv.PubKey,
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
	fmt.Printf("  address: %s\n", pv.Address.String())
	fmt.Printf("  pubkey:  %s\n", pv.PubKey.String())
	fmt.Printf("  name:    %s\n", name)
	fmt.Printf("  power:   %d\n", power)
	return nil
}
