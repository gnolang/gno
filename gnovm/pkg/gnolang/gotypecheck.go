package gnolang

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"go.uber.org/multierr"
)

// type checking (using go/types)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *gnovm.MemPackage
}

// TypeCheckMemPackage performs type validation and checking on the given
// mempkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
//
// If format is true, the code in msmpkg will be automatically updated with the
// formatted source code.
func TypeCheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, format bool) error {
	return typeCheckMemPackage(mempkg, getter, false, format)
}

// TypeCheckMemPackageTest performs the same type checks as [TypeCheckMemPackage],
// but allows re-declarations.
//
// Note: like TypeCheckMemPackage, this function ignores tests and filetests.
func TypeCheckMemPackageTest(mempkg *gnovm.MemPackage, getter MemPackageGetter) error {
	return typeCheckMemPackage(mempkg, getter, true, false)
}

func typeCheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, testing, format bool) error {
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
	fmt_ := false
	result, err := g.parseCheckMemPackage(mpkg, fmt_)
	g.cache[path] = gnoImporterResult{pkg: result, err: err}
	return result, err
}

func (g *gnoImporter) parseCheckMemPackage(mpkg *gnovm.MemPackage, fmt_ bool) (*types.Package, error) {
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
	const parseOpts = parser.ParseComments | parser.DeclarationErrors | parser.SkipObjectResolution
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

		f, err := parser.ParseFile(fset, path.Join(mpkg.Path, file.Name), file.Body, parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		//----------------------------------------
		// Non-logical formatting transforms

		if delFunc != nil {
			deleteOldIdents(delFunc, f)
		}

		// Enforce formatting.
		// This must happen before logical transforms.
		if fmt_ {
			var buf bytes.Buffer
			err = format.Node(&buf, fset, f)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			file.Body = buf.String()
		}

		//----------------------------------------
		// Logical transforms

		// No need to filter because of gnobuiltins.go.
		// But keep this code block for future transforms.
		/*
			// filter crossings for type checker
			if err := filterCrossing(f); err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
		*/

		files = append(files, f)
	}
	if errs != nil {
		return nil, errs
	}

	// Add builtins file.
	file := &gnovm.MemFile{
		Name: ".gnobuiltins.go",
		Body: fmt.Sprintf(`package %s

func istypednil(x any) bool { return false } // shim
func crossing() { } // shim
func cross[F any](fn F) F { return fn } // shim
func revive[F any](fn F) any { return nil } // shim
`, mpkg.Name),
	}
	f, err := parser.ParseFile(fset, path.Join(mpkg.Path, file.Name), file.Body, parseOpts)
	if err != nil {
		panic("error parsing gotypecheck gnobuiltins.go file")
	}
	files = append(files, f)

	pkg, err := g.cfg.Check(mpkg.Path, fset, files, nil)
	return pkg, err
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

/*
// This is how ast filtering would have worked.
// Keep this comment block around in case we need it.
func filterCrossing(f *ast.File) (err error) {
	astutil.Apply(f, nil, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.ExprStmt:
			if ce, ok := n.X.(*ast.CallExpr); ok {
				if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == "crossing" {
					// Validate syntax.
					if len(ce.Args) != 0 {
						err = errors.New("crossing called with non empty parameters")
					}
					// Delete statement 'crossing()'.
					c.Delete()
				}
			}
		case *ast.CallExpr:
			if id, ok := n.Fun.(*ast.Ident); ok && id.Name == "cross" {
				// Replace expression 'cross(x)' by 'x'.
				var a ast.Node
				if len(n.Args) == 1 {
					a = n.Args[0]
				} else {
					err = errors.New("cross called with invalid parameters")
				}
				c.Replace(a)
			}
		}
		return true
	})
	return err
}
*/
