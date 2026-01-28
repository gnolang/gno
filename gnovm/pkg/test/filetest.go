package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststdlibs "github.com/gnolang/gno/gnovm/tests/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/multierr"
)

// RunFiletest executes a gnovm internal filetest in test/files.
// If opts.Sync is enabled, and the filetest's golden output has changed,
// the first string is set to the new generated content of the file.
// Before the filetest is run it will be type-checked.
func (opts *TestOptions) RunFiletest(fname string, source []byte, tgs gno.Store) (string, error) {
	opts.outWriter.w = opts.Output
	opts.outWriter.errW = opts.Error
	tcheck := true // Go type-check filetests in test/files.
	return opts.runFiletest(fname, source, tgs, tcheck)
}

// tcheck: only set to false pkg/test.Test(), since `gno test`
// (cmd/gno/test.go) already type-checked the whole package.
// Go type-checking in filetests is only available for gnovm internal filetests
// in test/files.
func (opts *TestOptions) runFiletest(fname string, source []byte, tgs gno.Store, tcheck bool) (string, error) {
	dirs, err := ParseDirectives(bytes.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("error parsing directives: %w", err)
	}

	// Sanity check: type-check directives are not available
	// with `gno test` of user packages.
	if !tcheck && dirs.FirstDefault(DirectiveTypeCheckError, "") != "" {
		panic("type-check error directive is only available for gnovm internal test files")
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
	gasMeter := store.NewInfiniteGasMeter()
	// Create machine for execution and run test
	tcw := opts.BaseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        &opts.outWriter,
		Store:         tgs.BeginTransaction(tcw, tcw, gasMeter),
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
		GasMeter:      gasMeter,
		Debug:         opts.Debug,
		ReviveEnabled: true,
	})
	defer m.Release()

	// RUN THE FILETEST /////////////////////////////////////
	result := opts.runTest(m, pkgPath, fname, source, opslog, tcheck)

	// updated tells whether the directives have been updated, and as such
	// a new generated filetest should be returned.
	// returnErr is used as the return value, and may be a MultiError if
	// multiple mismatches occurred.
	updated := false
	var returnErr error
	// `match` verifies the content against dir.Content; if different,
	// either updates dir.Content (for opts.Sync) or appends a new returnErr.
	match := func(dir *Directive, actual string) {
		content := dir.Content
		actual = strings.TrimRight(actual, "\n")
		content = strings.TrimRight(content, "\n")
		if content != actual {
			if opts.Sync {
				dir.Content = actual
				updated = true
			} else {
				if dir.Name == DirectiveError {
					returnErr = multierr.Append(
						returnErr,
						fmt.Errorf("%s diff:\n%s\nstacktrace:\n%s\nstack:\n%v",
							dir.Name, unifiedDiff(content, actual),
							result.GnoStacktrace, string(result.GoPanicStack)),
					)
				} else {
					returnErr = multierr.Append(
						returnErr,
						fmt.Errorf("%s diff:\n%s", dir.Name, unifiedDiff(content, actual)),
					)
				}
			}
		}
	}

	// Ensure needed the directives are present.
	if result.Error != "" {
		// Ensure this error was supposed to happen.
		errDirective := dirs.First(DirectiveError)
		if errDirective == nil {
			if opts.Sync {
				dirs = append(dirs, Directive{
					Name:    DirectiveError,
					Content: "",
				})
			} else {
				return "", fmt.Errorf("unexpected panic: %s\noutput:\n%s\nstacktrace:\n%s\nstack:\n%v",
					result.Error, result.Output, result.GnoStacktrace, string(result.GoPanicStack))
			}
		}
	} else if result.Output != "" {
		outputDirective := dirs.First(DirectiveOutput)
		if outputDirective == nil {
			if opts.Sync {
				dirs = append(dirs, Directive{
					Name:    DirectiveOutput,
					Content: "",
				})
			} else {
				return "", fmt.Errorf("unexpected output:\n%s", result.Output)
			}
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

	// Set to true if there was a go-typecheck directive..
	var hasTypeCheckErrorDirective bool

	// Check through each directive and verify it against the values from the test.
	for idx := range dirs {
		dir := &dirs[idx]
		switch dir.Name {
		case DirectiveOutput:
			match(dir, trimTrailingSpaces(result.Output))
		case DirectiveError:
			match(dir, result.Error)
		case DirectiveRealm:
			res := opslog.(*bytes.Buffer).String()
			match(dir, res)
		case DirectiveEvents:
			events := m.Context.(*teststdlibs.TestExecContext).EventLogger.Events()
			evtjson, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				panic(err)
			}
			evtstr := string(evtjson)
			match(dir, evtstr)
		case DirectivePreprocessed:
			pn := m.Store.GetBlockNodeSafe(gno.PackageNodeLocation(pkgPath))
			if pn == nil {
				return "", fmt.Errorf("package %q not preprocessed: %s", pkgPath, result.Error)
			}
			pre := pn.(*gno.PackageNode).FileSet.Files[0].String()
			match(dir, pre)
		case DirectiveStacktrace:
			match(dir, result.GnoStacktrace)
		case DirectiveGas:
			match(dir, strconv.FormatInt(m.GasMeter.GasConsumed(), 10))
		case DirectiveStorage:
			rlmDiff := realmDiffsString(m.Store.RealmStorageDiffs())
			match(dir, rlmDiff)
		case DirectiveTypeCheckError:
			hasTypeCheckErrorDirective = true
			match(dir, result.TypeCheckError)
		}
	}

	if !hasTypeCheckErrorDirective && result.TypeCheckError != "" {
		dir := Directive{
			Name:    DirectiveTypeCheckError,
			Content: "",
		}
		match(&dir, result.TypeCheckError)
		dirs = append(dirs, dir)
	}

	if updated { // only true if sync == true
		return dirs.FileTest(), returnErr
	}

	return "", returnErr
}

