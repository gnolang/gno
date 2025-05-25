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
	"reflect"
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
*/

type ParseMode int

const (
	// no test files.
	ParseModeProduction ParseMode = iota
	// production and test files when xxx_test tests import xxx package.
	ParseModeIntegration
	// all files even including *_filetest.gno; for linting and testing.
	ParseModeAll
	// a directory of file tests. consider all to be filetests.
	ParseModeOnlyFiletests
)

// ========================================
// Go parse the Gno source in mpkg to Go's *token.FileSet and
// []ast.File with `go/parser`.
//
// Args:
//   - pmode: see documentation for ParseMode.
//
// Results:
//   - gofs: all normal .gno files (and _test.gno files if wtests).
//   - _gofs: all xxx_test package _test.gno files if wtests.
//   - tgofs: all _testfile.gno test files.
//
// XXX move to pkg/gnolang/gotypecheck.go?
func GoParseMemPackage(mpkg *std.MemPackage, pmode ParseMode) (
	gofset *token.FileSet, gofs, _gofs, tgofs []*ast.File, errs error,
) {
	gofset = token.NewFileSet()

	// This map is used to allow for function re-definitions, which are
	// allowed in Gno (testing context) but not in Go.  This map links
	// each function identifier with a closure to remove its associated
	// declaration.
	delFunc := make(map[string]func())

	// Go parse and collect files from mpkg.
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		// Ignore _test/_filetest.gno files depending.
		switch pmode {
		case ParseModeProduction:
			if strings.HasSuffix(file.Name, "_test.gno") ||
				strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case ParseModeIntegration:
			if strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case ParseModeAll, ParseModeOnlyFiletests:
			// include all
		default:
			panic("should not happen")

		}
		// Go parse file.
		const parseOpts = parser.ParseComments |
			parser.DeclarationErrors |
			parser.SkipObjectResolution
		gof, err := parser.ParseFile(
			gofset, path.Join(mpkg.Path, file.Name),
			file.Body,
			parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// The *ast.File passed all filters.
		if strings.HasSuffix(file.Name, "_filetest.gno") ||
			pmode == ParseModeOnlyFiletests {
			tgofs = append(tgofs, gof)
		} else if strings.HasSuffix(file.Name, "_test.gno") &&
			strings.HasSuffix(gof.Name.String(), "_test") {
			if pmode == ParseModeIntegration {
				// never wanted these gofs.
				// (we do want other *_test.gno in gofs)
			} else {
				deleteOldIdents(delFunc, gof)
				_gofs = append(_gofs, gof)
			}
		} else { // normal *_test.gno here for integration testing.
			deleteOldIdents(delFunc, gof)
			gofs = append(gofs, gof)
		}
	}
	if errs != nil {
		return gofset, gofs, _gofs, tgofs, errs
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
		err := prepareGno0p9_part1(gof)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}
	if errs != nil {
		return errs
	}
	// Write AST transforms to mpkg.
	err := WriteToMemPackage(gofset, gofs, mpkg)
	if err != nil {
		errs = multierr.Append(errs, err)
	}
	// NOTE: If there was a need to preserve the gofs,
	// a reversal can happen here as prepareGno0p2_part2().
	return errs
}

