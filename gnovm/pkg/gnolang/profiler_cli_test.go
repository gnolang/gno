package gnolang

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestProfilerCLI_BasicCommands tests basic profiler CLI commands
func TestProfilerCLI_BasicCommands(t *testing.T) {
	// Create a test profile
	profile := createTestProfileForCLI()

	tests := []struct {
		name    string
		input   string
		wantOut []string // expected output contains these strings
		wantErr bool
	}{
		{
			name:  "help command",
			input: "help\nquit\n",
			wantOut: []string{
				"Commands:",
				"top",
				"list",
				"help",
				"quit",
			},
		},
		{
			name:  "top command default",
			input: "top\nquit\n",
			wantOut: []string{
				"flat",
				"cum%",
				"main.main",
			},
		},
		{
			name:  "top command with count",
			input: "top 5\nquit\n",
			wantOut: []string{
				"flat",
				"cum%",
			},
		},
		{
			name:    "list command without argument",
			input:   "list\nquit\n",
			wantErr: true,
			wantOut: []string{
				"function name required",
			},
		},
		{
			name:  "list command with function",
			input: "list main.main\nquit\n",
			wantOut: []string{
				"ROUTINE",
				"main.main",
			},
		},
		{
			name:  "show settings",
			input: "show\nquit\n",
			wantOut: []string{
				"Current settings:",
				"Focus:",
				"Hide:",
				"Cumulative:",
			},
		},
		{
			name:  "focus command",
			input: "focus main.main\ntop\nquit\n",
			wantOut: []string{
				"Focused on: main.main",
				"main.main",
			},
		},
		{
			name:  "reset command",
			input: "focus main.main\nreset\nshow\nquit\n",
			wantOut: []string{
				"Reset all focus/ignore/hide settings",
				"Focus: \n", // empty after reset
			},
		},
		{
			name:    "unknown command",
			input:   "unknown\nquit\n",
			wantErr: true,
			wantOut: []string{
				"unknown command",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI with test input/output
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			cli := NewProfilerCLI(profile, nil)
			cli.SetInput(in)
			cli.SetOutput(out)

			// Run CLI
			err := cli.Run()

			// Check output
			output := out.String()
			for _, want := range tt.wantOut {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q\nGot:\n%s", want, output)
				}
			}

			// Check error expectation
			if tt.wantErr && err == nil && !strings.Contains(output, "Error:") {
				t.Errorf("expected error but got none")
			}
		})
	}
}

// TestProfilerCLI_Filtering tests focus/ignore/hide functionality
func TestProfilerCLI_Filtering(t *testing.T) {
	profile := createTestProfileForCLI()

	tests := []struct {
		name    string
		input   string
		wantOut []string
		notWant []string // should not contain these strings
	}{
		{
			name:  "hide function",
			input: "hide helper\ntop\nquit\n",
			wantOut: []string{
				"Hiding: helper",
				"main.main",
				"main.process",
			},
			notWant: []string{
				"main.helper", // should be hidden
			},
		},
		{
			name:  "ignore function",
			input: "ignore init\ntop\nquit\n",
			wantOut: []string{
				"Ignoring: init",
				"main.main",
				"main.process",
			},
			notWant: []string{
				"main.init", // should be ignored
			},
		},
		{
			name:  "focus on function",
			input: "focus process\ntop\nquit\n",
			wantOut: []string{
				"Focused on: process",
				"main.process", // process functions should appear
				"main.main",    // main.main calls process, so it appears too
			},
			notWant: []string{
				"main.init", // should not appear when focused on process
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			cli := NewProfilerCLI(profile, nil)
			cli.SetInput(in)
			cli.SetOutput(out)

			cli.Run()

			output := out.String()
			// Split output to get only the last top command output
			parts := strings.Split(output, "(pprof) ")
			if len(parts) >= 3 {
				// Get the last command output (before quit)
				lastOutput := parts[len(parts)-2]

				for _, want := range tt.wantOut {
					if !strings.Contains(output, want) {
						t.Errorf("output missing expected string %q\nGot:\n%s", want, output)
					}
				}

				for _, notWant := range tt.notWant {
					// Check only in the last output for notWant strings
					if strings.Contains(lastOutput, notWant) {
						t.Errorf("last output contains unexpected string %q\nGot:\n%s", notWant, lastOutput)
					}
				}
			} else {
				t.Errorf("unexpected output format")
			}
		})
	}
}

