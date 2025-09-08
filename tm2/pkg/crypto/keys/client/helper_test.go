package client

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPasswordIO is a test IO that simulates password masking
type mockPasswordIO struct {
	commands.IO
	passwords []string
	pwIndex   int
}

func (m *mockPasswordIO) GetPassword(prompt string, insecure bool) (string, error) {
	if m.pwIndex >= len(m.passwords) {
		return "", fmt.Errorf("no more passwords in mock (prompt: %q, index: %d/%d)", prompt, m.pwIndex, len(m.passwords))
	}
	pw := m.passwords[m.pwIndex]
	m.pwIndex++
	// fmt.Printf("DEBUG: GetPassword[%d] prompt=%q, returning: %q\n", m.pwIndex-1, prompt, pw)
	return pw, nil
}

func TestGenerateMnemonicWithCustomEntropy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		entropy       string
		confirmations []bool
		shouldError   bool
		errorContains string
	}{
		{
			name:          "good entropy with long input",
			entropy:       "this is some really good entropy with lots of randomness 12345!@#$%",
			confirmations: []bool{true},
			shouldError:   false,
		},
		{
			name:          "minimum acceptable entropy",
			entropy:       strings.Repeat("a", MinEntropyChars), // exactly MinEntropyChars
			confirmations: []bool{true},
			shouldError:   false,
		},
		{
			name:          "very short entropy",
			entropy:       "tiny",
			confirmations: []bool{},
			shouldError:   true,
			errorContains: "too short",
		},
		{
			name:          "generation rejected",
			entropy:       "good enough entropy here with sufficient length",
			confirmations: []bool{false},
			shouldError:   true,
			errorContains: "cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test IO with configured input for entropy and confirmation
			io := commands.NewTestIO()

			// Set up input stream with entropy and confirmation response
			inputData := tt.entropy + "\n"
			if len(tt.confirmations) > 0 {
				if tt.confirmations[0] {
					inputData += "y\n"
				} else {
					inputData += "n\n"
				}
			}
			io.SetIn(strings.NewReader(inputData))

			// Capture output
			var outBuf, errBuf bytes.Buffer
			io.SetOut(commands.WriteNopCloser(&outBuf))
			io.SetErr(commands.WriteNopCloser(&errBuf))

			mnemonic, err := GenerateMnemonicWithCustomEntropy(io, false)

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)

			// Verify the mnemonic is valid
			assert.True(t, bip39.IsMnemonicValid(mnemonic), "generated mnemonic is not valid BIP39")

			// Verify determinism - same entropy should produce same mnemonic
			io2 := commands.NewTestIO()
			inputData2 := tt.entropy + "\n"
			if len(tt.confirmations) > 0 && tt.confirmations[0] {
				inputData2 += "y\n"
			}
			io2.SetIn(strings.NewReader(inputData2))
			io2.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
			io2.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

			mnemonic2, err := GenerateMnemonicWithCustomEntropy(io2, false)
			require.NoError(t, err, "unexpected error on second generation")

			assert.Equal(t, mnemonic, mnemonic2, "same entropy produced different mnemonics")
		})
	}
}

func TestDeterministicMnemonicGeneration(t *testing.T) {
	t.Parallel()
	// Test that the same entropy always produces the same mnemonic
	testEntropy := "test entropy for deterministic generation 42"

	// Generate expected result
	hashedEntropy := sha256.Sum256([]byte(testEntropy))
	expectedMnemonic, err := bip39.NewMnemonic(hashedEntropy[:])
	require.NoError(t, err, "failed to generate expected mnemonic")

	// Create test IO with entropy and confirmation
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(testEntropy + "\ny\n"))
	io.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
	io.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

	actualMnemonic, err := GenerateMnemonicWithCustomEntropy(io, false)
	require.NoError(t, err, "failed to generate mnemonic")

	assert.Equal(t, expectedMnemonic, actualMnemonic, "mnemonic doesn't match expected deterministic result")
}

func TestMaskedEntropyInput(t *testing.T) {
	t.Parallel()
	// Test that masked input works correctly using a mock IO
	testEntropy := "this is a test entropy that should work when masked"

	// Create base test IO for output
	baseIO := commands.NewTestIO()
	baseIO.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
	baseIO.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

	// Create mock IO that simulates password input
	mockIO := &mockPasswordIO{
		IO:        baseIO,
		passwords: []string{testEntropy},
	}
	// For confirmation prompt
	mockIO.SetIn(strings.NewReader("y\n"))

	// Generate with masked = true using our mock
	mnemonic, err := GenerateMnemonicWithCustomEntropy(mockIO, true)
	require.NoError(t, err, "failed to generate mnemonic with masked input")
	assert.True(t, bip39.IsMnemonicValid(mnemonic), "invalid mnemonic generated with masked input")

	// Verify that same entropy produces same result whether masked or not
	io2 := commands.NewTestIO()
	io2.SetIn(strings.NewReader(testEntropy + "\ny\n"))
	io2.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
	io2.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

	mnemonic2, err := GenerateMnemonicWithCustomEntropy(io2, false)
	require.NoError(t, err, "failed to generate mnemonic without masking")
	assert.Equal(t, mnemonic, mnemonic2, "masked and unmasked entropy produced different mnemonics")
}

func TestEntropyHashingConsistency(t *testing.T) {
	t.Parallel()
	// Test specific entropy inputs to ensure consistent results
	testCases := []struct {
		input string
	}{
		{
			input: "42 dice rolls: 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16",
		},
	}

	for _, tc := range testCases {
		// First generation
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(tc.input + "\ny\n"))
		io.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
		io.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

		mnemonic, err := GenerateMnemonicWithCustomEntropy(io, false)
		require.NoError(t, err, "failed to generate mnemonic for input %q", tc.input)

		// Verify it's a valid mnemonic
		assert.True(t, bip39.IsMnemonicValid(mnemonic), "invalid mnemonic generated for input %q", tc.input)

		// Test consistency - run it again
		io2 := commands.NewTestIO()
		io2.SetIn(strings.NewReader(tc.input + "\ny\n"))
		io2.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
		io2.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

		mnemonic2, err := GenerateMnemonicWithCustomEntropy(io2, false)
		assert.NoError(t, err, "failed to generate mnemonic on second try")
		assert.Equal(t, mnemonic, mnemonic2, "inconsistent mnemonic generation for input %q", tc.input)
	}
}
