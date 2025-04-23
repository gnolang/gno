package main

import (
	"context"
	"flag"
	"fmt"
	goio "io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type testCfg struct {
	verbose             bool
	failfast            bool
	rootDir             string
	run                 string
	timeout             time.Duration
	updateGoldenTests   bool
	printRuntimeMetrics bool
	printEvents         bool
	debug               bool
	debugAddr           string
}

func newTestCmd(io commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags] <package> [<package>...]",
			ShortHelp:  "test packages",
			LongHelp: `Runs the tests for the specified packages.

'gno test' recompiles each package along with any files with names matching the
file pattern "*_test.gno" or "*_filetest.gno".

The <package> can be directory or file path (relative or absolute).

- "*_test.gno" files work like "*_test.go" files, but they contain only test
functions. Benchmark and fuzz functions aren't supported yet. Similarly, only
tests that belong to the same package are supported for now (no "xxx_test").

The package path used to execute the "*_test.gno" file is fetched from the
module name found in 'gno.mod', or else it is set to
"gno.land/r/txtar".

- "*_filetest.gno" files on the other hand are kind of unique. They exist to
provide a way to interact and assert a gno contract, thanks to a set of
specific directives that can be added using code comments.

"*_filetest.gno" must be declared in the 'main' package and so must have a
'main' function, that will be executed to test the target contract.

These single-line directives can set "input parameters" for the machine used
to perform the test:
	- "PKGPATH:" is a single line directive that can be used to define the
	package used to interact with the tested package. If not specified, "main" is
	used.
	- "MAXALLOC:" is a single line directive that can be used to define a limit
	to the VM allocator. If this limit is exceeded, the VM will panic. Default to
	0, no limit.
	- "SEND:" is a single line directive that can be used to send an amount of
	token along with the transaction. The format is for example "1000000ugnot".
	Default is empty.

These directives, instead, match the comment that follows with the result
of the GnoVM, acting as a "golden test":
	- "Output:" tests the following comment with the standard output of the
	filetest.
	- "Error:" tests the following comment with any panic, or other kind of
	error that the filetest generates (like a parsing or preprocessing error).
	- "Realm:" tests the following comment against the store log, which can show
	what realm information is stored.
	- "Stacktrace:" can be used to verify the following lines against the
	stacktrace of the error.
	- "Events:" can be used to verify the emitted events against a JSON.

To speed up execution, imports of pure packages are processed separately from
the execution of the tests. This makes testing faster, but means that the
initialization of imported pure packages cannot be checked in filetests.
`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTest(cfg, args, io)
		},
	)
}

