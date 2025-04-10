package gnolang

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strings"

	"github.com/gnolang/gno/gnovm"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"go.uber.org/multierr"
)

// type checking (using go/types)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *gnovm.MemPackage
}

const DEFAULT_MAX_GAS_UGNOT = 1_000_000 // 1Gnot aka 1e6 ugnots

// TypeCheckMemPackage performs type validation and checking on the given
// mempkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
//
// If format is true, the code will be automatically updated with the
// formatted source code.
//
// By default it uses a gas meter with `DEFAULT_MAX_GAS_UGNOT`.
func TypeCheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, format bool) error {
	return typeCheckMemPackage(mempkg, getter, false, format, storetypes.NewGasMeter(DEFAULT_MAX_GAS_UGNOT))
}

// TypeCheckMemPackageWithGasMeter is like TypeCheckMemPackage, except
// that it allows passing in the gas meter to use.
func TypeCheckMemPackageWithGasMeter(mempkg *gnovm.MemPackage, getter MemPackageGetter, format bool, gasMeter storetypes.GasMeter) error {
	return typeCheckMemPackage(mempkg, getter, false, format, gasMeter)
}

// TypeCheckMemPackageTest performs the same type checks as [TypeCheckMemPackage],
// but allows re-declarations.
//
// Note: like TypeCheckMemPackage, this function ignores tests and filetests.
func TypeCheckMemPackageTest(mempkg *gnovm.MemPackage, getter MemPackageGetter) error {
	return typeCheckMemPackage(mempkg, getter, true, false, storetypes.NewInfiniteGasMeter())
}

func typeCheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, testing, format bool, gasMeter storetypes.GasMeter) error {
	var errs error
	imp := &gnoImporter{
		getter: getter,
		cache:  map[string]gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				errs = multierr.Append(errs, err)
			},
		},
		allowRedefinitions: testing,
		gasMeter:           gasMeter,
	}
	imp.cfg.Importer = imp

	_, err := imp.parseCheckMemPackage(mempkg, format)
	// prefer to return errs instead of err:
	// err will generally contain only the first error encountered.
	if errs != nil {
		return errs
	}
	return err
}

type gnoImporterResult struct {
	pkg *types.Package
	err error
}

type gnoImporter struct {
	getter MemPackageGetter
	cache  map[string]gnoImporterResult
	cfg    *types.Config

	// allow symbol redefinitions? (test standard libraries)
	allowRedefinitions bool
	gasMeter           storetypes.GasMeter
}

// Unused, but satisfies the Importer interface.
func (g *gnoImporter) Import(path string) (*types.Package, error) {
	return g.ImportFrom(path, "", 0)
}

type importNotFoundError string

func (e importNotFoundError) Error() string { return "import not found: " + string(e) }

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (g *gnoImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := g.cache[path]; ok {
		return pkg.pkg, pkg.err
	}
	mpkg := g.getter.GetMemPackage(path)
	if mpkg == nil {
		err := importNotFoundError(path)
		g.cache[path] = gnoImporterResult{err: err}
		return nil, err
	}
	fmt := false
	result, err := g.parseCheckMemPackage(mpkg, fmt)
	g.cache[path] = gnoImporterResult{pkg: result, err: err}
	return result, err
}

func (g *gnoImporter) parseCheckMemPackage(mpkg *gnovm.MemPackage, fmt bool) (*types.Package, error) {
	// This map is used to allow for function re-definitions, which are allowed
	// in Gno (testing context) but not in Go.
	// This map links each function identifier with a closure to remove its
	// associated declaration.
	var delFunc map[string]func()
	if g.allowRedefinitions {
		delFunc = make(map[string]func())
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(mpkg.Files))
	var errs error
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		// TODO: support filetest type checking. (should probably handle as each its
		// own separate pkg, which should also be typechecked)
		if !strings.HasSuffix(file.Name, ".gno") ||
			strings.HasSuffix(file.Name, "_test.gno") ||
			strings.HasSuffix(file.Name, "_filetest.gno") {
			continue
		}

		const parseOpts = parser.ParseComments | parser.DeclarationErrors | parser.SkipObjectResolution
		f, err := parser.ParseFile(fset, path.Join(mpkg.Path, file.Name), file.Body, parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		chargeGasForTypecheck(g.gasMeter, f)

		if delFunc != nil {
			deleteOldIdents(delFunc, f)
		}

		// enforce formatting
		if fmt {
			var buf bytes.Buffer
			err = format.Node(&buf, fset, f)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			file.Body = buf.String()
		}

		files = append(files, f)
	}
	if errs != nil {
		return nil, errs
	}

	return g.cfg.Check(mpkg.Path, fset, files, nil)
}

func deleteOldIdents(idents map[string]func(), f *ast.File) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		// ignore methods and init functions
		//nolint:goconst
		if !ok || fd.Recv != nil || fd.Name.Name == "init" {
			continue
		}
		if del := idents[fd.Name.Name]; del != nil {
			del()
		}
		decl := decl
		idents[fd.Name.Name] = func() {
			// NOTE: cannot use the index as a file may contain multiple decls to be removed,
			// so removing one would make all "later" indexes wrong.
			f.Decls = slices.DeleteFunc(f.Decls, func(d ast.Decl) bool { return decl == d })
		}
	}
}

func chargeGasForTypecheck(gasMeter storetypes.GasMeter, f *ast.File) {
	ast.Walk(&astTraversingGasCharger{gasMeter}, f)
}

// astTraversingGasCharger is an ast.Visitor helper that statically traverses an AST
// charging gas for the respective typechecking operations so as to bear a cost
// and not let typechecking be abused
type astTraversingGasCharger struct {
	m storetypes.GasMeter
}

var _ ast.Visitor = (*astTraversingGasCharger)(nil)

func (atgc *astTraversingGasCharger) consumeGas(amount storetypes.Gas) {
	atgc.m.ConsumeGas(amount, "typeCheck")
}

const _BASIC_TYPECHECK_GAS_CHARGE = 5 // Arbitrary value, needs more research and derivation.

func (atgc *astTraversingGasCharger) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	case *ast.ImportSpec:
		// No need to charge gas for imports.
		return nil

	case *ast.UnaryExpr:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 2)

	case *ast.BinaryExpr:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)

	case *ast.BasicLit:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 2)

	case *ast.CompositeLit:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)

	case *ast.CallExpr:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 4)

	case *ast.ForStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 5)

	case *ast.RangeStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 6)
		// TODO: Alternate on the different type of range statements.

	case *ast.FuncDecl:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 6)

	case *ast.SwitchStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 4)

	case *ast.IfStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 5)

	case *ast.CaseClause:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)

	case *ast.BranchStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)

	case *ast.AssignStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 2)

	case *ast.Ident:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 1)

	case *ast.SelectorExpr:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 5)

	case *ast.ParenExpr:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)

	case *ast.ReturnStmt, *ast.DeferStmt:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 2)

	case nil:
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE / 2)

	default: // IndexExpr, StarExpr et al, all fall under defaults here.
		atgc.consumeGas(_BASIC_TYPECHECK_GAS_CHARGE * 3)
	}

	return atgc
}
