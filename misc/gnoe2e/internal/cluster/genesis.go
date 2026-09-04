package cluster

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	// validatorBalance is the ugnot every validator holds at genesis. Enough to
	// sign the genesis package transactions, which is all a validator key
	// spends.
	validatorBalance = 100_000_000

	// genesisDeployBudget is what one package deployed at genesis costs its
	// sender, at the fee BuildGenesis signs those transactions with.
	genesisDeployBudget = 50_000_000
)

// fundDeployer adds what deploying packages costs to whatever addr already
// holds. Added rather than set, because the deployer is validator 0 and
// clusterBalances funded it first.
func fundDeployer(balances gnoland.Balances, addr crypto.Address, packages int) {
	held, _ := balances.Get(addr)
	balances.Set(addr, held.Amount.Add(
		std.NewCoins(std.NewCoin("ugnot", int64(packages)*genesisDeployBudget))))
}

// clusterBalances funds every validator plus the accounts the caller named.
// That is the whole of a cluster's money: no other address has a key here.
func clusterBalances(validatorKeys []*signer.FileKey, extra map[string]int64) (gnoland.Balances, error) {
	balances := gnoland.NewBalances()

	for _, key := range validatorKeys {
		balances.Set(key.Address, std.NewCoins(std.NewCoin("ugnot", validatorBalance)))
	}

	for addr, amount := range extra {
		address, err := crypto.AddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid balance address %q: %w", addr, err)
		}
		balances.Set(address, std.NewCoins(std.NewCoin("ugnot", amount)))
	}

	return balances, nil
}

// settableParams is the module set `gnogenesis params set` writes, keyed the
// way it keys them, so "vm.chain_domain" means the same thing in both.
type settableParams struct {
	Auth *auth.Params `json:"auth"`
	VM   *vm.Params   `json:"vm"`
	Bank *bank.Params `json:"bank"`
}

func newSettableParams(genState *gnoland.GnoGenesisState) *settableParams {
	return &settableParams{&genState.Auth.Params, &genState.VM.Params, &genState.Bank.Params}
}

// GenesisDefaults lists every "genesis." key a scenario can set in its
// "-- cluster --" section, with the value the chain starts from.
func GenesisDefaults() []Override {
	genState := gnoland.DefaultGenState()
	return fieldDefaults(reflect.ValueOf(newSettableParams(&genState)).Elem(), genesisParamsSelector)
}

// applyGenesisParams sets a scenario's genesis param overrides on the genesis
// state, then validates the module each key touched.
//
// The validation is what keeps a bad value a harness error rather than a node
// that dies at boot for no stated reason: applyOverride parses the value into
// the field's type but consults neither Validate nor WillSetParam, and all of
// these params are consensus state.
//
// Only the module the key names, and only after the legacy defaulting the node
// itself applies -- validating all three would refuse a genesis the node boots
// fine, the way `gnogenesis params set` found.
func applyGenesisParams(genState *gnoland.GnoGenesisState, overrides []Override) error {
	if len(overrides) == 0 {
		return nil
	}

	root := reflect.ValueOf(newSettableParams(genState)).Elem()

	for _, o := range overrides {
		if err := applyOverride(root, genesisParamsSelector, o); err != nil {
			// The key is named the way the scenario wrote it, prefix included,
			// rather than as the path the params were traversed by.
			return fmt.Errorf("genesis.%w", err)
		}

		var verr error
		switch {
		case strings.HasPrefix(o.Key, "vm."):
			verr = genState.VM.Params.ApplyLegacyDefaults().Validate()
		case strings.HasPrefix(o.Key, "auth."):
			verr = genState.Auth.Params.Validate()
		case strings.HasPrefix(o.Key, "bank."):
			verr = genState.Bank.Params.Validate()
		}
		if verr != nil {
			return fmt.Errorf("invalid params after setting genesis.%s: %w", o.Key, verr)
		}
	}
	return nil
}

// CopySharedGenesis copies the shared genesis file to each node's directory.
func CopySharedGenesis(tempDir string, node *Node) error {
	sharedGenesisPath := filepath.Join(tempDir, "shared_genesis.json")

	genesisData, err := os.ReadFile(sharedGenesisPath)
	if err != nil {
		return fmt.Errorf("read shared genesis: %w", err)
	}

	if err := os.WriteFile(node.Genesis, genesisData, 0644); err != nil {
		return fmt.Errorf("write genesis to node: %w", err)
	}

	slog.Debug("copied genesis to node", "node_index", node.Index, "path", node.Genesis)
	return nil
}

// PrintGenesisConfig prints genesis configuration for debugging.
func PrintGenesisConfig(gen *bft.GenesisDoc) {
	slog.Debug("genesis config",
		"chain_id", gen.ChainID,
		"max_gas", gen.ConsensusParams.Block.MaxGas,
		"max_tx_bytes", gen.ConsensusParams.Block.MaxTxBytes,
		"time_iota_ms", gen.ConsensusParams.Block.TimeIotaMS,
		"validators", len(gen.Validators),
	)
	for i, val := range gen.Validators {
		slog.Debug("genesis validator", "index", i, "address", val.Address, "power", val.Power, "name", val.Name)
	}

	// BuildGenesis assigns the state by value, which is also what
	// bft.GenesisDocFromFile decodes it back to.
	if genState, ok := gen.AppState.(gnoland.GnoGenesisState); ok {
		slog.Debug("genesis state", "balances", len(genState.Balances), "txs", len(genState.Txs))
	}
}

// ValidateGenesisParams reports whether overrides resolve, parse and leave the
// modules they touch valid, checked against the genesis a cluster starts from.
//
// The counterpart to ValidateNodeConfig, and there for the same reason: a
// misspelled key fails when the scenario is read rather than when its own
// cluster is built.
func ValidateGenesisParams(overrides []Override) error {
	_, err := ResolveGenesisState(GenesisConfig{Params: overrides})
	return err
}
