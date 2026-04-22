// XXX: DEPRECATED, see cmd/gno/fix

package gnolang

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

/*
	Transpiling old Gno code to Gno 0.9.
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).

	ParseCheckGnoMod() defined in pkg/gnolang/gnomod.go.
	GoParseMemPackage() defined in pkg/gnolang/gotypecheck.go.
*/

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
	if errs != nil {
		return errs
	}
	// Write AST transforms to mpkg.
	err := WriteToMemPackage(gofset, gofs, mpkg, false)
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	return errs
}

// Minimal AST mutation(s) for Gno 0.9.
func prepareGno0p9(f *ast.File) (err error) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch gon := c.Node().(type) {
		case *ast.Ident:
			// XXX: optimistic.
			switch gon.Name {
			case "cross":
				// only exists in .gnobuiltins.gno for gno 0.0
				gon.Name = "_cross_gno0p0"
			case "realm":
				gon.Name = "realm_XXX"
			case "address":
				gon.Name = "address_XXX"
			case "gnocoin":
				gon.Name = "gnocoin_XXX"
			case "gnocoins":
				gon.Name = "gnocoins_XXX"
			case "cross_gno0p9":
				gon.Name = "cross"
			case "realm_gno0p9": // doesn't work, prepare in pkg/test/import.
				gon.Name = "realm"
			case "address_gno0p9":
				gon.Name = "address"
			case "gnocoin_gno0p9":
				gon.Name = "gnocoin"
			case "gnocoins_gno0p9":
				gon.Name = "gnocoins"
			}
		}
		return true
	}, nil)
	return err
}

//========================================
// Find Xforms for Gno 0.0 --> Gno 0.9.

// Xform represents a single needed transform.
type Xform struct {
	Type string
	Location
}

