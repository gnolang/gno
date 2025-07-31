package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodeSetup tests the basic node setup functionality
func TestNodeSetup(t *testing.T) {
	tempDir := t.TempDir()

	// ta := NewTestAssertion(t) // Not implemented yet

	// Test validator node setup
	validator := setupValidatorNode(t, tempDir, 0)
	assert.NotNil(t, validator, "validator should not be nil")
	assert.Equal(t, 0, validator.Index, "validator index should be 0")
	assert.Greater(t, validator.P2PPort, 0, "validator should have valid P2P port")
	assert.NotEmpty(t, validator.NodeID, "validator should have NodeID")
	assert.NotEmpty(t, validator.DataDir, "validator should have DataDir")

	// Test non-validator node setup
	nonValidator := setupNonValidatorNode(t, tempDir, 1)
	assert.NotNil(t, nonValidator, "non-validator should not be nil")
	assert.Equal(t, 1, nonValidator.Index, "non-validator index should be 1")
	assert.Greater(t, nonValidator.P2PPort, 0, "non-validator should have valid P2P port")
	assert.NotEmpty(t, nonValidator.NodeID, "non-validator should have NodeID")
	assert.NotEmpty(t, nonValidator.DataDir, "non-validator should have DataDir")

	// Test with testify assertions helper
	// nodes := []*Node{validator, nonValidator}
	// validateNodeSetup(t, nodes, 2) // Not implemented yet
}

// TestNodeTypeEnum tests the NodeType enum functionality
func TestNodeTypeEnum(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{ValidatorNode, "validator"},
		{NonValidatorNode, "non-validator"},
		{NodeType(999), "unknown"}, // Test unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.nodeType.String())
		})
	}
}

// TestConfigValidation tests the test configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *testCfg
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: &testCfg{
				numValidators:    2,
				numNonValidators: 3,
				targetHeight:     100,
				maxTestTime:      5 * time.Minute,
			},
			expectError: false,
		},
		{
			name: "zero validators",
			cfg: &testCfg{
				numValidators:    0,
				numNonValidators: 3,
				targetHeight:     100,
				maxTestTime:      5 * time.Minute,
			},
			expectError: true,
			errorMsg:    "at least 1 validator required",
		},
		{
			name: "negative non-validators",
			cfg: &testCfg{
				numValidators:    2,
				numNonValidators: -1,
				targetHeight:     100,
				maxTestTime:      5 * time.Minute,
			},
			expectError: true,
			errorMsg:    "non-validators must be >= 0",
		},
		{
			name: "target height too low",
			cfg: &testCfg{
				numValidators:    2,
				numNonValidators: 3,
				targetHeight:     5,
				maxTestTime:      5 * time.Minute,
			},
			expectError: true,
			errorMsg:    "target height must be >= 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTestConfig(tt.cfg)
			if tt.expectError {
				require.Error(t, err, "should have validation error")
				assert.Contains(t, err.Error(), tt.errorMsg, "error message should contain expected text")
			} else {
				require.NoError(t, err, "should not have validation error")
			}
		})
	}
}

// TestBinaryBuilding tests the gnoland binary building process
func TestBinaryBuilding(t *testing.T) {
	tempDir := t.TempDir()

	// Test binary building
	binaryPath, err := buildGnolandBinary(t, tempDir)
	require.NoError(t, err, "should build binary successfully")
	require.NotEmpty(t, binaryPath, "binary path should not be empty")

	// Verify binary exists
	_, err = os.Stat(binaryPath)
	require.NoError(t, err, "binary should exist at specified path")

	// Verify binary is executable (on Unix systems)
	info, err := os.Stat(binaryPath)
	require.NoError(t, err, "should get binary file info")
	assert.Greater(t, info.Mode()&0o111, os.FileMode(0), "binary should be executable")
}

// TestErrorHandling tests basic error handling patterns
func TestErrorHandling(t *testing.T) {
	// Test that require.NoError works correctly with nil error
	require.NotPanics(t, func() {
		err := (error)(nil)
		require.NoError(t, err, "This should not fail")
	})
}

// TestGenesisCreation tests the genesis file creation process
func TestGenesisCreation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test validator
	validator := setupValidatorNode(t, tempDir, 0)
	require.NotNil(t, validator, "should create validator")

	validators := []*Node{validator}

	// Test genesis creation
	createSharedGenesis(t, tempDir, validators)

	// Verify genesis file exists
	genesisPath := tempDir + "/shared_genesis.json"
	_, err := os.Stat(genesisPath)
	require.NoError(t, err, "shared genesis file should exist")

	// Test copying genesis to node
	copySharedGenesis(t, tempDir, validator)

	// Verify node genesis file exists
	_, err = os.Stat(validator.Genesis)
	require.NoError(t, err, "validator genesis file should exist")
}

// validateTestConfig validates the test configuration
func validateTestConfig(cfg *testCfg) error {
	if cfg.numValidators < 1 {
		return fmt.Errorf("at least 1 validator required")
	}
	if cfg.numNonValidators < 0 {
		return fmt.Errorf("non-validators must be >= 0")
	}
	if cfg.targetHeight < 10 {
		return fmt.Errorf("target height must be >= 10")
	}
	return nil
}

// BenchmarkNodeSetup benchmarks the node setup process
func BenchmarkNodeSetup(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = setupValidatorNode(b, tempDir, i)
	}
}

// Example of using testify suite for more complex testing scenarios
// This would require the suite package: import "github.com/stretchr/testify/suite"

/*
type E2ETestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *E2ETestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
}

func (suite *E2ETestSuite) TearDownTest() {
	// Cleanup is handled automatically by TestingT.TempDir()
}

func (suite *E2ETestSuite) TestMultipleValidators() {
	validators := make([]*Node, 3)
	for i := 0; i < 3; i++ {
		validators[i] = setupValidatorNode(suite.T(), suite.tempDir, i)
		suite.NotNil(validators[i])
	}
	
	// ta := NewTestAssertion(suite.T()) // Not implemented yet
	validateNodeSetup(ta, validators, 3)
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
*/