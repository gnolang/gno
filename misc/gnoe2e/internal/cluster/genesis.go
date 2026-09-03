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
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
)

// CreateEnhancedBalanceFile creates a balance file with validator + extra accounts.
func CreateEnhancedBalanceFile(tempDir string, validatorKeys []*signer.FileKey, extraBalances map[string]int64) (string, error) {
	balanceFile := filepath.Join(tempDir, "enhanced_genesis_balances.txt")

	balanceLines := make([]string, 0, len(validatorKeys)+5)

	// Validator accounts
	for i, key := range validatorKeys {
		balance := fmt.Sprintf("%s=100000000ugnot", key.Address.String())
		balanceLines = append(balanceLines, balance)
		slog.Debug("validator balance", "index", i+1, "balance", balance)
	}

	// Hardcoded test accounts from official genesis
	testAccounts := []struct {
		addr    string
		balance int64
	}{
		{"g1sj0p2u3u3ptdhxxgrntw2ylgpywnxcx0hxeejf", 39902556},
		{"g1htpr653j6q356wza4zvj2usghuhmtdqjdq7gl3", 9066},
		{"g1esgv6w2ya3hrxa5rhcyummh2al8w5snv535e2f", 14667793},
		{"g1tp63hd67kcg6zcvpn87mj58z59hr6suw5gjykt", 48239067},
		{"g1hh9zcupzrcaspgs8al3chaumhkskq5d02frg48", 12232135},
	}
	for _, account := range testAccounts {
		balance := fmt.Sprintf("%s=%dugnot", account.addr, account.balance)
		balanceLines = append(balanceLines, balance)
	}

	// Extra balances (test user, etc.)
	for addr, amount := range extraBalances {
		balance := fmt.Sprintf("%s=%dugnot", addr, amount)
		balanceLines = append(balanceLines, balance)
		slog.Debug("extra balance", "balance", balance)
	}

	content := strings.Join(balanceLines, "\n") + "\n"
	if err := os.WriteFile(balanceFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write balance file: %w", err)
	}

	slog.Debug("created balance file", "accounts", len(balanceLines), "path", balanceFile)
	return balanceFile, nil
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

	// The three settable modules, keyed the way `gnogenesis params set` keys
	// them, so "vm.chain_domain" means the same thing in both.
	params := struct {
		Auth *auth.Params `json:"auth"`
		VM   *vm.Params   `json:"vm"`
		Bank *bank.Params `json:"bank"`
	}{&genState.Auth.Params, &genState.VM.Params, &genState.Bank.Params}
	root := reflect.ValueOf(&params).Elem()

	for _, o := range overrides {
		// "json", the selector `gnogenesis params set` resolves with.
		if err := applyOverride(root, "json", o); err != nil {
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

	if genState, ok := gen.AppState.(*gnoland.GnoGenesisState); ok {
		slog.Debug("genesis state", "balances", len(genState.Balances), "txs", len(genState.Txs))
	}
}
