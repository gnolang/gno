// gnodebug/debug_test.go
package gnodebug

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DebugFlags
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: DebugFlags{},
		},
		{
			name:  "Single flag without value",
			input: "trace",
			expected: DebugFlags{
				"trace": "1",
			},
		},
		{
			name:  "Single flag with value",
			input: "level=5",
			expected: DebugFlags{
				"level": "5",
			},
		},
		{
			name:  "Multiple flags mixed",
			input: "trace,level=5,verbose",
			expected: DebugFlags{
				"trace":   "1",
				"level":   "5",
				"verbose": "1",
			},
		},
		{
			name:  "Empty values",
			input: ",,trace,,level=5,,",
			expected: DebugFlags{
				"trace": "1",
				"level": "5",
			},
		},
		{
			name:  "Empty key with value",
			input: "=123",
			expected: DebugFlags{
				"": "123",
			},
		},
		{
			name:  "Empty value",
			input: "key=",
			expected: DebugFlags{
				"key": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseFlags(tc.input)

			assert.Equal(t, tc.expected, got, "results should match")
		})
	}
}

func TestDebugFlags_Printf(t *testing.T) {
	originalOutput := Output
	t.Cleanup(func() {
		Output = originalOutput
	})

	tests := []struct {
		name          string
		flags         DebugFlags
		flagName      string
		format        string
		args          []any
		shouldPrint   bool
		printContains []string
	}{
		{
			name:        "No debug flag",
			flags:       DebugFlags{},
			flagName:    "trace",
			format:      "test message",
			shouldPrint: false,
		},
		{
			name:          "Enabled debug flag",
			flags:         DebugFlags{"trace": "1"},
			flagName:      "trace",
			format:        "test message",
			shouldPrint:   true,
			printContains: []string{"trace", "test message"},
		},
		{
			name:        "Disabled debug flag (not 1)",
			flags:       DebugFlags{"trace": "0"},
			flagName:    "trace",
			format:      "test message",
			shouldPrint: false,
		},
		{
			name:          "Empty flag name prints always",
			flags:         DebugFlags{},
			flagName:      "",
			format:        "test message",
			shouldPrint:   true,
			printContains: []string{"test message"},
		},
		{
			name:          "Format with arguments",
			flags:         DebugFlags{"trace": "1"},
			flagName:      "trace",
			format:        "value: %d, name: %s",
			args:          []any{42, "test"},
			shouldPrint:   true,
			printContains: []string{"value: 42, name: test"},
		},
		{
			name:          "Format without newline gets one added",
			flags:         DebugFlags{"trace": "1"},
			flagName:      "trace",
			format:        "no newline",
			shouldPrint:   true,
			printContains: []string{"no newline\n"},
		},
		{
			name:          "Format with newline keeps it",
			flags:         DebugFlags{"trace": "1"},
			flagName:      "trace",
			format:        "has newline\n",
			shouldPrint:   true,
			printContains: []string{"has newline\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Output = buf

			tc.flags.Printf(tc.flagName, tc.format, tc.args...)

			output := buf.String()
			if tc.shouldPrint {
				require.NotEmpty(t, output, "Expected output but got nothing")

				// Check the file:line format with regex
				assert.Regexp(t, `[^:]+:\d+:`, output, "Output missing file:line pattern")
				assert.Contains(t, output, tc.flagName+":", "does not contain flag name")

				for _, substr := range tc.printContains {
					assert.Contains(t, output, substr,
						"Output doesn't contain expected substring")
				}
			} else {
				assert.Empty(t, output, "Expected no output but got something")
			}
		})
	}
}
