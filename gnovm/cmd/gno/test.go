package main

import (
	"context"
	"flag"
	"fmt"
	goio "io"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type testCfg struct {
	verbose             bool
	rootDir             string
	run                 string
	timeout             time.Duration
	updateGoldenTests   bool
	printRuntimeMetrics bool
	printEvents         bool
}

func newTestCmd(io commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags] <package> [<package>...]",
			ShortHelp:  "runs the tests for the specified packages",
			LongHelp: `Runs the tests for the specified packages.

'gno test' recompiles each package along with any files with names matching the
file pattern "*_test.gno" or "*_filetest.gno".

The <package> can be directory or file path (relative or absolute).

- "*_test.gno" files work like "*_test.go" files, but they contain only test
functions. Benchmark and fuzz functions aren't supported yet. Similarly, only
tests that belong to the same package are supported for now (no "xxx_test").

The package path used to execute the "*_test.gno" file is fetched from the
module name found in 'gno.mod', or else it is randomly generated like
"gno.land/r/XXXXXXXX".

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
}

func execTest(cfg *testCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	// Find targets for test.
	conf := &packages.LoadConfig{
		IO:           io,
		Fetcher:      testPackageFetcher,
		DepsPatterns: []string{"./..."},
	}
	pkgs, err := packages.Load(conf, args...)
	if err != nil {
		return err
	}

	pkgsMap := map[string]*packages.Package{}
	packages.Inject(pkgsMap, pkgs)

	if cfg.timeout > 0 {
		go func() {
			time.Sleep(cfg.timeout)
			panic("test timed out after " + cfg.timeout.String())
		}()
	}

	// Set up options to run tests.
	stdout := goio.Discard
	if cfg.verbose {
		stdout = io.Out()
	}
	opts := test.NewTestOptions(cfg.rootDir, pkgsMap, io.In(), stdout, io.Err())
	opts.RunFlag = cfg.run
	opts.Sync = cfg.updateGoldenTests
	opts.Verbose = cfg.verbose
	opts.Metrics = cfg.printRuntimeMetrics
	opts.Events = cfg.printEvents

	buildErrCount := 0
	testErrCount := 0
	for _, pkg := range pkgs {
		// ignore deps
		if len(pkg.Match) == 0 {
			continue
		}

		if !pkg.Draft && pkg.Files.Size() == 0 {
			return fmt.Errorf("no Gno files in %s", pkg.Dir)
		}

		label := pkg.ImportPath
		if label == "" {
			label = tryRelativize(pkg.Dir)
		}

		if len(pkg.Files[packages.FileKindTest]) == 0 && len(pkg.Files[packages.FileKindXTest]) == 0 && len(pkg.Files[packages.FileKindFiletest]) == 0 {
			io.ErrPrintfln("?       %s \t[no test files]", label)
			continue
		}

		depsConf := *conf
		depsConf.Deps = true
		depsConf.Cache = pkgsMap
		deps, loadDepsErr := packages.Load(&depsConf, pkg.Dir)
		if loadDepsErr != nil {
			io.ErrPrintfln("%s: load deps: %v", label, err)
			buildErrCount++
			continue
		}
		packages.Inject(pkgsMap, deps)

		memPkg, err := gno.ReadMemPackage(pkg.Dir, label, conf.Fset)
		if err != nil {
			io.ErrPrintln(err)
			buildErrCount++
			continue
		}

		startedAt := time.Now()
		hasError := catchRuntimeError(pkg.Dir, io.Err(), func() {
			err = test.Test(memPkg, pkg.Dir, opts)
		})

		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		if hasError || err != nil {
			if err != nil {
				io.ErrPrintfln("%s: test pkg: %v", label, err)
			}
			io.ErrPrintfln("FAIL")
			io.ErrPrintfln("FAIL    %s \t%s", label, dstr)
			io.ErrPrintfln("FAIL")
			testErrCount++
		} else {
			io.ErrPrintfln("ok      %s \t%s", label, dstr)
		}
	}
	if testErrCount > 0 || buildErrCount > 0 {
		io.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d build errors, %d test errors", buildErrCount, testErrCount)
	}

	return nil
}
