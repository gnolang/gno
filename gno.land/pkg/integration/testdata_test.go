package integration

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

var (
	flagInMemoryTS = flag.Bool("ts-inmemory", false,
		"Make txtar run in memory, without running any side process, forcing test to run sequentially")
	flagSeqTS = flag.Bool("ts-seq", false,
		"forcing tests to run sequentially")
)

func init() {
	if t, err := strconv.ParseBool(os.Getenv("INMEMORY_TS")); err == nil {
		*flagInMemoryTS = t
	}

	if t, err := strconv.ParseBool(os.Getenv("SEQ_TS")); err == nil {
		*flagSeqTS = t
	}
}

// Specific file opts
type testFileOpts struct {
	FileOpts

	noParallel bool
	skip       bool
}

func newDefaultTestFileOpts() *testFileOpts {
	return &testFileOpts{FileOpts: NewDefaultFileOpts()}
}

func newTestFlagsOpts(fs *flag.FlagSet) *testFileOpts {
	topts := newDefaultTestFileOpts()

	// File specific opts
	fs.BoolVar(&topts.NoFormat, "no-fmt", topts.NoFormat,
		"Disable formatting of Gno files in this test file. Use this to preserve original formatting")
	fs.DurationVar(&topts.Timeout, "timeout", topts.Timeout,
		"Set a custom timeout for this test file")

	// Test specific opts
	fs.BoolVar(&topts.noParallel, "no-parallel", topts.noParallel,
		"Disable parallel execution for this test file. This is handled natively using t.Parallel")
	fs.BoolVar(&topts.skip, "skip", topts.skip,
		"Skip this test file entirely")

	return topts
}

func TestTestdata(t *testing.T) {
	t.Parallel()

	p := gno_integration.NewTestingParams(t, "testdata")
	p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	p.RequireUniqueNames = true

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript
	err := SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	// Setup setopts
	const command = "setopts"
	// Setup noop command, as this command is parsed before test is actually running
	p.Cmds[command] = func(ts *testscript.TestScript, neg bool, args []string) {}

	mf, err := ParseDirFlags(".txtar", command, p.Dir, newTestFlagsOpts)
	require.NoError(t, err)

	mode := CommandKindTesting
	if *flagInMemoryTS {
		mode = CommandKindInMemory
	}

	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		// set command kind based on mode
		SetEnvCommandKind(env, mode)

		// Save file opts if any
		name := strings.TrimPrefix(filepath.Base(env.WorkDir), "script-")
		if topts, ok := mf[name]; ok {
			SetEnvFileOpts(env, topts.FileOpts)
		}

		// Override origin setup
		if origSetup == nil {
			return nil
		}

		return origSetup(env)
	}

	ts := tShim{
		T:        t,
		forceSeq: *flagInMemoryTS || *flagSeqTS,
		mflags:   mf,
	}
	testscript.RunT(ts, p)
}

type tShim struct {
	*testing.T
	forceSeq bool // force sequential tests
	mflags   MapFlags[*testFileOpts]
}

func (ts tShim) getOpts() *testFileOpts {
	name := path.Base(ts.Name())
	if topts, ok := ts.mflags[name]; ok {
		return topts
	}

	return newDefaultTestFileOpts()
}

func (ts tShim) Parallel() {
	if ts.forceSeq {
		return
	}

	opts := ts.getOpts()
	if !opts.noParallel {
		ts.T.Parallel()
	} else {
		ts.Log("parallel testing is disable for this test")
	}
}

func (ts tShim) Run(name string, f func(testscript.T)) {
	ts.T.Run(name, func(t *testing.T) {
		ts := tShim{t, ts.forceSeq, ts.mflags}
		opts := ts.getOpts()
		if opts.skip {
			t.Skipf("skipping %q due to `-skip` flags", name)
		}

		f(ts)
	})
}

func (ts tShim) Verbose() bool {
	return testing.Verbose()
}

type MapFlags[T any] map[string]T

func ParseDirFlags[T any](ext, prefix, dir string, newT func(*flag.FlagSet) T) (MapFlags[T], error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading dir %q: %w", dir, err)
	}

	mf := MapFlags[T]{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ext) {
			continue
		}

		fpath := filepath.Join(dir, entry.Name())
		f, err := os.Open(fpath)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %q: %w", fpath, err)
		}

		args, err := captureTopLevelLineArgs(f, prefix)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("invalid file flags %q: %w", entry.Name(), err)
		}

		fs := flag.NewFlagSet(prefix, flag.ContinueOnError)
		t := newT(fs)
		if err := fs.Parse(args); err != nil {
			return nil, fmt.Errorf("unable to parse flags in %q: %w", entry.Name(), err)
		}

		name := strings.TrimSuffix(entry.Name(), ext)
		mf[name] = t
	}

	return mf, nil
}

// ParseTopLevelFlags parses top-level lines starting with # <prefix> <flags>.
func captureTopLevelLineArgs(r io.Reader, prefix string) ([]string, error) {
	scanner := bufio.NewScanner(r)

	args := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#") || len(line) == 0 { // skip comment line and empty line
			continue
		}

		if !strings.HasPrefix(line, prefix) {
			break // setopts as to be the top level commands
		}

		opts := strings.TrimSpace(line[len(prefix):])
		sargs, err := splitArgs(opts)
		if err != nil {
			return nil, fmt.Errorf("unable to split opts %q: %w", opts, err)
		}

		args = append(args, sargs...)
	}

	return args, nil
}
