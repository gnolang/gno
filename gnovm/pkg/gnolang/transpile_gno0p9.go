package gnolang

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	gofmt "go/format"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

/*
	Transpiling old Gno code to Gno 0.9.
	Refer to the [Lint and Transpile ADR](./adr/lint_transpile.md).

	ParseCheckGnoMod() defined in pkg/gnolang/gnomod.go.
*/

// ========================================
// Go parse the Gno source in mpkg to Go's *token.FileSet and
// []ast.File with `go/parser`.
//
// Args:
//   - wtests: if true also parses and includes all *_test.gno
//     and *_filetest.gno files.
func GoParseMemPackage(mpkg *std.MemPackage, wtests bool) (
	gofset *token.FileSet, gofs []*ast.File, errs error) {
	gofset = token.NewFileSet()

	// This map is used to allow for function re-definitions, which are
	// allowed in Gno (testing context) but not in Go.  This map links
	// each function identifier with a closure to remove its associated
	// declaration.
	var delFunc = make(map[string]func())

	// Go parse and collect files from mpkg.
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		// Ignore _test/_filetest.gno files unless wtests.
		if !wtests &&
			(false ||
				strings.HasSuffix(file.Name, "_test.gno") ||
				strings.HasSuffix(file.Name, "_filetest.gno")) {
			continue
		}
		// Go parse file.
		const parseOpts = parser.ParseComments |
			parser.DeclarationErrors |
			parser.SkipObjectResolution
		// fmt.Println("GO/PARSER", mpkg.Path, file.Name)
		var gof, err = parser.ParseFile(
			gofset, path.Join(mpkg.Path, file.Name),
			file.Body,
			parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		deleteOldIdents(delFunc, gof)
		// The *ast.File passed all filters.
		gofs = append(gofs, gof)
	}
	if errs != nil {
		return gofset, gofs, errs
	}
	// END processing all files.
	return
}

// ========================================
// Prepare Gno 0.0 for Gno 0.9.
//
// When Gno syntax breaks in higher versions, existing code must first be
// pre-transcribed such that the Gno preprocessor won't panic.  This allows
// old Gno code to be preprocessed and used by the Gno VM for static
// analysis.  More transpiling will happen later after the preprocessed Gno
// AST is scanned for mutations on the Go AST which follows.  Any changes are
// applied directly on the mempackage.
//
// * Renames 'realm' to 'realm_' to avoid conflict with new uverse type.
//
// Args:
//   - mpkg: writes (mutated) AST to mempackage if not nil.
//
// Results:
//   - errs: returned in aggregate as a multierr type.
func PrepareGno0p9(gofset *token.FileSet, gofs []*ast.File, mpkg *std.MemPackage) (errs error) {
	for _, gof := range gofs {
		// AST transform for Gno 0.9.
		err := prepareGno0p9(gof)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}
	errs = WriteToMemPackage(gofset, gofs, mpkg)
	return
}

// Minimal AST mutation(s) for Gno 0.9.
//   - Renames 'realm' to '_realm' to avoid conflict with new builtin "realm".
func prepareGno0p9(f *ast.File) (err error) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.Ident:
			if n.Name == "realm" {
				// XXX: optimistic.
				n.Name = "_realm"
			}
		}
		return true
	}, nil)
	return err
}

//========================================
// Find XItems for Gno 0.0 --> Gno 0.9.

// XItem represents a single needed transform.
type XItem struct {
	Type string
	Location
}

