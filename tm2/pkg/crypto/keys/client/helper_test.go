package client

import (
	"bytes"
	"crypto/sha256"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
)

// testMockIO implements commands.IO for testing
type testMockIO struct {
	inputs       []string
	inputIndex   int
	confirms     []bool
	confirmIndex int
	output       bytes.Buffer
}

func (m *testMockIO) In() io.Reader                                  { return &m.output }
func (m *testMockIO) Out() io.WriteCloser                            { return &testWriteCloser{&m.output} }
func (m *testMockIO) Err() io.WriteCloser                            { return &testWriteCloser{&m.output} }
func (m *testMockIO) SetIn(in io.Reader)                             {}
func (m *testMockIO) SetOut(out io.WriteCloser)                      {}
func (m *testMockIO) SetErr(err io.WriteCloser)                      {}
func (m *testMockIO) Println(args ...interface{})                    {}
func (m *testMockIO) Printf(format string, args ...interface{})      {}
func (m *testMockIO) Printfln(format string, args ...interface{})    {}
func (m *testMockIO) ErrPrintln(args ...interface{})                 {}
func (m *testMockIO) ErrPrintfln(format string, args ...interface{}) {}

type testWriteCloser struct {
	*bytes.Buffer
}

func (w *testWriteCloser) Close() error { return nil }

func (m *testMockIO) GetString(prompt string) (string, error) {
	if m.inputIndex >= len(m.inputs) {
		return "", nil
	}
	result := m.inputs[m.inputIndex]
	m.inputIndex++
	return result, nil
}

func (m *testMockIO) GetConfirmation(prompt string) (bool, error) {
	if m.confirmIndex >= len(m.confirms) {
		return false, nil
	}
	result := m.confirms[m.confirmIndex]
	m.confirmIndex++
	return result, nil
}

func (m *testMockIO) GetCheckPassword(prompts [2]string, insecure bool) (string, error) {
	return "password", nil
}

func (m *testMockIO) GetPassword(prompt string, insecure bool) (string, error) {
	return "password", nil
}

func TestGenerateMnemonicWithCustomEntropy(t *testing.T) {
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
			entropy:       "exactly 27 characters here!", // exactly 27 chars
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
			mockIO := &testMockIO{
				inputs:   []string{tt.entropy},
				confirms: tt.confirmations,
			}

			mnemonic, err := GenerateMnemonicWithCustomEntropy(mockIO)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the mnemonic is valid
			if !bip39.IsMnemonicValid(mnemonic) {
				t.Errorf("generated mnemonic is not valid BIP39")
			}

			// Verify determinism - same entropy should produce same mnemonic
			mockIO2 := &testMockIO{
				inputs:   []string{tt.entropy},
				confirms: tt.confirmations,
			}

			mnemonic2, err := GenerateMnemonicWithCustomEntropy(mockIO2)
			if err != nil {
				t.Errorf("unexpected error on second generation: %v", err)
				return
			}

			if mnemonic != mnemonic2 {
				t.Errorf("same entropy produced different mnemonics: %q vs %q", mnemonic, mnemonic2)
			}
		})
	}
}

func TestDeterministicMnemonicGeneration(t *testing.T) {
	// Test that the same entropy always produces the same mnemonic
	testEntropy := "test entropy for deterministic generation 42"

	// Generate expected result
	hashedEntropy := sha256.Sum256([]byte(testEntropy))
	expectedMnemonic, err := bip39.NewMnemonic(hashedEntropy[:])
	if err != nil {
		t.Fatalf("failed to generate expected mnemonic: %v", err)
	}

	mockIO := &testMockIO{
		inputs:   []string{testEntropy},
		confirms: []bool{true},
	}

	actualMnemonic, err := GenerateMnemonicWithCustomEntropy(mockIO)
	if err != nil {
		t.Fatalf("failed to generate mnemonic: %v", err)
	}

	if actualMnemonic != expectedMnemonic {
		t.Errorf("mnemonic doesn't match expected deterministic result")
		t.Errorf("expected: %s", expectedMnemonic)
		t.Errorf("actual:   %s", actualMnemonic)
	}
}

func TestEntropyHashingConsistency(t *testing.T) {
	// Test specific entropy inputs to ensure consistent results
	testCases := []struct {
		input string
	}{
		{
			input: "42 dice rolls: 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16",
		},
	}

	for _, tc := range testCases {
		mockIO := &testMockIO{
			inputs:   []string{tc.input},
			confirms: []bool{true},
		}

		mnemonic, err := GenerateMnemonicWithCustomEntropy(mockIO)
		if err != nil {
			t.Errorf("failed to generate mnemonic for input %q: %v", tc.input, err)
			continue
		}

		// Verify it's a valid mnemonic
		if !bip39.IsMnemonicValid(mnemonic) {
			t.Errorf("invalid mnemonic generated for input %q", tc.input)
		}

		// Test consistency - run it again
		mockIO2 := &testMockIO{
			inputs:   []string{tc.input},
			confirms: []bool{true},
		}

		mnemonic2, err := GenerateMnemonicWithCustomEntropy(mockIO2)
		if err != nil {
			t.Errorf("failed to generate mnemonic on second try: %v", err)
			continue
		}

		if mnemonic != mnemonic2 {
			t.Errorf("inconsistent mnemonic generation for input %q", tc.input)
		}
	}
}
