package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestdata(t *testing.T) {
	t.Parallel()

	RunGnolandTestscripts(t, "testdata")
}

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
