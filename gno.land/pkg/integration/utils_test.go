package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnquote(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input      string
		Expected   []string
		ShouldFail bool
	}{
		{"", []string{""}, false},
		{"g", []string{"g"}, false},
		{"Hello Gno", []string{"Hello", "Gno"}, false},
		{`"Hello" "Gno"`, []string{"Hello", "Gno"}, false},
		{`"Hel lo" "Gno"`, []string{"Hel lo", "Gno"}, false},
		{`"H e l l o\n" \nGno`, []string{"H e l l o\n", "\\nGno"}, false},
		{`"Hel\n"\nlo "  ""G"n"o"`, []string{"Hel\n\\nlo", "  Gno"}, false},
		{`"He said, \"Hello\"" "Gno"`, []string{`He said, "Hello"`, "Gno"}, false},
		{`"\n \t" \n\t`, []string{"\n \t", "\\n\\t"}, false},
		{`"Hel\\n"\t\\nlo " ""\\nGno"`, []string{"Hel\\n\\t\\\\nlo", " \\nGno"}, false},
		// errors:
		{`"Hello Gno`, []string{}, true},    // unfinished quote
		{`"Hello\e Gno"`, []string{}, true}, // unhandled escape sequence
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()

			// split by whitespace to simulate command-line arguments
			args := strings.Split(tc.Input, " ")
			unquotedArgs, err := unquote(args)
			if tc.ShouldFail {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, unquotedArgs)
		})
	}
}

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		// Basic cases
		{"plain", `--foo=bar --bar=42`, []string{"--foo=bar", "--bar=42"}, false},
		{"double_quotes", `--foo="bar baz" --bar=42`, []string{"--foo=bar baz", "--bar=42"}, false},
		{"single_quotes", `--foo='bar baz' --bar=42`, []string{"--foo=bar baz", "--bar=42"}, false},
		{"nested_single_in_double", `--foo="bar 'baz'" --bar=42`, []string{"--foo=bar 'baz'", "--bar=42"}, false},
		{"nested_double_in_single", `--foo='bar "baz"' --bar=42`, []string{"--foo=bar \"baz\"", "--bar=42"}, false},

		// Escaping in double quotes
		{"escaped_quote_in_double", `--foo="bar \"baz\" test"`, []string{"--foo=bar \"baz\" test"}, false},
		{"escaped_backslash_in_double", `--foo="bar \\ baz"`, []string{"--foo=bar \\ baz"}, false},

		// Escaping in single quotes
		{"escaped_single_in_single", `--foo='bar \'baz'\'`, []string{`--foo=bar \baz\`}, false},

		// Escaping outside quotes (not handled, so literal)
		{"escaped_space_outside_quotes", `foo\ bar`, []string{`foo bar`}, false},

		// Unclosed quotes (should error)
		{"unclosed_double", `--foo="bar baz`, nil, true},
		{"unclosed_single", `--foo='bar baz`, nil, true},

		// Quote mismatch (should not error, treated as literal)
		{"quote_mismatch", `--foo="bar 'baz"`, []string{"--foo=bar 'baz"}, false},

		// Multiple spaces
		{"extra_spaces", `    --foo=bar     --bar=42  `, []string{"--foo=bar", "--bar=42"}, false},

		// Empty input
		{"empty", ``, []string{}, false},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result, err := splitArgs(tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