// returns a sorted string representation of realm diffs map
func realmDiffsString(m map[string]int64) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s: %d\n", k, m[k]))
	}
	return sb.String()
}

func trimTrailingSpaces(in string) string {
	lines := strings.Split(in, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " ")
		lines[i] = line
	}
	return strings.Join(lines, "\n")
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

func (opts *TestOptions) runTest(m *gno.Machine, pkgPath, fname string, content []byte, opslog io.Writer, tcheck bool) (rr runResult) {
	pkgName := gno.Name(pkgPath[strings.LastIndexByte(pkgPath, '/')+1:])
	tcError := ""
	fname = filepath.Base(fname)
	if opts.tcCache == nil {
		opts.tcCache = make(gno.TypeCheckCache)
	}

	// Eagerly load imports.
	// LoadImports is run using opts.Store, rather than the transaction store;
	// it allows us to only have to load the imports once (and re-use the cached
	// versions). Running the tests in separate "transactions" means that they
	// don't get the parent store dirty.
	abortOnError := true
	if err := LoadImports(opts.TestStore, &std.MemPackage{
		Type: gno.MPFiletests,
		Name: string(pkgName),
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
			{Name: fname, Body: string(content)},
		},
	}, abortOnError); err != nil {
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
			rr.TypeCheckError = tcError
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

	// Remove filetest from name, as that can lead to the package not being
	// parsed correctly when using RunMemPackage.
	fname = strings.ReplaceAll(fname, "_filetest", "")

	// Use last element after / (works also if slash is missing).
	if !gno.IsRealmPath(pkgPath) { // Simple case - pure package.
		// Determine package type based on path
		mptype := gno.MPUserProd
		if strings.HasSuffix(pkgPath, "_test") {
			mptype = gno.MPUserIntegration
		}
		// Construct mem package for single filetest.
		mpkg := &std.MemPackage{
			Type: mptype,
			Name: string(pkgName),
			Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
				{Name: fname, Body: string(content)},
			},
		}
		// Validate Gno syntax and type check.
		if tcheck {
			if _, err := gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
				// Use Teststore to load imported packages,
				// mimicing the loading behavior with on-chain.
				// (if using m.Store, the realm package will
				// be preloaded during typecheck)
				Getter:     opts.TestStore,
				TestGetter: m.Store,
				Mode:       gno.TCLatestRelaxed,
				Cache:      opts.tcCache,
			}); err != nil {
				tcError = fmt.Sprintf("%v", err.Error())
			}
		}
		// Must parse before set pn&pv.
		fn := m.MustParseFile(fname, string(content))
		// Construct throwaway package and parse file.
		pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
		pv := pn.NewPackage(m.Alloc)
		m.Store.SetBlockNode(pn)
		m.Store.SetCachePackage(pv)
		m.SetActivePackage(pv)
		m.Context.(*teststdlibs.TestExecContext).OriginCaller = DefaultCaller
		// Run (add) file, and then run main().
		m.RunFiles(fn)
		m.RunMain()
	} else { // Realm case.
		gno.DisableDebug() // until main call.

		// Save package using realm crawl procedure.
		// Realms are always MPUserProd because they need to be stored
		mpkg := &std.MemPackage{
			Type: gno.MPUserProd,
			Name: string(pkgName),
			Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
				{Name: fname, Body: string(content)},
			},
		}
		// Start transaction store.
		orig, txs := m.Store, m.Store.BeginTransaction(nil, nil, nil)
		m.Store = txs
		// Validate Gno syntax and type check.
		if tcheck {
			if _, err := gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
				Getter:     m.Store,
				TestGetter: m.Store,
				Mode:       gno.TCLatestRelaxed,
				Cache:      opts.tcCache,
			}); err != nil {
				tcError = fmt.Sprintf("%v", err.Error())
			}
		}
		// Run decls and init functions.
		m.RunMemPackage(mpkg, true)

		// Clear store cache and reconstruct machine from committed info
		// (mimicking on-chain behaviour).
		// (jae) why is this needed?
		txs.Write()
		m.Store = orig
		pv2 := m.Store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2)
		m.Context.(*teststdlibs.TestExecContext).OriginCaller = DefaultCaller
		gno.EnableDebug()

		// Clear store.opslog from init function(s).
		m.Store.SetLogStoreOps(opslog) // resets.
		m.RunMainMaybeCrossing()
	}
	return runResult{
		Output:         opts.filetestBuffer.String(),
		GnoStacktrace:  m.Stacktrace().String(),
		TypeCheckError: tcError,
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
	DirectiveOutput         = "Output"
	DirectiveError          = "Error"
	DirectiveRealm          = "Realm"
	DirectiveEvents         = "Events"
	DirectivePreprocessed   = "Preprocessed"
	DirectiveStacktrace     = "Stacktrace"
	DirectiveGas            = "Gas"
	DirectiveStorage        = "Storage"
	DirectiveTypeCheckError = "TypeCheckError"
)

