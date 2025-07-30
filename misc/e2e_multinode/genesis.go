package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// createSharedGenesis creates a genesis file shared by all nodes
func createSharedGenesis(tempDir string, validators []*Node) error {
	// Read all validators' private keys
	validatorKeys := make([]*signer.FileKey, len(validators))
	for i, validator := range validators {
		validatorKeyPath := filepath.Join(validator.DataDir, "secrets", defaultValidatorKeyName)
		validatorFileKey, err := signer.LoadFileKey(validatorKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load validator key: %w", err)
		}
		validatorKeys[i] = validatorFileKey
	}

	// Create genesis document matching official configuration
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Date(2025, 7, 25, 7, 0, 0, 0, time.UTC) // Match official genesis
	gen.ChainID = "test7-determinism"                              // Similar to official chain ID

	// Match official consensus parameters
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:   1_000_000,     // 1MB - match official
			MaxDataBytes: 2_000_000,     // 2MB - match official
			MaxGas:       3_000_000_000, // 3B gas - match official
			TimeIotaMS:   100,           // 100ms - match official
		},
	}

	// Set up validators with realistic power (matching official genesis power=1)
	gen.Validators = make([]bft.GenesisValidator, len(validators))
	for i, validatorKey := range validatorKeys {
		gen.Validators[i] = bft.GenesisValidator{
			Address: validatorKey.Address,
			PubKey:  validatorKey.PubKey,
			Power:   1, // Realistic power matching official genesis
			Name:    fmt.Sprintf("testval%d", i+1),
		}
	}

	// Create enhanced balance file with multiple accounts (like official genesis)
	balanceFile := createEnhancedBalanceFile(tempDir, validatorKeys)
	balances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis balances: %w", err)
	}

	// Load packages from examples directory to create real transactions
	examplesDir := findExamplesDir()
	if examplesDir == "" {
		return fmt.Errorf("could not find examples directory")
	}

	txSender := validatorKeys[0].Address // Use first validator as transaction sender

	// Deploy fee matching official genesis
	deployFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	pkgsTxs, err := gnoland.LoadPackagesFromDir(examplesDir, txSender, deployFee)
	if err != nil {
		return fmt.Errorf("failed to load packages from examples: %w", err)
	}
	slog.Info("Loaded package transactions", "count", len(pkgsTxs), "source", "examples directory")

	// Sign genesis transactions with the first validator key
	err = gnoland.SignGenesisTxs(pkgsTxs, validatorKeys[0].PrivKey, gen.ChainID)
	if err != nil {
		return fmt.Errorf("failed to sign genesis transactions: %w", err)
	}

	// Ensure deployer has sufficient balance for all transactions
	deployerBalance := int64(len(pkgsTxs)) * 50_000_000 // ~50 GNOT per tx (match official)
	balances.Set(txSender, std.NewCoins(std.NewCoin("ugnot", deployerBalance)))

	// Create genesis state with real transactions
	defaultGenState := gnoland.DefaultGenState()
	defaultGenState.Balances = balances.List()
	defaultGenState.Txs = pkgsTxs // Include real package deployment transactions
	gen.AppState = defaultGenState

	// Write shared genesis to a common location
	sharedGenesisPath := filepath.Join(tempDir, "shared_genesis.json")
	if err := gen.SaveAs(sharedGenesisPath); err != nil {
		return fmt.Errorf("failed to save genesis: %w", err)
	}

	slog.Info("Created shared genesis", "path", sharedGenesisPath)

	// Print genesis configuration for debugging
	printGenesisConfig(gen)

	return nil
}

