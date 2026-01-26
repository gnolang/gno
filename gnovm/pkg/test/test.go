// Package test contains the code to parse and execute Gno tests and filetests.
package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/tests/stdlibs/chain/runtime"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"go.uber.org/multierr"
)

const (
	// DefaultHeight is the default height used in the [Context].
	DefaultHeight = 123
	// DefaultTimestamp is the Timestamp value used by default in [Context].
	DefaultTimestamp = 1234567890
	// DefaultCaller is the result of gno.DerivePkgBech32Addr("user1.gno"),
	// used as the default caller in [Context].
	DefaultCaller crypto.Bech32Address = "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
)

// Context returns a TestExecContext. Usable for test purpose only.
// The caller should be empty for package initialization.
// The returned context has a mock banker, params and event logger. It will give
// the pkgAddr the coins in `send` by default, and only that.
// The Height and Timestamp parameters are set to the [DefaultHeight] and
// [DefaultTimestamp].
func Context(caller crypto.Bech32Address, pkgPath string, send std.Coins) *runtime.TestExecContext {
	// FIXME: create a better package to manage this, with custom constructors
	pkgAddr := gno.DerivePkgBech32Addr(pkgPath) // the addr of the pkgPath called.

	banker := &runtime.TestBanker{
		CoinTable: map[crypto.Bech32Address]std.Coins{
			pkgAddr: send,
		},
	}
	ctx := stdlibs.ExecContext{
		ChainID:         "dev",
		ChainDomain:     "gno.land", // TODO: make this configurable
		Height:          DefaultHeight,
		Timestamp:       DefaultTimestamp,
		OriginCaller:    caller,
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		Banker:          banker,
		Params:          newTestParams(),
		EventLogger:     sdk.NewEventLogger(),
	}
	return &runtime.TestExecContext{
		ExecContext: ctx,
		RealmFrames: make(map[int]runtime.RealmOverride),
	}
}

// Machine is a minimal machine, set up with just the Store, Output and Context.
// It is only used for linting/preprocessing.
func Machine(testStore gno.Store, output io.Writer, pkgPath string, debug bool, gasMeter store.GasMeter) *gno.Machine {
	return gno.NewMachineWithOptions(gno.MachineOptions{
		Store:         testStore,
		Output:        output,
		Context:       Context("", pkgPath, nil),
		Debug:         debug,
		ReviveEnabled: true,
		GasMeter:      gasMeter,
	})
}

// OutputWithError returns an io.Writer that can be used as a [gno.Machine.Output],
// where the test standard libraries will write to errWriter when using
// os.Stderr.
func OutputWithError(output, errWriter io.Writer) io.Writer {
	return &outputWithError{output, errWriter}
}

type outputWithError struct {
	w    io.Writer
	errW io.Writer
}

func (o outputWithError) Write(p []byte) (int, error)       { return o.w.Write(p) }
func (o outputWithError) StderrWrite(p []byte) (int, error) { return o.errW.Write(p) }

// ----------------------------------------
// testParams

type testParams struct{}

func newTestParams() *testParams {
	return &testParams{}
}

func (tp *testParams) SetBool(key string, val bool)                     { /* noop */ }
func (tp *testParams) SetBytes(key string, val []byte)                  { /* noop */ }
func (tp *testParams) SetInt64(key string, val int64)                   { /* noop */ }
func (tp *testParams) SetUint64(key string, val uint64)                 { /* noop */ }
func (tp *testParams) SetString(key string, val string)                 { /* noop */ }
func (tp *testParams) SetStrings(key string, val []string)              { /* noop */ }
func (tp *testParams) UpdateStrings(key string, val []string, add bool) { /* noop */ }

// ----------------------------------------
// main test function

// TestOptions is a list of options that must be passed to [Test].
type TestOptions struct {
	// BaseStore / TestStore to use for the tests.
	BaseStore storetypes.CommitStore
	TestStore gno.Store
	// Gno root dir.
	RootDir string
	// Used for printing program output, during verbose logging.
	Output io.Writer
	// Used for os.Stderr, and for printing errors.
	Error io.Writer
	// Debug enables the interactive debugger on gno tests.
	Debug bool

	// Not set by NewTestOptions:

	// Flag to filter tests to run.
	RunFlag string
	// Flag to stop executing as soon a test fails.
	FailfastFlag bool
	// Whether to update filetest directives.
	Sync bool
	// Uses Error to print when starting a test, and prints test output directly,
	// unbuffered.
	Verbose bool
	// Uses Error to print runtime metrics for tests.
	Metrics bool
	// Uses Error to print the events emitted.
	Events bool

	filetestBuffer bytes.Buffer
	outWriter      proxyWriter
	tcCache        gno.TypeCheckCache
}