var allDirectives = []string{
	DirectivePkgPath,
	DirectiveMaxAlloc,
	DirectiveSend,
	DirectiveOutput,
	DirectiveError,
	DirectiveRealm,
	DirectiveEvents,
	DirectivePreprocessed,
	DirectiveStacktrace,
	DirectiveGas,
	DirectiveStorage,
	DirectiveTypeCheckError,
}

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
	for i, dir := range d {
		ll := ""
		if i < len(d)-1 {
			ll = "\n"
		}
		switch {
		case dir.Name == "":
			cnt := strings.TrimRight(dir.Content, "\n ")
			bld.WriteString(cnt + "\n" + ll)
		case strings.ToUpper(dir.Name) == dir.Name: // ALLCAPS:
			bld.WriteString("// " + dir.Name + ": " + dir.Content + ll)
		default:
			if dir.Content == "" || dir.Content == "\n" {
				continue
			}
			bld.WriteString("// " + dir.Name + ":\n")
			cnt := strings.TrimRight(dir.Content, "\n ")
			lines := strings.Split(cnt, "\n")
			for _, line := range lines {
				if line == "" {
					bld.WriteString("//\n")
					continue
				}
				bld.WriteString("// ")
				bld.WriteString(line)
				bld.WriteString("\n")
			}
			bld.WriteString(ll)
		}
	}
	return bld.String()
}