// createEnhancedBalanceFile creates a balance file with multiple accounts like official genesis
func createEnhancedBalanceFile(tempDir string, validatorKeys []*signer.FileKey) string {
	balanceFile := filepath.Join(tempDir, "enhanced_genesis_balances.txt")

	// Create content similar to official genesis with multiple funded accounts
	var balanceLines []string

	// Add validator accounts with substantial balances
	for i, key := range validatorKeys {
		balance := fmt.Sprintf("%s=100000000ugnot", key.Address.String())
		balanceLines = append(balanceLines, balance)
		slog.Debug("Added validator balance", "validator", i+1, "balance", balance)
	}

	// Add additional test accounts with various balances (using actual addresses from official genesis)
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

	// Write balance file
	content := strings.Join(balanceLines, "\n") + "\n"
	if err := os.WriteFile(balanceFile, []byte(content), 0644); err != nil {
		slog.Error("Failed to write balance file", "error", err)
		os.Exit(1)
	}

	slog.Info("Created enhanced balance file", "accounts", len(balanceLines), "path", balanceFile)

	return balanceFile
}

// findExamplesDir locates the examples directory using gnoenv.RootDir()
func findExamplesDir() string {
	// Use gnoenv.RootDir() to get the root of the gno project
	gnoRoot := gnoenv.RootDir()
	examplesPath := filepath.Join(gnoRoot, "examples")

	// Verify the examples directory exists
	if info, err := os.Stat(examplesPath); err == nil && info.IsDir() {
		slog.Info("Found examples directory", "path", examplesPath)
		return examplesPath
	}

	slog.Warn("Examples directory not found", "path", examplesPath)
	return ""
}

// copySharedGenesis copies the shared genesis file to each node's directory
func copySharedGenesis(tempDir string, node *Node) error {
	sharedGenesisPath := filepath.Join(tempDir, "shared_genesis.json")

	// Read shared genesis
	genesisData, err := os.ReadFile(sharedGenesisPath)
	if err != nil {
		return fmt.Errorf("failed to read shared genesis: %w", err)
	}

	// Write to node's genesis location
	if err := os.WriteFile(node.Genesis, genesisData, 0644); err != nil {
		return fmt.Errorf("failed to write genesis to node: %w", err)
	}

	slog.Debug("Copied shared genesis", "node_index", node.Index, "path", node.Genesis)

	return nil
}

// printGenesisConfig prints the genesis configuration without transactions for debugging
func printGenesisConfig(gen *bft.GenesisDoc) {
	slog.Info("📋 Genesis Configuration",
		"chain_id", gen.ChainID,
		"genesis_time", gen.GenesisTime.Format("2006-01-02T15:04:05Z"))

	slog.Info("Consensus Parameters",
		"max_tx_bytes", gen.ConsensusParams.Block.MaxTxBytes,
		"max_data_bytes", gen.ConsensusParams.Block.MaxDataBytes,
		"max_gas", gen.ConsensusParams.Block.MaxGas,
		"time_iota_ms", gen.ConsensusParams.Block.TimeIotaMS)

	if gen.ConsensusParams.Validator != nil {
		slog.Debug("Validator parameters",
			"pubkey_type_urls", gen.ConsensusParams.Validator.PubKeyTypeURLs)
	}

	slog.Info("Validators", "count", len(gen.Validators))
	for i, val := range gen.Validators {
		slog.Info("Validator",
			"index", i+1,
			"address", val.Address.String(),
			"power", val.Power,
			"name", val.Name,
			"pubkey", val.PubKey)

	}

	// Print balance and transaction summary
	if genState, ok := gen.AppState.(*gnoland.GnoGenesisState); ok {
		slog.Info("Genesis state",
			"balance_accounts", len(genState.Balances),
			"package_transactions", len(genState.Txs))

		// Show first few balances as examples
		slog.Debug("Sample balances")
		for i, balance := range genState.Balances {
			if i >= 3 { // Show only first 3 balances
				slog.Debug("Additional balances", "remaining", len(genState.Balances)-3)
				break
			}
			slog.Debug("Balance", "account", balance.String())
		}

		// Show first few transaction types as examples
		slog.Debug("Sample transactions")  
		for i, tx := range genState.Txs {
			if i >= 3 { // Show only first 3 transactions
				slog.Debug("Additional transactions", "remaining", len(genState.Txs)-3)
				break
			}
			slog.Debug("Transaction",
				"index", i+1,
				"msgs", len(tx.Tx.Msgs),
				"fee", fmt.Sprintf("%v", tx.Tx.Fee))
		}
	}
}
