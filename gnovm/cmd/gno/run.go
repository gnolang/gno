package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
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
	xVars     *xFlag
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

	c.xVars = newXFlag()
	fs.Var(
		c.xVars,
		"X",
		"set the value of a package-level string variable, e.g. -X main.myVar=override or "+
			"-X gno.land/r/demo/foo.Version=1.2.3 (may be repeated); like 'go build -ldflags \"-X ...\"', "+
			"the package path must be given, and only simple 'var name = \"...\"' declarations are patched",
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
	xOverrides := cfg.xVars.forPackage(pkgPath, pkgName)
	xt := newXTracker()
	files, err := parseFiles(m, args, stderr, xOverrides, xt)
	if err != nil {
		return err
	}
	if err := xt.unmatched(xOverrides); err != nil {
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

// derivePkgPath derives the package path of the files to run from the first
// argument: from its "// PKGPATH:" directive if it is a (file)test file, or
// else from the module path of the gnomod.toml file in its directory.
// It returns "" if neither is found.
func derivePkgPath(arg string) (string, error) {
	dir := arg
	if s, err := os.Stat(arg); err != nil {
		return "", err
	} else if !s.IsDir() {
		dir = filepath.Dir(arg)
		body, err := os.ReadFile(arg)
		if err != nil {
			return "", err
		}
		pkgPath, err := parsePkgPathDirective(string(body), "")
		if err != nil || pkgPath != "" {
			return pkgPath, err
		}
	}
	mod, err := gnomod.ParseDir(dir)
	switch {
	case errors.Is(err, gnomod.ErrNoModFile):
		return "", nil
	case err != nil:
		return "", err
	}
	return mod.Module, nil
}

// xTracker aggregates which -X overrides were matched (as a proper
// package-level string var) across all files in a package, so a name that
// isn't found in file1.gno but is in file2.gno doesn't get wrongly
// reported as unmatched, and a name that's genuinely never found anywhere
// is reported exactly once at the end, rather than per-file.
type xTracker struct {
	matched   map[string]bool
	wrongKind map[string]bool
}

func newXTracker() *xTracker {
	return &xTracker{matched: map[string]bool{}, wrongKind: map[string]bool{}}
}

func (t *xTracker) record(result xPatchResult) {
	for _, n := range result.Matched {
		t.matched[n] = true
	}
	// A name reported as WrongKind here might still turn out matched in
	// another file; it's only finalized as an explanation in unmatched()
	// once all files have been processed.
	for _, n := range result.WrongKind {
		t.wrongKind[n] = true
	}
}

// unmatched returns, for each requested override name not found as a
// proper package-level string var in any processed file, an error
// explaining why: either it was never found at all, or it was found but
// isn't a plain string var (a const, a different type, or missing an
// initializer).
func (t *xTracker) unmatched(requested map[string]string) error {
	var notFound, wrongKind []string

	for name := range requested {
		switch {
		case t.matched[name]:
			continue
		case t.wrongKind[name]:
			wrongKind = append(wrongKind, name)
		default:
			notFound = append(notFound, name)
		}
	}

	if len(notFound) == 0 && len(wrongKind) == 0 {
		return nil
	}

	sort.Strings(notFound)
	sort.Strings(wrongKind)

	var msgs []string
	for _, n := range notFound {
		msgs = append(msgs, fmt.Sprintf("-X: no such package-level var: %s", n))
	}
	for _, n := range wrongKind {
		msgs = append(msgs, fmt.Sprintf("-X: %s is not a package-level string var with a literal initializer", n))
	}

	return errors.New(strings.Join(msgs, "\n"))
}

func parseFiles(m *gno.Machine, fpaths []string, stderr io.WriteCloser, xOverrides map[string]string, xt *xTracker) ([]*gno.FileNode, error) {
	files := make([]*gno.FileNode, 0, len(fpaths))
	var didPanic bool
	for _, fpath := range fpaths {
		if s, err := os.Stat(fpath); err == nil && s.IsDir() {
			subFns, err := listNonTestFiles(fpath)
			if err != nil {
				return nil, err
			}
			subFiles, err := parseFiles(m, subFns, stderr, xOverrides, xt)
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
			files = append(files, mustReadAndPatchFile(m, fpath, xOverrides, xt))
		})
	}

	if didPanic {
		return nil, commands.ExitCodeError(1)
	}
	return files, nil
}

// mustReadAndPatchFile reads fpath, applies any -X overrides to its
// package-level string variable declarations (see patchXVars), and parses
// the (possibly patched) result. It panics on error, like
// (*gno.Machine).MustReadFile, which it replaces as the read path here so
// that overrides can be applied before parsing. Matches (and non-matches)
// are recorded into xt for aggregation across every file in the package.
func mustReadAndPatchFile(m *gno.Machine, fpath string, xOverrides map[string]string, xt *xTracker) *gno.FileNode {
	bz, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	body, result := patchXVars(fpath, string(bz), xOverrides)
	xt.record(result)

	return m.MustParseFile(fpath, body)
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