// Finds XItems for Gno 0.9 from the Gno AST and stores them pn
// ATTR_GNO0P9_XITEMS.
func FindXItemsGno0p9(store Store, pn *PackageNode, bn BlockNode) {
	// create stack of BlockNodes.
	var stack []BlockNode = make([]BlockNode, 0, 32)
	var last BlockNode = pn
	stack = append(stack, last)

	// Iterate over all nodes recursively.
	_ = Transcribe(bn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:
			pushInitBlock(n.(BlockNode), &last, &stack)

		// ----------------------------------------
		case TRANS_LEAVE:
			// Pop block from stack.
			// NOTE: DO NOT USE TRANS_SKIP WITHIN BLOCK
			// NODES, AS TRANS_LEAVE WILL BE SKIPPED; OR
			// POP BLOCK YOURSELF.
			defer func() {
				switch n.(type) {
				case BlockNode:
					stack = stack[:len(stack)-1]
					last = stack[len(stack)-1]
				}
			}()

			switch n := n.(type) {
			case *CallExpr:
				if _, ok := n.Func.(*constTypeExpr); ok {
					return n, TRANS_CONTINUE
				} else if cx, ok := n.Func.(*ConstExpr); ok {
					if cx.TypedValue.T.Kind() != FuncKind {
						return n, TRANS_CONTINUE
					}
					fv := cx.GetFunc()
					if fv.PkgPath == uversePkgPath && fv.Name == "cross" {
						// Add a nil realm as first argument.
						pc, ok := ns[len(ns)-1].(*CallExpr)
						if !ok {
							panic("cross(fn) must be followed by a call")
						}
						loc := last.GetLocation()
						addXItem(pn, "add nilrealm", loc.PkgPath, loc.File, pc.GetLine(), pc.GetColumn())
					} else if fv.PkgPath == uversePkgPath && fv.Name == "crossing" {
						if !IsRealmPath(pn.PkgPath) {
							panic("crossing() is only allowed in realm packages")
						}
						// Add `cur realm` as first argument to func decl.
						loc := last.GetLocation()
						addXItem(pn, "add curfunc", loc.PkgPath, loc.File, loc.Line, loc.Column)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "attach" {
						// reserve attach() so we can support it later.
						panic("attach() not yet supported")
					}
				} else {
					// Already handled, added "add nilrealm"
					// from the "cross" case above.
					if n.WithCross {
						// Is a cross(fn)(...) call.
						// Leave it alone.
						return n, TRANS_CONTINUE
					}
					pv := pn.NewPackage() // temporary
					store := store.BeginTransaction(nil, nil, nil)
					store.SetCachePackage(pv)
					m := NewMachine("x", store)
					defer m.Release()
					tv := TypedValue{}
					func() {
						// cannot be resolved statically
						defer func() {
							recover()
							//fmt.Println("FAILED TO EVALSTATIC", n.Func, r)
						}()
						// try to evaluate n.Func.
						tv = m.EvalStatic(last, n.Func)
					}()
					switch cv := tv.V.(type) {
					case nil:
						return n, TRANS_CONTINUE
					case TypeValue:
						panic("wtf")
					case *FuncValue:
						if cv.IsCrossing() {
							// Not cross-called, so add `cur` as first argument.
							loc := last.GetLocation()
							addXItem(pn, "add curcall", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
						}
					case *BoundMethodValue:
						if cv.IsCrossing() {
							// Not cross-called, so add `cur` as first argument.
							loc := last.GetLocation()
							addXItem(pn, "add curcall", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
						}
					}
				}
			}
			// end type switch statement
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

const ATTR_GNO0P9_XITEMS = "ATTR_GNO0P9_XITEMS"

// t: type, p: pkgpath, f: filename, l: line, c: column
func addXItem(n Node, t string, p string, f string, l int, c int) {
	x, _ := n.GetAttribute(ATTR_GNO0P9_XITEMS).(map[string]string) // p/f:l:c -> t
	if x == nil {
		x = make(map[string]string)
		n.SetAttribute(ATTR_GNO0P9_XITEMS, x)
	}
	key := fmt.Sprintf("%s/%s:%d:%d", p, f, l, c)
	x[key] = t
	fmt.Printf("Gno 0.9 transpile +%s:%s\n", t, key)
}

// ========================================
// Transpiles existing Gno code to Gno 0.9.
//
// Writes in place if dir is provided. Transpiled packages will have their
// gno.mod Gno version to 0.9.
//
// Args:
//   - dir: where to write to.
//   - xform: result of FindGno0p9XItems().
func TranspileGno0p9(mpkg *std.MemPackage, dir string, xform map[string]string) error {

	// Return if gno.mod is current.
	var mod *gnomod.File
	var err error
	mod, err = ParseCheckGnoMod(mpkg)
	if err == nil {
		return nil // already up-to-date.
	}

	// Go parse and collect files from mpkg.
	gofset := token.NewFileSet()
	var errs error
	for _, mfile := range mpkg.Files {
		// Ignore non-gno files.
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue
		}
		/*
			// Ignore _test/_filetest.gno files unless testing.
			if !testing {
				if strings.HasSuffix(mfile.Name, "_test.gno") ||
					strings.HasSuffix(mfile.Name, "_filetest.gno") {
					continue
				}
			}
		*/
		// Go parse file.
		const parseOpts = parser.ParseComments |
			parser.DeclarationErrors |
			parser.SkipObjectResolution
		var gof, err = parser.ParseFile(
			gofset,
			path.Join(mpkg.Path, mfile.Name),
			mfile.Body,
			parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Transpile Part 1: re-key xform by ast.Node.
		xform2, err := transpileGno0p9_part1(mpkg.Path, gofset, mfile.Name, gof, xform)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Transpile Part 2: main Go AST transform for Gno 0.9.
		if err := transpileGno0p9_part2(mpkg.Path, gofset, mfile.Name, gof, xform2); err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Write transformed Go AST to memfile.
		if err := WriteToMemFile(gofset, gof, mfile); err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}
	if errs != nil {
		return errs
	}
	// END processing all memfiles.

	// Write version to mod and to memfile named "gno.mod".
	mod.SetGno(GnoVersion)
	mpkg.SetFile("gno.mod", mod.WriteString())

	// Write mempackage to dir.
	err = mpkg.WriteTo(dir)
	return err
}

// Transpile Step 1: re-key xform by ast.Node.
//
// We can't just apply as we encounter matches in xform unfortunately because
// it causes the lines to shift.  So we first convert xform into a map keyed
// by node and then do the actual transpiling in step 2.
func transpileGno0p9_part1(pkgPath string, fs *token.FileSet, fname string, f *ast.File, xform map[string]string) (xform2 map[ast.Node]string, err error) {
	xform2 = make(map[ast.Node]string, len(xform))

	astutil.Apply(f, func(c *astutil.Cursor) bool {

		// Main switch on c.Node() type.
		switch n := c.Node().(type) {
		case *ast.FuncLit:
			pos := n.Pos()
			posn := fs.Position(pos)
			line, col := posn.Line, posn.Column
			key := fmt.Sprintf("%s/%s:%d:%d", pkgPath, fname, line, col)
			if xform != nil && xform[key] == "add curfunc" {
				xform2[n] = "add curfunc"
			}
		case *ast.FuncDecl:
			pos := n.Pos()
			posn := fs.Position(pos)
			line, col := posn.Line, posn.Column
			key := fmt.Sprintf("%s/%s:%d:%d", pkgPath, fname, line, col)
			if xform != nil && xform[key] == "add curfunc" {
				xform2[n] = "add curfunc"
			}
		case *ast.CallExpr:
			pos := n.Pos()
			posn := fs.Position(pos)
			line, col := posn.Line, posn.Column
			key := fmt.Sprintf("%s/%s:%d:%d", pkgPath, fname, line, col)
			if id, ok := n.Fun.(*ast.Ident); ok && id.Name == "cross" {
				return true // can be superimposed with parent call.
			}
			if xform != nil && xform[key] == "add curcall" {
				xform2[n] = "add curcall"
			} else if xform != nil && xform[key] == "add nilrealm" {
				xform2[n] = "add nilrealm"
			}
		}
		return true
	}, nil)
	return xform2, err
}

// The main Go AST transpiling logic to make Gno code Gno 0.9.
func transpileGno0p9_part2(pkgPath string, fs *token.FileSet, fname string, gof *ast.File, xform map[ast.Node]string) (err error) {

	var lastLine = 0
	var didRemoveCrossing = false
	var setLast = func(end token.Pos) {
		posn := fs.Position(end)
		lastLine = posn.Line
	}
	var getLine = func(pos token.Pos) int {
		return fs.Position(pos).Line
	}

	astutil.Apply(gof, func(c *astutil.Cursor) bool {

		// Handle newlines after crossing
		if didRemoveCrossing {
			n := c.Node()
			line := getLine(n.Pos())
			tf := fs.File(n.Pos())
			if lastLine < line {
				// lastLine - 1 is the deleted crossing().
				tf.MergeLine(lastLine - 1)
				// and the next empty line too.
				tf.MergeLine(lastLine)
			}
			didRemoveCrossing = false
		}

		// Main switch on c.Node() type.
		switch n := c.Node().(type) {
		case *ast.Ident:
			if n.Name == "realm_XXX_TRANSPILE" {
				// Impostor varname 'realm' will become
				// renamed, so reclaim 'realm'.
				n.Name = "realm"
			} else if n.Name == "realm" {
				// Rename name to _realm to avoid conflict with new builtin "realm".
				// XXX: optimistic.
				n.Name = "_realm"
			}
		case *ast.ExprStmt:
			if ce, ok := n.X.(*ast.CallExpr); ok {
				if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == "crossing" {
					// Validate syntax.
					if len(ce.Args) != 0 {
						err = errors.New("crossing called with non empty parameters")
					}
					// Delete statement 'crossing()'.
					c.Delete()
					didRemoveCrossing = true
					setLast(n.End())
					return false
				}
			}
		case *ast.FuncLit:
			if xform != nil && xform[n] == "add curfunc" {
				n.Type.Params.List = append([]*ast.Field{&ast.Field{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm_XXX_TRANSPILE"),
				}}, n.Type.Params.List...)
			}
		case *ast.FuncDecl:
			if xform != nil && xform[n] == "add curfunc" {
				n.Type.Params.List = append([]*ast.Field{&ast.Field{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm_XXX_TRANSPILE"),
				}}, n.Type.Params.List...)
			}
		case *ast.CallExpr:
			if xform != nil && xform[n] == "add curcall" {
				n.Args = append([]ast.Expr{ast.NewIdent("cur")}, n.Args...)
			} else if xform != nil && xform[n] == "add nilrealm" {
				n.Args = append([]ast.Expr{ast.NewIdent("nil")}, n.Args...)
			}
			if id, ok := n.Fun.(*ast.Ident); ok && id.Name == "cross" {
				// Replace expression 'cross(x)' by 'x'.
				// In Gno 0.9 @cross decorator is used instead.
				var gon ast.Node
				if len(n.Args) == 1 {
					gon = n.Args[0]
				} else {
					err = errors.New("cross called with invalid parameters")
				}
				c.Replace(gon)
				return true
			}
		}
		return true
	}, nil)
	return err
}

// ========================================
// WriteToMemPackage writes Go AST to a mempackage
// This is useful for preparing prior version code for the preprocessor.
func WriteToMemPackage(gofset *token.FileSet, gofs []*ast.File, mpkg *std.MemPackage) error {
	for _, gof := range gofs {
		fpath := gofset.File(gof.Pos()).Name()
		_, fname := filepath.Split(fpath)
		mfile := mpkg.GetFile(fname)
		if mfile == nil {
			if strings.HasPrefix(fname, ".") {
				// Hidden files like .gnobuiltins.gno that
				// start with a dot should not get written to
				// the mempackage.
				continue
			} else {
				return fmt.Errorf("missing memfile %q", mfile)
			}
		}
		err := WriteToMemFile(gofset, gof, mfile)
		if err != nil {
			return fmt.Errorf("writing to mempackage %q: %w",
				mpkg.Path, err)
		}
	}
	return nil
}

func WriteToMemFile(gofset *token.FileSet, gof *ast.File, mfile *std.MemFile) error {
	var buf bytes.Buffer
	err := gofmt.Node(&buf, gofset, gof)
	if err != nil {
		return fmt.Errorf("writing to memfile %q: %w",
			mfile.Name, err)
	}
	mfile.Body = buf.String()
	return nil
}
