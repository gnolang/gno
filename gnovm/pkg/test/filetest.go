package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/multierr"
)

// RunFiletest executes the program in source as a filetest.
// If opts.Sync is enabled, and the filetest's golden output has changed,
// the first string is set to the new generated content of the file.
func (opts *TestOptions) RunFiletest(filename string, source []byte) (string, error) {
	opts.outWriter.w = opts.Output
	opts.outWriter.errW = opts.Error

	return opts.runFiletest(filename, source)
}

var reEndOfLineSpaces = func() *regexp.Regexp {
	re := regexp.MustCompile(" +\n")
	re.Longest()
	return re
}()

func (opts *TestOptions) runFiletest(filename string, source []byte) (string, error) {
	dirs, err := ParseDirectives(bytes.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("error parsing directives: %w", err)
	}

	// Initialize Machine.Context and Machine.Alloc according to the input directives.
	pkgPath := dirs.FirstDefault(DirectivePkgPath, "main")
	coins, err := std.ParseCoins(dirs.FirstDefault(DirectiveSend, ""))
	if err != nil {
		return "", err
	}
	ctx := Context("", pkgPath, coins)
	maxAllocRaw := dirs.FirstDefault(DirectiveMaxAlloc, "0")
	maxAlloc, err := strconv.ParseInt(maxAllocRaw, 10, 64)
	if err != nil {
		return "", fmt.Errorf("could not parse MAXALLOC directive: %w", err)
	}

	var opslog io.Writer
	if dirs.First(DirectiveRealm) != nil {
		opslog = new(bytes.Buffer)
	}

	// Create machine for execution and run test
	cw := opts.BaseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        &opts.outWriter,
		Store:         opts.TestStore.BeginTransaction(cw, cw, nil),
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
		Debug:         opts.Debug,
	})
	defer m.Release()
	result := opts.runTest(m, pkgPath, filename, source, opslog)

	// If there was a type-check error, return immediately.
	if result.TypeCheckError != "" {
		return "", fmt.Errorf("typecheck error: %s", result.TypeCheckError)
	}

	// updated tells whether the directives have been updated, and as such
	// a new generated filetest should be returned.
	// returnErr is used as the return value, and may be a MultiError if
	// multiple mismatches occurred.
	updated := false
	var returnErr error
	// match verifies the content against dir.Content; if different,
	// either updates dir.Content (for opts.Sync) or appends a new returnErr.
	match := func(dir *Directive, actual string) {
		// Remove end-of-line spaces, as these are removed from `fmt` in the filetests anyway.
		actual = reEndOfLineSpaces.ReplaceAllString(actual, "\n")
		if dir.Content != actual {
			if opts.Sync {
				dir.Content = actual
				updated = true
			} else {
				returnErr = multierr.Append(
					returnErr,
					fmt.Errorf("%s diff:\n%s", dir.Name, unifiedDiff(dir.Content, actual)),
				)
			}
		}
	}

	// First, check if we have an error, whether we're supposed to get it.
	if result.Error != "" {
		// Ensure this error was supposed to happen.
		errDirective := dirs.First(DirectiveError)
		if errDirective == nil {
			return "", fmt.Errorf("unexpected panic: %s\noutput:\n%s\nstacktrace:%s\nstack:\n%v",
				result.Error, result.Output, result.GnoStacktrace, string(result.GoPanicStack))
		}

		// The Error directive (and many others) will have one trailing newline,
		// which is not in the output - so add it there.
		match(errDirective, result.Error+"\n")
	} else if result.Output != "" {
		outputDirective := dirs.First(DirectiveOutput)
		if outputDirective == nil {
			return "", fmt.Errorf("unexpected output:\n%s", result.Output)
		}
	} else {
		err = m.CheckEmpty()
		if err != nil {
			return "", fmt.Errorf("machine not empty after main: %w", err)
		}
		if gno.HasDebugErrors() {
			return "", fmt.Errorf("got unexpected debug error(s): %v", gno.GetDebugErrors())
		}
	}

	// Check through each directive and verify it against the values from the test.
	for idx := range dirs {
		dir := &dirs[idx]
		switch dir.Name {
		case DirectiveOutput:
			if !strings.HasSuffix(result.Output, "\n") {
				result.Output += "\n"
			}
			match(dir, result.Output)
		case DirectiveRealm:
			res := opslog.(*bytes.Buffer).String()
			match(dir, res)
		case DirectiveEvents:
			events := m.Context.(*teststd.TestExecContext).EventLogger.Events()
			evtjson, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				panic(err)
			}
			evtstr := string(evtjson) + "\n"
			match(dir, evtstr)
		case DirectivePreprocessed:
			pn := m.Store.GetBlockNode(gno.PackageNodeLocation(pkgPath))
			pre := pn.(*gno.PackageNode).FileSet.Files[0].String() + "\n"
			match(dir, pre)
		case DirectiveStacktrace:
			match(dir, result.GnoStacktrace)
		}
	}

	if updated { // only true if sync == true
		return dirs.FileTest(), returnErr
	}

	return "", returnErr
}