// WriterForStore is the writer that should be passed to [Store], so that
// [Test] is then able to swap it when needed.
func (opts *TestOptions) WriterForStore() io.Writer {
	return &opts.outWriter
}

// NewTestOptions sets up TestOptions, filling out all "required" parameters.
func NewTestOptions(rootDir string, stdout, stderr io.Writer, pkgs packages.PkgList) *TestOptions {
	opts := &TestOptions{
		RootDir: rootDir,
		Output:  stdout,
		Error:   stderr,
	}
	opts.BaseStore, opts.TestStore = StoreWithOptions(
		rootDir, opts.WriterForStore(), StoreOptions{
			WithExtern: false,
			Testing:    true,
			Packages:   pkgs,
		})
	return opts
}

// proxyWriter is a simple wrapper around a io.Writer, it exists so that the
// underlying writer can be swapped with another when necessary.
type proxyWriter struct {
	w    io.Writer
	errW io.Writer
}

func (p *proxyWriter) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

// StderrWrite implements the interface specified in tests/stdlibs/os/os.go,
// which if found in Machine.Output allows to write to stderr from Gno.
func (p *proxyWriter) StderrWrite(b []byte) (int, error) {
	return p.errW.Write(b)
}

// tee temporarily appends the writer w to an underlying MultiWriter, which
// should then be reverted using revert().
func (p *proxyWriter) tee(w io.Writer) (revert func()) {
	rev := tee(&p.w, w)
	revErr := tee(&p.errW, w)
	return func() {
		rev()
		revErr()
	}
}

func tee(ptr *io.Writer, dst io.Writer) (revert func()) {
	save := *ptr
	if save == io.Discard {
		*ptr = dst
	} else {
		*ptr = io.MultiWriter(save, dst)
	}
	return func() {
		*ptr = save
	}
}

