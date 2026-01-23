package integration

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// listTestdataFiles returns all .txtar files in testdata, excluding the bench subdirectory.
func listTestdataFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir("testdata", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the bench subdirectory entirely
		if d.IsDir() && d.Name() == "bench" {
			return filepath.SkipDir
		}

		// Only include .txtar files
		if !d.IsDir() && filepath.Ext(path) == ".txtar" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("failed to list testdata files: %v", err)
	}

	return files
}

func TestTestdata(t *testing.T) {
	t.Parallel()

	flagInMemoryTS, _ := strconv.ParseBool(os.Getenv("INMEMORY_TS"))
	flagSeqTS, _ := strconv.ParseBool(os.Getenv("SEQ_TS"))

	// List files explicitly, excluding the bench directory
	files := listTestdataFiles(t)
	if len(files) == 0 {
		t.Skip("no testdata files found")
	}

	p := gno_integration.NewTestingParams(t, "")
	p.Files = files

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript
	err := SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	mode := commandKindTesting
	if flagInMemoryTS {
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

	if flagInMemoryTS || flagSeqTS {
		testscript.RunT(tSeqShim{t}, p)
	} else {
		testscript.Run(t, p)
	}
}

type tSeqShim struct{ *testing.T }

// noop Parallel method allow us to run test sequentially
func (tSeqShim) Parallel() {}

func (t tSeqShim) Run(name string, f func(testscript.T)) {
	t.T.Run(name, func(t *testing.T) {
		f(tSeqShim{t})
	})
}

func (t tSeqShim) Verbose() bool {
	return testing.Verbose()
}
