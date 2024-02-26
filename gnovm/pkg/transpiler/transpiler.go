// Package transpiler implements a source-to-source compiler for translating Gno
// code into Go code.
package transpiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	goscanner "go/scanner"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ImportPrefix is the import path to the root of the gno repository, which should
// be used to create go import paths.
const ImportPrefix = "github.com/gnolang/gno"

// TranspileImportPath takes an import path s, and converts it into the full
// import path relative to the Gno repository.
func TranspileImportPath(s string) string {
	return ImportPrefix + "/" + PackageDirLocation(s)
}

// IsStdlib determines whether s is a pkgpath for a standard library.
func IsStdlib(s string) bool {
	// NOTE(morgan): this is likely to change in the future as we add support for
	// IBC/ICS and we allow import paths to other chains. It might be good to
	// follow the same rule as Go, which is: does the first element of the
	// import path contain a dot?
	return !strings.HasPrefix(s, "gno.land/")
}

// PackageDirLocation provides the supposed directory of the package, relative to the root dir.
//
// TODO(morgan): move out, this should go in a "resolver" package.
func PackageDirLocation(s string) string {
	switch {
	case !IsStdlib(s):
		return "examples/" + s
	default:
		return "gnovm/stdlibs/" + s
	}
}

// Result is returned by Transpile, returning the file's imports and output
// out the transpilation.
type Result struct {
	Imports    []*ast.ImportSpec
	Translated string
	File       *ast.File
}

// TODO: func TranspileFile: supports caching.
// TODO: func TranspilePkg: supports directories.

// TranspiledFilenameAndTags returns the filename and tags for transpiled files.
func TranspiledFilenameAndTags(gnoFilePath string) (targetFilename, tags string) {
	nameNoExtension := strings.TrimSuffix(filepath.Base(gnoFilePath), ".gno")
	switch {
	case strings.HasSuffix(gnoFilePath, "_filetest.gno"):
		tags = "gno && filetest"
		targetFilename = "." + nameNoExtension + ".gno.gen.go"
	case strings.HasSuffix(gnoFilePath, "_test.gno"):
		tags = "gno && test"
		targetFilename = "." + nameNoExtension + ".gno.gen_test.go"
	default:
		tags = "gno"
		targetFilename = nameNoExtension + ".gno.gen.go"
	}
	return
}

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// gnolang.Store, separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *std.MemPackage
}

// DefaultGetter is used by [TranspileAndCheckMempkg] when a nil getter is passed.
// It resolves paths on-the-fly from the root directory, using gnolang.ReadMemPackage.
// If rootDir == "", then it will be set to the value of gnoenv.RootDir.
func DefaultGetter(rootDir string) MemPackageGetter {
	if rootDir == "" {
		rootDir = gnoenv.RootDir()
	}
	return defaultGetter{rootDir}
}

type defaultGetter struct {
	rootDir string
}

func (dg defaultGetter) GetMemPackage(path string) *std.MemPackage {
	dir := filepath.Join(dg.rootDir, PackageDirLocation(path))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	defer func() {
		// TODO(morgan): use variant that doesn't panic. *rolls eyes*
		err := recover()
		if err != nil {
			log.Printf("import %q: %v", path, err)
		}
	}()
	return gnolang.ReadMemPackage(dir, path)
}

// TranspileAndCheckMempkg converts each of the files in mempkg to Go, and
// performs static checking using Go's type checker.
func TranspileAndCheckMempkg(mempkg *std.MemPackage, getter MemPackageGetter) error {
	if getter == nil {
		getter = DefaultGetter("")
	}
	imp := &transpImporter{
		getter: getter,
		cache:  map[string]interface{}{},
		cfg:    &types.Config{},
	}

	imp.cfg.Importer = imp
	_, err := imp.transpileParseMemPkg(mempkg)
	if err != nil {
		return err
	}

	return nil
}

type transpImporter struct {
	getter MemPackageGetter
	cache  map[string]any // *types.Package or error
	cfg    *types.Config
}

