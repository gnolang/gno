package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/std"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/multierr"
)

type FileTestOptions struct {
	Store     gno.Store
	BaseStore storetypes.CommitStore

	// The Store should reference this, for instance in gonative definitions of
	// os.Stdout or fmt.Println. It is Reset on each run.
	Stdout bytes.Buffer
}

// RunSync executes the program in source as a filetest.
// RunSync returns additionally a string, which is the updated filetest should its "golden"
// outputs, error or other directives change.
func (opts *FileTestOptions) RunSync(filename string, source []byte) (string, error) {
	return opts.run(filename, source, true)
}

// Run executes the program in source as a filetest.
func (opts *FileTestOptions) Run(filename string, source []byte) error {
	_, err := opts.run(filename, source, false)
	return err
}

func (opts *FileTestOptions) run(filename string, source []byte, sync bool) (string, error) {
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
	ctx := TestContext(
		pkgPath,
		coins,
	)
	maxAllocRaw := dirs.FirstDefault(DirectiveMaxAlloc, "0")
	maxAlloc, err := strconv.ParseInt(maxAllocRaw, 10, 64)
	if err != nil {
		return "", fmt.Errorf("could not parse MAXALLOC directive: %w", err)
	}

	// Create machine for execution and run test
	cw := opts.BaseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        &opts.Stdout,
		Store:         opts.Store.BeginTransaction(cw, cw),
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
	})
	defer m.Release()
	result := opts.runTest(m, pkgPath, filename, source)

	// These are used to generate the output.
	updated := false
	var returnErr error
	match := func(dir *Directive, actual string) {
		if dir.Content != actual {
			if sync {
				dir.Content = actual
				updated = true
			} else {
				returnErr = multierr.Append(returnErr,
					fmt.Errorf("%s diff:\n%s", dir.Name, unifiedDiff(dir.Content, actual)))
			}
		}
	}

	// First, check if we have an error, whether we're supposed to get it.
	if result.Error != "" {
		// ensure this error was supposed to happen.
		errDirective := dirs.First(DirectiveError)
		if errDirective == nil {
			return "", fmt.Errorf("unexpected panic: %s\noutput:\n%s\nstack:\n%v",
				result.Error, result.Output, string(result.GoPanicStack))
		}

		// The Error directive will have one trailing newline, which is not in
		// the output - so add it there.
		match(errDirective, result.Error+"\n")
	} else {
		err = m.CheckEmpty()
		if err != nil {
			return "", fmt.Errorf("machine not empty after main: %w", err)
		}
		if gno.HasDebugErrors() {
			return "", fmt.Errorf("got unexpected debug error(s): %v", gno.GetDebugErrors())
		}
	}

	for idx := range dirs {
		dir := &dirs[idx]
		switch dir.Name {
		case DirectiveOutput:
			match(dir, result.Output)
		case DirectiveRealm:
			sops := m.Store.SprintStoreOps() + "\n"
			match(dir, sops)
		case DirectiveEvents:
			events := m.Context.(*teststd.TestExecContext).EventLogger.Events()
			evtjson, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				panic(err)
			}
			evtstr := string(evtjson)
			match(dir, evtstr)
		case DirectivePreprocessed:
			pn := m.Store.GetBlockNode(gno.PackageNodeLocation(pkgPath))
			pre := pn.(*gno.PackageNode).FileSet.Files[0].String() + "\n"
			match(dir, pre)
		case DirectiveStacktrace:
			match(dir, result.GnoStacktrace)
		}
	}

	if updated {
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
	// Set if there was a panic within gno code.
	GnoStacktrace string
	// Set if this was recovered from a panic.
	GoPanicStack []byte
}