// Minimal AST mutation(s) for Gno 0.9.
func prepareGno0p9_part1(f *ast.File) (err error) {
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch gon := c.Node().(type) {
		case *ast.Ident:
			// XXX: optimistic.
			switch gon.Name {
			case "cross":
				gon.Name = "_cross_gno0p0" // only exists in .gnobuiltins.gno for gno 0.0
			case "realm":
				gon.Name = "realm_XXX"
			case "realm_gno0p9": // not used
				gon.Name = "realm"
			case "address":
				gon.Name = "address_XXX"
			case "address_gno0p9": // not used
				gon.Name = "address"
			case "gnocoin":
				gon.Name = "gnocoin_XXX"
			case "gnocoin_gno0p9": // not used
				gon.Name = "gnocoin"
			case "gnocoins":
				gon.Name = "gnocoins_XXX"
			case "gnocoins_gno0p9": // not used
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
// ATTR_GNO0P9_XFORMS.
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
						fname := last.GetLocation().File
						addXform1(pn, fname, pc, XTYPE_ADD_NILREALM)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "crossing" {
						if !IsRealmPath(pn.PkgPath) {
							panic("crossing() is only allowed in realm packages")
						}
						// Add `cur realm` as first argument to func decl.
						fname := last.GetLocation().File
						addXform1(pn, fname, last, XTYPE_ADD_CURFUNC)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "attach" {
						// reserve attach() so we can support it later.
						panic("attach() not yet supported")
					}
					return n, TRANS_CONTINUE
				} else {
					// Already handled, added XTYPE_ADD_NILREALM
					// from the "cross" case above.
					if n.WithCross {
						// Is a cross(fn)(...) call.
						// Leave it alone.
						return n, TRANS_CONTINUE
					}
					// Try to evaluate statically n.Func; may fail.
					ftv, err := tryEvalStatic(store, pn, last, n.Func)
					if false { // for debugging:
						fmt.Println("FAILED TO EVALSTATIC", n.Func, err)
					}
					var isCrossing bool
					switch fv := ftv.V.(type) {
					case nil:
						return n, TRANS_CONTINUE
					case TypeValue:
						return n, TRANS_CONTINUE
					case *FuncValue:
						fd := fv.GetSource(store)
						if fd.GetBody().isCrossing_gno0p0() {
							// Not cross-called, so add `cur` as first argument.
							fname := last.GetLocation().File
							addXform1(pn, fname, n, XTYPE_ADD_CURCALL)
							isCrossing = true
						}
					case *BoundMethodValue:
						md := fv.Func.GetSource(store)
						if md.GetBody().isCrossing_gno0p0() {
							// Not cross-called, so add `cur` as first argument.
							fname := last.GetLocation().File
							addXform1(pn, fname, n, XTYPE_ADD_CURCALL)
							isCrossing = true
						}
					}
					if isCrossing {
						// If `cur` isn't available, it needs to be included
						// in the outer-most containing func decl/expr.
						if last.GetValueRef(store, Name(`cur`), true) == nil {
							fn, _, ok := findLastFunction(last, pn)
							if ok {
								// NOTE: will also add to init/main,
								// but gnovm knows how to call them.
								fname := last.GetLocation().File
								addXform1(pn, fname, fn, XTYPE_ADD_CURFUNC)
							} else {
								panic("`cur` can only be used in a func body")
							}
						}
					}
					return n, TRANS_CONTINUE
				}
			} // END switch n.(type) {}
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

type xtype string

const (
	XTYPE_ADD_CURCALL  xtype = "ADD_CURCALL"
	XTYPE_ADD_CURFUNC  xtype = "ADD_CURFUNC"
	XTYPE_ADD_NILREALM xtype = "ADD_NILREALM"
)

const ATTR_GNO0P9_XFORMS = "ATTR_GNO0P9_XFORMS" // all on package node
const ATTR_GNO0P9_XFORM = "ATTR_GNO0P9_XFORM"   // one per node

// Called from FindXformsGno0p9().
// pn: package node to write xform1s.
// f: filename
// n: node to transform.
// x: transform type.
func addXform1(pn *PackageNode, f string, n Node, x xtype) {
	var s = n.GetSpan()
	var p = pn.PkgPath
	// key: p/f:s+x
	var xforms1, _ = pn.GetAttribute(ATTR_GNO0P9_XFORMS).(map[string]struct{})
	if xforms1 == nil {
		xforms1 = make(map[string]struct{})
		pn.SetAttribute(ATTR_GNO0P9_XFORMS, xforms1)
	}
	var xform1 = fmt.Sprintf("%s/%s:%v+%s", p, f, s, x)
	if _, exists := xforms1[xform1]; exists {
		// panic("cannot trample existing item")
		return // allow duplicates.
	}
	xforms1[xform1] = struct{}{}
	n.SetAttribute(ATTR_GNO0P9_XFORM, x)
	fmt.Printf("xpiling to Gno 0.9: +%q\n", xform1)
}

// Called from transpileGno0p9_part1 to translate p/f:l:c+x to n.
func addXform2IfMatched(
	xforms1 map[string]struct{},
	xforms2 map[ast.Node]string,
	gon ast.Node, p string, f string, s Span, x xtype,
) {
	var xform1 = fmt.Sprintf("%s/%s:%v+%s", p, f, s, x)
	if _, exists := xforms1[xform1]; exists {
		if prior, exists := xforms2[gon]; exists {
			fmt.Println("xform2 already exists. prior:", prior, "new:", xform1)
			panic("oops, need to refactor xforms2 to allow multiple xforms per node?")
		}
		xforms2[gon] = xform1
	} else { // debugging:
		// fmt.Println("not found", xform1)
	}
}

// XXX Rename, only spreads XTYPE_ADD_CURFUNC (because it's a type change?) and
// only applies for ATTR_GNO0P9_XFORM attr. So this general name is not ideal,
// but it will be renamed once it becomes more clear how to move forward.
func spreadXform(lhs, rhs Node) (more Node, cmp int) {
	var attrl = lhs.GetAttribute(ATTR_GNO0P9_XFORM)
	var attrr = rhs.GetAttribute(ATTR_GNO0P9_XFORM)
	if attrl == nil && attrr != nil {
		if attrr != XTYPE_ADD_CURFUNC {
			return
		}
		lhs.SetAttribute(ATTR_GNO0P9_XFORM, attrr)
		more, cmp = lhs, -1
	} else if attrl != nil && attrr == nil {
		if attrr != XTYPE_ADD_CURFUNC {
			return
		}
		rhs.SetAttribute(ATTR_GNO0P9_XFORM, attrl)
		more, cmp = rhs, 1
	} else if attrl != attrr {
		panic("conflicting attributes not yet handled")
	} else { // attrl == attr
		more, cmp = nil, 0
	}
	return
}

// XXX Memoize.
func tryEvalStatic(store Store, pn *PackageNode, last BlockNode, n Node) (tv TypedValue, err error) {
	pv := pn.NewPackage() // throwaway
	store = store.BeginTransaction(nil, nil, nil)
	store.SetCachePackage(pv)
	var m = NewMachine("x", store)
	defer m.Release()
	func() {
		// cannot be resolved statically
		defer func() {
			r := recover()
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered panic with: %v", r)
			}
		}()
		// try to evaluate n
		tv = m.EvalStatic(last, n)
	}()
	return
}

// Apply the found xforms and xform more.
func FindMoreXformsGno0p9(store Store, pn *PackageNode, last BlockNode, n Node) {
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
				dbn := last.GetBlockNodeForPath(store, n.Path)
				src := dbn.GetNameSources()[n.Path.Index] // name expr
				// Spread attribute.
				_, cmp := spreadXform(n, src)
				if cmp > 0 {
					// recurse again in dbn.
					dbnLast := dbn.GetParentNode(store)
					FindMoreXformsGno0p9(store, pn, dbnLast, dbn)
				}
			case *AssignStmt:
				// XXX if RHS has attribute, apply attribute to LHS.
				lhs := n.Lhs
				rhs := n.Rhs
				if len(lhs) == len(rhs) { // a, b, c [:]= 1, 2, 3
					for i, lhx := range lhs {
						rhx := rhs[i]
						more, _ := spreadXform(lhx, rhx)
						if more != nil { // recurse
							FindMoreXformsGno0p9(store, pn, last, more)
						}
					}
				} else if len(lhs) > 1 && len(rhs) == 1 {
					// XXX not yet supported.
				} else {
					panic("should not happen")
				}
			case *CompositeLitExpr:
				clt := evalStaticType(store, last, n.Type)
				// NOTE: Types are interchangeable, or should
				// be, so they should not be used for acquiring
				// the source in general, unless it is a struct
				// or interface type which may have unexposed
				// names; and for these they also need to keep
				// .PkgPath; but still should not be relied on
				// for acquiring the source.
				//
				// In FindXformsGno0p9 tryEvalStatic >
				// FuncValue.GetSource() is used to get the
				// source, but getting the type is a little
				// tricker.
				// XXX Keep 'type' about interchangeable types,
				// and refine usage to do what is wanted here.
				switch cltx := n.Type.(type) {
				case *constTypeExpr:
					nx, ok := cltx.Source.(*NameExpr)
					if !ok {
						return n, TRANS_CONTINUE // XXX ?
					}
					// Find the block where type is defined.
					dbn := last.GetBlockNodeForPath(store, nx.Path)
					dnp := dbn.GetNameParents()[nx.Path.Index] // type decl/expr
					_, ok = clt.(*DeclaredType)
					if !ok { // XXX support more types.
						return n, TRANS_CONTINUE
					}
					fn, decl, ok := pn.GetDeclForSafe(nx.Name)
					if !ok {
						// e.g. dnp declared in a func.
					} else if dnp != *decl {
						// This check exists to verify correctness of
						// assumptions. Keep it for a while until it
						// is replaced with a finalized spec and docs.
						panic(fmt.Sprintf("decl mismatch", *decl, dnp))
					}
					_, ok = baseOf(clt).(*StructType)
					if !ok { // XXX support more types
						return n, TRANS_CONTINUE
					}
					// Iterate over CompositeLitExpr key:value elements
					// and match them against the declaration fields.
					for i, kvx := range n.Elts {
						// .Type of composite lit expr is pre-evaluated.
						ctx := dnp.(*TypeDecl).Type.(*constTypeExpr)
						fields := ctx.Source.(*StructTypeExpr).Fields
						var ftx *FieldTypeExpr = nil
						if n.IsKeyed() {
							ftx = fields.GetFieldTypeExpr(kvx.Key.(*NameExpr).Name)
							if ftx == nil { // Key not used in CompositeLitExpr.
								continue
							}
						} else {
							ftx = &fields[i]
						}
						// Spread xform attribute from value to type expr field.
						_, cmp := spreadXform(&ftx.NameExpr, kvx.Value)
						switch {
						case cmp < 0: // name expr <<< value
							_, cmp = spreadXform(ftx.Type, kvx.Value)
							if cmp >= 0 {
								panic("expected spread xform to type expr")
							}
							var fname string
							if fn != nil {
								fname = string(fn.Name)
							} else {
								loc := dbn.GetLocation()
								fname = loc.File
							}
							// XXX Get the xtype from spreadXform,
							// or otherwise check XTYPE_ADD_CURFUNC is good.
							addXform1(pn, fname, ftx.Type, XTYPE_ADD_CURFUNC)
							// Dive into the param type? maybe useful later.
							FindMoreXformsGno0p9(store, pn, last, ftx.Type)
						case cmp > 0:
							// Find more in kvx.
							FindMoreXformsGno0p9(store, pn, last, kvx.Value)
						}
					}
					return n, TRANS_CONTINUE
				case *SelectorExpr:
					return n, TRANS_CONTINUE // XXX implement
				case TypeExpr:
					return n, TRANS_CONTINUE // XXX implement
				default:
					panic(fmt.Sprintf("unexpected composite lit type %v\n%v\n%v", n, n.Type, reflect.TypeOf(n.Type)))
				}
			case *CallExpr:
				if _, ok := n.Func.(*constTypeExpr); ok {
					return n, TRANS_CONTINUE
				} else if _, ok := n.Func.(*ConstExpr); ok {
					return n, TRANS_CONTINUE
				} else {
					// Try to evaluate statically n.Func; may fail.
					// NOTE: Document some of the reasons why it may fail,
					// and find intuitve ways to solve for them.
					tv, err := tryEvalStatic(store, pn, last, n.Func)
					// Get the source of the function.
					if false { // for debugging:
						fmt.Println("FAILED TO EVALSTATIC", n.Func, err)
					}
					// Find func source and func type.
					var src FuncNode
					var ft *FuncType
					var ftx *FuncTypeExpr
					switch cv := tv.V.(type) {
					case nil:
						return n, TRANS_CONTINUE
					case TypeValue:
						return n, TRANS_CONTINUE
					case *FuncValue:
						src = cv.GetSource(store).(FuncNode)
						ft = cv.GetType(store)
						if src.GetIsMethod() {
							ftx = src.(*FuncDecl).GetUnboundTypeExpr()
						} else {
							ftx = src.GetFuncTypeExpr()
						}
					case *BoundMethodValue:
						src = cv.Func.GetSource(store).(FuncNode)
						ft = cv.Func.GetType(store).BoundType()
						ftx = src.GetFuncTypeExpr()
					}
					if ft.HasVarg() { // not yet supported
						return n, TRANS_CONTINUE
					}

					spn := packageOf(src)
					var didFix bool // did fix func type expr.
					for i, argx := range n.Args {
						_, cmp := spreadXform(&ftx.Params[i].NameExpr, argx)
						switch {
						case cmp < 0: // param <<< argx
							ptx, cmp := spreadXform(ftx.Params[i].Type, argx)
							if cmp >= 0 {
								panic("expected spread xform to type expr")
							}
							fname := src.GetLocation().File
							addXform1(spn, fname, ptx, XTYPE_ADD_CURFUNC)
							// Dive into the param type? maybe useful later.
							FindMoreXformsGno0p9(store, spn, src, ptx)
							didFix = true
						case cmp > 0: // param >>> argx
							FindMoreXformsGno0p9(store, pn, last, argx)
						}
					}
					// Recurse again in source body. Since Params[i].NameExpr/.Type
					// were set ADD_CURFUNC, more calls may spread more.
					// NOTE: Generally this isn't the most computationally efficient,
					// but it's simple enough and works for the most part.
					if didFix {
						parent := src.GetParentNode(store)
						FindMoreXformsGno0p9(store, pn, parent, src)
					}
					return n, TRANS_CONTINUE
				}
			} // END switch n.(type) {}
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
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
func TranspileGno0p9(mpkg *std.MemPackage, dir string, pn *PackageNode, fnames []Name, xforms1 map[string]struct{}) error {
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
	var gofset = token.NewFileSet()
	var errs error
	var xall int = 0 // number translated from part 1
	for _, fname := range fnames {
		if !strings.HasSuffix(string(fname), ".gno") {
			panic(fmt.Sprintf("expected a .gno file but got %q", fname))
		}
		mfile := mpkg.GetFile(string(fname))
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
		xnum, xforms2, err := transpileGno0p9_part1(pkgPath, gofset, mfile.Name, gof, xforms1)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		xall += xnum
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
	if xall != len(xforms1) {
		// this is likely some bug in find* or part 1.
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
func transpileGno0p9_part1(pkgPath string, gofs *token.FileSet, fname string, gof *ast.File, xforms1 map[string]struct{}) (xfound int, xforms2 map[ast.Node]string, err error) {
	xforms2 = make(map[ast.Node]string, len(xforms1))

	astutil.Apply(gof, func(c *astutil.Cursor) bool {
		// Main switch on c.Node() type.
		switch gon := c.Node().(type) {
		case *ast.FuncLit:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CURFUNC)
		case *ast.FuncDecl:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CURFUNC)
		case *ast.FuncType:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CURFUNC)
		case *ast.CallExpr:
			span := SpanFromGo(gofs, gon)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_CURCALL)
			addXform2IfMatched(
				xforms1, xforms2, gon,
				pkgPath, fname, span,
				XTYPE_ADD_NILREALM)
		}
		return true
	}, nil)

	// Check that all xforms1 items were translated to xforms2.
	xfound = checkXforms(xforms1, xforms2, fname)
	return xfound, xforms2, err
}

