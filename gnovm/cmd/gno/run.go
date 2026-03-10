package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/lint"
	"github.com/gnolang/gno/gnovm/pkg/lint/reporters"
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
	pkgPath   string
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

	fs.StringVar(
		&c.pkgPath,
		"pkgpath",
		"",
		"run with this package path, overriding the \"// PKGPATH:\" file directive and the gnomod.toml module path",
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
	_, testStore := test.ProdStore(cfg.rootDir, output, nil)

	if len(args) == 0 {
		args = []string{"."}
	}

	var send std.Coins
	pkgName, err := packageNameFromFiles(args)
	if err != nil {
		return err
	}

	// The package path is given by the -pkgpath flag if set, else derived
	// from the first argument's "// PKGPATH:" directive or gnomod.toml
	// module path, else it defaults to the package name.
	pkgPath := cfg.pkgPath
	if pkgPath == "" {
		if pkgPath, err = derivePkgPath(args[0]); err != nil {
			return err
		}
	}
	if pkgPath == "" {
		pkgPath = pkgName
	}

	// Realm packages persist state; run them in a transaction store.
	realmMode := gno.IsRealmPath(pkgPath)
	store := testStore
	if realmMode {
		store = testStore.BeginTransaction(nil, nil, nil, nil)
	}

	ctx := test.Context("", pkgPath, send)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        output,
		Input:         stdin,
		Store:         store,
		MaxAllocBytes: maxAllocRun,
		Context:       ctx,
		Debug:         cfg.debug || cfg.debugAddr != "",
	})

	defer m.Release()

	// Construct the package to run; don't use MachineOptions.PkgPath, which
	// would load an existing package at the same path from the store,
	// conflicting with the files to run ("package fork" simulation).
	pn := gno.NewPackageNode(gno.Name(pkgName), pkgPath, &gno.FileSet{})
	pv := pn.NewPackage(m.Alloc)
	m.Store.SetBlockNode(pn)
	m.Store.SetCachePackage(pv)
	m.SetActivePackage(pv)

	if cfg.debug {
		// Provide a helper to access sources of stdlibs and examples
		// packages, so that the debugger can list them.
		m.Debugger.Enable(stdin, output, func(ppath, name string) string {
			p := filepath.Join(cfg.rootDir, ppath, name)
			b, err := os.ReadFile(p)
			if err != nil {
				p = filepath.Join(cfg.rootDir, "gnovm", "stdlibs", ppath, name)
				b, err = os.ReadFile(p)
			}
			if err != nil {
				p = filepath.Join(cfg.rootDir, "examples", ppath, name)
				b, err = os.ReadFile(p)
			}
			if err != nil {
				return ""
			}
			return string(b)
		})
	}

	// read files
	files, err := parseFiles(m, args, reporters.NewDirectReporter(stderr))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no files to run")
	}

	// If the debug address is set, the debugger waits for a remote client to connect to it.
	if cfg.debugAddr != "" {
		if err := m.Debugger.Serve(cfg.debugAddr); err != nil {
			return err
		}
	}

	if realmMode {
		// Set the origin caller before running package init, so that
		// package-level initializers see the proper caller (matches the
		// filetest ordering in gnovm/pkg/test).
		ctx.OriginCaller = test.DefaultCaller
	}

	// run files
	m.RunFiles(files...)

	if realmMode {
		// Reconstruct the active package from the store, following realm
		// finalization by RunFiles.
		m.SetActivePackage(m.Store.GetPackage(pkgPath, false))
	}
	return runExpr(m, cfg.expr)
}

func parseFiles(m *gno.Machine, fpaths []string, reporter lint.Reporter) ([]*gno.FileNode, error) {
	files := make([]*gno.FileNode, 0, len(fpaths))
	var didPanic bool
	for _, fpath := range fpaths {
		if s, err := os.Stat(fpath); err == nil && s.IsDir() {
			subFns, err := listNonTestFiles(fpath)
			if err != nil {
				return nil, err
			}
			subFiles, err := parseFiles(m, subFns, reporter)
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
		didPanic = catchPanicWithReporter(reporter, dir, fname, func() {
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
	// If the expression is a call to a crossing function of the package
	// (e.g. `main(cur realm)`), prepend `.cur` as the first argument.
	m.MaybeInjectCurForEval(ex)
	m.Eval(ex)
	return nil
}

const maxAllocRun = 500_000_000
