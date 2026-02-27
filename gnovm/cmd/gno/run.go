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
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type runCmd struct {
	verbose    bool
	rootDir    string
	expr       string
	debug      bool
	debugAddr  string
	dapMode    bool
	attachMode bool
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

	fs.BoolVar(
		&c.dapMode,
		"dap",
		false,
		"enable Debug Adapter Protocol mode for IDE integration",
	)

	fs.BoolVar(
		&c.attachMode,
		"attach",
		false,
		"attach to an already running debug session (requires --dap)",
	)
}

func packageNameFromFiles(args []string) (string, error) {
	var (
		firstPkgName string
		firstPkgFile string
		foundAny     bool
	)

	for _, arg := range args {
		s, err := os.Stat(arg)
		if err != nil {
			return "", err
		}

		// ---- Directory case ----
		if s.IsDir() {
			files, err := os.ReadDir(arg)
			if err != nil {
				return "", err
			}

			dirFoundAny := false

			for _, f := range files {
				n := f.Name()
				if !isGnoFile(f) ||
					strings.HasSuffix(n, "_test.gno") ||
					strings.HasSuffix(n, "_filetest.gno") {
					continue
				}

				fullPath := filepath.Join(arg, n)
				firstPkgName, firstPkgFile, err = updatePackageInfo(fullPath, firstPkgName, firstPkgFile)
				if err != nil {
					return "", err
				}
				foundAny = true
				dirFoundAny = true
			}

			// when directory has only test files
			if !dirFoundAny {
				return "", fmt.Errorf("gno: no non-test Gno files in %s", arg)
			}

			continue
		}

		// ---- File case ----
		n := filepath.Base(arg)
		if strings.HasSuffix(n, "_test.gno") || strings.HasSuffix(n, "_filetest.gno") {
			return "", fmt.Errorf("gno run: cannot run test files (%s), use gno test instead", n)
		}

		firstPkgName, firstPkgFile, err = updatePackageInfo(arg, firstPkgName, firstPkgFile)
		if err != nil {
			return "", err
		}
		foundAny = true
	}

	if !foundAny {
		return "", fmt.Errorf("no valid gno file found")
	}

	return firstPkgName, nil
}

// updatePackageInfo parses the package name of a given .gno file
// and compares it with the first known package. It returns updated values
// for firstPkgName and firstPkgFile, or an error if a mismatch is found.
func updatePackageInfo(
	path string,
	firstPkgName, firstPkgFile string,
) (string, string, error) {
	pkgName, err := gno.ParseFilePackageName(path)
	if err != nil {
		return firstPkgName, firstPkgFile, err
	}

	if firstPkgName == "" {
		// First valid file sets the base package
		return pkgName, path, nil
	}

	if pkgName != firstPkgName {
		return firstPkgName, firstPkgFile, fmt.Errorf(
			"found mismatched packages %s (%s) and %s (%s)",
			firstPkgName, filepath.Base(firstPkgFile),
			pkgName, filepath.Base(path),
		)
	}

	return firstPkgName, firstPkgFile, nil
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

	var send std.Coins
	pkgPath, err := packageNameFromFiles(args)
	if err != nil {
		return err
	}
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

	// read files
	files, err := parseFiles(m, args, stderr)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no files to run")
	}

	// If the debug address is set, the debugger waits for a remote client to connect to it.
	if cfg.debugAddr != "" {
		if cfg.dapMode {
			// In DAP mode, run the program in a goroutine and start the DAP server
			// The DAP server will control program execution via debug commands
			programDone := make(chan error, 1)

			if !cfg.attachMode {
				// Launch mode: start program execution automatically
				go func() {
					// Wait for DAP server to be initialized
					for m.Debugger.DAPServer() == nil {
						time.Sleep(10 * time.Millisecond)
					}
					// Small additional delay to ensure DAP server is fully ready
					time.Sleep(50 * time.Millisecond)

					m.RunFiles(files...)
					err := runExpr(m, cfg.expr)

					// Send terminated event if DAP server is active
					if dapServer := m.Debugger.DAPServer(); dapServer != nil {
						dapServer.SendTerminatedEvent()
					}

					programDone <- err
				}()
			}
			// In attach mode, don't start program execution - wait for attach request

			// Start DAP server mode in another goroutine
			serverDone := make(chan error, 1)
			go func() {
				serverDone <- m.Debugger.ServeDAP(m, cfg.debugAddr, cfg.attachMode, files, cfg.expr)
			}()

			// Wait for either program completion or server error
			select {
			case err := <-programDone:
				return err
			case err := <-serverDone:
				return err
			}
		} else {
			// Start traditional debugger mode
			if err := m.Debugger.Serve(cfg.debugAddr); err != nil {
				return err
			}
		}
	}

	// run files
	m.RunFiles(files...)
	return runExpr(m, cfg.expr)
}

func parseFiles(m *gno.Machine, fpaths []string, stderr io.WriteCloser) ([]*gno.FileNode, error) {
	files := make([]*gno.FileNode, 0, len(fpaths))
	var didPanic bool
	for _, fpath := range fpaths {
		if s, err := os.Stat(fpath); err == nil && s.IsDir() {
			subFns, err := listNonTestFiles(fpath)
			if err != nil {
				return nil, err
			}
			subFiles, err := parseFiles(m, subFns, stderr)
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
