package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	goio "io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/cmdutil"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

/*
	Linting.
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
*/

type lintCmd struct {
	verbose    bool
	rootDir    string
	autoGnomod bool
	// min_confidence: minimum confidence of a problem to print it
	// (default 0.8) auto-fix: apply suggested fixes automatically.
}

func newLintCmd(io commands.IO) *commands.Command {
	cmd := &lintCmd{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "runs the linter for the specified packages",
		},
		cmd,
		func(_ context.Context, args []string) error {
			return execLint(cmd, args, io)
		},
	)
}

func (c *lintCmd) RegisterFlags(fs *flag.FlagSet) {
	rootdir := gnoenv.RootDir()

	fs.BoolVar(&c.verbose, "v", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.autoGnomod, "auto-gnomod", true, "auto-generate gnomod.toml file if not already present")
}

func execLint(cmd *lintCmd, args []string, io commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	// Guess opts.RootDir.
	if cmd.rootDir == "" {
		cmd.rootDir = gnoenv.RootDir()
	}

	loadCfg := packages.LoadConfig{
		Fetcher:    testPackageFetcher,
		Deps:       true,
		Test:       true,
		Out:        io.Err(),
		AllowEmpty: true,
		GnoRoot:    cmd.rootDir,
	}
	pkgs, err := packages.Load(loadCfg, args...)
	if err != nil {
		return err
	}

	hasError := false

	prodbs, prodgs := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true, WithExtern: false, WithExamples: true, Testing: false, Packages: pkgs},
	)
	testbs, testgs := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{
			PreprocessOnly: true,
			WithExtern:     false,
			WithExamples:   true,
			Testing:        true,
			SourceStore:    prodgs,
			Packages:       pkgs,
		},
	)
	ppkgs := map[string]cmdutil.ProcessedPackage{}
	cache := make(gno.TypeCheckCache)

	if cmd.verbose {
		targetsNames := []string{}
		for _, pkg := range pkgs {
			if len(pkg.Match) == 0 {
				continue
			}
			targetsNames = append(targetsNames, lintTargetName(pkg))
		}
		io.ErrPrintfln("linting packages: %v", targetsNames)
	}
	//----------------------------------------
	// LINT STAGE 1: Typecheck and lint.
	for _, pkg := range pkgs {
		// ignore dependencies
		if len(pkg.Match) == 0 {
			continue
		}

		if cmd.verbose {
			io.ErrPrintfln("linting %q", lintTargetName(pkg))
		}

		// XXX Currently the linter only supports linting directories.
		// In order to support linting individual files, we need to
		// refactor this code to work with mempackages, not dirs, and
		// cmd/gno/util.go needs to be refactored to return mempackages
		// rather than dirs. Commands like `gno lint a.gno b.gno`
		// should create a temporary package from just those files. We
		// could also load mempackages lazily for memory efficiency.
		// Alternative: support `command-line-arguments` in packages.Load
		dir := pkg.Dir

		// Read and parse gnomod.toml directly.
		fpath := filepath.Join(dir, "gnomod.toml")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			// TODO: gno.mod is deprecated, but we still support it for now.
			// if gno.mod exists -> port
			if cmd.autoGnomod {
				modulePath, _ := determinePkgPath(nil, dir, cmd.rootDir)
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
		if err != nil {
			issue := gnoIssue{
				Code:       gnoGnoModError,
				Confidence: 1, // ??
				Location:   fpath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
			return commands.ExitCodeError(1)
		}

		// Skip processing for ignored modules
		if mod.Ignore {
			if cmd.verbose {
				io.ErrPrintfln("%s: module is ignored, skipping", dir)
			}
			continue
		}

		// See adr/pr4264_lint_transpile.md
		// LINT STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPAnyAll)
		if err != nil {
			printError(io.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Check package name matches path element.
		// This is also enforced in ValidateMemPackageAny, but we check here
		// to provide a specific lint error code before other processing.
		// See https://github.com/gnolang/gno/issues/1571
		if err := gno.ValidatePkgNameMatchesPath(gno.Name(mpkg.Name), mpkg.Path); err != nil {
			issue := gnoIssue{
				Code:       gnoPackageNameMismatchError,
				Confidence: 1,
				Location:   dir,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
			// Skip the remaining lint steps: type-checking would only
			// cascade-fail on the same mismatched package name.
			continue
		}

		// Perform imports using the parent store.
		abortOnError := true
		if err := test.LoadImports(testgs, mpkg, abortOnError); err != nil {
			printError(io.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Wrap in cache wrap so execution of the linter
		// doesn't impact other packages.
		newProdGnoStore := func() gno.Store {
			pcw := prodbs.CacheWrap()
			pgs := prodgs.BeginTransaction(pcw, pcw, nil, nil)
			return pgs
		}
		injectTmpkg := func(tgs gno.Store) {
			// NOTE: if we don't do it lazily like this, otherwise there
			// needs to be a hook from original store creation
			// (complicated), or, if not done lazily we won't get the Go
			// typecheck error we prefer.
			tgetter := tgs.GetPackageGetter()
			tgs.SetPackageGetter(func(pkgPath string, store gno.Store) (
				*gno.PackageNode, *gno.PackageValue,
			) {
				if pkgPath == mpkg.Path {
					tmpkg := gno.MPFTest.FilterMemPackage(mpkg)
					m2 := gno.NewMachineWithOptions(gno.MachineOptions{
						PkgPath:     pkgPath,
						Output:      goio.Discard,
						Store:       tgs,
						SkipPackage: true,
					})
					// Use the actual type of the filtered package
					tmpkgType := tmpkg.Type.(gno.MemPackageType)
					m2.Store.AddMemPackage(tmpkg, tmpkgType)
					return m2.PreprocessFiles(tmpkg.Name, tmpkg.Path,
						m2.ParseMemPackageAsType(tmpkg, tmpkgType), true, true)
				} else {
					return tgetter(pkgPath, store)
				}
			})
		}
		newTestGnoStore := func(withTmpkg bool) gno.Store {
			tcw := testbs.CacheWrap()
			tgs := testgs.BeginTransaction(tcw, tcw, nil, nil)
			if withTmpkg {
				injectTmpkg(tgs)
			}
			return tgs
		}

		// Handle runtime errors
		didPanic := catchPanic(dir, pkgPath, io.Err(), func() {
			// Memo process results here.
			ppkg := cmdutil.ProcessedPackage{MPkg: mpkg, Dir: dir}

			// Run type checking
			// LINT STEP 2: ParseGnoMod()
			// STEP 3: GoParse*()
			//
			// lintTypeCheck(mpkg) -->
			//   TypeCheckMemPackage(mpkg) -->
			//     imp.typeCheckMemPackage(mpkg)
			//       ParseGnoMod(mpkg);
			//       GoParseMemPackage(mpkg);
			//       g.cmd.Check();

			tcmode := gno.TCLatestStrict
			if cmd.autoGnomod {
				tcmode = gno.TCLatestRelaxed
			}
			tcFset := token.NewFileSet()
			var lintFiles []*ast.File
			var lintInfo *types.Info
			tcPkg, errs := lintTypeCheck(io, dir, mpkg, gno.TypeCheckOptions{
				Getter:     newProdGnoStore(),
				TestGetter: newTestGnoStore(true),
				Mode:       tcmode,
				Cache:      cache,
				Fset:       tcFset,
				LintInfo: func(files []*ast.File, info *types.Info) {
					lintFiles, lintInfo = files, info
				},
			})
			if errs != nil {
				// io.ErrPrintln(errs) printed above.
				hasError = true
				return
			}

			// other realms cannot hold this realm's struct values (warning)
			lintRealmValueAPI(io, tcPkg, tcFset)
			lintRealmBoxedResult(io, tcPkg, tcFset, lintFiles, lintInfo)
			lintRealmFuncParams(io, tcPkg, tcFset, lintFiles, lintInfo)

			// ensure the 'Render' function is correct
			err = lintRenderSignature(io, tcPkg, tcFset)
			if err != nil {
				// io.ErrPrintln(err) printed above.
				hasError = true
				return
			}

			// Construct machine for testing.
			tm := test.Machine(newProdGnoStore(), goio.Discard, pkgPath, false, nil)
			defer tm.Release()

			// LINT STEP 4: re-parse for preprocessor.
			// While lintTypeCheck > TypeCheckMemPackage will find
			// most issues, the preprocessor may have additional
			// checks.
			// Gno parse source fileset and test filesets.
			_, fset, tfset, _tests, ftests := sourceAndTestFileset(mpkg, false)

			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess fset files (no test files)
				tm.Store = newProdGnoStore()
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false)
				ppkg.AddNormal(pn, fset)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess fset files (w/ some *_test.gno).
				tm.Store = newTestGnoStore(false)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, tfset, false, false)
				ppkg.AddTest(pn, fset)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _test files (all xxx_test *_test.gno).
				tm.Store = newTestGnoStore(true)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name+"_test", mpkg.Path+"_test", _tests, false, false)
				ppkg.AddUnderscoreTests(pn, _tests)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _filetest.gno files.
				for i, fset := range ftests {
					tm.Store = newTestGnoStore(true)
					fname := fset.Files[0].FileName
					mfile := mpkg.GetFile(fname)
					pkgPath := fmt.Sprintf("%s_filetest%d", mpkg.Path, i)
					pkgPath, err = parsePkgPathDirective(mfile.Body, pkgPath)
					if err != nil {
						io.ErrPrintln(err)
						hasError = true
						continue
					}
					// A filetest may assert a preprocess-time failure
					// via // Error:; in that case, swallow any panic
					// here. Exact-message verification is `gno test`'s
					// job — lint just needs to not flag the expected
					// failure as a lint error.
					expectsErr, derr := hasErrorDirective(mfile.Body)
					if derr != nil {
						io.ErrPrintln(derr)
						hasError = true
						continue
					}
					pkgName := string(fset.Files[0].PkgName)
					func() {
						if expectsErr {
							defer func() { _ = recover() }()
						}
						pn, _ := tm.PreprocessFiles(pkgName, pkgPath, fset, false, false)
						ppkg.AddFileTest(pn, fset)
					}()
				}
			}

			// Record results.
			ppkgs[dir] = ppkg
		})
		if didPanic {
			hasError = true
		}
	}
	if hasError {
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// LINT STAGE 2: Write.
	// Must be a separate stage to prevent partial writes.
	for _, pkg := range pkgs {
		// ignore dependencies
		if len(pkg.Match) == 0 {
			continue
		}

		ppkg, ok := ppkgs[pkg.Dir]
		if !ok {
			// Skip directories that were not processed (e.g., ignored modules)
			continue
		}

		// LINT STEP 6: write the processed mempackage back, keeping
		// filetests in their filetests/ subdir (mempackage file names
		// are flat; the generic MPkg.WriteTo would scatter them into
		// the package root).
		err := gno.WriteMemPackageTo(ppkg.MPkg, pkg.Dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Wrapper around TypeCheckMemPackage() to io.ErrPrintln(gnoIssue{}).
// Prints and returns errors. Panics upon an unexpected error.
func lintTypeCheck(
	// Args:
	io commands.IO,
	dir string,
	mpkg *std.MemPackage,
	opts gno.TypeCheckOptions) (
	// Results:
	tcPkg *types.Package,
	lerr error,
) {
	// gno.TypeCheckMemPackage(mpkg, testStore).
	tcPkg, tcErrs := gno.TypeCheckMemPackage(mpkg, opts)

	// Print errors, and return the first unexpected error.
	errors := multierr.Errors(tcErrs)
	for _, err := range errors {
		printError(io.Err(), dir, mpkg.Path, err)
	}

	lerr = tcErrs
	return
}

func lintTargetName(pkg *packages.Package) string {
	if pkg.ImportPath != "" {
		return pkg.ImportPath
	}

	return tryRelativizePath(pkg.Dir)
}

// lintRenderSignature checks if a Render function in the package has the expected signature
// Returns error if the signature is incorrect.
func lintRenderSignature(io commands.IO, pkg *types.Package, fset *token.FileSet) error {
	// ignore pure package and ephemeral realms
	if pkg == nil || !gno.IsRealmPath(pkg.Path()) {
		return nil
	}

	o := pkg.Scope().Lookup("Render")
	if o == nil {
		return nil
	}

	fn, ok := o.(*types.Func)
	if !ok {
		return nil
	}

	s, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil
	}

	if s.Recv() != nil {
		return nil
	}

	// Accept both the legacy non-crossing form `Render(string) string`
	// and the crossing form `Render(cur realm, string) string`. The
	// crossing form is auto-injected with .cur by the chain query layer
	// (see MaybeInjectCurForEval).
	params, results := s.Params(), s.Results()
	isString := func(t *types.Tuple, i int) bool {
		return t != nil && i < t.Len() && t.At(i) != nil && t.At(i).Type().String() == "string"
	}
	isRealm := func(t *types.Tuple, i int) bool {
		if t == nil || i >= t.Len() || t.At(i) == nil {
			return false
		}
		// gno's `realm` type is exposed via the gnobuiltins shim as
		// gnobuiltins/gno0p9.realm (alias .Realm). Match by suffix to
		// tolerate qualifier variants.
		ts := t.At(i).Type().String()
		return strings.HasSuffix(ts, ".realm") || strings.HasSuffix(ts, ".Realm")
	}

	validResults := results != nil && results.Len() == 1 && isString(results, 0)
	validNonCrossing := params != nil && params.Len() == 1 && isString(params, 0)
	validCrossing := params != nil && params.Len() == 2 && isRealm(params, 0) && isString(params, 1)

	if !validResults || (!validNonCrossing && !validCrossing) {
		location := pkg.Path()
		if fset != nil && fn.Pos().IsValid() {
			pos := fset.Position(fn.Pos())
			location = fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
		}

		err := fmt.Errorf("invalid signature for the realm's Render function; must be `func Render(string) string` or `func Render(cur realm, string) string`")
		fmt.Fprintln(io.Err(), gnoIssue{
			Code:       gnoLintError,
			Msg:        err.Error(),
			Confidence: 1,
			Location:   location,
		})
		return err
	}

	return nil
}

// lintRealmValueAPI reports exported functions, methods and variables of a
// realm that pass the realm's own struct or array types by value: no other
// realm may hold such a value, so callers can only read fields of it.
func lintRealmValueAPI(io commands.IO, pkg *types.Package, fset *token.FileSet) {
	if pkg == nil || !gno.IsRealmPath(pkg.Path()) {
		return
	}
	report := func(obj types.Object, what string, t types.Type) {
		location := pkg.Path()
		if fset != nil && obj.Pos().IsValid() {
			pos := fset.Position(obj.Pos())
			location = fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
		}
		fmt.Fprintln(io.Err(), gnoIssue{
			Code:       gnoLintWarning,
			Msg:        fmt.Sprintf("exported %s %s passes %s by value; other realms cannot hold it (use a pointer)", what, obj.Name(), t),
			Confidence: 1,
			Location:   location,
		})
	}
	checkSig := func(obj types.Object, sig *types.Signature, what string) {
		for _, tup := range []*types.Tuple{sig.Params(), sig.Results()} {
			for i := 0; i < tup.Len(); i++ {
				if t := ownRealmValue(tup.At(i).Type(), pkg); t != nil {
					report(obj, what, t)
					return
				}
			}
		}
	}
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Func:
			checkSig(o, o.Type().(*types.Signature), "function")
		case *types.Var:
			if t := ownRealmValue(o.Type(), pkg); t != nil {
				report(o, "variable", t)
			}
		case *types.TypeName:
			if named, ok := o.Type().(*types.Named); ok {
				for i := 0; i < named.NumMethods(); i++ {
					if m := named.Method(i); m.Exported() {
						checkSig(m, m.Type().(*types.Signature), "method")
					}
				}
			}
		}
	}
}

// ownRealmValue returns a struct/array type declared in pkg that t holds by value.
func ownRealmValue(t types.Type, pkg *types.Package) types.Type {
	switch u := types.Unalias(t).(type) {
	case *types.Named:
		if u.Obj().Pkg() == pkg {
			switch u.Underlying().(type) {
			case *types.Struct, *types.Array:
				return u
			}
		}
	case *types.Array:
		return ownRealmValue(u.Elem(), pkg)
	case *types.Slice:
		return ownRealmValue(u.Elem(), pkg)
	case *types.Map:
		if k := ownRealmValue(u.Key(), pkg); k != nil {
			return k
		}
		return ownRealmValue(u.Elem(), pkg)
	case *types.Chan:
		return ownRealmValue(u.Elem(), pkg)
	case *types.Struct:
		for i := 0; i < u.NumFields(); i++ {
			if ft := ownRealmValue(u.Field(i).Type(), pkg); ft != nil {
				return ft
			}
		}
	}
	return nil
}

// lintRealmBoxedResult warns on exported functions and methods that return
// one of the realm's own struct or array values boxed in an interface result.
func lintRealmBoxedResult(io commands.IO, pkg *types.Package, fset *token.FileSet, files []*ast.File, info *types.Info) {
	if pkg == nil || info == nil || !gno.IsRealmPath(pkg.Path()) {
		return
	}
	decls := funcDecls(files, info)
	report := func(obj types.Object, what string, t, iface types.Type) {
		location := pkg.Path()
		if fset != nil && obj.Pos().IsValid() {
			pos := fset.Position(obj.Pos())
			location = fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
		}
		fmt.Fprintln(io.Err(), gnoIssue{
			Code:       gnoLintWarning,
			Msg:        fmt.Sprintf("exported %s %s returns %s by value inside %s; other realms cannot hold it (return a pointer)", what, obj.Name(), t, iface),
			Confidence: 1,
			Location:   location,
		})
	}
	walkExportedFuncs(pkg, func(fn *types.Func) {
		fd := decls[fn]
		if fd == nil || fd.Body == nil {
			return
		}
		sig := fn.Type().(*types.Signature)
		results := sig.Results()
		if results.Len() == 0 {
			return
		}
		what := "function"
		if sig.Recv() != nil {
			what = "method"
		}
		reported := false
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if reported {
				return false
			}
			ret, ok := n.(*ast.ReturnStmt)
			if !ok || len(ret.Results) != results.Len() {
				return true // bare or multi-value spread returns are not followed
			}
			for i, expr := range ret.Results {
				rt := results.At(i).Type()
				if !types.IsInterface(rt) {
					continue
				}
				if t := ownRealmValue(info.TypeOf(expr), pkg); t != nil {
					report(fn, what, t, rt)
					reported = true
					return false
				}
			}
			return true
		})
	})
}

// funcDecls maps each declared function or method object to its FuncDecl.
func funcDecls(files []*ast.File, info *types.Info) map[types.Object]*ast.FuncDecl {
	m := map[types.Object]*ast.FuncDecl{}
	for _, f := range files {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name == nil {
				continue
			}
			if obj := info.Defs[fd.Name]; obj != nil {
				m[obj] = fd
			}
		}
	}
	return m
}

