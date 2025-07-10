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
	"time"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// Specific file opts
type testFileOpts struct {
	FileOpts

	noParallel bool
	skip       bool
	timeout    time.Duration
}

func newDefaultTestFileOpts() *testFileOpts {
	return &testFileOpts{FileOpts: NewDefaultFileOpts()}
}

func newTestFlagsOpts(fs *flag.FlagSet) *testFileOpts {
	topts := newDefaultTestFileOpts()
	fs.BoolVar(&topts.noParallel, "no-parallel", topts.noParallel, "disable parallel testing")
	fs.BoolVar(&topts.skip, "skip", topts.skip, "skip this test")
	fs.BoolVar(&topts.NoFormat, "no-fmt", topts.NoFormat, "disable format in this file")
	fs.DurationVar(&topts.timeout, "timeout", 0, "configure file test timeout")

	return topts
}

func TestTestdata(t *testing.T) {
	t.Parallel()

	flagInMemoryTS, _ := strconv.ParseBool(os.Getenv("INMEMORY_TS"))
	flagSeqTS, _ := strconv.ParseBool(os.Getenv("SEQ_TS"))

	p := gno_integration.NewTestingParams(t, "testdata")
	p.RequireUniqueNames = true

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	mf, err := ParseDirFlags(".txtar", "txtar:opts", p.Dir, newTestFlagsOpts)
	require.NoError(t, err)

	// Set up gnoland for testscript
	err = SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	mode := CommandKindTesting
	if flagInMemoryTS {
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
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		return nil
	}

	ts := tShim{
		T:        t,
		forceSeq: flagInMemoryTS || flagSeqTS,
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

		fs := flag.NewFlagSet("txtar:opts", flag.ContinueOnError)
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
		if !strings.HasPrefix(line, "#") {
			break // Stop at first non-# line
		}

		// Remove leading #
		line = strings.TrimSpace(line[1:])

		if strings.HasPrefix(line, prefix) {
			opts := strings.TrimSpace(line[len(prefix):])
			sargs, err := splitArgs(opts)
			if err != nil {
				return nil, fmt.Errorf("unable to split opts %q: %w", opts, err)
			}

			args = append(args, sargs...)
		}
	}

	return args, nil
}