func (g *transpImporter) Import(path string) (*types.Package, error) {
	return g.ImportFrom(path, "", 0)
}

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (g *transpImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := g.cache[path]; ok {
		switch ret := pkg.(type) {
		case *types.Package:
			return ret, nil
		case error:
			return nil, ret
		default:
			panic(fmt.Sprintf("invalid type in transpImporter.cache %T", ret))
		}
	}
	mpkg := g.getter.GetMemPackage(path)
	if mpkg == nil {
		g.cache[path] = (*types.Package)(nil)
		return nil, nil
	}
	result, err := g.transpileParseMemPkg(mpkg)
	if err != nil {
		g.cache[path] = err
		return nil, err
	}
	g.cache[path] = result
	return result, nil
}

func (g *transpImporter) transpileParseMemPkg(mpkg *std.MemPackage) (*types.Package, error) {
	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(mpkg.Files))
	var errs error
	for _, file := range mpkg.Files {
		// include go files to have native bindings checked.
		if !strings.HasSuffix(file.Name, ".gno") && !strings.HasSuffix(file.Name, ".go") {
			continue // skip spurious file.
		}
		// TODO: because this is in-memory, could avoid header.
		res, err := transpileWithFset(fset, file.Body, "gno", file.Name)
		if err != nil {
			err = multierr.Append(errs, err)
		}
		files = append(files, res.File)
	}
	if errs != nil {
		return nil, errs
	}

	// TODO g.cfg.Error
	return g.cfg.Check(mpkg.Path, fset, files, nil)
}

// Transpile performs transpilation on the given source code. tags can be used
// to specify build tags; and filename helps generate useful error messages and
// discriminate between test and normal source files.
func Transpile(source, tags, filename string) (*Result, error) {
	fset := token.NewFileSet()
	return transpileWithFset(fset, source, tags, filename)
}

func transpileWithFset(fset *token.FileSet, source, tags, filename string) (*Result, error) {
	f, err := parser.ParseFile(fset, filename, source, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	isTestFile := strings.HasSuffix(filename, "_test.gno") || strings.HasSuffix(filename, "_filetest.gno")
	ctx := &transpileCtx{
		rootDir: gnoenv.RootDir(),
	}
	stdlibPrefix := filepath.Join(ctx.rootDir, "gnovm", "stdlibs")
	if isTestFile {
		// XXX(morgan): this disables checking that a package exists (in examples or stdlibs)
		// when transpiling a test file. After all Gno functions, including those in
		// tests/imports.go are converted to native bindings, support should
		// be added for transpiling stdlibs only available in tests/stdlibs, and
		// enable as such "package checking" also on test files.
		ctx.rootDir = ""
	}
	if strings.HasPrefix(filename, stdlibPrefix) {
		// this is a standard library. Mark it in the options so the native
		// bindings resolve correctly.
		path := strings.TrimPrefix(filename, stdlibPrefix)
		path = filepath.Dir(path)
		path = strings.Replace(path, string(filepath.Separator), "/", -1)
		path = strings.TrimLeft(path, "/")

		ctx.stdlibPath = path
	}

	transformed, err := ctx.transformFile(fset, f)
	if err != nil {
		return nil, fmt.Errorf("transpileAST: %w", err)
	}

	header := "// Code generated by github.com/gnolang/gno. DO NOT EDIT.\n\n"
	if tags != "" {
		header += "//go:build " + tags + "\n\n"
	}
	var out bytes.Buffer
	out.WriteString(header)
	err = format.Node(&out, fset, transformed)
	if err != nil {
		return nil, fmt.Errorf("format.Node: %w", err)
	}

	res := &Result{
		Imports:    f.Imports,
		Translated: out.String(),
		File:       transformed,
	}
	return res, nil
}

// TranspileBuildPackage tries to run `go build` against the transpiled .go files.
//
// This method is the most efficient to detect errors but requires that
// all the import are valid and available.
func TranspileBuildPackage(fileOrPkg, goBinary string) error {
	// TODO: use cmd/compile instead of exec?
	// TODO: find the nearest go.mod file, chdir in the same folder, rim prefix?
	// TODO: temporarily create an in-memory go.mod or disable go modules for gno?
	// TODO: ignore .go files that were not generated from gno?
	// TODO: automatically transpile if not yet done.

	files := []string{}

	info, err := os.Stat(fileOrPkg)
	if err != nil {
		return fmt.Errorf("invalid file or package path %s: %w", fileOrPkg, err)
	}
	if !info.IsDir() {
		file := fileOrPkg
		files = append(files, file)
	} else {
		pkgDir := fileOrPkg
		goGlob := filepath.Join(pkgDir, "*.go")
		goMatches, err := filepath.Glob(goGlob)
		if err != nil {
			return fmt.Errorf("glob %s: %w", goGlob, err)
		}
		for _, goMatch := range goMatches {
			switch {
			case strings.HasPrefix(goMatch, "."): // skip
			case strings.HasSuffix(goMatch, "_filetest.go"): // skip
			case strings.HasSuffix(goMatch, "_filetest.gno.gen.go"): // skip
			case strings.HasSuffix(goMatch, "_test.go"): // skip
			case strings.HasSuffix(goMatch, "_test.gno.gen.go"): // skip
			default:
				if !filepath.IsAbs(pkgDir) {
					// Makes clear to go compiler that this is a relative path,
					// rather than a path to a package/module.
					// can't use filepath.Join as it cleans its results.
					goMatch = "." + string(filepath.Separator) + goMatch
				}
				files = append(files, goMatch)
			}
		}
	}

	sort.Strings(files)
	args := append([]string{"build", "-tags=gno"}, files...)
	cmd := exec.Command(goBinary, args...)
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		// exit error
		return parseGoBuildErrors(string(out))
	}
	return err
}

