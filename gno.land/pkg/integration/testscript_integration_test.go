package integration

import (
	"os"
	"strconv"
	"strings"
	"testing"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var debugTs = false

func init() { debugTs, _ = strconv.ParseBool(os.Getenv("DEBUG_TS")) }

func TestTestdata(t *testing.T) {
	t.Parallel()

	p := gno_integration.NewTestingParams(t, "testdata")

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript
	err := SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	if debugTs {
		RunInMemoryTestscripts(t, p)
	} else {
		testscript.Run(t, p)
	}

	// Run testscript
	// XXX: We have to use seqshim for now as tests don't run well in parallel

	// RunSeqShimTestscripts(t, p)
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
