package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
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
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(tt.stdin)
			
			err := run(stdin, &stdout, &stderr, tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("run() error = %v, want error containing %q", err, tt.errMsg)
			}
			
			if tt.wantInOut != "" && !strings.Contains(stdout.String(), tt.wantInOut) {
				t.Errorf("run() output = %q, want output containing %q", stdout.String(), tt.wantInOut)
			}
		})
	}
}

func TestDeterministicOutput(t *testing.T) {
	entropy := "my entropy seed with sufficient randomness from dice rolls 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 4 8 12 3 7 19 2 11 15 18 5 9 14 6 1 20 13 10 17 4 8 16"
	expectedMnemonic := "nominee spring term very amazing start rebel slogan breeze across appear hospital emotion rabbit snack please loop real inmate pet unusual any journey avocado"
	
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(entropy)
	
	err := run(stdin, &stdout, &stderr, []string{"-quiet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	got := strings.TrimSpace(stdout.String())
	if got != expectedMnemonic {
		t.Errorf("got mnemonic %q, want %q", got, expectedMnemonic)
	}
}

func TestFullInteractiveOutput(t *testing.T) {
	entropy := "my entropy seed with sufficient randomness from dice rolls 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 4 8 12 3 7 19 2 11 15 18 5 9 14 6 1 20 13 10 17 4 8 16"
	
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(entropy)
	
	err := run(stdin, &stdout, &stderr, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
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
		if !strings.Contains(output, part) {
			t.Errorf("output missing expected part: %q", part)
		}
	}
}