var (
	errorRe   = regexp.MustCompile(`(?m)^(\S+):(\d+):(\d+): (.+)$`)
	commentRe = regexp.MustCompile(`(?m)^#.*$`)
)

// parseGoBuildErrors returns a scanner.ErrorList filled with all errors found
// in out, which is supposed to be the output of the `go build` command.
// Each errors are translated into their correlated gno files, by:
// - changing the filename from *.gno.gen.go to *.gno
// - shifting line number according to the added header in generated go files
// (see [Transpile] for that header).
func parseGoBuildErrors(out string) error {
	var errList goscanner.ErrorList
	matches := errorRe.FindAllStringSubmatch(out, -1)
	for _, match := range matches {
		filename := match[1]
		line, err := strconv.Atoi(match[2])
		if err != nil {
			return fmt.Errorf("parse line go build error %s: %w", match, err)
		}

		column, err := strconv.Atoi(match[3])
		if err != nil {
			return fmt.Errorf("parse column go build error %s: %w", match, err)
		}
		msg := match[4]
		errList.Add(token.Position{
			// Remove .gen.go extension, we want to target the gno file
			Filename: strings.TrimSuffix(filename, ".gen.go"),
			// Shift the 4 lines header added in *.gen.go files.
			// NOTE(tb): the 4 lines shift below assumes there's always a //go:build
			// directive. But the tags are optional in the Transpile() function
			// so that leaves some doubts... We might want something more reliable than
			// constants to shift lines.
			Line:   line - 4,
			Column: column,
		}, msg)
	}

	replaced := errorRe.ReplaceAllLiteralString(out, "")
	replaced = commentRe.ReplaceAllString(replaced, "")
	replaced = strings.TrimSpace(replaced)
	if replaced != "" {
		errList.Add(token.Position{}, "Additional go build errors:\n"+replaced)
	}

	return errList.Err()
}

type transpileCtx struct {
	// If rootDir is given, we will check that the directory of the import path
	// exists (using rootDir/packageDirLocation()).
	rootDir string
	// This should be set if we're working with a file from a standard library.
	// This allows us to easily check if a function has a native binding, and as
	// such modify its call expressions appropriately.
	stdlibPath string

	stdlibImports map[string]string // symbol -> import path
}