// Check that xforms1 items were translated to xforms2 items for file named fname.
// Returns the number of items matched for file.
func checkXforms(xforms1 map[string]struct{}, xforms2 map[ast.Node]string, fname string) int {
	mismatch := false
	found := 0
XFORMS1_LOOP:
	for xform1 := range xforms1 {
		if !strings.Contains(xform1, "/"+fname) {
			continue
		}
		for _, xform2 := range xforms2 {
			if xform1 == xform2 {
				// good.
				found += 1
				continue XFORMS1_LOOP
			}
		}
		fmt.Println("xform2 item not found for xform1:", xform1, len(xforms2))
		for _, xform2 := range xforms2 {
			fmt.Println("xform2:", xform2)
		}
		mismatch = true
	}
	if mismatch {
		for xform1 := range xforms1 {
			fmt.Println("xform1:", xform1)
		}
		for n2, xform2 := range xforms2 {
			fmt.Println("xform2:", xform2, n2)
		}
		panic("xforms1 and xforms2 don't match")
	}
	/*
		if len(xforms1) != len(xforms2) {
			panic("xforms1 and xforms2 length don't match")
		}
	*/
	return found // good
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
		if strings.HasSuffix(value, "+"+string(x)) {
			return true
		}
		return false
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
			if inXforms2(gon, XTYPE_ADD_CURFUNC) {
				gon.Type.Params.List = append([]*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm"),
				}}, gon.Type.Params.List...)
				delete(xforms2, gon)
			}
		case *ast.FuncDecl:
			if inXforms2(gon, XTYPE_ADD_CURFUNC) {
				gon.Type.Params.List = append([]*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("cur")},
					Type:  ast.NewIdent("realm"),
				}}, gon.Type.Params.List...)
				delete(xforms2, gon)
			}
		case *ast.FuncType:
			if inXforms2(gon, XTYPE_ADD_CURFUNC) {
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
			if inXforms2(gon, XTYPE_ADD_CURCALL) {
				gon.Args = append([]ast.Expr{ast.NewIdent("cur")}, gon.Args...)
				delete(xforms2, gon)
			} else if inXforms2(gon, XTYPE_ADD_NILREALM) {
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
