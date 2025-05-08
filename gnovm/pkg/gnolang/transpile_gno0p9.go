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
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"
)

// Transpiles existing Gno code to Gno 0.9, the one with @cross decorators, not
// cross(fn)(...). (duh).
//
// Writes in place if dirPath is provided.
// Files without `// Gno 0.9` as the first line are considered to be for Gno
// 0.x < 0.9.
//
// xform: result of FindGno0p9XItems().
func TranspileToGno0p9(mpkg *std.MemPackage, dirPath string, testing, format bool, xform map[string]string) error {
	// This map is used to allow for function re-definitions, which are allowed
	// in Gno (testing context) but not in Go.
	// This map links each function identifier with a closure to remove its
	// associated declaration.
	var delFunc map[string]func()
	if testing {
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
		if format && xform == nil {
			var buf bytes.Buffer
			err = gofmt.Node(&buf, fset, f)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			file.Body = buf.String()
		}

		//----------------------------------------
		// Logical transforms

		if xform != nil {
			// AST transform for Gno 0.9.
			if err := transpileToGno0p9(mpkg.Path, fset, file.Name, f, xform); err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			// Write transformed AST to Go to file.
			var buf bytes.Buffer
			err = gofmt.Node(&buf, fset, f)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			file.Body = buf.String()
		}
		files = append(files, f)
	}
	if errs != nil {
		return errs
	}
	// END processing all files.

	// Write to dirPath.
	err := mpkg.WriteTo(dirPath)
	return err
}

func transpileToGno0p9(pkgPath string, fs *token.FileSet, fileName string, f *ast.File, xform map[string]string) (err error) {

	var lastLine = 0
	var didRemoveCrossing = false
	var setLast = func(end token.Pos) {
		posn := fs.Position(end)
		lastLine = posn.Line
	}
	var getLine = func(pos token.Pos) int {
		return fs.Position(pos).Line
	}

	astutil.Apply(f, func(c *astutil.Cursor) bool {

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
			if n.Name == "realm" {
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
		case *ast.FuncDecl:
			pos := n.Pos()
			posn := fs.Position(pos)
			line, col := posn.Line, posn.Column
			key := fmt.Sprintf("%s/%s:%d:%d", pkgPath, fileName, line, col)
			if xform[key] == "add curfunc" {
				n.Type.Params.List = append([]*ast.Field{&ast.Field{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm"),
				}}, n.Type.Params.List...)
			}
		case *ast.CallExpr:
			pos := n.Pos()
			posn := fs.Position(pos)
			line, col := posn.Line, posn.Column
			key := fmt.Sprintf("%s/%s:%d:%d", pkgPath, fileName, line, col)
			if id, ok := n.Fun.(*ast.Ident); ok && id.Name == "cross" {
				// Replace expression 'cross(x)' by 'x'.
				// In Gno 0.9 @cross decorator is used instead.
				var a ast.Node
				if len(n.Args) == 1 {
					a = n.Args[0]
				} else {
					err = errors.New("cross called with invalid parameters")
				}
				c.Replace(a)
				return true
			}
			if xform[key] == "add curcall" {
				n.Args = append([]ast.Expr{ast.NewIdent("cur")}, n.Args...)
			} else if xform[key] == "add nilrealm" {
				n.Args = append([]ast.Expr{ast.NewIdent("nil")}, n.Args...)
			}
		}
		return true
	}, nil)
	return err
}

//========================================
// Find Gno0.9 XItems

// Represents a needed transform.
type XItem struct {
	Type string
	Location
}

// Finds XItems for Gno 0.9 from the Gno AST and
// stores them pn ATTR_GNO0P9_XITEMS
// Then TranspileToGno0p9() applies them to Go AST and writes them.
func FindGno0p9XItems(store Store, pn *PackageNode, bn BlockNode) {
	// create stack of BlockNodes.
	var stack []BlockNode = make([]BlockNode, 0, 32)
	var last BlockNode = pn
	stack = append(stack, last)

	// Iterate over all nodes recursively.
	_ = Transcribe(bn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("FindXItems %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

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
	fmt.Printf("%s:%s\n", t, key)
}
