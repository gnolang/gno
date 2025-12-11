package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	goio "io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type testCmd struct {
	verbose             bool
	failfast            bool
	rootDir             string
	autoGnomod          bool
	run                 string
	timeout             time.Duration
	updateGoldenTests   bool
	printRuntimeMetrics bool
	printEvents         bool
	debug               bool
	debugAddr           string
}

func newTestCmd(io commands.IO) *commands.Command {
	cmd := &testCmd{}

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
		cmd,
		func(_ context.Context, args []string) error {
			return execTest(cmd, args, io)
		},
	)
}

func (c *testCmd) RegisterFlags(fs *flag.FlagSet) {
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

	fs.BoolVar(
		&c.autoGnomod,
		"auto-gnomod",
		true,
		"auto-generate gno.mod file if not already present.",
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

func execTest(cmd *testCmd, args []string, io commands.IO) error {
	// Default to current directory if no args provided
	if len(args) == 0 {
		args = []string{"."}
	}

	// Guess opts.RootDir.
	if cmd.rootDir == "" {
		cmd.rootDir = gnoenv.RootDir()
	}

	loadConf := packages.LoadConfig{
		Fetcher:    testPackageFetcher,
		Out:        io.Err(),
		Deps:       true,
		Test:       true,
		AllowEmpty: true,
	}
	pkgs, err := packages.Load(loadConf, args...)
	if err != nil {
		return err
	}

	if len(pkgs) == 0 {
		io.ErrPrintln("no packages to test")
		return nil
	}

	if cmd.timeout > 0 {
		go func() {
			time.Sleep(cmd.timeout)
			panic("test timed out after " + cmd.timeout.String())
		}()
	}

	// Set up options to run tests.
	stdout := goio.Discard
	if cmd.verbose {
		stdout = io.Out()
	}
	opts := test.NewTestOptions(cmd.rootDir, stdout, io.Err(), pkgs)
	opts.RunFlag = cmd.run
	opts.Sync = cmd.updateGoldenTests
	opts.Verbose = cmd.verbose
	opts.Metrics = cmd.printRuntimeMetrics
	opts.Events = cmd.printEvents
	opts.Debug = cmd.debug
	opts.FailfastFlag = cmd.failfast
	cache := make(gno.TypeCheckCache, 64)

	// test.ProdStore() is suitable for type-checking prod (non-test) files.
	// _, pgs := test.ProdStore(cmd.rootDir, opts.WriterForStore())

	buildErrCount := 0
	testErrCount := 0
	fail := func() error {
		io.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d build errors, %d test errors", buildErrCount, testErrCount)
	}

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			io.ErrPrintfln("%s", err.Error())
			buildErrCount++
		}
		// don't test packages with load errors
		if len(pkg.Errors) != 0 {
			continue
		}
		// don't test packages not listed in patterns
		if len(pkg.Match) == 0 {
			continue
		}

		// Relativize and prepend dot to pkg dir if possible
		// We ignore errors since it's a cosmetic thing
		// XXX: use pkg import path instead of this when printing if possible
		prettyDir := pkg.Dir
		if filepath.IsAbs(pkg.Dir) {
			cwd, err := os.Getwd()
			if err == nil {
				relDir, err := filepath.Rel(cwd, pkg.Dir)
				if err == nil {
					prettyDir = relDir
					if prettyDir != "." && !strings.HasPrefix(prettyDir, "."+string(filepath.Separator)) {
						prettyDir = "." + string(filepath.Separator) + prettyDir
					}
				}
			}
		}

		if len(pkg.Files[packages.FileKindTest]) == 0 && len(pkg.Files[packages.FileKindXTest]) == 0 && len(pkg.Files[packages.FileKindFiletest]) == 0 {
			io.ErrPrintfln("?       %s \t[no test files]", prettyDir)
			continue
		}

		// Read and parse gnomod.toml directly.
		fpath := filepath.Join(pkg.Dir, "gnomod.toml")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			if cmd.autoGnomod {
				modulePath, _ := determinePkgPath(nil, pkg.Dir, cmd.rootDir)
				modstr := gno.GenGnoModLatest(modulePath)
				mod, err = gnomod.ParseBytes("gnomod.toml", []byte(modstr))
				if err != nil {
					panic(fmt.Errorf("unexpected panic parsing default gnomod.toml bytes: %w", err))
				}
				io.ErrPrintfln("auto-generated %q", fpath)
				err = mod.WriteFile(fpath)
				if err != nil {
					panic(fmt.Errorf("unexpected panic writing to %q: %w", fpath, err))
				}
				// err == nil.
			}
		}

		// Determine pkgPath from gno.mod.
		pkgPath, ok := determinePkgPath(mod, pkg.Dir, cmd.rootDir)
		if !ok {
			io.ErrPrintfln("WARNING: unable to read package path from gno.mod or gno root directory; try creating a gno.mod file")
		}

		// Read MemPackage with all files.
		mpkg := gno.MustReadMemPackage(pkg.Dir, pkgPath, gno.MPAnyAll)
		var didPanic, didError bool
		startedAt := time.Now()
		didPanic = catchPanic(pkg.Dir, pkgPath, io.Err(), func() {
			if mod == nil || !mod.Ignore {
				_, errs := lintTypeCheck(io, pkg.Dir, mpkg, gno.TypeCheckOptions{
					Getter:     opts.TestStore,
					TestGetter: opts.TestStore,
					Mode:       gno.TCLatestRelaxed,
					Cache:      cache,
				})
				if errs != nil {
					didError = true
					// already printed in lintTypeCheck.
					// io.ErrPrintln(errs)
					return
				}
			} else if cmd.verbose {
				io.ErrPrintfln("%s: module is ignore, skipping type check", pkgPath)
			}

			///////////////////////////////////
			// Run the tests found in the mpkg.
			errs := test.Test(mpkg, prettyDir, opts)
			if errs != nil {
				didError = true
				io.ErrPrintln(errs)
				return
			}
		})

		// Print status with duration.
		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)
		if didPanic || didError {
			io.ErrPrintfln("FAIL    %s \t%s", prettyDir, dstr)
			testErrCount++
			if cmd.failfast {
				return fail()
			}
		} else {
			io.ErrPrintfln("ok      %s \t%s", prettyDir, dstr)
		}
	}
	if testErrCount > 0 || buildErrCount > 0 {
		return fail()
	}

	return nil
}

func determinePkgPath(mod *gnomod.File, dir, rootDir string) (string, bool) {
	if mod != nil {
		return mod.Module, true
	}
	if pkgPath := pkgPathFromRootDir(dir, rootDir); pkgPath != "" {
		return pkgPath, true
	}
	// unable to read pkgPath from gno.mod, use a deterministic path.
	return "gno.land/r/test", false // XXX: gno.land hardcoded for convenience.
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
