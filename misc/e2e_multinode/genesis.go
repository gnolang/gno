package main

import (
	"fmt"
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
	"github.com/stretchr/testify/require"
)

// createSharedGenesis creates a genesis file shared by all nodes
func createSharedGenesis(t TestingT, tempDir string, validators []*Node) {
	// Read all validators' private keys
	validatorKeys := make([]*signer.FileKey, len(validators))
	for i, validator := range validators {
		validatorKeyPath := filepath.Join(validator.DataDir, "secrets", defaultValidatorKeyName)
		validatorFileKey, err := signer.LoadFileKey(validatorKeyPath)
		require.NoError(t, err, "failed to load validator key")
		validatorKeys[i] = validatorFileKey
	}

	// Create genesis document matching official configuration
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Date(2025, 7, 25, 7, 0, 0, 0, time.UTC) // Match official genesis
	gen.ChainID = "test-e2e"

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
	balanceFile := createEnhancedBalanceFile(t, tempDir, validatorKeys)
	balances, err := gnoland.LoadGenesisBalancesFile(balanceFile)
	require.NoError(t, err, "failed to load genesis balances")

	// Load packages from examples directory to create real transactions
	examplesDir := findExamplesDir(t)
	require.NotEmpty(t, examplesDir, "could not find examples directory")

	txSender := validatorKeys[0].Address // Use first validator as transaction sender

	// Deploy fee matching official genesis
	deployFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
	pkgsTxs, err := gnoland.LoadPackagesFromDir(examplesDir, txSender, deployFee)
	require.NoError(t, err, "failed to load packages from examples")
	t.Logf("Loaded package transactions, count: %d, source: examples directory", len(pkgsTxs))

	// Sign genesis transactions with the first validator key
	err = gnoland.SignGenesisTxs(pkgsTxs, validatorKeys[0].PrivKey, gen.ChainID)
	require.NoError(t, err, "failed to sign genesis transactions")

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
	err = gen.SaveAs(sharedGenesisPath)
	require.NoError(t, err, "failed to save genesis")

	t.Logf("Created shared genesis, path: %s", sharedGenesisPath)

	// Print genesis configuration for debugging
	printGenesisConfig(t, gen)
}

// createEnhancedBalanceFile creates a balance file with multiple accounts like official genesis
// XXX: generate this part
func createEnhancedBalanceFile(t TestingT, tempDir string, validatorKeys []*signer.FileKey) string {
	balanceFile := filepath.Join(tempDir, "enhanced_genesis_balances.txt")

	// Create content similar to official genesis with multiple funded accounts
	balanceLines := make([]string, 0, len(validatorKeys)+5)

	// Add validator accounts with substantial balances
	for i, key := range validatorKeys {
		balance := fmt.Sprintf("%s=100000000ugnot", key.Address.String())
		balanceLines = append(balanceLines, balance)
		t.Logf("Validator %d balance: %s", i+1, balance)
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
	err := os.WriteFile(balanceFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to write balance file")

	t.Logf("Created enhanced balance file, accounts: %d, path: %s", len(balanceLines), balanceFile)

	return balanceFile
}

// findExamplesDir locates the examples directory using gnoenv.RootDir()
func findExamplesDir(t TestingT) string {
	// Use gnoenv.RootDir() to get the root of the gno project
	gnoRoot := gnoenv.RootDir()
	examplesPath := filepath.Join(gnoRoot, "examples")

	// Verify the examples directory exists
	if info, err := os.Stat(examplesPath); err == nil && info.IsDir() {
		t.Logf("Found examples directory, path: %s", examplesPath)
		return examplesPath
	}

	t.Logf("WARNING: Examples directory not found, path: %s", examplesPath)
	return ""
}

// copySharedGenesis copies the shared genesis file to each node's directory
func copySharedGenesis(t TestingT, tempDir string, node *Node) {
	sharedGenesisPath := filepath.Join(tempDir, "shared_genesis.json")

	// Read shared genesis
	genesisData, err := os.ReadFile(sharedGenesisPath)
	require.NoError(t, err, "failed to read shared genesis")

	// Write to node's genesis location
	err = os.WriteFile(node.Genesis, genesisData, 0644)
	require.NoError(t, err, "failed to write genesis to node")

	t.Logf("Node %d genesis: %s", node.Index, node.Genesis)
}

// printGenesisConfig prints the genesis configuration without transactions for debugging
func printGenesisConfig(t TestingT, gen *bft.GenesisDoc) {
	t.Logf("📋 Genesis Configuration, chain_id: %s, genesis_time: %s",
		gen.ChainID, gen.GenesisTime.Format("2006-01-02T15:04:05Z"))

	t.Logf("Consensus Parameters, max_tx_bytes: %d, max_data_bytes: %d, max_gas: %d, time_iota_ms: %d",
		gen.ConsensusParams.Block.MaxTxBytes, gen.ConsensusParams.Block.MaxDataBytes,
		gen.ConsensusParams.Block.MaxGas, gen.ConsensusParams.Block.TimeIotaMS)

	if gen.ConsensusParams.Validator != nil {
		t.Logf("Validator params: %v", gen.ConsensusParams.Validator.PubKeyTypeURLs)
	}

	t.Logf("Validators, count: %d", len(gen.Validators))
	for i, val := range gen.Validators {
		t.Logf("Validator, index: %d, address: %s, power: %d, name: %s, pubkey: %v",
			i+1, val.Address.String(), val.Power, val.Name, val.PubKey)
	}

	// Print balance and transaction summary
	if genState, ok := gen.AppState.(*gnoland.GnoGenesisState); ok {
		t.Logf("Genesis state, balance_accounts: %d, package_transactions: %d",
			len(genState.Balances), len(genState.Txs))

		// Show sample balances and transactions
		for i, balance := range genState.Balances {
			if i >= 3 {
				t.Logf("Additional balances: %d", len(genState.Balances)-3)
				break
			}
			t.Logf("Balance: %s", balance.String())
		}

		for i, tx := range genState.Txs {
			if i >= 3 {
				t.Logf("Additional transactions: %d", len(genState.Txs)-3)
				break
			}
			t.Logf("Tx %d: %d msgs, fee: %v", i+1, len(tx.Tx.Msgs), tx.Tx.Fee)
		}
	}
}