func (c *testCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)

	fs.BoolVar(
		&c.failfast,
		"failfast",
		false,
		"do not start new tests after the first test failure",
	)

	fs.BoolVar(
		&c.updateGoldenTests,
		"update-golden-tests",
		false,
		`writes actual as wanted for "golden" directives in filetests`,
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
	)

	fs.StringVar(
		&c.run,
		"run",
		"",
		"test name filtering pattern",
	)

	fs.DurationVar(
		&c.timeout,
		"timeout",
		0,
		"max execution time",
	)

	fs.BoolVar(
		&c.printRuntimeMetrics,
		"print-runtime-metrics",
		false,
		"print runtime metrics (gas, memory, cpu cycles)",
	)

	fs.BoolVar(
		&c.printEvents,
		"print-events",
		false,
		"print emitted events",
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

func execTest(cfg *testCfg, args []string, io commands.IO) error {
	// Default to current directory if no args provided
	if len(args) == 0 {
		args = []string{"."}
	}

	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("list targets from patterns: %w", err)
	}

	if len(paths) == 0 {
		io.ErrPrintln("no packages to test")
		return nil
	}

	if cfg.timeout > 0 {
		go func() {
			time.Sleep(cfg.timeout)
			panic("test timed out after " + cfg.timeout.String())
		}()
	}

	subPkgs, err := gnomod.SubPkgsFromPaths(paths)
	if err != nil {
		return fmt.Errorf("list sub packages: %w", err)
	}

	// Set up options to run tests.
	stdout := goio.Discard
	if cfg.verbose {
		stdout = io.Out()
	}
	opts := test.NewTestOptions(cfg.rootDir, stdout, io.Err())
	opts.RunFlag = cfg.run
	opts.Sync = cfg.updateGoldenTests
	opts.Verbose = cfg.verbose
	opts.Metrics = cfg.printRuntimeMetrics
	opts.Events = cfg.printEvents
	opts.Debug = cfg.debug
	opts.FailfastFlag = cfg.failfast

	buildErrCount := 0
	testErrCount := 0
	fail := func() error {
		io.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d build errors, %d test errors", buildErrCount, testErrCount)
	}

	for _, pkg := range subPkgs {
		if len(pkg.TestGnoFiles) == 0 && len(pkg.FiletestGnoFiles) == 0 {
			io.ErrPrintfln("?       %s \t[no test files]", pkg.Dir)
			continue
		}
		// Determine gnoPkgPath by reading gno.mod
		var gnoPkgPath string
		modfile, err := gnomod.ParseAt(pkg.Dir)
		if err == nil {
			gnoPkgPath = modfile.Module.Mod.Path
		} else {
			gnoPkgPath = pkgPathFromRootDir(pkg.Dir, cfg.rootDir)
			if gnoPkgPath == "" {
				// unable to read pkgPath from gno.mod, use a deterministic path.
				io.ErrPrintfln("--- WARNING: unable to read package path from gno.mod or gno root directory; try creating a gno.mod file")
				gnoPkgPath = "gno.land/r/txtar" // XXX: gno.land hardcoded for convenience.
			}
		}

		memPkg := gno.MustReadMemPackage(pkg.Dir, gnoPkgPath)

		var hasError bool

		startedAt := time.Now()
		runtimeError := catchRuntimeError(gnoPkgPath, io.Err(), func() {
			if modfile == nil || !modfile.Draft {
				foundErr, lintErr := lintTypeCheck(io, memPkg, opts.TestStore)
				if lintErr != nil {
					io.ErrPrintln(lintErr)
					hasError = true
				} else if foundErr {
					hasError = true
				}
			} else if cfg.verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", gnoPkgPath)
			}
			err = test.Test(memPkg, pkg.Dir, opts)
		})
		hasError = hasError || runtimeError

		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		if hasError || err != nil {
			if err != nil {
				io.ErrPrintfln("%s: test pkg: %v", pkg.Dir, err)
			}
			io.ErrPrintfln("FAIL    %s \t%s", pkg.Dir, dstr)
			testErrCount++
			if cfg.failfast {
				return fail()
			}
		} else {
			io.ErrPrintfln("ok      %s \t%s", pkg.Dir, dstr)
		}
	}
	if testErrCount > 0 || buildErrCount > 0 {
		return fail()
	}

	return nil
}

// attempts to determine the full gno pkg path by analyzing the directory.
func pkgPathFromRootDir(pkgPath, rootDir string) string {
	abPkgPath, err := filepath.Abs(pkgPath)
	if err != nil {
		log.Printf("could not determine abs path: %v", err)
		return ""
	}
	abRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Printf("could not determine abs path: %v", err)
		return ""
	}
	abRootDir += string(filepath.Separator)
	if !strings.HasPrefix(abPkgPath, abRootDir) {
		return ""
	}
	impPath := strings.ReplaceAll(abPkgPath[len(abRootDir):], string(filepath.Separator), "/")
	for _, prefix := range [...]string{
		"examples/",
		"gnovm/stdlibs/",
		"gnovm/tests/stdlibs/",
	} {
		if strings.HasPrefix(impPath, prefix) {
			return impPath[len(prefix):]
		}
	}
	return ""
}
