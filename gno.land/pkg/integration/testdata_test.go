package integration

import (
	"os"
	"strconv"
	"testing"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
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

	mode := commandKindTesting
	if debugTs {
		mode = commandKindInMemory
	}

	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		env.Values[envKeyExecCommand] = mode
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		return nil
	}

	if debugTs {
		testscript.RunT(tSeqShim{t}, p)
	} else {
		testscript.Run(t, p)
	}
}
