package cluster

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
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