// walkExportedFuncs visits the exported functions and the exported methods of
// exported types, the same set lintRealmValueAPI judges.
func walkExportedFuncs(pkg *types.Package, fn func(*types.Func)) {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Func:
			fn(o)
		case *types.TypeName:
			if named, ok := o.Type().(*types.Named); ok {
				for i := 0; i < named.NumMethods(); i++ {
					if m := named.Method(i); m.Exported() {
						fn(m)
					}
				}
			}
		}
	}
}

// lintRealmFuncParams warns on crossing functions that call a function value
// reached through a parameter of non-function type: a caller can hand in this
// realm's own stored functions that way.
func lintRealmFuncParams(io commands.IO, pkg *types.Package, fset *token.FileSet, files []*ast.File, info *types.Info) {
	if pkg == nil || info == nil || !gno.IsRealmPath(pkg.Path()) {
		return
	}
	for _, file := range files {
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil || !isCrossingFunc(fd, info) {
				continue
			}
			lintCrossingCalls(io, pkg, fset, fd, info)
		}
	}
}

// isCrossingFunc: the first parameter is of type realm.
func isCrossingFunc(fd *ast.FuncDecl, info *types.Info) bool {
	if fd.Type.Params == nil || len(fd.Type.Params.List) == 0 {
		return false
	}
	return isRealmType(info.TypeOf(fd.Type.Params.List[0].Type))
}

