package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type runCmd struct {
	verbose   bool
	rootDir   string
	expr      string
	debug     bool
	debugAddr string
}

func newRunCmd(cio commands.IO) *commands.Command {
	cfg := &runCmd{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <file> [<file>...]",
			ShortHelp:  "run gno packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(cfg, args, cio)
		},
	)
}

func (c *runCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno binary tries to guess it)",
	)

	fs.StringVar(
		&c.expr,
		"expr",
		"main()",
		"value of expression to evaluate. Defaults to executing function main() with no args",
	)

	fs.BoolVar(
		&c.debug,
		"debug",
		false,
		"enable interactive debugger using stdin and stdout",
	)

	fs.StringVar(
		&c.debugAddr,
		"debug-addr",
		"",
		"enable interactive debugger using tcp address in the form [host]:port",
	)
}

func execRun(cfg *runCmd, args []string, cio commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	stdin := cio.In()
	stdout := cio.Out()
	stderr := cio.Err()

	// init store and machine
	output := test.OutputWithError(stdout, stderr)
	_, testStore := test.ProdStore(
		cfg.rootDir, output, nil)

	if len(args) == 0 {
		args = []string{"."}
	}

	// read files
	files, err := parseFiles(args, stderr)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no files to run")
	}

	var send std.Coins
	pkgPath := string(files[0].PkgName)
	ctx := test.Context("", pkgPath, send)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       pkgPath,
		Output:        output,
		Input:         stdin,
		Store:         testStore,
		MaxAllocBytes: maxAllocRun,
		Context:       ctx,
		Debug:         cfg.debug || cfg.debugAddr != "",
	})

	defer m.Release()

	// If the debug address is set, the debugger waits for a remote client to connect to it.
	if cfg.debugAddr != "" {
		if err := m.Debugger.Serve(cfg.debugAddr); err != nil {
			return err
		}
	}

	// run files
	m.RunFiles(files...)
	return runExpr(m, cfg.expr)
}

func parseFiles(fpaths []string, stderr io.WriteCloser) ([]*gno.FileNode, error) {
	files := make([]*gno.FileNode, 0, len(fpaths))
	var didPanic bool
	var m *gno.Machine
	for _, fpath := range fpaths {
		if s, err := os.Stat(fpath); err == nil && s.IsDir() {
			subFns, err := listNonTestFiles(fpath)
			if err != nil {
				return nil, err
			}
			subFiles, err := parseFiles(subFns, stderr)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
			continue
		} else if err != nil {
			// either not found or some other kind of error --
			// in either case not a file we can parse.
			return nil, err
		}

		dir, fname := filepath.Split(fpath)
		didPanic = catchPanic(dir, fname, stderr, func() {
			files = append(files, m.MustReadFile(fpath))
		})
	}

	if didPanic {
		return nil, commands.ExitCodeError(1)
	}
	return files, nil
}

func listNonTestFiles(dir string) ([]string, error) {
	fs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fn := make([]string, 0, len(fs))
	for _, f := range fs {
		n := f.Name()
		if isGnoFile(f) &&
			!strings.HasSuffix(n, "_test.gno") &&
			!strings.HasSuffix(n, "_filetest.gno") {
			fn = append(fn, filepath.Join(dir, n))
		}
	}
	return fn, nil
}

func runExpr(m *gno.Machine, expr string) (err error) {
	ex, err := m.ParseExpr(expr)
	if err != nil {
		return fmt.Errorf("could not parse expression: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case gno.UnhandledPanicError:
				err = fmt.Errorf("panic running expression %s: %v\nStacktrace:\n%s",
					expr, r.Error(), m.ExceptionStacktrace())
			default:
				err = fmt.Errorf("panic running expression %s: %v\nStacktrace:\n%s",
					expr, r, m.Stacktrace().String())
			}
		}
	}()
	m.Eval(ex)
	return nil
}

const maxAllocRun = 500_000_000
