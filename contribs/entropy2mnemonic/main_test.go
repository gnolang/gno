package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		args      []string
		stdin     string
		wantErr   bool
		errMsg    string
		wantInOut string // partial string that should be in output
	}{
		{
			name:      "quiet mode with valid entropy",
			args:      []string{"-quiet"},
			stdin:     "my entropy seed with sufficient randomness from dice rolls",
			wantErr:   false,
			wantInOut: "secret acid ritual soldier laugh error reveal sample tower plug various correct leaf pass harsh tragic struggle cement legend degree auction young lamp asset\n",
		},
		{
			name:    "insufficient entropy",
			args:    []string{"-quiet"},
			stdin:   "short",
			wantErr: true,
			errMsg:  "entropy too short (5 characters)",
		},
		{
			name:    "no entropy provided",
			args:    []string{"-quiet"},
			stdin:   "",
			wantErr: true,
			errMsg:  "no entropy provided",
		},
		{
			name:      "interactive mode shows guidance",
			args:      []string{},
			stdin:     "my entropy seed with sufficient randomness from dice rolls",
			wantErr:   false,
			wantInOut: "=== ENTROPY TO MNEMONIC CONVERTER ===",
		},
		{
			name:      "command line entropy",
			args:      []string{"-quiet", "my entropy seed with sufficient randomness from dice rolls"},
			stdin:     "",
			wantErr:   false,
			wantInOut: "secret acid ritual soldier laugh error reveal sample tower plug various correct leaf pass harsh tragic struggle cement legend degree auction young lamp asset\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(tt.stdin)

			err := run(stdin, &stdout, &stderr, tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.wantInOut != "" {
				assert.Contains(t, stdout.String(), tt.wantInOut)
			}
		})
	}
}

func TestDeterministicOutput(t *testing.T) {
	t.Parallel()
	entropy := "my entropy seed with sufficient randomness from dice rolls 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 4 8 12 3 7 19 2 11 15 18 5 9 14 6 1 20 13 10 17 4 8 16"
	expectedMnemonic := "nominee spring term very amazing start rebel slogan breeze across appear hospital emotion rabbit snack please loop real inmate pet unusual any journey avocado"

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(entropy)

	err := run(stdin, &stdout, &stderr, []string{"-quiet"})
	require.NoError(t, err)

	got := strings.TrimSpace(stdout.String())
	assert.Equal(t, expectedMnemonic, got)
}

func TestFullInteractiveOutput(t *testing.T) {
	t.Parallel()
	entropy := "my entropy seed with sufficient randomness from dice rolls 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 4 8 12 3 7 19 2 11 15 18 5 9 14 6 1 20 13 10 17 4 8 16"

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(entropy)

	err := run(stdin, &stdout, &stderr, []string{})
	require.NoError(t, err)

	output := stdout.String()

	// Check key parts of the output
	expectedParts := []string{
		"=== ENTROPY TO MNEMONIC CONVERTER ===",
		"Enter your entropy (press Enter when done):",
		"Entropy received:",
		"Length: 164 characters",
		"SHA-256: 95fa677df9707da96cce5f1b80482a36f48d6073453383f65dd151bee4145e18",
		"Generated mnemonic (24 words):",
		"nominee spring term very amazing start rebel slogan breeze across appear hospital emotion rabbit snack please loop real inmate pet unusual any journey avocado",
		"IMPORTANT: Store this mnemonic securely. It cannot be recovered!",
	}

	for _, part := range expectedParts {
		assert.Contains(t, output, part, "output missing expected part: %q", part)
	}
}