func isRealmType(t types.Type) bool {
	if t == nil {
		return false
	}
	s := types.TypeString(t, nil)
	return strings.HasSuffix(s, ".realm") || strings.HasSuffix(s, ".Realm")
}

func isFuncType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := types.Unalias(t).Underlying().(*types.Signature)
	return ok
}

// lintCrossingCalls reports calls in a crossing function whose callee is a
// function value reached through a non-function parameter, directly or through
// one level of local assignment (index, field, deref, range, or a call result).
func lintCrossingCalls(io commands.IO, pkg *types.Package, fset *token.FileSet, fd *ast.FuncDecl, info *types.Info) {
	nonFunc := map[string]bool{}
	if fd.Type.Params != nil {
		for _, field := range fd.Type.Params.List {
			t := info.TypeOf(field.Type)
			for _, n := range field.Names {
				if n.Name != "_" {
					nonFunc[n.Name] = !isFuncType(t)
				}
			}
		}
	}
	origin := map[string]string{}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range x.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || i >= len(x.Rhs) {
					continue
				}
				if p, ok := reachedParam(x.Rhs[i], nonFunc, origin, info); ok {
					origin[id.Name] = p
				}
			}
		case *ast.ValueSpec:
			for i, id := range x.Names {
				if i >= len(x.Values) {
					continue
				}
				if p, ok := reachedParam(x.Values[i], nonFunc, origin, info); ok {
					origin[id.Name] = p
				}
			}
		case *ast.RangeStmt:
			p, ok := paramOrigin(x.X, nonFunc, origin)
			if !ok {
				return true
			}
			for _, e := range []ast.Expr{x.Key, x.Value} {
				if id, ok := e.(*ast.Ident); ok && id.Name != "_" {
					origin[id.Name] = p
				}
			}
		}
		return true
	})
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		p, ok := calleeParam(call.Fun, nonFunc, origin, info)
		if !ok {
			return true
		}
		location := pkg.Path()
		if fset != nil && call.Pos().IsValid() {
			pos := fset.Position(call.Pos())
			location = fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
		}
		fmt.Fprintln(io.Err(), gnoIssue{
			Code:       gnoLintWarning,
			Msg:        fmt.Sprintf("crossing function %s calls a function reached through parameter %s; a caller can pass this realm's own stored functions (take the function itself, or look it up by key)", fd.Name.Name, p),
			Confidence: 1,
			Location:   location,
		})
		return true
	})
}