// TestProfilerCLI_ToggleCommands tests toggle commands
func TestProfilerCLI_ToggleCommands(t *testing.T) {
	profile := createTestProfileForCLI()

	tests := []struct {
		name    string
		input   string
		wantOut []string
	}{
		{
			name:  "toggle cumulative",
			input: "cum\nshow\nquit\n",
			wantOut: []string{
				"Cumulative mode: false",
				"Cumulative: false",
			},
		},
		{
			name:  "toggle flat",
			input: "flat\nshow\nquit\n",
			wantOut: []string{
				"Flat mode: true",
				"Flat: true",
			},
		},
		{
			name:  "set node count",
			input: "nodecount 20\nshow\nquit\n",
			wantOut: []string{
				"Node count set to: 20",
				"Node count: 20",
			},
		},
		{
			name:  "set unit",
			input: "unit ms\nshow\nquit\n",
			wantOut: []string{
				"Unit set to: ms",
				"Unit: ms",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			cli := NewProfilerCLI(profile, nil)
			cli.SetInput(in)
			cli.SetOutput(out)

			cli.Run()

			output := out.String()
			for _, want := range tt.wantOut {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

// TestProfilerCLI_EmptyLineRepeat tests that empty line repeats last command
func TestProfilerCLI_EmptyLineRepeat(t *testing.T) {
	profile := createTestProfileForCLI()

	in := strings.NewReader("top 3\n\nquit\n")
	out := &bytes.Buffer{}

	cli := NewProfilerCLI(profile, nil)
	cli.SetInput(in)
	cli.SetOutput(out)

	cli.Run()

	output := out.String()
	// Should see "top" output twice
	topCount := strings.Count(output, "flat  flat%")
	if topCount < 2 {
		t.Errorf("expected top command to be executed at least twice, but only found %d", topCount)
	}
}

// TestProfilerCLI_Aliases tests command aliases
func TestProfilerCLI_Aliases(t *testing.T) {
	profile := createTestProfileForCLI()

	tests := []struct {
		alias   string
		fullCmd string
	}{
		{"t", "top"},
		{"l", "list"},
		{"h", "help"},
		{"q", "quit"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			var input string
			if tt.alias == "l" {
				input = tt.alias + " main.main\nquit\n"
			} else if tt.alias == "q" {
				input = tt.alias + "\n"
			} else {
				input = tt.alias + "\nquit\n"
			}

			in := strings.NewReader(input)
			out := &bytes.Buffer{}

			cli := NewProfilerCLI(profile, nil)
			cli.SetInput(in)
			cli.SetOutput(out)

			cli.Run()

			// Just verify it doesn't error out with "unknown command"
			output := out.String()
			if strings.Contains(output, "unknown command") {
				t.Errorf("alias %q not recognized", tt.alias)
			}
		})
	}
}

// Helper function to create a test profile
func createTestProfileForCLI() *Profile {
	p := &Profile{
		Type:          ProfileCPU,
		TimeNanos:     time.Now().UnixNano(),
		DurationNanos: int64(10 * time.Second),
		Samples:       []ProfileSample{},
	}

	// Add some sample data
	samples := []struct {
		functions []string
		cycles    int64
	}{
		{[]string{"main.main", "main.process", "main.helper"}, 1000},
		{[]string{"main.main", "main.process"}, 500},
		{[]string{"main.main"}, 300},
		{[]string{"main.init"}, 100},
		{[]string{"main.process", "main.helper"}, 200},
	}

	for _, s := range samples {
		locations := make([]ProfileLocation, len(s.functions))
		for i, fn := range s.functions {
			locations[i] = ProfileLocation{
				Function: fn,
				File:     "main.go",
				Line:     i*10 + 10,
			}
		}

		p.Samples = append(p.Samples, ProfileSample{
			Location: locations,
			Value:    []int64{1, s.cycles},
			Label:    make(map[string][]string),
			NumLabel: map[string][]int64{
				"cycles": {s.cycles},
			},
			SampleType: ProfileCPU,
		})
	}

	return p
}