func unifiedDiff(wanted, actual string) string {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(wanted),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})
	if err != nil {
		panic(fmt.Errorf("error generating unified diff: %w", err))
	}
	return diff
}

type runResult struct {
	Output string
	Error  string
	// Set if there was an issue with type-checking.
	TypeCheckError string
	// Set if there was a panic within gno code.
	GnoStacktrace string
	// Set if this was recovered from a panic.
	GoPanicStack []byte
}

func (opts *TestOptions) runTest(m *gno.Machine, pkgPath, filename string, content []byte, opslog io.Writer) (rr runResult) {
	pkgName := gno.Name(pkgPath[strings.LastIndexByte(pkgPath, '/')+1:])

	// Eagerly load imports.
	// This is executed using opts.Store, rather than the transaction store;
	// it allows us to only have to load the imports once (and re-use the cached
	// versions). Running the tests in separate "transactions" means that they
	// don't get the parent store dirty.
	if err := LoadImports(opts.TestStore, &gnovm.MemPackage{
		Name: string(pkgName),
		Path: pkgPath,
		Files: []*gnovm.MemFile{
			{Name: filename, Body: string(content)},
		},
	}); err != nil {
		// NOTE: we perform this here, so we can capture the runResult.
		if swe, ok := err.(*stackWrappedError); ok {
			return runResult{Error: err.Error(), GoPanicStack: swe.stack}
		}
		return runResult{Error: err.Error()}
	}

	// Reset and start capturing stdout.
	opts.filetestBuffer.Reset()
	revert := opts.outWriter.tee(&opts.filetestBuffer)
	defer revert()

	defer func() {
		if r := recover(); r != nil {
			rr.Output = opts.filetestBuffer.String()
			rr.GoPanicStack = debug.Stack()
			switch v := r.(type) {
			case *gno.TypedValue:
				rr.Error = v.Sprint(m)
			case *gno.PreprocessError:
				rr.Error = v.Unwrap().Error()
			case gno.UnhandledPanicError:
				rr.Error = v.Error()
				rr.GnoStacktrace = m.ExceptionStacktrace()
			default:
				rr.Error = fmt.Sprint(v)
				rr.GnoStacktrace = m.Stacktrace().String()
			}
		}
	}()

	// Use last element after / (works also if slash is missing).
	if !gno.IsRealmPath(pkgPath) {
		// Simple case - pure package.
		pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
		pv := pn.NewPackage()
		m.Store.SetBlockNode(pn)
		m.Store.SetCachePackage(pv)
		m.SetActivePackage(pv)
		m.Context.(*teststd.TestExecContext).OriginCaller = DefaultCaller
		n := gno.MustParseFile(filename, string(content))
		m.RunFiles(n)
		m.RunStatement(gno.StageRun, gno.S(gno.Call(gno.X("main"))))
	} else {
		// Realm case.
		gno.DisableDebug() // until main call.

		// Remove filetest from name, as that can lead to the package not being
		// parsed correctly when using RunMemPackage.
		filename = strings.ReplaceAll(filename, "_filetest", "")

		// save package using realm crawl procedure.
		memPkg := &gnovm.MemPackage{
			Name: string(pkgName),
			Path: pkgPath,
			Files: []*gnovm.MemFile{
				{
					Name: filename,
					Body: string(content),
				},
			},
		}
		orig, tx := m.Store, m.Store.BeginTransaction(nil, nil, nil)
		m.Store = tx

		// Validate Gno syntax and type check.
		if err := gno.TypeCheckMemPackageTest(memPkg, m.Store); err != nil {
			return runResult{TypeCheckError: err.Error()}
		}

		// Run decls and init functions.
		m.RunMemPackage(memPkg, true)
		// Clear store cache and reconstruct machine from committed info
		// (mimicking on-chain behaviour).
		tx.Write()
		m.Store = orig

		pv2 := m.Store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2) // XXX should it set the realm?
		m.Context.(*teststd.TestExecContext).OriginCaller = DefaultCaller
		gno.EnableDebug()
		// clear store.opslog from init function(s).
		m.Store.SetLogStoreOps(opslog) // resets.
		// Call main() like withrealm(main)().
		// This will switch the realm to the package.
		// main() must start with crossing().
		m.RunStatement(gno.StageRun, gno.S(gno.Call(gno.Call(gno.X("cross"), gno.X("main"))))) // switch realm.
	}

	return runResult{
		Output:        opts.filetestBuffer.String(),
		GnoStacktrace: m.Stacktrace().String(),
	}
}