func (ctx *transpileCtx) transformFile(fset *token.FileSet, f *ast.File) (*ast.File, error) {
	var errs goscanner.ErrorList

	imports := astutil.Imports(fset, f)
	ctx.stdlibImports = make(map[string]string)

	// rewrite imports to point to stdlibs/ or examples/
	for _, paragraph := range imports {
		for _, importSpec := range paragraph {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				errs.Add(fset.Position(importSpec.Pos()), fmt.Sprintf("can't unquote import path %s: %v", importSpec.Path.Value, err))
				continue
			}

			if ctx.rootDir != "" {
				dirPath := filepath.Join(ctx.rootDir, PackageDirLocation(importPath))
				if _, err := os.Stat(dirPath); err != nil {
					if !os.IsNotExist(err) {
						return nil, err
					}
					errs.Add(fset.Position(importSpec.Pos()), fmt.Sprintf("import %q does not exist", importPath))
					continue
				}
			}

			// Create mapping
			if IsStdlib(importPath) {
				if importSpec.Name != nil {
					ctx.stdlibImports[importSpec.Name.Name] = importPath
				} else {
					// XXX: imperfect, see comment on transformCallExpr
					ctx.stdlibImports[path.Base(importPath)] = importPath
				}
			}

			transp := TranspileImportPath(importPath)
			if !astutil.RewriteImport(fset, f, importPath, transp) {
				errs.Add(fset.Position(importSpec.Pos()), fmt.Sprintf("failed to replace the %q package with %q", importPath, transp))
			}
		}
	}

	// custom handler
	node := astutil.Apply(f,
		// pre
		func(c *astutil.Cursor) bool {
			node := c.Node()
			// is function declaration without body?
			// -> delete (native binding)
			if fd, ok := node.(*ast.FuncDecl); ok && fd.Body == nil {
				c.Delete()
				return false // don't attempt to traverse children
			}

			// is function call to a native function?
			// -> rename if unexported, apply `nil,` for the first arg if necessary
			if ce, ok := node.(*ast.CallExpr); ok {
				return ctx.transformCallExpr(c, ce)
			}

			return true
		},

		// post
		func(c *astutil.Cursor) bool {
			return true
		},
	)
	return node.(*ast.File), errs.Err()
}

func (ctx *transpileCtx) transformCallExpr(_ *astutil.Cursor, ce *ast.CallExpr) bool {
	switch fe := ce.Fun.(type) {
	case *ast.SelectorExpr:
		// XXX: This is not correct in 100% of cases. If I shadow the `std` symbol, and
		// its replacement is a type with the method AssertOriginCall, this system
		// will incorrectly add a `nil` as the first argument.
		// A full fix requires understanding scope; the Go standard library recommends
		// using go/types, which for proper functioning requires an importer
		// which can work with Gno. This is deferred for a future PR.
		id, ok := fe.X.(*ast.Ident)
		if !ok {
			break
		}
		ip, ok := ctx.stdlibImports[id.Name]
		if !ok {
			break
		}
		if stdlibs.HasMachineParam(ip, gnolang.Name(fe.Sel.Name)) {
			// Because it's an import, the symbol is always exported, so no need for the
			// X_ prefix we add below.
			ce.Args = append([]ast.Expr{ast.NewIdent("nil")}, ce.Args...)
		}

	case *ast.Ident:
		// Is this a native binding?
		// Note: this is only useful within packages like `std` and `math`.
		// The logic here is not robust to be generic. It does not account for locally
		// defined scope. However, because native bindings have a narrowly defined and
		// controlled scope (standard libraries) this will work for our usecase.
		if ctx.stdlibPath != "" &&
			stdlibs.HasNativeBinding(ctx.stdlibPath, gnolang.Name(fe.Name)) {
			if stdlibs.HasMachineParam(ctx.stdlibPath, gnolang.Name(fe.Name)) {
				ce.Args = append([]ast.Expr{ast.NewIdent("nil")}, ce.Args...)
			}
			if !fe.IsExported() {
				// Prefix unexported names with X_, per native binding convention
				// (to export the symbol within Go).
				fe.Name = "X_" + fe.Name
			}
		}
	}
	return true
}