// Test runs tests on the specified mpkg.
// fsDir is the directory on filesystem of package; it's used in case opts.Sync
// is enabled, and points to the directory where the files are contained if they
// are to be updated.
// opts is a required set of options, which is often shared among different
// tests; you can use [NewTestOptions] for a common base configuration.
// It is assumed that mpkg is already type-checked, even filetests.
func Test(mpkg *std.MemPackage, fsDir string, opts *TestOptions) error {
	opts.outWriter.w = opts.Output
	opts.outWriter.errW = opts.Error

	var errs error

	// Create a common tcw/tgs for both the `pkg` tests as well as the
	// `pkg_test` tests. This allows us to "export" symbols from the pkg
	// tests and import them from the `pkg_test` tests.
	tcw := opts.BaseStore.CacheWrap()
	tgs := opts.TestStore.BeginTransaction(tcw, tcw, nil)

	// Let opts.TestStore load itself.
	// This needs to happen before LoadImports, as LoadImports will
	// otherwise only load without *_test.gno files (but we want them for
	// mpkg since we're running tests on them).
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: mpkg.Path,
		Output:  opts.WriterForStore(),
		Store:   tgs,
		Context: Context("", mpkg.Path, nil),
		// When testing examples we will find them, so pv, pn, file
		// block nodes would otherwise become set, but for running
		// tests on packages not known by the store, it will construct
		// new packages by default, which we don't want.  Instead we
		// will run the mempackage ourselves in the next line.
		SkipPackage: true,
	})
	// Filter out xxx_test *_test.gno and *_filetest.gno and run.
	// If testing with only filetests, there will be no files.
	tmpkg := gno.MPFTest.FilterMemPackage(mpkg)
	if !tmpkg.IsEmptyOf(".gno") {
		_, _ = m2.RunMemPackageWithOverrides(tmpkg, true)
	}

	// Eagerly load imports.
	abortOnError := true
	if err := LoadImports(tgs, mpkg, abortOnError); err != nil {
		return err
	}

	// Stands for "test", "integration test", and "filetest".
	// "integration test" are the test files with `package xxx_test` (they are
	// not necessarily integration tests, it's just for our internal reference.)
	tset, itset, itfiles, ftfiles := parseMemPackageTests(mpkg)

	// Testing with *_test.gno
	if len(tset.Files)+len(itset.Files) > 0 {
		// Run test files in pkg.
		if len(tset.Files) > 0 {
			err := opts.runTestFiles(mpkg, tset, tgs)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}

		// Test xxx_test pkg.
		if len(itset.Files) > 0 {
			var mpkgType gno.MemPackageType
			if gno.IsStdlib(mpkg.Path) {
				mpkgType = gno.MPStdlibIntegration
			} else {
				mpkgType = gno.MPUserIntegration
			}
			itmpkg := &std.MemPackage{
				Type:  mpkgType,
				Name:  mpkg.Name + "_test",
				Path:  mpkg.Path + "_test",
				Files: itfiles,
			}

			err := opts.runTestFiles(itmpkg, itset, tgs)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}
	}

	// Testing with *_filetest.gno.
	if len(ftfiles) > 0 {
		filter := splitRegexp(opts.RunFlag)
		for _, testFile := range ftfiles {
			testFileName := testFile.Name
			testFilePath := filepath.Join(fsDir, testFileName)
			// XXX consider this
			testName := fsDir + "/" + testFileName
			// testName := "file/" + testFileName
			if !shouldRun(filter, testName) {
				continue
			}

			startedAt := time.Now()
			if opts.Verbose {
				fmt.Fprintf(opts.Error, "=== RUN   %s\n", testName)
			}
			tcheck := false // already type-checked e.g. by cmd/gno/test.go
			// We can not use shared tx gno store (tgs) between _filetest.gno since we need to
			// isolate the state between them
			changed, gas, err := opts.runFiletest(
				testFileName, []byte(testFile.Body), opts.TestStore, tcheck)
			if changed != "" {
				// Note: changed always == "" if opts.Sync == false.
				err = os.WriteFile(testFilePath, []byte(changed), 0o644)
				if err != nil {
					panic(fmt.Errorf("could not fix golden file: %w", err))
				}
			}

			duration := time.Since(startedAt)
			dstr := fmtDuration(duration)
			if err != nil {
				fmt.Fprintf(opts.Error, "--- FAIL: %s (elapsed: %s, gas: %d)\n", testName, dstr, gas)
				fmt.Fprintln(opts.Error, err.Error())
				errs = multierr.Append(errs, fmt.Errorf("%s failed", testName))
			} else if opts.Verbose {
				fmt.Fprintf(opts.Error, "--- PASS: %s (elapsed: %s, gas: %d)\n", testName, dstr, gas)
			}

			// XXX: add per-test metrics
		}
	}

	return errs
}