// reachedParam resolves an expression to the non-function parameter it reaches.
func reachedParam(e ast.Expr, nonFunc map[string]bool, origin map[string]string, info *types.Info) (string, bool) {
	switch x := ast.Unparen(e).(type) {
	case *ast.Ident:
		if p, ok := origin[x.Name]; ok {
			return p, true
		}
	case *ast.IndexExpr:
		return paramOrigin(x.X, nonFunc, origin)
	case *ast.StarExpr:
		return paramOrigin(x.X, nonFunc, origin)
	case *ast.SelectorExpr:
		if sel := info.Selections[x]; sel != nil && sel.Kind() == types.FieldVal {
			return paramOrigin(x.X, nonFunc, origin)
		}
	case *ast.CallExpr:
		return calleeOnParam(x.Fun, nonFunc, origin)
	}
	return "", false
}

// calleeOnParam: a call on a non-function parameter (a returned function).
func calleeOnParam(fun ast.Expr, nonFunc map[string]bool, origin map[string]string) (string, bool) {
	switch x := ast.Unparen(fun).(type) {
	case *ast.SelectorExpr:
		return paramOrigin(x.X, nonFunc, origin)
	case *ast.IndexExpr:
		return paramOrigin(x.X, nonFunc, origin)
	}
	return "", false
}

// calleeParam: the call's callee is a function value reached through a parameter.
func calleeParam(fun ast.Expr, nonFunc map[string]bool, origin map[string]string, info *types.Info) (string, bool) {
	switch x := ast.Unparen(fun).(type) {
	case *ast.Ident:
		if p, ok := origin[x.Name]; ok {
			return p, true
		}
	case *ast.IndexExpr:
		return paramOrigin(x.X, nonFunc, origin)
	case *ast.StarExpr:
		return paramOrigin(x.X, nonFunc, origin)
	case *ast.SelectorExpr:
		if sel := info.Selections[x]; sel != nil && sel.Kind() == types.FieldVal {
			return paramOrigin(x.X, nonFunc, origin)
		}
	}
	return "", false
}

// paramOrigin: an ident naming a non-function parameter or a local traced to one.
func paramOrigin(e ast.Expr, nonFunc map[string]bool, origin map[string]string) (string, bool) {
	id, ok := ast.Unparen(e).(*ast.Ident)
	if !ok {
		return "", false
	}
	if nonFunc[id.Name] {
		return id.Name, true
	}
	if p, ok := origin[id.Name]; ok {
		return p, true
	}
	return "", false
}