func (opts *FileTestOptions) loadImports(filename string, content []byte) (rr runResult) {
	defer func() {
		// This is slightly different from the handling below; we do not have a
		// machine to work with, as this comes from an import; so we need
		// "machine-less" alternatives. (like v.String instead of v.Sprint)
		if r := recover(); r != nil {
			rr.GoPanicStack = debug.Stack()
			switch v := r.(type) {
			case *gno.TypedValue:
				rr.Error = v.String()
			case *gno.PreprocessError:
				rr.Error = v.Unwrap().Error()
			case gno.UnhandledPanicError:
				rr.Error = v.Error()
			default:
				rr.Error = fmt.Sprint(v)
			}
		}
	}()

	fl, err := parser.ParseFile(token.NewFileSet(), filename, content, parser.ImportsOnly)
	if err != nil {
		return runResult{Error: fmt.Sprintf("parse failure: %v", err)}
	}
	for _, imp := range fl.Imports {
		impPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return runResult{Error: fmt.Sprintf("unexpected invalid import path: %v", impPath)}
		}
		pkg := opts.Store.GetPackage(impPath, true)
		if pkg == nil {
			return runResult{Error: fmt.Sprintf("package not found: %v", impPath)}
		}
	}
	return runResult{}
}

func (opts *FileTestOptions) runTest(m *gno.Machine, pkgPath, filename string, content []byte) (rr runResult) {
	// Eagerly load imports.
	// This is executed using opts.Store, rather than the transaction store;
	// it allows us to only have to load the imports once (and re-use the cached
	// versions). Running the tests in separate "transactions" means that they
	// don't get the parent store dirty.
	if importRes := opts.loadImports(filename, content); importRes.Error != "" {
		return importRes
	}

	// imports loaded - reset stdout.
	opts.Stdout.Reset()

	defer func() {
		if r := recover(); r != nil {
			rr.Output = opts.Stdout.String()
			rr.GoPanicStack = debug.Stack()
			switch v := r.(type) {
			case *gno.TypedValue:
				rr.Error = v.Sprint(m)
			case *gno.PreprocessError:
				rr.Error = v.Unwrap().Error()
			case gno.UnhandledPanicError:
				rr.Error = v.Error()
				rr.GnoStacktrace = m.ExceptionsStacktrace()
			default:
				rr.Error = fmt.Sprint(v)
				rr.GnoStacktrace = m.Stacktrace().String()
			}
		}
	}()

	// use last element after / (works also if slash is missing)
	pkgName := gno.Name(pkgPath[strings.LastIndexByte(pkgPath, '/')+1:])
	if !gno.IsRealmPath(pkgPath) {
		// simple case - pure package.
		pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
		pv := pn.NewPackage()
		m.Store.SetBlockNode(pn)
		m.Store.SetCachePackage(pv)
		m.SetActivePackage(pv)
		n := gno.MustParseFile(filename, string(content))
		m.RunFiles(n)
		m.RunStatement(gno.S(gno.Call(gno.X("main"))))
	} else {
		// realm case.
		gno.DisableDebug() // until main call.

		// remove filetest from name, as that can lead to the package not being
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
		orig, tx := m.Store, m.Store.BeginTransaction(nil, nil)
		m.Store = tx
		// run decls and init functions.
		m.RunMemPackage(memPkg, true)
		// clear store cache and reconstruct machine from committed info
		// (mimicking on-chain behaviour).
		tx.Write()
		m.Store = orig

		pv2 := m.Store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2)
		gno.EnableDebug()
		// clear store.opslog from init function(s),
		// and PreprocessAllFilesAndSaveBlockNodes().
		m.Store.SetLogStoreOps(true) // resets.
		m.RunStatement(gno.S(gno.Call(gno.X("main"))))
	}

	return runResult{
		Output:        opts.Stdout.String(),
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
			cnt = strings.ReplaceAll(cnt, "\n", "\n// ")
			bld.WriteString("// ")
			bld.WriteString(cnt)
			bld.WriteByte('\n')
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

		// Find if there is a colon (indicating a possible directive)
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
				continue
			}
			// append to last line's content.
			last.Content += comment + "\n"
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