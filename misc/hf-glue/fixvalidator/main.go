// fixvalidator rewrites the validator set in a gnoland genesis.json.
// Input modes (mutually exclusive):
//   - priv_validator_key.json path (single validator)
//   - bech32 address + pubkey pair (single validator, key-less environments)
//   - valset-list file (multi-validator, gno-cluster-style lines)
//
// Usage:
//
//	fixvalidator --priv-key <path>                          --genesis <path> [--name NAME] [--power N]
//	fixvalidator --address g1... --pubkey gpub1...          --genesis <path> [--name NAME] [--power N]
//	fixvalidator --valset-list <path>                       --genesis <path>
//
// valset-list format: one validator per line, three whitespace-separated
// fields — "<name> <power> <pubkey>". Blank lines and lines starting with '#'
// are ignored. Addresses are derived from pubkeys. This matches gno-cluster's
// INITIAL_VALSET output (strip the `INITIAL_VALSET=(` wrapper and quotes
// before feeding it in).
//
// This is testbed glue (misc/hf-glue). Not intended to be installed.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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
		valsetList  string
		genesisPath string
		name        string
		power       int64
		emitJSON    bool
	)

	flag.StringVar(&privPath, "priv-key", "", "path to priv_validator_key.json (single-validator mode)")
	flag.StringVar(&addrStr, "address", "", "validator bech32 address g1... (requires --pubkey; single-validator mode)")
	flag.StringVar(&pubkeyStr, "pubkey", "", "validator bech32 pubkey gpub1... (requires --address)")
	flag.StringVar(&valsetList, "valset-list", "", "path to valset-list file (multi-validator mode; '<name> <power> <pubkey>' per line)")
	flag.StringVar(&genesisPath, "genesis", "", "path to genesis.json to rewrite in place (omit with --emit-json)")
	flag.StringVar(&name, "name", "hf-glue-local", "validator name (single-validator mode only)")
	flag.Int64Var(&power, "power", 10, "validator voting power (single-validator mode only)")
	flag.BoolVar(&emitJSON, "emit-json", false, "print the resolved valset as hf-glue's NEW_VALSET_JSON format to stdout and exit (no --genesis needed)")
	flag.Parse()

	if !emitJSON && genesisPath == "" {
		fmt.Fprintln(os.Stderr, "--genesis is required (or use --emit-json)")
		os.Exit(2)
	}

	validators, err := resolveValidators(privPath, addrStr, pubkeyStr, valsetList, name, power)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	if emitJSON {
		if err := emitNewValsetJSON(os.Stdout, validators); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	if err := run(genesisPath, validators); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// emitNewValsetJSON writes the validator set as a JSON array in the same
// shape hf-glue's migrate.sh and build.sh expect for NEW_VALSET_JSON —
// {address, pub_key, voting_power, name}.
func emitNewValsetJSON(w io.Writer, validators []bft.GenesisValidator) error {
	type entry struct {
		Address     string `json:"address"`
		PubKey      string `json:"pub_key"`
		VotingPower int64  `json:"voting_power"`
		Name        string `json:"name"`
	}
	entries := make([]entry, len(validators))
	for i, v := range validators {
		entries[i] = entry{
			Address:     v.Address.String(),
			PubKey:      v.PubKey.String(),
			VotingPower: v.Power,
			Name:        v.Name,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// resolveValidators chooses the input mode and returns the new valset. Modes
// are mutually exclusive; priv-key > address+pubkey > valset-list in precedence.
func resolveValidators(privPath, addrStr, pubkeyStr, valsetList, name string, power int64) ([]bft.GenesisValidator, error) {
	switch {
	case privPath != "":
		pv, err := signer.LoadFileKey(privPath)
		if err != nil {
			return nil, fmt.Errorf("load priv key: %w", err)
		}
		return []bft.GenesisValidator{{Address: pv.Address, PubKey: pv.PubKey, Power: power, Name: name}}, nil

	case addrStr != "" || pubkeyStr != "":
		if addrStr == "" || pubkeyStr == "" {
			return nil, fmt.Errorf("--address and --pubkey must be provided together")
		}
		address, err := crypto.AddressFromBech32(addrStr)
		if err != nil {
			return nil, fmt.Errorf("parse address %q: %w", addrStr, err)
		}
		pubkey, err := crypto.PubKeyFromBech32(pubkeyStr)
		if err != nil {
			return nil, fmt.Errorf("parse pubkey %q: %w", pubkeyStr, err)
		}
		if derived := pubkey.Address(); address != derived {
			return nil, fmt.Errorf("--address %s does not match --pubkey (derives %s)", address, derived)
		}
		return []bft.GenesisValidator{{Address: address, PubKey: pubkey, Power: power, Name: name}}, nil

	case valsetList != "":
		f, err := os.Open(valsetList)
		if err != nil {
			return nil, fmt.Errorf("open valset-list %q: %w", valsetList, err)
		}
		defer f.Close()
		return parseValsetList(f)

	default:
		return nil, fmt.Errorf("one of --priv-key, --address/--pubkey, or --valset-list is required")
	}
}

// parseValsetList reads the multi-validator list format. Format per line:
//
//	<name> <power> <pubkey>
//
// Blank lines and lines whose first non-space character is '#' are ignored.
// Pubkey addresses are derived via crypto.PubKey.Address().
func parseValsetList(r io.Reader) ([]bft.GenesisValidator, error) {
	var out []bft.GenesisValidator
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %d: want 3 fields '<name> <power> <pubkey>', got %d", lineNo, len(fields))
		}
		name, powerStr, pubkeyStr := fields[0], fields[1], fields[2]
		power, err := strconv.ParseInt(powerStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid power %q: %w", lineNo, powerStr, err)
		}
		pubkey, err := crypto.PubKeyFromBech32(pubkeyStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid pubkey %q: %w", lineNo, pubkeyStr, err)
		}
		out = append(out, bft.GenesisValidator{
			Address: pubkey.Address(),
			PubKey:  pubkey,
			Power:   power,
			Name:    name,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan valset-list: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("valset-list is empty")
	}
	return out, nil
}

func run(genesisPath string, validators []bft.GenesisValidator) error {
	genDoc, err := bft.GenesisDocFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("load genesis: %w", err)
	}

	oldCount := len(genDoc.Validators)
	genDoc.Validators = validators

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

	fmt.Printf("replaced %d validator(s) with %d new validator(s):\n", oldCount, len(validators))
	for i, v := range validators {
		fmt.Printf("  [%d] %s (%s), power=%d, name=%q\n", i, v.Address, v.PubKey, v.Power, v.Name)
	}
	return nil
}