// Directive represents a directive in a filetest.
// A [Directives] slice may also contain directives with empty Names:
// these compose the source file itself, and are used to re-construct the file
// when a directive is changed.
type Directive struct {
	Name     string
	Content  string
	Complete bool
	LastLine string
}

// Allows either a `ALLCAPS: content` on a single line, or a `PascalCase:`,
// with content on the following lines.
var reDirectiveLine = regexp.MustCompile("^(?:([A-Za-z]*):|([A-Z]+): ?(.*))$")

// ParseDirectives parses all the directives in the filetest given at source.
func ParseDirectives(source io.Reader) (Directives, error) {
	sc := bufio.NewScanner(source)
	parsed := make(Directives, 0, 8)
	parsed = append(parsed, Directive{LastLine: "// FAUX: faux directive", Complete: true}) // faux directive.
	for sc.Scan() {
		last := &parsed[len(parsed)-1]
		txt := sc.Text()
		if !strings.HasPrefix(txt, "//") {
			// If we're already in an incomplete text directive, simply append there.
			if last.Name == "" && !last.Complete {
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			}
			// Otherwise make a new directive.
			parsed = append(parsed,
				Directive{
					Content:  txt + "\n",
					LastLine: txt,
				})
			continue
		}

		comment := txt[2:]                         // leading double slash
		comment = strings.TrimPrefix(comment, " ") // leading space (if any)

		// Special case if following an incomplete comment line,
		// always append to it even if it looks like `// TODO: ...`.
		if strings.HasPrefix(txt, "//") &&
			strings.HasPrefix(last.LastLine, "//") &&
			!last.Complete {
			if last.Name == "" {
				// Just append text to it.
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			} else {
				// Just append comment to it.
				last.Content += comment + "\n"
				last.LastLine = txt
				continue
			}
		}

		// Find if there is a colon (indicating a possible directive).
		subm := reDirectiveLine.FindStringSubmatch(comment)
		if subm != nil && slices.Contains(allDirectives, subm[1]) {
			// CamelCase directive.
			parsed = append(parsed,
				Directive{
					Name:     subm[1],
					LastLine: txt,
				})
			continue
		}
		if subm != nil && slices.Contains(allDirectives, subm[2]) {
			// APPCAPS directive.
			parsed = append(parsed,
				Directive{
					Name:     subm[2],
					Content:  subm[3],
					Complete: true,
				})
			continue
		}

		// Not a directive, just a comment.
		// If we're already in an incomplete directive, simply append there.
		if !last.Complete {
			if last.Name == "" {
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			} else {
				last.Content += comment + "\n"
				last.LastLine = txt
				continue
			}
		}
		// Otherwise make a new directive.
		parsed = append(parsed,
			Directive{
				Content:  txt + "\n",
				LastLine: txt,
			})
	}

	// Remove trailing (newline|space)* and filter empty directives.
	result := make([]Directive, 0, len(parsed))
	parsed = parsed[1:] // remove faux directive
	for _, dir := range parsed {
		content := dir.Content
		content = strings.TrimRight(content, "\n ")
		if content == "" {
			continue
		}
		dir.Content = content
		result = append(result, dir)
		// fmt.Printf("#%d %s: [[[%s]]]\n", i, dir.Name, dir.Content)
	}

	return result, sc.Err()
}