// Finds Xforms for Gno 0.9 from the Gno AST and stores them pn
// ATTR_PN_XFORMS.
func FindXformsGno0p9(store Store, pn *PackageNode, fn *FileNode) {
	// Iterate over all file nodes recursively.
	_ = TranscribeB(pn, fn, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		switch stage {
		// ----------------------------------------
		case TRANS_LEAVE:
			switch n := n.(type) {
			case *CallExpr: // TRANS_LEAVE
				if _, ok := n.Func.(*constTypeExpr); ok {
					return n, TRANS_CONTINUE
				} else if cx, ok := n.Func.(*ConstExpr); ok {
					if cx.TypedValue.T.Kind() != FuncKind {
						return n, TRANS_CONTINUE
					}
					fv := cx.GetFunc()
					if fv.PkgPath == uversePkgPath && fv.Name == "_cross_gno0p0" {
						// Add a nil realm as first argument.
						pc, ok := ns[len(ns)-1].(*CallExpr)
						if !ok {
							panic("cross(fn) must be followed by a call")
						}
						addXform1(pn, fileNameOf(last), pc, XTYPE_ADD_CROSS_CALL, nil)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "crossing" {
						if !IsRealmPath(pn.PkgPath) {
							panic("crossing() is only allowed in realm packages")
						}
						// Add `cur realm` as first argument to func decl.
						addXform1(pn, fileNameOf(last), last, XTYPE_ADD_CUR_FUNC, nil)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "attach" {
						// reserve attach() so we can support it later.
						panic("attach() not yet supported")
					}
					return n, TRANS_CONTINUE
				}
				if n.WithCross { // (cross-called)
					// Already handled, added XTYPE_ADD_CROSS_CALL
					// cross(fn)(...) --> fn(cross,...)
					return n, TRANS_CONTINUE
				}
				// Add xform to call expr n if the body is crossing.
				// The rest will be handled by FindMore.
				// Try to evaluate statically n.Func; may fail.
				ftv, err := tryEvalStatic(store, pn, last, n.Func)
				if false { // for debugging:
					fmt.Println("FAILED TO EVALSTATIC", n.Func, err)
				}
				switch fv := ftv.V.(type) {
				case nil:
					return n, TRANS_CONTINUE
				case TypeValue:
					return n, TRANS_CONTINUE
				case *FuncValue:
					fn := fv.GetSource(store).(FuncNode)
					if fn.GetBody().isCrossing_gno0p0() {
						// Not cross-called, so add `cur` as first argument.
						addXform1(pn, fileNameOf(last), n, XTYPE_ADD_CUR_CALL, nil)
						ensureCurFunc(store, pn, last, nil)
					}
				case *BoundMethodValue:
					fn := fv.Func.GetSource(store).(FuncNode)
					if fn.GetBody().isCrossing_gno0p0() {
						// Not cross-called, so add `cur` as first argument.
						addXform1(pn, fileNameOf(last), n, XTYPE_ADD_CUR_CALL, nil)
						ensureCurFunc(store, pn, last, nil)
					}
				}
				return n, TRANS_CONTINUE
			} // END switch n.(type) {}
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

type xtype string

const (
	XTYPE_ADD_CUR_FUNC   xtype = "ADD_CUR_FUNC"   // add `cur realm` to func signature
	XTYPE_ADD_CUR_CALL   xtype = "ADD_CUR_CALL"   // add `cur` as first arg in call
	XTYPE_ADD_CROSS_CALL xtype = "ADD_CROSS_CALL" // add `cross` as first arg in call
)

const (
	ATTR_PN_XFORMS = "ATTR_PN_XFORMS" // all on package node
	ATTR_XFORM     = "ATTR_XFORM"     // one per node
)

// Called from FindXformsGno0p9().
// pn: package node to write xform1s.
// f: filename
// n: node to transform.
// x: transform type.
func addXform1(pn *PackageNode, f string, n Node, x xtype, xnew *int) {
	s := n.GetSpan()
	p := pn.PkgPath
	// key: p/f:s+x
	xforms1, _ := pn.GetAttribute(ATTR_PN_XFORMS).(map[string]struct{})
	if xforms1 == nil {
		xforms1 = make(map[string]struct{})
		pn.SetAttribute(ATTR_PN_XFORMS, xforms1)
	}
	xform1 := fmt.Sprintf("%s/%s:%v+%s", p, f, s, x)
	if _, exists := xforms1[xform1]; exists {
		return // ignore duplicates.
	}
	xforms1[xform1] = struct{}{}
	n.SetAttribute(ATTR_XFORM, x)
	if xnew != nil {
		*xnew++
	}
	fmt.Printf("xpiling to Gno 0.9: +%q\n", xform1)
}

// Called from transpileGno0p9_part1 to translate p/f:l:c+x to n.
func addXform2IfMatched(
	xforms1 map[string]struct{},
	xforms2 map[ast.Node]string,
	gon ast.Node, p string, f string, s Span, x xtype,
) {
	xform1 := fmt.Sprintf("%s/%s:%v+%s", p, f, s, x)
	if _, exists := xforms1[xform1]; exists {
		if prior, exists := xforms2[gon]; exists {
			fmt.Println("xform2 already exists. prior:", prior, "new:", xform1)
			panic("oops, need to refactor xforms2 to allow multiple xforms per node?")
		}
		xforms2[gon] = xform1
	}
}

// XXX Rename, only spreads XTYPE_ADD_CUR_FUNC (because it's a type change?) and
// only applies for ATTR_XFORM attr. So this general name is not ideal,
// but it will be renamed once it becomes more clear how to move forward.
func spreadXform(lhs, rhs Node, xnew *int) (more Node, cmp int) {
	attrl := lhs.GetAttribute(ATTR_XFORM)
	attrr := rhs.GetAttribute(ATTR_XFORM)
	if attrl == nil && attrr != nil {
		if attrr != XTYPE_ADD_CUR_FUNC {
			return
		}
		lhs.SetAttribute(ATTR_XFORM, attrr)
		more, cmp = lhs, -1
		*xnew++
	} else if attrl != nil && attrr == nil {
		if attrr != XTYPE_ADD_CUR_FUNC {
			return
		}
		rhs.SetAttribute(ATTR_XFORM, attrl)
		more, cmp = rhs, 1
		*xnew++
	} else if attrl != attrr {
		panic("conflicting attributes not yet handled")
	} else { // attrl == attr
		more, cmp = nil, 0
	}
	return
}

// Apply the found xform's and xform more.  The ultimate goal is to find more
// pn xforms which result in actual transforms in part 2.
// Returns number of new xforms; need to run across all files until all files return 0.
func FindMoreXformsGno0p9(store Store, pn *PackageNode, last BlockNode, n Node) (xnew int) {
	// Iterate over all nodes recursively.
	_ = TranscribeB(last, n, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		switch stage {
		// ----------------------------------------
		case TRANS_LEAVE:
			switch n := n.(type) {
			case *NameExpr:
				// NOTE: Keep in sync maybe with preprocess.go/TRANS_LEAVE *NameExpr.
				// Ignore non-block type paths
				if n.Path.Type != VPBlock {
					return n, TRANS_CONTINUE
				}
				// Ignore blank identifers
				if n.Name == blankIdentifier {
					return n, TRANS_CONTINUE
				}
				// Ignore package names
				if n.GetAttribute(ATTR_PACKAGE_REF) != nil {
					return n, TRANS_CONTINUE
				}
				// Ignore decls names.
				if ftype == TRANS_VAR_NAME {
					return n, TRANS_CONTINUE
				}
				// Ignore := defines, etc.
				if n.Type != NameExprTypeNormal {
					return n, TRANS_CONTINUE
				}
				// Find the block where name is defined.
				dbn, _, nsrc := last.GetNameSourceForPath(store, n.Path)
				// Spread attribute.
				_, cmp := spreadXform(n, nsrc.NameExpr, &xnew)
				if cmp > 0 {
					// recurse again in dbn.
					dbnLast := dbn.GetParentNode(store)
					FindMoreXformsGno0p9(store, pn, dbnLast, dbn)
				}
				return n, TRANS_CONTINUE
			case *AssignStmt:
				// XXX if RHS has attribute, apply attribute to LHS.
				lhs := n.Lhs
				rhs := n.Rhs
				if len(lhs) == len(rhs) { // a, b, c [:]= 1, 2, 3
					for i, lhx := range lhs {
						rhx := rhs[i]
						more, _ := spreadXform(lhx, rhx, &xnew)
						if more != nil { // recurse
							FindMoreXformsGno0p9(store, pn, last, more)
						}
					}
				} else if len(lhs) > 1 && len(rhs) == 1 {
					// XXX not yet supported.
				} else {
					panic("should not happen")
				}
				return n, TRANS_CONTINUE
			case *CompositeLitExpr:
				cltx := unconst(n.Type)
				_, tx := last.GetTypeExprForExpr(store, cltx)
				switch tx := tx.(type) {
				case *StructTypeExpr:
					fields := tx.Fields
					// Iterate over CompositeLitExpr key:value elements
					// and match them against the declaration fields.
					for i, kvx := range n.Elts {
						var ftx *FieldTypeExpr
						if n.IsKeyed() {
							ftx = fields.GetFieldTypeExpr(
								kvx.Key.(*NameExpr).Name)
							if ftx == nil { // key omitted
								continue
							}
						} else {
							ftx = &fields[i]
						}
						// Spread xform attribute from value to type expr field.
						_, cmp := spreadXform(&ftx.NameExpr, kvx.Value, &xnew)
						switch {
						case cmp < 0: // name expr <<< value (add cur to func)
							_, cmp = spreadXform(ftx.Type, kvx.Value, &xnew)
							if cmp >= 0 {
								panic("expected spread xform to type expr")
							}
							addXform1(pn, fileNameOf(last), ftx.Type, XTYPE_ADD_CUR_FUNC, &xnew)
							// Dive into the param type? maybe useful later.
							FindMoreXformsGno0p9(store, pn, last, ftx.Type)
						case cmp > 0: // name expr >>> value (add cur to func)
							// Find more in kvx.
							FindMoreXformsGno0p9(store, pn, last, kvx.Value)
						}
					}
					return n, TRANS_CONTINUE
				default: // XXX implement more type exprs
					return n, TRANS_CONTINUE
				}
			case *CallExpr:

				//--------------------------------------------------------------------------------
				// These vars are what can be determined statically.
				var cfn Expr               // either cfn(...) or cross(cfn)(...)
				var ft *FuncType           // if cfn is actually a function
				var ns NameSource          // if cfn is name expr
				var sx *SelectorExpr       // if cfn is a selector
				var sxt Type               // if sx, static type of sx.X
				var ipn *PackageNode       // if sxt is interface, interface decl package
				var itx *InterfaceTypeExpr // if sxt is interface, type expr of iface type decl

				// fill `cfn`
				if n.WithCross { // cross(n.Func)(...)
					ccfx := gno0p0CrossCallFunc(n)
					if ccfx == nil {
						panic("should not happen")
					}
					cfn = ccfx
				} else {
					ccfx := gno0p0CrossCallFunc(n)
					if ccfx != nil {
						panic("should not happen")
					}
					cfn = n.Func
				}
				// fill `ft`
				ft, ok := evalStaticTypeOf(store, last, cfn).(*FuncType)
				if !ok {
					// conversions not handled "yet".
					return n, TRANS_CONTINUE
				}
				// fill `ns`
				if nx, ok := cfn.(*NameExpr); ok {
					_, _, ns = last.GetNameSourceForPath(store, nx.Path)
				}
				// fill `sx*`
				if sx2, ok := cfn.(*SelectorExpr); ok {
					sx = sx2
					sxt = evalStaticTypeOf(store, last, sx.X)
				}
				// fill `i*` and do some work.
				if sx != nil && sxt.Kind() == InterfaceKind {
					it, ok := sxt.(*DeclaredType)
					if !ok {
						return n, TRANS_CONTINUE
						// panic("anonymous interfaces not supported by gno fix")
					}
					if it.PkgPath == ".uverse" {
						return n, TRANS_CONTINUE
					}
					ipn = store.GetPackageNode(it.PkgPath)
					inx1 := unconst(Preprocess(store, ipn, Nx(it.Name), nil).(Expr)).(*NameExpr)
					ifn1, itx1 := ipn.GetTypeExprForExpr(store, inx1)
					if ipn != skipFile(ifn1) {
						panic("package mismatch")
					}
					itx = itx1.(*InterfaceTypeExpr)
					// ipn2 : package node of type decl.
					// ifn2: file node of type decl.
					// idn2: the (iface) type decl.
					// inx2: the name expr of type decl.
					ipn2, ifn2, ns2 := ipn.GetNameSourceForPath(store, inx1.Path)
					if ipn != ipn2 {
						panic("package mismatch")
					}
					// NOTE: inx1 was temporary, ns2.NameExpr is the correct one.
					if false {
						println(ns2)
					}

					mfnt := it.Base.(*InterfaceType).GetMethodFieldType(sx.Sel)
					if mfnt == nil {
						// XXX embedded interface methods not supported;
						// e.g. `.Write()` in `interface {io.Writer; io.Closer}`
						return n, TRANS_CONTINUE
					}
					mft := mfnt.Type.(*FuncType)
					mftx := itx.Methods.GetFieldTypeExpr(sx.Sel).Type.(*FuncTypeExpr)
					// If the method was called with cross, add xform to itx.
					if n.WithCross {
						if mft.IsCrossing() {
							// weird. it might have an attribute,
							// but this shouldn't be happening
							panic("should not happen")
						} else {
							// add xform to method's type expr.
							addXform1(ipn, ifn2.FileName, mftx, XTYPE_ADD_CUR_FUNC, &xnew)
						}
					} else {
						// otherwise if non-cross calling an xform'd type,
						// add `cur` as first argument
						_, cmp := spreadXform(cfn, mftx, &xnew)
						if cmp > 0 {
							panic("should not happen")
						}
					}
				}

				//----------------------------------------
				// If the func is from a variable...
				switch ns.Origin.(type) {
				case *AssignStmt:
					// TODO
				case *ValueDecl:
					// TODO
				default:
					// could support more...
				}

				//----------------------------------------
				// Apply xform from n.Func to pn xforms.
				if n.Func.GetAttribute(ATTR_XFORM) == XTYPE_ADD_CUR_FUNC {
					// Not cross-called, so add `cur` as first argument.
					addXform1(pn, fileNameOf(last), n, XTYPE_ADD_CUR_CALL, &xnew)
					ensureCurFunc(store, pn, last, &xnew)
				}

				//--------------------------------------------------------------------------------
				// These vars are what can be determined with a statically determined func node.
				var fn FuncNode
				var fpn *PackageNode
				var ftx *FuncTypeExpr

				// fill `fn` if possible.
				fn, err := last.GetFuncNodeForExpr(store, cfn)
				if err != nil {
					// There's nothing more to do.
					return n, TRANS_CONTINUE
				}

				// fill `fpn` and `ftx`.
				fpn = packageOf(fn)
				if fn.GetIsMethod() {
					ftx = fn.GetFuncTypeExpr()
					if len(ftx.Params) == len(ft.Params) {
						// good, leave as is.
					} else if len(ftx.Params) == len(ft.Params)-1 {
						// e.g. DeclaredType.Method(recv, ...)
						ftx = fn.(*FuncDecl).GetUnboundTypeExpr()
					} else {
						panic("unexpected func type param length")
					}
				} else {
					ftx = fn.GetFuncTypeExpr()
				}

				/* XXX delete
				// XXX use sb.GetFuncNodeForExpr() and simplify.
				if _, ok := n.Func.(*constTypeExpr); ok {
					// TODO: handle conversions.
					return n, TRANS_CONTINUE
				}
				// Try to evaluate statically n.Func; may fail.
				// NOTE: Document some of the reasons why it may fail,
				// and find intuitve ways to solve for them.
				tv, err := tryEvalStatic(store, pn, last, n.Func)
				// Get the source of the function.
				if false { // for debugging:
					fmt.Println("FAILED TO EVALSTATIC", n.Func, err)
				}
				// Find func source and func type.
				switch cv := tv.V.(type) {
				case nil:
					return n, TRANS_CONTINUE
				case TypeValue:
					return n, TRANS_CONTINUE
				case *FuncValue:
					fn = cv.GetSource(store).(FuncNode)
					ft = cv.GetType(store)
					if fn.GetIsMethod() {
						ftx = fn.(*FuncDecl).GetUnboundTypeExpr()
					} else {
						ftx = fn.GetFuncTypeExpr()
					}
				case *BoundMethodValue:
					fn = cv.Func.GetSource(store).(FuncNode)
					ft = cv.Func.GetType(store).BoundType()
					ftx = fn.GetFuncTypeExpr()
				}
				*/
				/*
					if n.WithCross {
						// Already handled, added XTYPE_ADD_CROSS_CALL
						// cross(fn)(...) --> fn(cross,...)
						return n, TRANS_CONTINUE
					}
				*/
				if ft.HasVarg() { // not yet supported
					return n, TRANS_CONTINUE
				}

				// Spread xform from func param types to arg exprs.
				var didFixArg bool
				for i, argx := range n.Args {
					_, cmp := spreadXform(&ftx.Params[i].NameExpr, argx, &xnew)
					switch {
					case cmp < 0: // param <<< argx
						ptx, cmp := spreadXform(ftx.Params[i].Type, argx, &xnew)
						if cmp >= 0 {
							panic("expected spread xform to type expr")
						}
						fname := fn.GetLocation().File
						addXform1(fpn, fname, ptx, XTYPE_ADD_CUR_FUNC, &xnew)
						// Dive into the param type? maybe useful later.
						FindMoreXformsGno0p9(store, fpn, fn, ptx)
						didFixArg = true
					case cmp > 0: // param >>> argx
						FindMoreXformsGno0p9(store, pn, last, argx)
					}
				}

				// If any arg types changed, recurse again in
				// source body. Since Params[i].NameExpr/.Type
				// were set ADD_CUR_FUNC, more calls may spread
				// more.
				// NOTE: Generally this isn't the most
				// computationally efficient, but it's simple.
				if didFixArg {
					fparent := fn.GetParentNode(store)
					FindMoreXformsGno0p9(store, fpn, fparent, fn)
				}
				return n, TRANS_CONTINUE
			} // END switch n.(type) {}
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
	return
}

// If gno 0.0 cross(fn)(...) call, return fn or nil.
func gno0p0CrossCallFunc(cx *CallExpr) Expr {
	innercx, ok := unconst(cx.Func).(*CallExpr)
	if !ok {
		return nil
	}
	innerfn, ok := unconst(innercx.Func).(*NameExpr)
	if !ok || innerfn.Name != Name("_cross_gno0p0") {
		return nil
	}
	if len(innercx.Args) == 0 {
		panic("invalid cross() with no fn argument found, expected cross(fn)(...)")
	}
	return innercx.Args[0]
}

func ensureCurFunc(store Store, pn *PackageNode, last BlockNode, xnew *int) {
	// If `cur` isn't available, it needs to be included
	// in the outer-most containing func decl/expr.
	if last.GetSlot(store, Name(`cur`), true) == nil {
		fn, _, ok := findLastFunction(last, pn)
		if !ok {
			panic("`cur` can only be used in a func body")
		}
		/* NOTE: Uncomment the following to disable
			TestFoo(cur realm, t *testing.T).
			Also see pkg/test/test.go *_cur, and
			preprocessor special case for "testing/base".
		// Test functions cannot take cur;
		// tests don't have a "caller". A default
		// caller is convenient but surprising.
		if strings.HasSuffix(fileNameOf(last), "_test.gno") &&
			name := string(fn.GetName())
			strings.HasPrefix(name, "Test") {
			// TODO: improve test cross utils.
			fmt.Printf("illegal `cur` in test %s.\n",
				name) // just a warning.
			return n, TRANS_CONTINUE
		}
		*/
		// NOTE: will also add to init/main,
		// but gnovm knows how to call them.
		addXform1(pn, fileNameOf(last), fn, XTYPE_ADD_CUR_FUNC, xnew)
	}
}

// ========================================
// Transpiles existing Gno code to Gno 0.9.
//
// Writes in place if dir is provided. Transpiled packages will have their
// gno.mod Gno version to 0.9.
//
// Args:
//   - dir: where to write to.
//   - pn: package node of fnames
//   - fnames: file names (subset of mpkg) to transpile.
//   - xforms1: result of FindGno0p9Xforms().
func TranspileGno0p9(mpkg *std.MemPackage, dir string, pn *PackageNode, fnames []string, xforms1 map[string]struct{}) error {
	// NOTE: The pkgPath may be different than mpkg.Path
	// e.g. for filetests or xxx_test tests.
	pkgPath := pn.PkgPath

	// Return if gno.mod is current.
	var mod *gnomod.File
	var err error
	mod, err = ParseCheckGnoMod(mpkg)
	if err != nil {
		panic(fmt.Errorf("unhandled error %w", err))
	}
	if mod != nil && mod.GetGno() != GnoVerMissing {
		return fmt.Errorf("cannot transpile to gno 0.9: expected gno 0.0 but got %s",
			mod.GetGno())
	}

	// Go parse and collect files from mpkg.
	gofset := token.NewFileSet()
	var errs error
	xall := 0 // number translated from part 1
	xforms12 := make(map[string]struct{})
	for _, fname := range fnames {
		if !strings.HasSuffix(fname, ".gno") {
			panic(fmt.Sprintf("expected a .gno file but got %q", fname))
		}
		mfile := mpkg.GetFile(fname)
		// Go parse file.
		const parseOpts = parser.ParseComments |
			parser.DeclarationErrors |
			parser.SkipObjectResolution
		gof, err := parser.ParseFile(
			gofset,
			path.Join(mpkg.Path, mfile.Name),
			mfile.Body,
			parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Transpile Part 1: re-key xforms1 by ast.Node.
		xforms2, err := transpileGno0p9_part1(pkgPath, gofset, mfile.Name, gof, xforms1)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		for _, xform := range xforms2 {
			if _, exists := xforms12[xform]; exists {
				panic("duplicate xform: " + xform)
			}
			xforms12[xform] = struct{}{}
		}
		xall += len(xforms2)
		// Transpile Part 2: main Go AST transform for Gno 0.9.
		if err := transpileGno0p9_part2(pkgPath, gofset, mfile.Name, gof, xforms2); err != nil {
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

	// Ensure that all xforms were translated.
	checkMismatch := func(xforms1, xforms12 map[string]struct{}, verbose bool) (mismatch bool) {
		// this is likely some bug in find* or part 1.
		for xform1 := range xforms1 {
			_, seen := xforms12[xform1]
			if seen {
				if verbose {
					fmt.Println("xform:", xform1, " (OK)")
				}
			} else {
				if verbose {
					fmt.Println("xform:", xform1, " (NOT FOUND IN xforms2)")
				}
				mismatch = true
			}
		}
		for xform2 := range xforms12 {
			_, seen := xforms1[xform2]
			if !seen {
				if verbose {
					fmt.Println("xform:", xform2, " (NOT PRESENT IN xforms1)")
				}
				mismatch = true
			}
		}
		return mismatch
	}
	mismatch := checkMismatch(xforms1, xforms12, false)
	if mismatch {
		checkMismatch(xforms1, xforms12, true)
		panic("some xform items were not translated")
	}

	return nil
}

// Transpile Step 1: re-key xforms1 by ast.Node.
//
// We can't just apply as we encounter matches in xforms1 unfortunately because
// it causes the lines to shift.  So we first convert xforms1 into a map keyed
// by node and then do the actual transpiling in step 2.
//
// Results:
//   - xfound: number of items matched for file with name `fname` (for integrity)
func transpileGno0p9_part1(pkgPath string, gofs *token.FileSet, fname string, gof *ast.File, xforms1 map[string]struct{}) (xforms2 map[ast.Node]string, err error) {
	xforms2 = make(map[ast.Node]string, len(xforms1))

	astutil.Apply(gof, func(c *astutil.Cursor) bool {
		// Main switch on c.Node() type.
		switch gon := c.Node().(type) {
		case *ast.FuncLit:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CUR_FUNC)
		case *ast.FuncDecl:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CUR_FUNC)
		case *ast.FuncType:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CUR_FUNC)
		case *ast.CallExpr:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CUR_CALL)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CROSS_CALL)
		}
		return true
	}, nil)

	// Check that all xforms1 items were translated to xforms2.
	checkXforms(xforms1, xforms2, fname)
	return xforms2, err
}

// Check that xforms1 items were translated to xforms2 items for file named fname.
// Returns the number of items matched for file.
func checkXforms(xforms1 map[string]struct{}, xforms2 map[ast.Node]string, fname string) {
	mismatch := false
XFORMS1_LOOP:
	for xform1 := range xforms1 {
		if !strings.Contains(xform1, "/"+fname) {
			continue
		}
		for _, xform2 := range xforms2 {
			if xform1 == xform2 {
				// good.
				continue XFORMS1_LOOP
			}
		}
		fmt.Println("xform2 item not found for xform1:", xform1, len(xforms2))
		mismatch = true
	}
	if mismatch {
		for xform1 := range xforms1 {
			fmt.Println("xform1:", xform1)
		}
		for n2, xform2 := range xforms2 {
			fmt.Println("xform2:", xform2, n2)
		}
		panic("xforms1 and xforms2 mismatch")
	}
	/*
		if len(xforms1) != len(xforms2) {
			panic("xforms1 and xforms2 length don't match")
		}
	*/
	// all good, return
}

// The main Go AST transpiling logic to make Gno code Gno 0.9.
func transpileGno0p9_part2(pkgPath string, fs *token.FileSet, fname string, gof *ast.File, xforms2 map[ast.Node]string) (err error) {
	lastLine := 0
	didRemoveCrossing := false
	setLast := func(end token.Pos) {
		posn := fs.Position(end)
		lastLine = posn.Line
	}
	getLine := func(pos token.Pos) int {
		return fs.Position(pos).Line
	}
	inXforms2 := func(gon ast.Node, x xtype) bool {
		if xforms2 == nil {
			return false
		}
		value := xforms2[gon]
		return strings.HasSuffix(value, "+"+string(x))
	}

	astutil.Apply(gof, func(c *astutil.Cursor) bool {
		// Handle newlines after crossing
		if didRemoveCrossing {
			gon := c.Node()
			line := getLine(gon.Pos())
			tf := fs.File(gon.Pos())
			if lastLine < line {
				// lastLine - 1 is the deleted crossing().
				tf.MergeLine(lastLine - 1)
				// and the next empty line too.
				tf.MergeLine(lastLine)
			}
			didRemoveCrossing = false
		}

		// Main switch on c.Node() type.
		switch gon := c.Node().(type) {
		case *ast.ExprStmt:
			if ce, ok := gon.X.(*ast.CallExpr); ok {
				if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == "crossing" {
					// Validate syntax.
					if len(ce.Args) != 0 {
						err = errors.New("crossing called with non empty parameters")
					}
					// Delete statement 'crossing()'.
					c.Delete()
					didRemoveCrossing = true
					setLast(gon.End())
					return false
				}
			}
		case *ast.FuncLit:
			if inXforms2(gon, XTYPE_ADD_CUR_FUNC) {
				gon.Type.Params.List = append([]*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm"),
				}}, gon.Type.Params.List...)
				delete(xforms2, gon)
			}
		case *ast.FuncDecl:
			if inXforms2(gon, XTYPE_ADD_CUR_FUNC) {
				gon.Type.Params.List = append([]*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm"),
				}}, gon.Type.Params.List...)
				delete(xforms2, gon)
			}
		case *ast.FuncType:
			if inXforms2(gon, XTYPE_ADD_CUR_FUNC) {
				names := []*ast.Ident(nil)
				for _, param := range gon.Params.List {
					if len(param.Names) > 0 {
						names = []*ast.Ident{ast.NewIdent("cur")}
					}
				}
				gon.Params.List = append([]*ast.Field{{
					Names: names,
					Type:  ast.NewIdent("realm"),
				}}, gon.Params.List...)
				delete(xforms2, gon)
			}
		case *ast.CallExpr:
			if inXforms2(gon, XTYPE_ADD_CUR_CALL) {
				gon.Args = append([]ast.Expr{ast.NewIdent("cur")}, gon.Args...)
				delete(xforms2, gon)
			} else if inXforms2(gon, XTYPE_ADD_CROSS_CALL) {
				gon.Args = append([]ast.Expr{ast.NewIdent("cross")}, gon.Args...)
				delete(xforms2, gon)
			}
			if id, ok := gon.Fun.(*ast.Ident); ok && id.Name == "_cross_gno0p0" {
				// Replace expression 'cross(x)' by 'x'.
				// Gno 0.0: `cross(fn)(...)`, Gno 0.9 : `fn(cross,...)`
				if len(gon.Args) == 1 {
					c.Replace(gon.Args[0])
				} else {
					err = errors.New("cross called with invalid parameters")
				}
				return true
			}
		}
		return true
	}, nil)

	// Verify that all xforms2 items were consumed.
	if len(xforms2) > 0 {
		for gon, xform2 := range xforms2 {
			fmt.Println("xform2 left over:", xform2, gon)
		}
		panic("Xform items left over")
	}

	return err
}