// Runs *_test.go tests.
// Not the same as pkg/test/filetest runFiletests()
// which runs *_filetest.go tests.
func (opts *TestOptions) runTestFiles(
	mpkg *std.MemPackage,
	files *gno.FileSet,
	tgs gno.TransactionStore,
) (errs error) {
	var m *gno.Machine
	defer func() {
		if r := recover(); r != nil {
			if st := m.ExceptionStacktrace(); st != "" {
				errs = multierr.Append(errors.New(st), errs)
			}
			errs = multierr.Append(
				fmt.Errorf("panic: %v\ngo stacktrace:\n%v\ngno machine: %v\ngno stacktrace:\n%v",
					r, string(debug.Stack()), m.String(), m.Stacktrace()),
				errs,
			)
		}
	}()

	tests := loadTestFuncs(mpkg.Name, files)

	var alloc *gno.Allocator
	if opts.Metrics {
		alloc = gno.NewAllocator(math.MaxInt64)
	}
	// reset store ops, if any - we only need them for some filetests.
	opts.TestStore.SetLogStoreOps(nil)

	// Check if we already have the package - it may have been eagerly loaded.
	m = Machine(tgs, opts.WriterForStore(), mpkg.Path, opts.Debug, nil)
	m.Alloc = alloc
	if tgs.GetMemPackage(mpkg.Path) == nil {
		m.RunMemPackage(mpkg, false)
	} else {
		m.SetActivePackage(tgs.GetPackage(mpkg.Path, false))
	}
	pv := m.Package

	// Load the test files into package and save.
	// m.RunFiles(files.Files...)

	for _, tf := range tests {
		// TODO(morgan): we could theoretically use wrapping on the baseStore
		// and gno store to achieve per-test isolation. However, that requires
		// some deeper changes, as ideally we'd:
		// - Run the MemPackage independently (so it can also be run as a
		//   consequence of an import)
		// - Run the test files before this for loop (but persist it to store;
		//   RunFiles doesn't do that currently)
		// - Wrap here.
		m = Machine(tgs, opts.WriterForStore(), mpkg.Path, opts.Debug, store.NewInfiniteGasMeter())
		m.Alloc = alloc.Reset()
		m.SetActivePackage(pv)

		testingpv := m.Store.GetPackage("testing", false)
		testingtv := gno.TypedValue{T: &gno.PackageType{}, V: testingpv}
		testingcx := &gno.ConstExpr{TypedValue: testingtv}
		testfv := m.Eval(gno.Nx(tf.Name))[0].GetFunc()

		var runTestX gno.Expr
		var runTest gno.TypedValue
		var runTestF string
		var runTestCur gno.Expr
		if testfv.IsCrossing() {
			// Run a test with cur passed a special way.
			//
			// > TestSomething(cur realm, t *testing.T) {...}
			//
			// Normally this isn't possible because
			// stdlibs/testing is a non-realm, so it cannot
			// have `cur`. And while a realm could call `func(cur
			// realm){...}(cross)`, some *_test.gno test cases want
			// `cur` to refer to the realm package, while
			// `cur.Previous()` to refer to no realm--while the
			// `func(cur realm){...}(cross)` method would have both
			// previous to be the current realm.

			// Extract unexposed testing.runTestWithRealm.
			m.SetActivePackage(testingpv)
			runTestX = gno.Nx("runTest_cur")
			runTest = m.Eval(runTestX)[0]
			runTestF = "F_cur"
			runTestCur = gno.NewConstExpr(gno.Nx(".cur"), gno.NewConcreteRealm(mpkg.Path))
			m.SetActivePackage(pv)
		} else {
			// The normal way to test if `cur` isn't needed such as
			// in p package tests, or in realm package tests where
			// no non-crossing calls are made directly in the body
			// of the test func decl.
			//
			// > TestSomething(t *testing.T) {...}
			runTestX = gno.Sel(testingcx, "RunTest")
			runTest = m.Eval(runTestX)[0]
			runTestF = "F"
			runTestCur = gno.Nx("nil")
		}
		runTestCX := gno.NewConstExpr(runTestX, runTest)

		if opts.Debug {
			fileContent := func(ppath, name string) string {
				p := filepath.Join(opts.RootDir, ppath, name)
				b, err := os.ReadFile(p)
				if err != nil {
					p = filepath.Join(opts.RootDir, "gnovm", "stdlibs", ppath, name)
					b, err = os.ReadFile(p)
				}
				if err != nil {
					p = filepath.Join(opts.RootDir, "examples", ppath, name)
					b, _ = os.ReadFile(p)
				}
				return string(b)
			}
			m.Debugger.Enable(os.Stdin, os.Stdout, fileContent)
		}

		eval := m.Eval(gno.Call(
			runTestCX,                                     // Call testing.RunTest
			gno.Str(opts.RunFlag),                         // run flag
			gno.Nx(strconv.FormatBool(opts.Verbose)),      // is verbose?
			gno.Nx(strconv.FormatBool(opts.FailfastFlag)), // stop as soon as a test fails
			&gno.CompositeLitExpr{ // Third param, the testing.InternalTest
				Type: gno.Sel(testingcx, "InternalTest"),
				Elts: gno.KeyValueExprs{
					// XXX Consider this.
					// {Key: gno.X("Name"), Value: gno.Str(mpkg.Path + "/" + tf.Filename + "." + tf.Name)},
					{Key: gno.X("Name"), Value: gno.Str(tf.Name)},
					{Key: gno.X(runTestF), Value: gno.Nx(tf.Name)},
					{Key: gno.X("Cur"), Value: runTestCur},
				},
			},
		))
		fmt.Fprintf(opts.Error, "--- GAS:  %d\n", m.GasMeter.GasConsumed())

		if opts.Events {
			events := m.Context.(*runtime.TestExecContext).EventLogger.Events()
			if events != nil {
				res, err := json.Marshal(events)
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(opts.Error, "EVENTS: %s\n", string(res))
			}
		}

		ret := eval[0].GetString()
		if ret == "" {
			err := fmt.Errorf("failed to execute unit test: %q", tf.Name)
			errs = multierr.Append(errs, err)
			fmt.Fprintf(opts.Error, "--- FAIL: %s [internal gno testing error]", tf.Name)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err := json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			fmt.Fprintf(opts.Error, "--- FAIL: %s [internal gno testing error]", tf.Name)
			continue
		}

		if rep.Failed {
			err := fmt.Errorf("failed: %q", tf.Name)
			errs = multierr.Append(errs, err)
			if opts.FailfastFlag {
				return errs
			}
		}

		if opts.Metrics {
			// XXX: store changes
			// XXX: max mem consumption
			allocsVal := "n/a"
			if m.Alloc != nil {
				maxAllocs, allocs := m.Alloc.Status()
				allocsVal = fmt.Sprintf("%s(%.2f%%)",
					prettySize(allocs),
					float64(allocs)/float64(maxAllocs)*100,
				)
			}
			fmt.Fprintf(opts.Error, "---       runtime: cycle=%s allocs=%s\n",
				prettySize(m.Cycles),
				allocsVal,
			)
		}
	}

	return errs
}