// ---------------------------------------
// directives and directive parsing

const (
	// These directives are used to set input variables, which should change for
	// the specific filetest. They must be specified on a single line.
	DirectivePkgPath  = "PKGPATH"
	DirectiveMaxAlloc = "MAXALLOC"
	DirectiveSend     = "SEND"

	// These are used to match the result of the filetest against known golden
	// values.
	DirectiveOutput       = "Output"
	DirectiveError        = "Error"
	DirectiveRealm        = "Realm"
	DirectiveEvents       = "Events"
	DirectivePreprocessed = "Preprocessed"
	DirectiveStacktrace   = "Stacktrace"
)

// Directives contains the directives of a file.
// It may also contains directives with empty names, to indicate parts of the
// original source file (used to re-construct the filetest at the end).
type Directives []Directive

// First returns the first directive with the corresponding name.
func (d Directives) First(name string) *Directive {
	if name == "" {
		return nil
	}
	for i := range d {
		if d[i].Name == name {
			return &d[i]
		}
	}
	return nil
}

// FirstDefault returns the [Directive.Content] of First(name); if First(name)
// returns nil, then defaultValue is returned.
func (d Directives) FirstDefault(name, defaultValue string) string {
	v := d.First(name)
	if v == nil {
		return defaultValue
	}
	return v.Content
}

// FileTest re-generates the filetest from the given directives; the inverse of ParseDirectives.
func (d Directives) FileTest() string {
	var bld strings.Builder
	for _, dir := range d {
		switch {
		case dir.Name == "":
			bld.WriteString(dir.Content)
		case strings.ToUpper(dir.Name) == dir.Name: // is it all uppercase?
			bld.WriteString("// " + dir.Name + ": " + dir.Content + "\n")
		default:
			bld.WriteString("// " + dir.Name + ":\n")
			cnt := strings.TrimSuffix(dir.Content, "\n")
			lines := strings.Split(cnt, "\n")
			for _, line := range lines {
				if line == "" {
					bld.WriteString("//\n")
					continue
				}
				bld.WriteString("// ")
				bld.WriteString(strings.TrimRight(line, " "))
				bld.WriteByte('\n')
			}
		}
	}
	return bld.String()
}

// Directive represents a directive in a filetest.
// A [Directives] slice may also contain directives with empty Names:
// these compose the source file itself, and are used to re-construct the file
// when a directive is changed.
type Directive struct {
	Name    string
	Content string
}

// Allows either a `ALLCAPS: content` on a single line, or a `PascalCase:`,
// with content on the following lines.
var reDirectiveLine = regexp.MustCompile("^(?:([A-Z][a-z]*):|([A-Z]+): ?(.*))$")

// ParseDirectives parses all the directives in the filetest given at source.
func ParseDirectives(source io.Reader) (Directives, error) {
	sc := bufio.NewScanner(source)
	parsed := make(Directives, 0, 8)
	for sc.Scan() {
		// Re-append trailing newline.
		// Useful as we always use it anyway.
		txt := sc.Text() + "\n"
		if !strings.HasPrefix(txt, "//") {
			if len(parsed) == 0 || parsed[len(parsed)-1].Name != "" {
				parsed = append(parsed, Directive{Content: txt})
				continue
			}
			parsed[len(parsed)-1].Content += txt
			continue
		}

		comment := txt[2 : len(txt)-1]             // leading double slash, trailing \n
		comment = strings.TrimPrefix(comment, " ") // leading space (if any)

		// If we're already in a directive, simply append there.
		if len(parsed) > 0 && parsed[len(parsed)-1].Name != "" {
			parsed[len(parsed)-1].Content += comment + "\n"
			continue
		}

		// Find if there is a colon (indicating a possible directive).
		subm := reDirectiveLine.FindStringSubmatch(comment)
		switch {
		case subm == nil:
			// Not found; append to parsed as a line, or to the previous
			// directive if it exists.
			if len(parsed) == 0 {
				parsed = append(parsed, Directive{Content: txt})
				continue
			}
			last := &parsed[len(parsed)-1]
			if last.Name == "" {
				last.Content += txt
			} else {
				last.Content += comment + "\n"
			}
		case subm[1] != "": // output directive, with content on newlines
			parsed = append(parsed, Directive{Name: subm[1]})
		default: // subm[2] != "", all caps
			parsed = append(parsed,
				Directive{Name: subm[2], Content: subm[3]},
				// enforce new directive later
				Directive{},
			)
		}
	}
	return parsed, sc.Err()
}
