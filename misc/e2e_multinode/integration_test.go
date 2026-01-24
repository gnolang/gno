package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestFullE2EIntegration runs a complete end-to-end integration.
func TestFullE2EIntegration(t *testing.T) {
	// Create test configuration
	cfg := &testCfg{
		numValidators:    2,
		numNonValidators: 1,
		numTransactions:  3,
		targetHeight:     50, // Lower height for faster test
		maxTestTime:      5 * time.Minute,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.maxTestTime)
	defer cancel()

	t.Logf("Configuration: %d validators, %d non-validators, target height: %d",
		cfg.numValidators, cfg.numNonValidators, cfg.targetHeight)

	runDeterminismTest(t, ctx, cfg)
}

// TestIntegrationWithMultipleValidators tests with multiple validators
func TestIntegrationWithMultipleValidators(t *testing.T) {
	cfg := &testCfg{
		numValidators:    3,
		numNonValidators: 2,
		numTransactions:  5,
		targetHeight:     30,
		maxTestTime:      4 * time.Minute,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.maxTestTime)
	defer cancel()

	runDeterminismTest(t, ctx, cfg)
}

// TestIntegrationSetupOnly tests only the setup phase without running nodes
func TestIntegrationSetupOnly(t *testing.T) {
	cfg := &testCfg{
		numValidators:    1,
		numNonValidators: 1,
		numTransactions:  1,
		targetHeight:     20,
		maxTestTime:      3 * time.Minute,
	}

	tempDir := t.TempDir()

	// Build gnoland binary
	binaryPath, err := buildGnolandBinary(t, tempDir)
	require.NoError(t, err, "failed to build gnoland binary")
	require.NotEmpty(t, binaryPath)

	// Create validator nodes
	validators := make([]*Node, cfg.numValidators)
	for i := 0; i < cfg.numValidators; i++ {
		validators[i] = setupValidatorNode(t, tempDir, i)
		t.Logf("Created validator %d - ID: %s, Port: %d", i+1, validators[i].NodeID, validators[i].P2PPort)
	}

	// Create non-validator nodes
	nonValidators := make([]*Node, cfg.numNonValidators)
	for i := 0; i < cfg.numNonValidators; i++ {
		nonValidators[i] = setupNonValidatorNode(t, tempDir, cfg.numValidators+i)
		t.Logf("Created non-validator %d - ID: %s, Port: %d", i+1, nonValidators[i].NodeID, nonValidators[i].P2PPort)
	}

	// Combine all nodes
	nodes := append(validators, nonValidators...)

	// Create shared genesis
	t.Logf("Genesis will include %d validators", cfg.numValidators)
	createSharedGenesis(t, tempDir, validators)

	// Copy genesis to all nodes
	for _, node := range nodes {
		copySharedGenesis(t, tempDir, node)
	}

	// Configure P2P topology
	configureP2PTopology(t, validators, nonValidators)

	// Configure consensus settings
	for _, node := range nodes {
		configureConsensusForSync(t, node)
	}

	// Print node configurations
	printNodeConfigurations(t, nodes, cfg)
}

// TestConfigValidationIntegration tests configuration validation
func TestConfigValidationIntegration(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         *testCfg
		expectPanic bool
	}{
		{
			name: "valid config",
			cfg: &testCfg{
				numValidators:    1,
				numNonValidators: 1,
				targetHeight:     15,
				maxTestTime:      2 * time.Minute,
			},
			expectPanic: false,
		},
		{
			name: "zero validators should cause panic",
			cfg: &testCfg{
				numValidators:    0, // Invalid - should cause panic
				numNonValidators: 1,
				targetHeight:     15,
				maxTestTime:      2 * time.Minute,
			},
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.cfg.maxTestTime)
			defer cancel()

			if tc.expectPanic {
				require.Panics(t, func() {
					runDeterminismTest(t, ctx, tc.cfg)
				}, "Expected configuration validation to cause panic")
			} else {
				require.NotPanics(t, func() {
					runDeterminismTest(t, ctx, tc.cfg)
				}, "Valid configuration should not cause panic")
			}
		})
	}
}

// TestComponentIntegration tests individual components work together
func TestComponentIntegration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("binary building", func(t *testing.T) {
		binaryPath, err := buildGnolandBinary(t, tempDir)
		require.NoError(t, err)
		require.NotEmpty(t, binaryPath)
		t.Logf("Built binary: %s", binaryPath)
	})

	t.Run("node setup and configuration", func(t *testing.T) {
		// Test validator setup
		validator := setupValidatorNode(t, tempDir, 0)
		require.NotNil(t, validator)
		require.Equal(t, 0, validator.Index)
		require.NotEmpty(t, validator.NodeID)

		// Test non-validator setup
		nonValidator := setupNonValidatorNode(t, tempDir, 1)
		require.NotNil(t, nonValidator)
		require.Equal(t, 1, nonValidator.Index)
		require.NotEmpty(t, nonValidator.NodeID)

		t.Logf("Validator: %+v", validator)
		t.Logf("Non-validator: %+v", nonValidator)
	})

	t.Run("genesis creation", func(t *testing.T) {
		validator := setupValidatorNode(t, tempDir, 0)
		validators := []*Node{validator}

		// Test genesis creation
		require.NotPanics(t, func() {
			createSharedGenesis(t, tempDir, validators)
		})

		// Test genesis copying
		require.NotPanics(t, func() {
			copySharedGenesis(t, tempDir, validator)
		})

		t.Log("Genesis creation and copying completed successfully")
	})

	t.Run("p2p configuration", func(t *testing.T) {
		validator1 := setupValidatorNode(t, tempDir, 0)
		validator2 := setupValidatorNode(t, tempDir, 1)
		nonValidator := setupNonValidatorNode(t, tempDir, 2)

		validators := []*Node{validator1, validator2}
		nonValidators := []*Node{nonValidator}

		require.NotPanics(t, func() {
			configureP2PTopology(t, validators, nonValidators)
		})

		t.Log("P2P configuration completed successfully")
	})

	t.Run("consensus configuration", func(t *testing.T) {
		validator := setupValidatorNode(t, tempDir, 0)

		require.NotPanics(t, func() {
			configureConsensusForSync(t, validator)
		})

		t.Log("Consensus configuration completed successfully")
	})
}