// report is a mirror of Gno's stdlibs/testing.Report.
type report struct {
	Failed  bool
	Skipped bool
}

type testFunc struct {
	Package  string
	Name     string
	Filename string
}

func loadTestFuncs(pkgName string, tfiles *gno.FileSet) (rt []testFunc) {
	for _, tf := range tfiles.Files {
		for _, d := range tf.Decls {
			if fd, ok := d.(*gno.FuncDecl); ok {
				if fd.IsMethod {
					continue
				}
				fname := string(fd.Name)
				if strings.HasPrefix(fname, "Test") {
					tf := testFunc{
						Package:  pkgName,
						Name:     fname,
						Filename: tf.FileName,
					}
					rt = append(rt, tf)
				}
			}
		}
	}
	return
}

// parseMemPackageTests parses test files (skipping filetests) in the mpkg.
func parseMemPackageTests(mpkg *std.MemPackage) (tset, itset *gno.FileSet, itfiles, ftfiles []*std.MemFile) {
	tset = &gno.FileSet{}
	itset = &gno.FileSet{}
	var errs error
	var m *gno.Machine
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip this file.
		}

		n, err := m.ParseFile(mfile.Name, mfile.Body)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		if n == nil {
			panic("should not happen")
		}
		switch {
		case strings.HasSuffix(mfile.Name, "_filetest.gno"):
			ftfiles = append(ftfiles, mfile)
		case strings.HasSuffix(mfile.Name, "_test.gno") && mpkg.Name == string(n.PkgName):
			tset.AddFiles(n)
		case strings.HasSuffix(mfile.Name, "_test.gno") && mpkg.Name+"_test" == string(n.PkgName):
			itset.AddFiles(n)
			itfiles = append(itfiles, mfile)
		case mpkg.Name == string(n.PkgName):
			// normal package file
		default:
			panic(fmt.Sprintf(
				"expected package name [%s] or [%s_test] but got [%s] file [%s]",
				mpkg.Name, mpkg.Name, n.PkgName, mfile))
		}
	}
	if errs != nil {
		panic(errs)
	}
	return
}

func shouldRun(filter filterMatch, path string) bool {
	if filter == nil {
		return true
	}
	elem := strings.Split(path, "/")
	ok, _ := filter.matches(elem, matchString)
	return ok
}

// Adapted from https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func prettySize(nb int64) string {
	const unit = 1000
	if nb < unit {
		return fmt.Sprintf("%d", nb)
	}
	div, exp := int64(unit), 0
	for n := nb / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(nb)/float64(div), "kMGTPE"[exp])
}

func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}
