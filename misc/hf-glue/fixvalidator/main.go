// fixvalidator rewrites the validator set in a gnoland genesis.json to
// one or more validators loaded from priv_validator_key.json files.
//
// Usage:
//
//	fixvalidator --priv-key <path> [--priv-key <path>...] --genesis <path> [--name NAME] [--power N]
//
// Names are auto-suffixed (-0, -1, ...) when more than one priv-key is
// passed. Power is applied identically to each validator.
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

type stringList []string

func (s *stringList) String() string { return fmt.Sprint(*s) }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		privPaths   stringList
		genesisPath string
		name        string
		power       int64
	)

	flag.Var(&privPaths, "priv-key", "path to priv_validator_key.json (repeatable)")
	flag.StringVar(&genesisPath, "genesis", "", "path to genesis.json to rewrite in place")
	flag.StringVar(&name, "name", "hf-glue-local", "validator name (suffixed -N when multiple)")
	flag.Int64Var(&power, "power", 10, "voting power (applied to each)")
	flag.Parse()

	if len(privPaths) == 0 || genesisPath == "" {
		fmt.Fprintln(os.Stderr, "--priv-key (>=1) and --genesis are required")
		os.Exit(2)
	}

	if err := run(privPaths, genesisPath, name, power); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(privPaths []string, genesisPath, name string, power int64) error {
	genDoc, err := bft.GenesisDocFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}

	oldCount := len(genDoc.Validators)
	newVals := make([]bft.GenesisValidator, 0, len(privPaths))
	for i, p := range privPaths {
		pv, err := signer.LoadFileKey(p)
		if err != nil {
			return fmt.Errorf("load priv key %s: %w", p, err)
		}
		n := name
		if len(privPaths) > 1 {
			n = fmt.Sprintf("%s-%d", name, i)
		}
		newVals = append(newVals, bft.GenesisValidator{
			Address: pv.Address,
			PubKey:  pv.PubKey,
			Power:   power,
			Name:    n,
		})
	}
	genDoc.Validators = newVals

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

	fmt.Printf("replaced %d validator(s) with %d new validator(s):\n", oldCount, len(newVals))
	for _, v := range newVals {
		fmt.Printf("  - %s  addr=%s  pub=%s  power=%d\n", v.Name, v.Address, v.PubKey, v.Power)
	}
	return nil
}
