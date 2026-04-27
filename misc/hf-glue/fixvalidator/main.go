// fixvalidator rewrites the validator set in a gnoland genesis.json.
//
// Two input modes (mutually exclusive):
//
//  1. priv-key mode — read a priv_validator_key.json (repeatable for
//     multi-validator):
//
//     fixvalidator --priv-key <path> [--priv-key <path>...] \
//     --genesis <path> [--name NAME] [--power N]
//
//  2. keyless mode — supply bech32 address + pubkey directly, e.g. for a
//     remote-signed validator (gnokms) where the priv key is not on disk:
//
//     fixvalidator --address g1... --pubkey gpub1... \
//     --genesis <path> [--name NAME] [--power N]
//
//     The address is cross-checked against the pubkey's derived address;
//     mismatches are rejected so a typo can't ship a wrong validator.
//
// Names are auto-suffixed (-0, -1, ...) when multiple priv-keys are passed.
// Power is applied identically to each validator.
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
		// Keyless mode (mutually exclusive with --priv-key).
		addrBech string
		pubBech  string
	)

	flag.Var(&privPaths, "priv-key", "path to priv_validator_key.json (repeatable)")
	flag.StringVar(&genesisPath, "genesis", "", "path to genesis.json to rewrite in place")
	flag.StringVar(&name, "name", "hf-glue-local", "validator name (suffixed -N when multiple)")
	flag.Int64Var(&power, "power", 10, "voting power (applied to each)")
	flag.StringVar(&addrBech, "address", "", "validator bech32 address (g1...) for keyless mode")
	flag.StringVar(&pubBech, "pubkey", "", "validator bech32 pubkey (gpub1...) for keyless mode")
	flag.Parse()

	if genesisPath == "" {
		fmt.Fprintln(os.Stderr, "--genesis is required")
		os.Exit(2)
	}
	keyless := addrBech != "" || pubBech != ""
	if keyless && len(privPaths) > 0 {
		fmt.Fprintln(os.Stderr, "--address/--pubkey is mutually exclusive with --priv-key")
		os.Exit(2)
	}
	if !keyless && len(privPaths) == 0 {
		fmt.Fprintln(os.Stderr, "either --priv-key (>=1) or --address+--pubkey is required")
		os.Exit(2)
	}
	if keyless && (addrBech == "" || pubBech == "") {
		fmt.Fprintln(os.Stderr, "keyless mode requires both --address and --pubkey")
		os.Exit(2)
	}

	var err error
	if keyless {
		err = runKeyless(addrBech, pubBech, genesisPath, name, power)
	} else {
		err = run(privPaths, genesisPath, name, power)
	}
	if err != nil {
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
	return writeGenesis(genDoc, oldCount, newVals, genesisPath)
}

func runKeyless(addrBech, pubBech, genesisPath, name string, power int64) error {
	genDoc, err := bft.GenesisDocFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}

	addr, err := crypto.AddressFromBech32(addrBech)
	if err != nil {
		return fmt.Errorf("parse address %q: %w", addrBech, err)
	}
	pub, err := crypto.PubKeyFromBech32(pubBech)
	if err != nil {
		return fmt.Errorf("parse pubkey %q: %w", pubBech, err)
	}
	// Cross-check: derived address from the pubkey must match the bech32
	// address. Catches the typo case where a stale --address is paired with
	// a freshly-rotated --pubkey.
	if derived := pub.Address(); derived != addr {
		return fmt.Errorf("address/pubkey mismatch: --address=%s but pubkey derives to %s",
			addr, derived)
	}

	oldCount := len(genDoc.Validators)
	newVals := []bft.GenesisValidator{{
		Address: addr,
		PubKey:  pub,
		Power:   power,
		Name:    name,
	}}
	return writeGenesis(genDoc, oldCount, newVals, genesisPath)
}

func writeGenesis(genDoc *bft.GenesisDoc, oldCount int, newVals []bft.GenesisValidator, genesisPath string) error {
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
