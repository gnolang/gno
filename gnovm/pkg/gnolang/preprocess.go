package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	blankIdentifier           = "_"
	debugFind                 = false // toggle when debugging.
	AttrPreprocessFuncLitExpr = "FuncLitExpr"
	TestingBasePkgPath        = "testing"
)

// Predefine (initStaticBlocks) and partially evaluates all names
// not inside function bodies.
//
// This function must be called on *FileSets because declarations
// in file sets may be unordered.
func PredefineFileSet(store Store, pn *PackageNode, fset *FileSet) {
	// fmt.Println("---Start PredefineFileSet...")
	// defer func() {
	// 	fmt.Println("---Finish PredefineFileSet...")
	// }()
	// First, initialize all file nodes and connect to package node.
	// This will also reserve names on BlockNode.StaticBlock by
	// calling StaticBlock.Reserve().
	//
	// NOTE: The calls to .Reserve() in initStaticBlock() below only
	// reserve the name for determining the value path index. What comes
	// later after this for-loop is a partial definition for *FuncDecl
	// function declarations which will get filled out later during
	// preprocessing.
	for _, fn := range fset.Files {
		setNodeLines(fn)
		setNodeLocations(pn.PkgPath, fn.FileName, fn)
		initStaticBlocks(store, pn, fn)
	}
	// NOTE: much of what follows is duplicated for a single *FileNode
	// in the main Preprocess translation function.  Keep synced.

	// Predefine all import decls first.
	// This must be done before TypeDecls, as it may recursively
	// depend on names (even in other files) that depend on imports.
	for _, fn := range fset.Files {
		for i := range fn.Decls {
			d := fn.Decls[i]
			switch d.(type) {
			case *ImportDecl:
				if d.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a
					// dependent)
					continue
				}

				// recursively predefine dependencies.
				predefineRecursively(store, fn, d)
				fn.Decls[i] = d
			}
		}
	}
	// Predefine all type decls decls.
	for _, fn := range fset.Files {
		for i := range fn.Decls {
			d := fn.Decls[i]
			switch d.(type) {
			case *TypeDecl:
				if d.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a
					// dependent)
					continue
				}

				// recursively predefine dependencies.
				predefineRecursively(store, fn, d)
				fn.Decls[i] = d
			}
		}
	}
	// Then, predefine all func/method decls.
	for _, fn := range fset.Files {
		for i := range fn.Decls {
			d := fn.Decls[i]
			switch d.(type) {
			case *FuncDecl:
				if d.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a
					// dependent)
					continue
				}

				// recursively predefine dependencies.
				predefineRecursively(store, fn, d)
				fn.Decls[i] = d
			}
		}
	}
	// Then, predefine other decls and
	// preprocess ValueDecls..
	for _, fn := range fset.Files {
		for i := 0; i < len(fn.Decls); i++ {
			d := fn.Decls[i]
			if d.GetAttribute(ATTR_PREDEFINED) == true {
				// skip declarations already predefined (e.g.
				// through recursion for a dependent)
				continue
			}

			// Split multiple value decls into separate ones.
			// NOTE: As a strange quirk of Go, intra-dependent
			// value declarations like `var a, b, c = 1, a, b` is
			// only allowed at the package level, but not within a
			// function. Gno2 may allow this type of declaration,
			// and when it does the following splitting logic will
			// need to be duplicated in the Preprocessor.
			if vd, ok := d.(*ValueDecl); ok && len(vd.NameExprs) > 1 && len(vd.Values) == len(vd.NameExprs) {
				var iota_ any
				if len(vd.Attributes.data) > 0 {
					// TODO: use a uint64 bitflag and faster operations.
					keys := vd.GetAttributeKeys()
					for _, key := range keys {
						switch key {
						case ATTR_IOTA:
							iota_ = vd.GetAttribute(key)
						default:
							// defensive.
							panic("unexpected attribute")
						}
					}
				}
				split := make([]Decl, len(vd.NameExprs))
				for j := range vd.NameExprs {
					part := &ValueDecl{
						NameExprs: NameExprs{{
							Attributes: vd.NameExprs[j].Attributes,
							Path:       vd.NameExprs[j].Path,
							Name:       vd.NameExprs[j].Name,
							Type:       NameExprTypeDefine,
						}},
						Type:   vd.Type,
						Values: Exprs{vd.Values[j].Copy().(Expr)},
						Const:  vd.Const,
					}
					if iota_ != nil {
						part.SetAttribute(ATTR_IOTA, iota_)
					}
					predefineRecursively(store, fn, part)
					split[j] = part
				}
				fn.Decls = append(fn.Decls[:i], append(split, fn.Decls[i+1:]...)...) //nolint:makezero
				i += len(vd.NameExprs) - 1
				continue
			} else {
				// recursively predefine dependencies.
				predefineRecursively(store, fn, d)
				continue
			}
		}
	}
}

// Initialize static blocks, and also reserves all names.
// TODO: ensure and keep idempotent.
// PrpedefineFileSet may precede Preprocess.
func initStaticBlocks(store Store, ctx BlockNode, nn Node) {
	// create stack of BlockNodes.
	last := ctx
	stack := append(make([]BlockNode, 0, 32), last)

	// iterate over all nodes recursively.
	_ = Transcribe(nn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)
		if debug {
			debug.Printf("initStaticBlocks %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {
			case *AssignStmt:
				// fmt.Println("---initStaticBlocks, assignStmt: ", n)
				if n.Op == DEFINE {
					for i, lx := range n.Lhs {
						nx := lx.(*NameExpr)
						ln := nx.Name
						if ln == blankIdentifier {
							continue
						}
						if !isLocallyDefined2(last, ln) {
							// fmt.Println("---not locally defined, Reserve it, ln: ", ln)
							if ftype == TRANS_FOR_INIT {
								ln = Name(fmt.Sprintf(".loopvar_%s", ln))
								nx.Type = NameExprTypeLoopVarDefine // demote if shadowed by i := i
								// fmt.Println("---ln: ", ln)
								nx.Name = ln // rename

								// nx0 := Nx(".redefine_i")
								// nx0.Type = NameExprTypeDefine
								// last.Reserve(false, nx0, n, NSDefine, i)
								// last.Reserve(false, nx, n, NSDefine, i+1)
								last.Reserve(false, nx, n, NSDefine, i)
							} else {
								// if loop extern, will promote to
								// NameExprTypeHeapDefine later.
								nx.Type = NameExprTypeDefine
								last.Reserve(false, nx, n, NSDefine, i)
							}
						} else {
							// fmt.Println("---locally defined, Ignore it, ln: ", ln)
						}
					}
				}
			case *ImportDecl:
				nx := &n.NameExpr
				nn := nx.Name
				if nn == "." {
					panic("dot imports not allowed in gno")
				}
				if nn == "" { // use default
					pv := store.GetPackage(n.PkgPath, true)
					if pv == nil {
						panic(fmt.Sprintf(
							"unknown import path %s",
							n.PkgPath))
					}
					exp, ok := expectedPkgName(pv.PkgPath)
					if !ok {
						panic(fmt.Sprintf("invalid pkg path: %q", pv.PkgPath))
					}
					if exp != string(pv.PkgName) {
						panic(fmt.Sprintf(
							"package name for %q (%q) doesn't match its expected identifier %q; "+
								"the import declaration must specify an identifier", pv.PkgPath, pv.PkgName, exp))
					}
					nn = pv.PkgName
					nx.Name = nn
				}
				if nn != blankIdentifier {
					nx.Type = NameExprTypeDefine
					last.Reserve(false, nx, n, NSImportDecl, -1)
				}
			case *ValueDecl:
				last2 := skipFile(last)
				for i := range n.NameExprs {
					nx := &n.NameExprs[i]
					nn := nx.Name
					if nn == blankIdentifier {
						continue
					}
					nx.Type = NameExprTypeDefine
					last2.Reserve(n.Const, nx, n, NSValueDecl, i)
				}
			case *TypeDecl:
				last2 := skipFile(last)
				nx := &n.NameExpr
				nx.Type = NameExprTypeDefine
				last2.Reserve(true, nx, n, NSTypeDecl, -1)
			case *FuncDecl:
				if n.IsMethod {
					if n.Recv.Name == "" || n.Recv.Name == blankIdentifier {
						// create a hidden var with leading dot.
						// NOTE: document somewhere.
						n.Recv.Name = ".recv"
					}
				} else {
					pkg := skipFile(last).(*PackageNode)
					// special case: if n.Name == "init", assign unique suffix.
					switch n.Name {
					case "init":
						idx := pkg.GetNumNames()
						// NOTE: use a dot for init func suffixing.
						// this also makes them unreferenceable.
						dname := Name(fmt.Sprintf("init.%d", idx))
						n.Name = dname
					case blankIdentifier:
						idx := pkg.GetNumNames()
						dname := Name(fmt.Sprintf("._%d", idx))
						n.Name = dname
					}
					nx := &n.NameExpr
					nx.Type = NameExprTypeDefine
					pkg.Reserve(false, nx, n, NSFuncDecl, -1)
					pkg.UnassignableNames = append(pkg.UnassignableNames, n.Name)
				}
			case *FuncTypeExpr:
				for i := range n.Params {
					p := &n.Params[i]
					if p.Name == "" || p.Name == blankIdentifier {
						// create a hidden var with leading dot.
						// NOTE: document somewhere.
						pn := fmt.Sprintf(".arg_%d", i)
						p.Name = Name(pn)
					}
				}
				for i := range n.Results {
					r := &n.Results[i]
					if r.Name == "" {
						// create an unnamed name with leading dot.
						r.Name = Name(fmt.Sprintf("%s%d", missingResultNamePrefix, i))
					}
					if r.Name == blankIdentifier {
						// create an underscore name with leading dot.
						r.Name = Name(fmt.Sprintf("%s%d", underscoreResultNamePrefix, i))
					}
				}
			}
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_BLOCK:
			pushInitBlock(n.(BlockNode), &last, &stack)
			switch n := n.(type) {
			case *IfCaseStmt:
				// parent if statement.
				ifs := ns[len(ns)-1].(*IfStmt)
				// anything declared in ifs are copied.
				for _, ns := range ifs.GetNameSources() {
					last.Reserve(false, ns.NameExpr, ns.Origin, ns.Type, ns.Index)
				}
			case *RangeStmt:
				if n.Op == DEFINE {
					if n.Key != nil {
						nx := n.Key.(*NameExpr)
						if nx.Name != blankIdentifier {
							nx.Type = NameExprTypeDefine
							last.Reserve(false, nx, n, NSRangeKey, -1)
						}
					}
					if n.Value != nil {
						nx := n.Value.(*NameExpr)
						if nx.Name != blankIdentifier {
							nx.Type = NameExprTypeDefine
							last.Reserve(false, nx, n, NSRangeValue, -1)
						}
					}
				}
			case *FuncLitExpr:
				for i := range n.Type.Params {
					px := &n.Type.Params[i].NameExpr
					px.Type = NameExprTypeDefine
					last.Reserve(false, px, &n.Type, NSFuncParam, i)
				}
				for i := range n.Type.Results {
					rx := &n.Type.Results[i].NameExpr
					if rx.Name == "" {
						rn := fmt.Sprintf("%s%d", missingResultNamePrefix, i)
						rx.Name = Name(rn)
					}
					rx.Type = NameExprTypeDefine
					last.Reserve(false, rx, &n.Type, NSFuncResult, i)
				}
			case *SwitchStmt:
				// n.Varname is declared in each clause.
			case *SwitchClauseStmt:
				blen := len(n.Body)
				if blen > 0 {
					n.Body[blen-1].SetAttribute(ATTR_LAST_BLOCK_STMT, true)
				}

				// parent switch statement.
				ss := ns[len(ns)-1].(*SwitchStmt)
				// anything declared in ss.init are copied.
				for _, ns := range ss.GetNameSources() {
					last.Reserve(false, ns.NameExpr, ns.Origin, ns.Type, ns.Index)
				}
				if ss.IsTypeSwitch {
					if ss.VarName != "" {
						last.Reserve(false, &NameExpr{Name: ss.VarName}, ss, NSTypeSwitch, -1)
					}
				} else {
					if ss.VarName != "" {
						panic("should not happen")
					}
				}
			case *FuncDecl:
				if n.IsMethod {
					n.Recv.NameExpr.Type = NameExprTypeDefine
					n.Reserve(false, &n.Recv.NameExpr, n, NSFuncReceiver, -1)
				}
				for i := range n.Type.Params {
					px := &n.Type.Params[i].NameExpr
					if px.Name == "" {
						panic("should not happen")
					}
					px.Type = NameExprTypeDefine
					n.Reserve(false, px, &n.Type, NSFuncParam, i)
				}
				for i := range n.Type.Results {
					rx := &n.Type.Results[i].NameExpr
					if rx.Name == "" {
						rn := fmt.Sprintf("%s%d", missingResultNamePrefix, i)
						rx.Name = Name(rn)
					}
					rx.Type = NameExprTypeDefine
					n.Reserve(false, rx, &n.Type, NSFuncResult, i)
				}
			}
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_LEAVE:
			// Pop block from stack.
			// NOTE: DO NOT USE TRANS_SKIP WITHIN BLOCK
			// NODES, AS TRANS_LEAVE WILL BE SKIPPED; OR
			// POP BLOCK YOURSELF.
			switch n.(type) {
			case BlockNode:
				stack = stack[:len(stack)-1]
				last = stack[len(stack)-1]
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

func doRecover(stack []BlockNode, n Node) {
	if r := recover(); r != nil {
		if _, ok := r.(*PreprocessError); ok {
			// re-panic directly if this is a PreprocessError already.
			panic(r)
		}

		// before re-throwing the error, append location information to message.
		last := stack[len(stack)-1]
		loc := last.GetLocation()
		if !n.GetSpan().IsZero() {
			loc.SetSpanOverride(n.GetSpan())
		}

		var err error
		rerr, ok := r.(error)
		if ok {
			err = errors.Wrap(rerr, loc.String())
		} else {
			err = fmt.Errorf("%s: %v", loc.String(), r)
		}

		// Re-throw the error after wrapping it with the preprocessing stack information.
		panic(&PreprocessError{
			err:   err,
			stack: stack,
		})
	}
}

// This counter ensures (during testing) that certain functions
// (like ConvertUntypedTo() for bigints and strings)
// are only called during the preprocessing stage.
// It is a counter because Preprocess() is recursive.
// As a global counter, use lockless atomic to support concurrency.
var preprocessing atomic.Int32

// Preprocess n whose parent block node is ctx. If any names
// are defined in another file, generally you must call
// PredefineFileSet() on the whole fileset first before calling
// Preprocess.
//
// The ctx passed in may be mutated if there are any statements
// or declarations. The file or package which contains ctx may
// be mutated if there are any file-level declarations.
//
// Store is used to load external package values, but otherwise
// the package and newly created blocks/values are expected
// to be non-RefValues -- in some cases, nil is passed for store
// to enforce this.
//
// List of what Preprocess() does:
//   - Assigns BlockValuePath to NameExprs.
//   - TODO document what it does.
func Preprocess(store Store, ctx BlockNode, n Node) Node {
	// fmt.Println("---Preprocess, n: ", n, reflect.TypeOf(n))
	// PrintCaller(2, 5)
	// defer func() {
	// 	fmt.Println("---Preprocess complete...")
	// }()
	// PrintCaller(2, 5)
	// First init static blocks of blocknodes.
	// This may have already happened.
	// Keep this function idemponent.
	// NOTE: need to use Transcribe() here instead of `bn, ok := n.(BlockNode)`
	// because say n may be a *CallExpr containing an anonymous function.
	initStaticBlocks(store, ctx, n)
	defer func() {
		Transcribe(n,
			func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
				if stage != TRANS_ENTER {
					return n, TRANS_CONTINUE
				}
				n.DelAttribute(ATTR_PREPROCESS_SKIPPED)
				n.DelAttribute(ATTR_PREPROCESS_INCOMPLETE)
				return n, TRANS_CONTINUE
			})
	}()

	// Bulk of the preprocessor function
	n = preprocess1(store, ctx, n)

	// fmt.Println("---After prerocess1: ", n, reflect.TypeOf(n))
	// XXX check node lines and locations
	checkNodeLinesLocations("XXXpkgPath", "XXXfileName", n)

	// "coda" means "conclusion".
	// NOTE: need to use Transcribe() here instead of `bn, ok := n.(BlockNode)`
	// because say n may be a *CallExpr containing an anonymous function.
	transformHeapItem := func(n Node) {
		Transcribe(n,
			func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
				if stage != TRANS_ENTER {
					return n, TRANS_CONTINUE
				}
				if _, ok := n.(*FuncLitExpr); ok {
					if n.GetAttribute(ATTR_PREPROCESS_SKIPPED) == AttrPreprocessFuncLitExpr {
						return n, TRANS_SKIP
					}
				}
				if bn, ok := n.(BlockNode); ok {
					// findGotoLoopDefines(ctx, bn)
					findHeapDefinesByUse(ctx, bn)
					findHeapUsesDemoteDefines(ctx, bn)
					findPackageSelectors(bn)
					return n, TRANS_SKIP
				}
				return n, TRANS_CONTINUE
			})
	}

	transformLoopvar := func(n Node) {
		var found bool
		// fmt.Println("---transformLoopvar, n: ", n, reflect.TypeOf(n))
		Transcribe(n,
			func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
				if stage != TRANS_LEAVE {
					return n, TRANS_CONTINUE
				}
				// fmt.Println("---Trans_Leave bn: ", n)
				if bn, ok := n.(BlockNode); ok {
					// fmt.Println("---TransLoopvar, bn: ", bn)
					if _, ok := bn.(*ForStmt); ok {
						found2 := findHeapDefinedLoopvarByUse(ctx, bn)
						if found2 {
							found = true
						}
					}
					return n, TRANS_CONTINUE
				} else {
					// fmt.Println("---Not BlockNode, do nothing...")
				}
				return n, TRANS_CONTINUE
			})

		// find capture .loopvar_i
		// inject scope var
		if found {
			fmt.Println("---found: ", found)
			Transcribe(n,
				func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
					if stage != TRANS_LEAVE {
						return n, TRANS_CONTINUE
					}
					if bn, ok := n.(BlockNode); ok {
						if _, ok := bn.(*ForStmt); ok {
							shadowLoopvar(ctx, bn)
						}
						return n, TRANS_CONTINUE
					} else {
					}
					return n, TRANS_CONTINUE
				})

			Transcribe(n,
				func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
					if stage != TRANS_LEAVE {
						return n, TRANS_CONTINUE
					}
					if bn, ok := n.(BlockNode); ok {
						if _, ok := bn.(*ForStmt); ok {
							renameLoopvarUse(ctx, bn)
						}
						return n, TRANS_CONTINUE
					} else {
					}
					return n, TRANS_CONTINUE
				})

			fmt.Println("---Current n: ", n)
			// restore ATTR_PREPROCESSED attribute
			// fmt.Println("---restore node...")
			RestoreNode(ctx, n)
			// fmt.Println("---after restore node...")
			// re-preprocess1
			n = preprocess1(store, ctx, n)
			fmt.Println("---after re-preprocess1..., n: ", n)
			// panic("!!!")
		}
	}

	// _ = transformLoopvar

	transform := func() {
		transformLoopvar(n)
		transformHeapItem(n)
	}

	// do transform
	transform()
	fmt.Println("---transform complete..., n: ", n)

	return n
}

func preprocess1(store Store, ctx BlockNode, n Node) Node {
	// fmt.Println("---Start preprocess1, n: ", n, reflect.TypeOf(n))
	// Increment preprocessing counter while preprocessing.
	preprocessing.Add(1)
	defer preprocessing.Add(-1)

	// if n is file node, set node locations recursively.
	if fn, ok := n.(*FileNode); ok {
		pkgPath := ctx.(*PackageNode).PkgPath
		fileName := fn.FileName
		setNodeLines(fn)
		setNodeLocations(pkgPath, fileName, fn)
	}

	// create stack of BlockNodes.
	last := ctx
	stack := append(make([]BlockNode, 0, 32), last)
	ctxpn := packageOf(ctx)

	// iterate over all nodes recursively
	nn := Transcribe(n, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		// if already preprocessed, skip it.
		if n.GetAttribute(ATTR_PREPROCESSED) == true {
			return n, TRANS_SKIP
		}

		defer doRecover(stack, n)
		if debug {
			debug.Printf("Preprocess %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {
			// TRANS_ENTER -----------------------
			case *AssignStmt:
				checkValDefineMismatch(n)

				if n.Op == DEFINE {
					for i, lx := range n.Lhs {
						lnx := lx.(*NameExpr)
						if lnx.Name == blankIdentifier {
							// ignore.
						} else if strings.HasPrefix(string(lnx.Name), ".decompose_") {
							_, ok := last.GetLocalIndex(lnx.Name)
							if !ok {
								// initial declaration to be re-defined.
								last.Reserve(false, lnx, n, NSDefine, i)
							}
							// else, do not redeclare.
						}
					}
				}
				// else, nothing defined.
			// TRANS_ENTER -----------------------
			case *ImportDecl, *ValueDecl, *TypeDecl, *FuncDecl:
				// NOTE func decl usually must happen with a
				// file, and so last is usually a *FileNode,
				// but for testing convenience we allow
				// importing directly onto the package.
				// Uverse requires this.
				if n.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a dependent)
				} else {
					d := n.(Decl)
					if cd, ok := d.(*ValueDecl); ok {
						checkValDefineMismatch(cd)
					}

					// recursively predefine dependencies.
					preprocessed := predefineRecursively(store, last, d)
					if preprocessed {
						return d, TRANS_SKIP
					} else {
						return d, TRANS_CONTINUE
					}
				}

			// TRANS_ENTER -----------------------
			case *FuncTypeExpr:
				for i := range n.Params {
					p := &n.Params[i]
					if p.Name == "" || p.Name == blankIdentifier {
						panic("arg name should have been set in initStaticBlocks")
					}
					if nx, ok := p.Type.(*NameExpr); ok && nx.Name == Name("realm") {
						// don't alow confusion by e.g. declaring a crossing function like
						// func something(prev realm, x int) { ... }.
						if i == 0 {
							// XXX refactor .arg stuff; see isUnnamedResult.
							if p.Name != "cur" && !strings.HasPrefix(string(p.Name), ".arg") {
								panic("a crossing function's first realm argument must have name `cur`")
							}
						} else {
							if p.Name == "cur" {
								panic("only the first realm type argument of a crossing function may have name `cur`")
							}
						}
					}
				}
				for i := range n.Results {
					r := &n.Results[i]
					if r.Name == blankIdentifier {
						panic("result name should have been set in initStaticBlock")
					}
				}
			}

			// TRANS_ENTER -----------------------
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_BLOCK:

			switch n := n.(type) {
			// TRANS_BLOCK -----------------------
			case *BlockStmt:
				fmt.Println("---PushInitBlockStmt..., n: ", n)
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *ForStmt:
				fmt.Println("---PushInitBlockForStmt..., n: ", n)
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *IfStmt:
				// create faux block to store .Init.
				// the contents are copied onto the case block
				// in the if case below for .Body and .Else.
				// NOTE: similar to *SwitchStmt.
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *IfCaseStmt:
				pushInitBlockAndCopy(n, &last, &stack)
				// parent if statement.
				ifs := ns[len(ns)-1].(*IfStmt)
				// anything declared in ifs are copied.
				for _, n := range ifs.GetBlockNames() {
					tv := ifs.GetSlot(nil, n, false)
					last.Define(n, *tv)
				}

			// TRANS_BLOCK -----------------------
			case *RangeStmt:
				pushInitBlock(n, &last, &stack)
				// NOTE: preprocess it here, so type can
				// be used to set n.IsMap/IsString and
				// define key/value.
				n.X = Preprocess(store, last, n.X).(Expr)
				xt := evalStaticTypeOf(store, last, n.X)
				if xt == nil {
					panic("cannot range over nil")
				}

				switch xt.Kind() {
				case MapKind:
					n.IsMap = true
				case StringKind:
					n.IsString = true
				case PointerKind:
					if xt.Elem().Kind() != ArrayKind {
						panic("range iteration over pointer requires array elem type")
					}
					xt = xt.Elem()
					n.IsArrayPtr = true
				}
				// key value if define.
				if n.Op == DEFINE {
					if xt.Kind() == MapKind {
						if n.Key != nil {
							kt := baseOf(xt).(*MapType).Key
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(kt))
						}
						if n.Value != nil {
							vt := baseOf(xt).(*MapType).Value
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(vt))
						}
					} else if xt.Kind() == StringKind {
						if n.Key != nil {
							it := IntType
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(it))
						}
						if n.Value != nil {
							et := Int32Type
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(et))
						}
					} else {
						if n.Key != nil {
							it := IntType
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(it))
						}
						if n.Value != nil {
							et := xt.Elem()
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(et))
						}
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncLitExpr:
				// retrieve cached function type.
				ft := evalStaticType(store, last, &n.Type).(*FuncType)
				if n.GetAttribute(ATTR_PREPROCESS_SKIPPED) == AttrPreprocessFuncLitExpr {
					// This preprocessing was initiated by predefineRecursively().
					// When predefined from a package/file, the dependant names may
					// be declared out of order, so preprocessing the body of
					// function literals is not performed.  When predefined from
					// within a function body, the depdent names must already be
					// declared, so preprocessing is not skipped.
					//
					// NOTE: Machine still needs the attribute later for evalStatic*.
					// Clear it @ initStaticBlocks.
					// n.DelAttribute(ATTR_PREPROCESS_SKIPPED)
					for _, p := range ns {
						// Prevent parents from being marked as attr
						// preprocessed.  Be not concerned about any inner
						// Preprocess() calls for newly constructed nodes:
						// either this node will be inside it, in which case it
						// too will be marked incomplete, or, it won't be
						// inside it, in which case it will complete.
						p.SetAttribute(ATTR_PREPROCESS_INCOMPLETE, true)
					}
					// n.SetAttribute(ATTR_PREPROCESS_INCOMPLETE, true)
					return n, TRANS_SKIP
				}
				// a crossing function can only be declared in a realm.
				if ft.IsCrossing() && !isRealm(ctx) {
					panic(fmt.Sprintf("crossing function literal (realm first argument) declared in non-realm package: %v", n))
				}
				// push func body block.
				pushInitBlock(n, &last, &stack)
				// define parameters in new block.
				for i, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
					n.Type.Params[i].Path = n.GetPathForName(nil, p.Name)
				}
				// define results in new block.
				for i, rf := range ft.Results {
					name := rf.Name
					last.Define(name, anyValue(rf.Type))
					n.Type.Results[i].Path = n.GetPathForName(nil, name)
				}

			// TRANS_BLOCK -----------------------
			case *SelectCaseStmt:
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *SwitchStmt:
				// create faux block to store .Init/.Varname.
				// the contents are copied onto the case block
				// in the switch case below for switch cases.
				// NOTE: similar to *IfStmt, but with the major
				// difference that each clause block may have
				// different number of values.
				// To support the .Init statement and for
				// conceptual simplicity, we create a block in
				// OpExec.SwitchStmt, but since we don't initially
				// know which clause will match, we expand the
				// block once a clause has matched.
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *SwitchClauseStmt:
				pushInitBlockAndCopy(n, &last, &stack)
				// parent switch statement.
				ss := ns[len(ns)-1].(*SwitchStmt)
				// anything declared in ss.Init are copied.
				for _, n := range ss.GetBlockNames() {
					tv := ss.GetSlot(nil, n, false)
					last.Define(n, *tv)
				}
				if ss.IsTypeSwitch {
					if len(n.Cases) == 0 {
						// evaluate default case.
						if ss.VarName != "" {
							// The type is the tag type.
							tt := evalStaticTypeOf(store, last, ss.X)
							last.Define(
								ss.VarName, anyValue(tt))
						}
					} else {
						// evaluate case types.
						for i, cx := range n.Cases {
							cx = Preprocess(
								store, last, cx).(Expr)
							var ct Type
							if cxx, ok := cx.(*ConstExpr); ok {
								if cxx.IsUndefined() {
									// TODO: shouldn't cxx.T be TypeType?
									// Don't change cxx.GetType() for defensiveness.
									ct = nil
								} else {
									ct = cxx.GetType()
								}
							} else {
								ct = evalStaticType(store, last, cx)
							}
							n.Cases[i] = toConstTypeExpr(last, cx, ct)
							// maybe type-switch def.
							if ss.VarName != "" {
								if len(n.Cases) == 1 {
									// If there is only 1 case, the
									// define applies with type.
									// (re-definition).
									last.Define(
										ss.VarName, anyValue(ct))
								} else {
									// If there are 2 or more
									// cases, the type is the tag type.
									tt := evalStaticTypeOf(store, last, ss.X)
									last.Define(
										ss.VarName, anyValue(tt))
								}
							}
						}
					}
				} else {
					// evaluate tag type
					tt := evalStaticTypeOf(store, last, ss.X)
					// check or convert case types to tt.
					for i, cx := range n.Cases {
						cx = Preprocess(
							store, last, cx).(Expr)
						checkOrConvertType(store, last, n, &cx, tt) // #nosec G601
						n.Cases[i] = cx
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncDecl:
				// retrieve cached function type.
				// the type and receiver are already set in predefineRecursively.
				ft := getType(&n.Type).(*FuncType)
				// a crossing function can only be declared in a realm.
				if ft.IsCrossing() && !isRealm(ctx) {
					panic(fmt.Sprintf("crossing function (realm first argument) declared in non-realm package: %v", n))
				}
				// push func body block.
				pushInitBlock(n, &last, &stack)
				// define receiver in new block, if method.
				if n.IsMethod {
					name := n.Recv.Name
					if name != "" {
						rft := getType(&n.Recv).(FieldType)
						rt := rft.Type
						last.Define(name, anyValue(rt))
						n.Recv.Path = n.GetPathForName(nil, name)
					}
				}
				// define parameters in new block.
				for i, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
					n.Type.Params[i].Path = n.GetPathForName(nil, p.Name)
				}
				// define results in new block.
				for i, rf := range ft.Results {
					name := rf.Name
					last.Define(name, anyValue(rf.Type))
					n.Type.Results[i].Path = n.GetPathForName(nil, name)
				}
				// functions that don't return a value do not need termination analysis
				// functions that are externally defined or builtin implemented in the vm can't be analysed
				if len(ft.Results) > 0 && ctxpn.PkgPath != uversePkgPath && n.Body != nil {
					errs := Analyze(n)
					if len(errs) > 0 {
						panic(fmt.Sprintf("%+v\n", errs))
					}
				}
			// TRANS_BLOCK -----------------------
			case *FileNode:
				// only for imports.
				pushInitBlock(n, &last, &stack)
				{
					// This logic supports out-of-order
					// declarations.  (this must happen
					// after pushInitBlock above, otherwise
					// it would happen @ *FileNode:ENTER)

					// Predefine all import decls.
					for i := range n.Decls {
						d := n.Decls[i]
						switch d.(type) {
						case *ImportDecl:
							if d.GetAttribute(ATTR_PREDEFINED) == true {
								// skip declarations already
								// predefined (e.g. through
								// recursion for a dependent)
							} else {
								// recursively predefine
								// dependencies.
								predefineRecursively(store, n, d)
								n.Decls[i] = d
							}
						}
					}
					// Predefine all type decls.
					for i := range n.Decls {
						d := n.Decls[i]
						switch d.(type) {
						case *TypeDecl:
							if d.GetAttribute(ATTR_PREDEFINED) == true {
								// skip declarations already
								// predefined (e.g. through
								// recursion for a dependent)
							} else {
								// recursively predefine
								// dependencies.
								predefineRecursively(store, n, d)
								n.Decls[i] = d
							}
						}
					}
					// Then, predefine all func/method decls.
					for i := range n.Decls {
						d := n.Decls[i]
						switch d.(type) {
						case *FuncDecl:
							if d.GetAttribute(ATTR_PREDEFINED) == true {
								// skip declarations already
								// predefined (e.g. through
								// recursion for a dependent)
							} else {
								// recursively predefine
								// dependencies.
								predefineRecursively(store, n, d)
								n.Decls[i] = d
							}
						}
					}
					// Finally, predefine other decls and
					// preprocess ValueDecls..
					for i := range n.Decls {
						d := n.Decls[i]
						if d.GetAttribute(ATTR_PREDEFINED) == true {
							// skip declarations already
							// predefined (e.g. through
							// recursion for a dependent)
						} else {
							// recursively predefine
							// dependencies.
							predefineRecursively(store, n, d)
							n.Decls[i] = d
						}
					}
				}

			// TRANS_BLOCK -----------------------
			default:
				panic("should not happen")
			}
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_BLOCK2:

			// The main TRANS_BLOCK2 switch.
			switch n := n.(type) {
			// TRANS_BLOCK2 -----------------------
			case *SwitchStmt:

				// NOTE: TRANS_BLOCK2 ensures after .Init.
				// Preprocess and convert tag if const.
				if n.X != nil {
					n.X = Preprocess(store, last, n.X).(Expr)
					convertIfConst(store, last, n, n.X)
				}
			}
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_LEAVE:
			// mark as preprocessed so that it can be used
			// in evalStaticType(store,).
			setPreprocessed(n)

			// Defer pop block from stack.
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

			// While leaving nodes w/ TRANS_COMPOSITE_TYPE,
			// (regardless of whether name or literal), any elided
			// type names are inserted. (This works because the
			// transcriber leaves the composite type before
			// entering the kv elements.)
			defer func() {
				switch ftype {
				// TRANS_LEAVE (deferred)---------
				case TRANS_COMPOSITE_TYPE:
					// fill elided element composite lit type exprs
					clx := ns[len(ns)-1].(*CompositeLitExpr)
					// get or evaluate composite type.
					clt := evalStaticType(store, last, n.(Expr))
					// elide composite lit element (nested) composite types.
					elideCompositeElements(last, clx, clt)
				}
			}()

			// The main TRANS_LEAVE switch.
			switch n := n.(type) {
			// TRANS_LEAVE -----------------------
			case *NameExpr:
				// fmt.Println("---Trans_Leave NameExpr: ", n)
				if isBlankIdentifier(n) {
					switch ftype {
					case TRANS_ASSIGN_LHS, TRANS_RANGE_KEY, TRANS_RANGE_VALUE, TRANS_VAR_NAME:
						// can use _ as value or type in these contexts
					default:
						panic("cannot use _ as value or type")
					}
				}
				// Validity: check that name isn't reserved.
				if isReservedName(n.Name) {
					panic(fmt.Sprintf(
						"should not happen: name %q is reserved", n.Name))
				}
				// Special case if struct composite key.
				if ftype == TRANS_COMPOSITE_KEY {
					clx := ns[len(ns)-1].(*CompositeLitExpr)
					clt := evalStaticType(store, last, clx.Type)
					switch bt := baseOf(clt).(type) {
					case *StructType:
						n.Path = bt.GetPathForName(n.Name)
						return n, TRANS_CONTINUE
					case *ArrayType, *SliceType:
						fillNameExprPath(last, n, false)
						if last.GetIsConst(store, n.Name) {
							cx := evalConst(store, last, n)
							return cx, TRANS_CONTINUE
						}
						// If name refers to a package, and this is not in
						// the context of a selector, fail. Packages cannot
						// be used as a value, for go compatibility but also
						// to preserve the security expectation regarding imports.
						nt := evalStaticTypeOf(store, last, n)
						if nt.Kind() == PackageKind {
							panic(fmt.Sprintf(
								"package %s cannot only be referred to in a selector expression",
								n.Name))
						}
						return n, TRANS_CONTINUE
					}
				}
				// specific and general cases
				switch n.Name {
				case blankIdentifier:
					n.Path = NewValuePathBlock(0, 0, blankIdentifier)
					return n, TRANS_CONTINUE
				case "iota":
					pd := lastDecl(ns)
					io := pd.GetAttribute(ATTR_IOTA).(int)
					cx := constUntypedBigint(n, int64(io))
					return cx, TRANS_CONTINUE
				case ".cur":
					// special name that is defined in uverse as undefined,
					// but should not be statically evaluated.
					// (will be replaced TRANS_LEAVE *CallExpr).
					return n, TRANS_CONTINUE
				case "cross":
					// Special case for gno 0.0.
					if ctxpn.GetAttribute(ATTR_FIX_FROM) == GnoVerMissing {
						// Do nothing here, TRANS_LEAVE *CallExpr will handle it.
						return n, TRANS_CONTINUE
					}
					// `cross` can only be used as the first argument
					// to a crossing function, for cross-calling.
					if ftype != TRANS_CALL_ARG || index != 0 {
						panic("cross can only be used as the first argument to a crossing function (since gno 0.9)")
					}
					// special name that is defined in uverse as undefined,
					// but should not be statically evaluated.
					// (will be replaced TRANS_LEAVE *CallExpr).
					return n, TRANS_CONTINUE
				case nilStr:
					// nil will be converted to
					// typed-nils when appropriate upon
					// leaving the expression nodes that
					// contain nil nodes.
					fallthrough
				default:
					switch ftype {
					case TRANS_ASSIGN_LHS:
						// fmt.Println("---Trans_Assign_LHS")
						as := ns[len(ns)-1].(*AssignStmt)
						// for i := n; i > 0; i >>= 1 {
						TagLoopvar(last, n)
						fillNameExprPath(last, n, as.Op == DEFINE)
						return n, TRANS_CONTINUE
					case TRANS_VAR_NAME:
						// fmt.Println("---Trans_Var_Name")
						fillNameExprPath(last, n, true)
						return n, TRANS_CONTINUE
					default:
						// happens before fill path
						// rename ensure name match.
						TagLoopvar(last, n)
						fillNameExprPath(last, n, false)
					}
					// If uverse, return a *ConstExpr.
					if n.Path.Depth == 0 { // uverse
						cx := evalConst(store, last, n)
						// built-in functions must be called.
						if !cx.IsUndefined() &&
							cx.T.Kind() == FuncKind &&
							ftype != TRANS_CALL_FUNC {
							panic(fmt.Sprintf(
								"use of builtin %s not in function call",
								n.Name))
						}
						if !cx.IsUndefined() && cx.T.Kind() == TypeKind {
							return toConstTypeExpr(last, n, cx.GetType()), TRANS_CONTINUE
						}
						return cx, TRANS_CONTINUE
					}
					// Is const decl or type decl. Not (import) packages.
					if last.GetIsConst(store, n.Name) {
						// n.Name may refer to either
						// a value OR a type. But don't change
						// this behavior, it's reasonable for
						// a name to be a value by default.
						cx := evalConst(store, last, n)
						/*
							if !cx.IsUndefined() && cx.T.Kind() == TypeKind && ftype == TRANS_TYPE_TYPE {
								return toConstTypeExpr(last, n, cx.GetType()), TRANS_CONTINUE
							}
						*/
						return cx, TRANS_CONTINUE
					}
					// Special handling of packages
					nt := evalStaticTypeOf(store, last, n)
					if nt == nil {
						// this is fine, e.g. for TRANS_ASSIGN_LHS (define) etc.
					} else if nt.Kind() == PackageKind {
						// If name refers to a package, and this is not in
						// the context of a selector, fail. Packages cannot
						// be used as a value, for go compatibility but also
						// to preserve the security expectation regarding imports.
						if ftype != TRANS_SELECTOR_X {
							panic(fmt.Sprintf(
								"package %s cannot only be referred to in a selector expression",
								n.Name))
						}
						// Remember the package path
						// for codaPackageSelectors().
						pvc := evalConst(store, last, n)
						pv, ok := pvc.V.(*PackageValue)
						if !ok {
							panic(fmt.Sprintf(
								"missing package %s",
								n.String()))
						}
						pref := RefValueFromPackage(pv)
						n.SetAttribute(ATTR_PACKAGE_REF, pref)
						n.SetAttribute(ATTR_PACKAGE_PATH, pv.PkgPath)
					}
				}

			// TRANS_LEAVE -----------------------
			case *BasicLitExpr:
				// Replace with *ConstExpr.
				cx := evalConst(store, last, n)
				return cx, TRANS_CONTINUE

			// TRANS_LEAVE -----------------------
			case *BinaryExpr:
				lt := evalStaticTypeOf(store, last, n.Left)
				rt := evalStaticTypeOf(store, last, n.Right)

				lcx, lic := n.Left.(*ConstExpr)
				rcx, ric := n.Right.(*ConstExpr)

				if debug {
					debug.Printf("Trans_leave, BinaryExpr, OP: %v, lx: %v, rx: %v, lt: %v, rt: %v, isLeftConstExpr: %v, isRightConstExpr %v, isLeftUntyped: %v, isRightUntyped: %v \n", n.Op, n.Left, n.Right, lt, rt, lic, ric, isUntyped(lt), isUntyped(rt))
				}

				// Special (recursive) case if shift and right isn't uint.
				isShift := n.Op == SHL || n.Op == SHR
				if isShift {
					// check LHS type compatibility
					n.assertShiftExprCompatible1(store, last, lt, rt)
					// checkOrConvert RHS
					if baseOf(rt) != UintType {
						// convert n.Right to (gno) uint type,
						rn := Expr(Call("uint", n.Right))
						// reset/create n2 to preprocess right child.
						n2 := &BinaryExpr{
							Left:  n.Left,
							Op:    n.Op,
							Right: rn,
						}
						n2.Right.SetAttribute(ATTR_SHIFT_RHS, true)
						resn := Preprocess(store, last, n2)
						return resn, TRANS_CONTINUE
					}
					// Then, evaluate the expression.
					if lic && ric {
						cx := evalConst(store, last, n)
						return cx, TRANS_CONTINUE
					}
					return n, TRANS_CONTINUE
				}

				// general cases
				n.AssertCompatible(lt, rt) // check compatibility against binaryExprs other than shift expr
				// General case.
				if lic {
					if ric {
						// Left const, Right const ----------------------
						// Replace with *ConstExpr if const operands.
						//
						// First, convert untyped as necessary.
						// If either is interface type no conversion is required.
						if (lt == nil || lt.Kind() != InterfaceKind) &&
							(rt == nil || rt.Kind() != InterfaceKind) {
							if !shouldSwapOnSpecificity(lcx.T, rcx.T) {
								// convert n.Left to right type.
								checkOrConvertType(store, last, n, &n.Left, rcx.T)
							} else {
								// convert n.Right to left type.
								checkOrConvertType(store, last, n, &n.Right, lcx.T)
							}
						}
						// Then, evaluate the expression.
						cx := evalConst(store, last, n)
						return cx, TRANS_CONTINUE
					} else if isUntyped(lcx.T) {
						// Left untyped const, Right not ----------------
						// right is untyped const, left is not const, typed/untyped
						checkUntypedShiftExpr := func(x Expr) {
							if bx, ok := x.(*BinaryExpr); ok {
								slt := evalStaticTypeOf(store, last, bx.Left)
								if bx.Op == SHL || bx.Op == SHR {
									srt := evalStaticTypeOf(store, last, bx.Right)
									bx.assertShiftExprCompatible1(store, last, slt, srt)
								}
							}
						}

						if !isUntyped(rt) { // right is typed
							checkOrConvertType(store, last, n, &n.Left, rt)
						} else {
							if shouldSwapOnSpecificity(lt, rt) {
								checkUntypedShiftExpr(n.Right)
							} else {
								checkUntypedShiftExpr(n.Left)
							}
						}
					} else if lcx.T == nil { // LHS is nil.
						// convert n.Left to typed-nil type.
						checkOrConvertType(store, last, n, &n.Left, rt)
					} else {
						if isUntyped(rt) {
							checkOrConvertType(store, last, n, &n.Right, lt)
						}
					}
				} else if ric { // right is const, left is not
					if isUntyped(rcx.T) {
						// Left not, Right untyped const ----------------
						// right is untyped const, left is not const, typed or untyped
						checkUntypedShiftExpr := func(x Expr) {
							if bx, ok := x.(*BinaryExpr); ok {
								if bx.Op == SHL || bx.Op == SHR {
									srt := evalStaticTypeOf(store, last, bx.Right)
									bx.assertShiftExprCompatible1(store, last, rt, srt)
								}
							}
						}
						// both untyped, e.g. 1<<s != 1.0
						if !isUntyped(lt) { // left is typed
							checkOrConvertType(store, last, n, &n.Right, lt)
						} else { // if one side is untyped shift expression, check type with lower specificity
							if shouldSwapOnSpecificity(lt, rt) {
								checkUntypedShiftExpr(n.Right)
							} else {
								checkUntypedShiftExpr(n.Left)
							}
						}
					} else if rcx.T == nil { // RHS is nil
						// refer to tests/files/types/eql_0f20.gno
						checkOrConvertType(store, last, n, &n.Right, lt)
					} else { // left is not const, right is typed const
						if isUntyped(lt) {
							checkOrConvertType(store, last, n, &n.Left, rt)
						}
					}
				} else {
					// Left not const, Right not const ------------------
					// non-shift non-const binary operator.
					liu, riu := isUntyped(lt), isUntyped(rt)
					if liu {
						if riu {
							if lt.TypeID() != rt.TypeID() {
								panic(fmt.Sprintf(
									"incompatible types in binary expression: %v %v %v",
									lt.TypeID(), n.Op, rt.TypeID()))
							}
							// convert untyped to typed
							checkOrConvertType(store, last, n, &n.Left, defaultTypeOf(lt))
							checkOrConvertType(store, last, n, &n.Right, defaultTypeOf(rt))
						} else { // left untyped, right typed
							checkOrConvertType(store, last, n, &n.Left, rt)
						}
					} else if riu { // left typed, right untyped
						checkOrConvertType(store, last, n, &n.Right, lt)
					} else { // both typed, refer to 0a1g.gno
						if !shouldSwapOnSpecificity(lt, rt) {
							checkOrConvertType(store, last, n, &n.Left, rt)
						} else {
							checkOrConvertType(store, last, n, &n.Right, lt)
						}
					}
				}
			// TRANS_LEAVE -----------------------
			case *CallExpr:
				// Func type evaluation.
				nft := evalStaticTypeOf(store, last, n.Func)
				switch bnft := baseOf(nft).(type) {
				case *TypeType:
					// Not a func type, but a type conversion.
					if len(n.Args) != 1 {
						panic("type conversion requires single argument")
					}
					n.NumArgs = 1
					arg0 := n.Args[0]
					ct := evalStaticType(store, last, n.Func)
					at := evalStaticTypeOf(store, last, n.Args[0])

					// OPTIMIZATION: Skip redundant type conversions when source and target types are identical
					if at != nil && ct.TypeID() == at.TypeID() && !isUntyped(at) {
						n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
						return n.Args[0], TRANS_CONTINUE
					}

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR CONST/BINARY/UNARY CASES:
					//----------------------------------------

					// Special case for const value conversions.
					var constConverted bool
					switch arg0 := arg0.(type) {
					case *ConstExpr:
						// As a special case, if a decimal cannot
						// be represented as an integer, it cannot be converted to one,
						// and the error is handled here.
						// Out of bounds errors are usually handled during evalConst().
						if isIntNum(ct) {
							if bd, ok := arg0.TypedValue.V.(BigdecValue); ok {
								if !isInteger(bd.V) {
									panic(fmt.Sprintf(
										"cannot convert %s to integer type",
										arg0))
								}
							}
							if isNumeric(at) {
								convertConst(store, last, n, arg0, ct)
								constConverted = true
							}
						} else if ct.Kind() == SliceKind {
							switch ct.Elem().Kind() {
							case Uint8Kind, Int32Kind:
								// The inverse conversion is handled later, see "[]byte/rune CASE".
								// The conversion is legal, set the target type.
								n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
								return n, TRANS_CONTINUE
							}
						}
						// (const) untyped decimal -> float64.
						// (const) untyped bigint -> int.
						if !constConverted {
							convertConst(store, last, n, arg0, nil)
						}

						// check legal type for nil
						if arg0.IsUndefined() {
							switch ct.Kind() { // special case for nil conversion check.
							case SliceKind, PointerKind, FuncKind, MapKind, InterfaceKind, ChanKind:
								convertConst(store, last, n, arg0, ct)
							default:
								panic(fmt.Sprintf(
									"cannot convert %v to %v",
									arg0, ct.Kind()))
							}
						}

						// evaluate the new expression.
						cx := evalConst(store, last, n)
						// The conversion is legal, set the target type.
						// Though cx may be undefined if ct is interface,
						// the ATTR_TYPEOF_VALUE is still interface.
						cx.SetAttribute(ATTR_TYPEOF_VALUE, ct)
						return cx, TRANS_CONTINUE
					case *BinaryExpr:
						// special case to evaluate type of binaryExpr/UnaryExpr which has untyped shift nested.
						if isUntyped(at) {
							switch arg0.Op {
							case EQL, NEQ, LSS, GTR, LEQ, GEQ:
								assertAssignableTo(n, at, ct)
							default:
								checkOrConvertType(store, last, n, &n.Args[0], ct)
							}
							// The conversion is legal, set the target type.
							n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
							return n, TRANS_CONTINUE
						}
						// continue...
					case *UnaryExpr:
						if isUntyped(at) {
							checkOrConvertType(store, last, n, &n.Args[0], ct)
							// The conversion is legal, set the target type.
							n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
							return n, TRANS_CONTINUE
						}
						// continue...
					} // escapes...

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR INTERFACE CASES:
					// (when either is an interface)
					//----------------------------------------

					ctBase := baseOf(ct)
					atBase := baseOf(at)
					_, ctIface := ctBase.(*InterfaceType)
					_, atIface := atBase.(*InterfaceType)
					if ctIface {
						// e.g. <iface type>(...)
						assertAssignableTo(n, at, ct)
						// The conversion is legal, set the target type.
						n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
						return n, TRANS_CONTINUE
					} else if atIface {
						// e.g. <concrete type>(<iface type>)
						panic(fmt.Sprintf("cannot convert %v to %v: need type assertion",
							at.TypeID(), ct.TypeID()))
					} // no escape.

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR NUMERIC CASE:
					//----------------------------------------

					if ctBase == StringType && isWhole(atBase) {
						// The conversion is legal, set the target type.
						n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
						return n, TRANS_CONTINUE
					}
					if isNumeric(ctBase) {
						if isNumeric(atBase) {
							// The conversion is legal, set the target type.
							n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
							return n, TRANS_CONTINUE
						} else {
							panic(fmt.Sprintf("cannot convert %v to %v: non-numeric to numeric",
								at, ct))
						}
					} else {
						if isNumeric(atBase) {
							panic(fmt.Sprintf("cannot convert %v to %v: non-numeric to numeric",
								at, ct))
						}
					} // escapes...

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR []byte/rune CASE:
					//----------------------------------------

					if ast, ok := atBase.(*SliceType); ok {
						switch ast.Elem().Kind() {
						case Uint8Kind, Int32Kind:
							if ct.Kind() == StringKind {
								// The conversion is legal, set the target type.
								n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
								return n, TRANS_CONTINUE
							}
						default: // continue...
						}
					} else if cst, ok := ctBase.(*SliceType); ok {
						switch cst.Elem().Kind() {
						case Uint8Kind, Int32Kind:
							if at.Kind() == StringKind {
								// The conversion is legal, set the target type.
								n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
								return n, TRANS_CONTINUE
							}
						default: // continue...
						}
					} // escapes...

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR POINTER CASE:
					// (*Foo)(&Bar{}} is legal, but only one level.
					//----------------------------------------

					if apt, ok := atBase.(*PointerType); ok {
						if cpt, ok := ctBase.(*PointerType); ok {
							if baseOf(apt.Elem()).TypeID() != baseOf(cpt.Elem()).TypeID() {
								panic(fmt.Sprintf("cannot convert %v (of type %v) to type %v",
									arg0, at, ct))
							} else {
								// The conversion is legal, set the target type.
								n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
								return n, TRANS_CONTINUE
							}
						}
					}

					//----------------------------------------
					// LEAVE (TYPE) CALL EXPR GENERAL CASE:
					//----------------------------------------

					if ctBase.TypeID() != atBase.TypeID() {
						panic(fmt.Sprintf("cannot convert %v (of type %v) to type %v",
							arg0, at, ct))
					}

					// The conversion is legal, set the target type.
					n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
					return n, TRANS_CONTINUE

				case *FuncType:
					//----------------------------------------
					// LEAVE (FUNC) CALL EXPR SPECIAL CASES:
					//----------------------------------------

					// NOTE: these appear to be actually special cases in go.
					// In general, a string is not assignable to []bytes
					// without conversion.
					if cx, ok := n.Func.(*ConstExpr); ok {
						fv := cx.GetFunc()
						if fv.PkgPath == uversePkgPath && fv.Name == "append" {
							if n.Varg && len(n.Args) == 2 {
								// If the second argument is a string,
								// convert to byteslice.
								args1 := n.Args[1]
								if evalStaticTypeOf(store, last, args1).Kind() == StringKind {
									fn, err := last.GetFuncNodeForExpr(store, n.Func)
									if err != nil {
										panic(fmt.Errorf("unexpected: %w", err))
									}
									ftx := fn.GetFuncTypeExpr()
									etx := ftx.Params[1].Type
									bsx := toConstTypeExpr(last, etx, gByteSliceType)
									args1 = Call(bsx, args1)
									args1 = Preprocess(nil, last, args1).(Expr)
									n.Args[1] = args1
								}
							} else {
								var tx *constTypeExpr // array type expr, lazily initialized
								// Another special case for append: adding untyped constants.
								// They must be converted to the array type for consistency.
								for i, arg := range n.Args[1:] {
									if _, ok := arg.(*ConstExpr); !ok {
										// Consider only constant expressions.
										continue
									}
									if t1 := evalStaticTypeOf(store, last, arg); t1 != nil && !isUntyped(t1) {
										// Consider only untyped values (including nil).
										continue
									}
									if tx == nil {
										// Get the array type from the first argument.
										s0 := evalStaticTypeOf(store, last, n.Args[0])
										tx = toConstTypeExpr(last, arg, s0.Elem())
									}
									// Convert to the array type.
									// NOTE: append([]<iface>{}, nil) remains nil arg.
									arg1 := Call(tx, arg)
									n.Args[i+1] = Preprocess(nil, last, arg1).(Expr)
								}
							}
						} else if fv.PkgPath == uversePkgPath && fv.Name == "copy" {
							if len(n.Args) == 2 {
								// If the second argument is a string,
								// convert to byteslice.
								args1 := n.Args[1]
								if evalStaticTypeOf(store, last, args1).Kind() == StringKind {
									bsx := toConstTypeExpr(last, args1, gByteSliceType)
									args1 = Call(bsx, args1)
									args1 = Preprocess(nil, last, args1).(Expr)
									n.Args[1] = args1
								}
							}
						} else if fv.PkgPath == uversePkgPath && fv.Name == "cross" {
							panic("cross(fn)(...) syntax is deprecated, use fn(cross,...)")
						} else if fv.PkgPath == uversePkgPath && fv.Name == "_cross_gno0p0" {
							if ctxpn.GetAttribute(ATTR_FIX_FROM) == GnoVerMissing {
								// This is only backwards compatibility for the gno 0.9
								// transpiler/fixer.  cross() is no longer used.
								// See adr/pr4264_lint_transpile.md for more info.
								//
								// Memoize *CallExpr.WithCross.
								pc, ok := ns[len(ns)-1].(*CallExpr)
								if !ok {
									panic("cross(fn) must be followed by a call")
								}
								pc.WithCross = true // bypass method with checks.
							} else {
								// only way _cross_gno0p0 appears is
								panic("_cross_gno0p0 is reserved")
							}
						} else if fv.PkgPath == uversePkgPath && fv.Name == "crossing" {
							if ctxpn.GetAttribute(ATTR_FIX_FROM) != GnoVerMissing {
								panic("crossing() is reserved and deprecated")
							}
						} else if fv.PkgPath == uversePkgPath && fv.Name == "attach" {
							// reserve attach() so we can support it later.
							panic("attach() not yet supported")
						}
					}

					//----------------------------------------
					// LEAVE (FUNC) CALL EXPR GENERAL CASE:
					//----------------------------------------

					ft := bnft
					if ft.IsCrossing() {
						// Special case when ctxpn.PkgPath is "testing",
						// pkg/test/test.gno will pass toConstExpr(Nx(`.cur`), NewConcreteRealm())
						// of the callee function's realm. Normally this isn't possible,
						// as non-realm packages do not have access to `cur`,
						// and you cannot pass `cur` to an external realm anyways.
						if ctxpn.PkgPath == TestingBasePkgPath {
							goto LEAVE_CALL_EXPR_END_CHECK_CROSSING
						}
						// START Check validity of crossing arg n.Args[0].(*NameExpr).
						// TODO: Refactor this out into a function call.
						{
							// a non-crossing call of a crossing function (passing in `cur`) may
							// only happen with a `cur` declared as the first realm argument
							// of a containing function.
							if len(n.Args) == 0 {
								panic(fmt.Sprintf("missing realm argument in calling crossing function call %v (expected cur or cross)", n))
							}
							nx, ok := n.Args[0].(*NameExpr)
							if !ok || nx.Name != Name("cur") && nx.Name != Name(".cur") && nx.Name != Name("cross") {
								panic(fmt.Sprintf("only `cur` and `cross` are allowed as the first argument to a crossing function but got %s", n.Args[0]))
							}
							switch nx.Name {
							case Name(".cur"):
								if _, ok := skipFile(last).(*PackageNode); !ok {
									// .cur should only be used from
									// m.RunMainMaybeCrossing() or m.runFunc(maybeCrossing=true),
									// or as a special case in TestFoo(cur realm, t *testing.T),
									// but in the latter case nx.Name isn't `.cur`, but
									// eval(n.Args[0]).(*ConstExpr) and handled separately above.
									panic(fmt.Sprintf("unexpected context for .cur; wanted last.(*PackageNode), got last.(%T)", last))
								}
								// evaluation was skipped TRANS_LEAVE *NameExpr.
								crlm := NewConcreteRealm(ctxpn.PkgPath)
								n.Args[0] = toConstExpr(nx, crlm)
							case Name("cur"): // non-crossing call of a crossing function.
								// Try to check that called function is local.
								// NOTE: We don't necessaryily know statically
								// whether n.Func is a local realm function or external,
								// so this check must also happen at runtime.
								ftv, err := tryEvalStatic(store, ctxpn, last, n.Func)
								if err == nil {
									// This is fine; e.g. somefunc()(cur,...)
								} else if ftv.IsUndefined() {
									// Interface... what can we do?
								} else {
									fpp := ftv.GetUnboundFunc().PkgPath
									if fpp != ctxpn.PkgPath {
										panic(fmt.Sprintf("cannot cur-call to external realm function %s.%v from %v",
											fpp, n.Func, ctxpn.PkgPath))
									}
								}
								// Check `cur` directly from parent crossing function's argument.
								dbn := last.GetBlockNodeForPath(store, nx.Path)
								switch dbn := dbn.(type) {
								case *FuncDecl:
									// dft := evalStaticType(store, last, &dbn.Type).(*FuncType)
									dft := getType(&dbn.Type).(*FuncType)
									if !dft.IsCrossing() {
										panic("only the `cur` argument of a containing crossing function maybe passed by cross-call")
									}
									// at this point we know that `cur` is from a containing crossing function.
									// NOTE: TRANS_ENTER *FuncTypeExpr ensures that `cur realm` is the first
									// argument of the crossing function.
								case *FuncLitExpr:
									// dft := evalStaticType(store, last, &dbn.Type).(*FuncType)
									dft := getType(&dbn.Type).(*FuncType)
									if !dft.IsCrossing() {
										panic("only the `cur` argument of a containing crossing function maybe passed by cross-call")
									}
									// at this point we know that `cur` is from a containing crossing function.
									// NOTE: TRANS_ENTER *FuncTypeExpr ensures that `cur realm` is the first
									// argument of the crossing function.
								default:
									panic("only the `cur` argument of a containing crossing function maybe passed by cross-call")
								}
								// the `cur` cannot be passed as capture through a func lit.
								fle, _, found := findFirstClosure(stack, dbn)
								if found && dbn != fle {
									panic(fmt.Sprintf("`cur realm` cannot be used as a closure capture, but found %v", fle))
								}
							case Name("cross"):
								// n is a valid crossing call of a crossing function.
								n.SetWithCross()
								// evaluation was skipped TRANS_LEAVE *NameExpr.
								n.Args[0] = constNil(nx)
							default:
								panic("should not happen")
							} // END Check validity of crossing arg n.Args[0].(*NameExpr).
						}
					LEAVE_CALL_EXPR_END_CHECK_CROSSING:
					} // END if ft.IsCrossing()

					hasVarg := ft.HasVarg()
					isVarg := n.Varg
					embedded := false
					var argTVs []TypedValue
					minArgs := len(ft.Params)
					if hasVarg {
						minArgs--
					}
					numArgs := countNumArgs(store, last, n) // isVarg?
					n.NumArgs = numArgs

					// Check input arg count.
					if len(n.Args) == 1 && numArgs > 1 {
						// special case of x(f()) form:
						// use the number of results instead.
						if isVarg {
							panic("should not happen")
						}
						embedded = true
						pcx := n.Args[0].(*CallExpr)
						argTVs = getResultTypedValues(pcx)
						if !hasVarg {
							if numArgs != len(ft.Params) {
								panic(fmt.Sprintf(
									"wrong argument count in call to %s; want %d got %d (with embedded call expr as arg)",
									n.Func.String(),
									len(ft.Params),
									numArgs,
								))
							}
						} else if hasVarg && !isVarg {
							if numArgs < len(ft.Params)-1 {
								panic(fmt.Sprintf(
									"not enough arguments in call to %s; want %d (besides variadic) got %d (with embedded call expr as arg)",
									n.Func.String(),
									len(ft.Params)-1,
									numArgs))
							}
						}
					} else if !hasVarg {
						argTVs = evalStaticTypedValues(store, last, n.Args...)
						if len(n.Args) != len(ft.Params) {
							panic(fmt.Sprintf(
								"wrong argument count in call to %s; want %d got %d",
								n.Func.String(),
								len(ft.Params),
								len(n.Args),
							))
						}
					} else if hasVarg && !isVarg {
						argTVs = evalStaticTypedValues(store, last, n.Args...)
						if len(n.Args) < len(ft.Params)-1 {
							panic(fmt.Sprintf(
								"not enough arguments in call to %s; want %d (besides variadic) got %d",
								n.Func.String(),
								len(ft.Params)-1,
								len(n.Args)))
						}
					} else if hasVarg && isVarg {
						argTVs = evalStaticTypedValues(store, last, n.Args...)
						if len(n.Args) != len(ft.Params) {
							panic(fmt.Sprintf(
								"not enough arguments in call to %s; want %d (including variadic) got %d",
								n.Func.String(),
								len(ft.Params),
								len(n.Args)))
						}
					} else {
						panic("should not happen")
					}
					// Specify function param/result generics.
					sft := ft.Specify(store, n, argTVs, isVarg)
					spts := sft.Params
					srts := FieldTypeList(sft.Results).Types()
					// If generics were specified, override attr
					// and constexpr with specified types.  Also
					// copy the function value with updated type.
					n.Func.SetAttribute(ATTR_TYPEOF_VALUE, sft)
					if cx, ok := n.Func.(*ConstExpr); ok {
						fv := cx.V.(*FuncValue)
						fv2 := fv.Copy(nilAllocator)
						fv2.Type = sft
						cx.T = sft
						cx.V = fv2
					} else if sft.TypeID() != ft.TypeID() {
						panic("non-const function value should have no generics")
					}
					n.SetAttribute(ATTR_TYPEOF_VALUE, &tupleType{Elts: srts})
					// Check given argument type against required.
					// Also replace const Args with *ConstExpr unless embedded.
					if embedded {
						if isVarg {
							panic("should not happen")
						}
						for i, tv := range argTVs {
							if hasVarg {
								if (len(spts) - 1) <= i {
									assertAssignableTo(n, tv.T, spts[len(spts)-1].Type.Elem())
								} else {
									assertAssignableTo(n, tv.T, spts[i].Type)
								}
							} else {
								assertAssignableTo(n, tv.T, spts[i].Type)
							}
						}
					} else {
						for i := range n.Args { // iterate args
							if hasVarg {
								if (len(spts) - 1) <= i {
									if isVarg {
										if len(spts) <= i {
											panic("expected final vargs slice but got many")
										}
										checkOrConvertType(store, last, n, &n.Args[i], spts[i].Type)
									} else {
										checkOrConvertType(store, last, n, &n.Args[i],
											spts[len(spts)-1].Type.Elem())
									}
								} else {
									checkOrConvertType(store, last, n, &n.Args[i], spts[i].Type)
								}
							} else {
								checkOrConvertType(store, last, n, &n.Args[i], spts[i].Type)
							}
						}
					}
				default:
					panic(fmt.Sprintf(
						"unexpected func type %v (%v)",
						nft, reflect.TypeOf(nft)))
				}

			// TRANS_LEAVE -----------------------
			case *IndexExpr:
				dt := evalStaticTypeOf(store, last, n.X)
				if dt.Kind() == PointerKind {
					// if a is a pointer to an array,
					// a[low : high : max] is shorthand
					// for (*a)[low : high : max]
					dt = dt.Elem()
					n.X = &StarExpr{X: n.X}
					setPreprocessed(n.X)
				}
				switch dt.Kind() {
				case StringKind, ArrayKind, SliceKind:
					// Replace const index with int *ConstExpr,
					// or if not const, assert integer type..
					checkOrConvertIntegerKind(store, last, n, n.Index)
				case MapKind:
					mt := baseOf(dt).(*MapType)
					checkOrConvertType(store, last, n, &n.Index, mt.Key)
				default:
					panic(fmt.Sprintf(
						"unexpected index base kind for type %s",
						dt.String()))
				}

			// TRANS_LEAVE -----------------------
			case *SliceExpr:
				// Replace const L/H/M with int *ConstExpr,
				// or if not const, assert integer type..
				checkOrConvertIntegerKind(store, last, n, n.Low)
				checkOrConvertIntegerKind(store, last, n, n.High)
				checkOrConvertIntegerKind(store, last, n, n.Max)

				t := evalStaticTypeOf(store, last, n.X)

				// if n.X is untyped, convert to corresponding type
				if isUntyped(t) {
					dt := defaultTypeOf(t)
					checkOrConvertType(store, last, n, &n.X, dt)
				}

			// TRANS_LEAVE -----------------------
			case *TypeAssertExpr:
				if n.Type == nil {
					panic("should not happen")
				}

				// Type assertions on the blank identifier are illegal.

				if isBlankIdentifier(n.X) {
					panic("cannot use _ as value or type")
				}

				// ExprStmt of form `x.(<type>)`,
				// or special case form `c, ok := x.(<type>)`.
				t := evalStaticTypeOf(store, last, n.X)
				baseType := baseOf(t) // The base type of the asserted value must be an interface.
				switch baseType.(type) {
				case *InterfaceType:
					break
				default:
					panic(
						fmt.Sprintf(
							"invalid operation: %s (variable of type %s) is not an interface",
							n.X.String(),
							t.String(),
						),
					)
				}

			// TRANS_LEAVE -----------------------
			case *UnaryExpr:
				xt := evalStaticTypeOf(store, last, n.X)
				n.AssertCompatible(xt)

				// Replace with *ConstExpr if const X.
				if isConst(n.X) {
					cx := evalConst(store, last, n)
					return cx, TRANS_CONTINUE
				}

			// TRANS_LEAVE -----------------------
			case *CompositeLitExpr:
				// Get or evaluate composite type.
				clt := evalStaticType(store, last, n.Type)
				n.Type = toConstTypeExpr(last, n.Type, clt)
				// Replace const Elts with default *ConstExpr.
				switch cclt := baseOf(clt).(type) {
				case *StructType:
					if n.IsKeyed() {
						for i := range n.Elts {
							key := n.Elts[i].Key.(*NameExpr).Name
							path := cclt.GetPathForName(key)
							ft := cclt.GetStaticTypeOfAt(path)
							checkOrConvertType(store, last, n, &n.Elts[i].Value, ft)
						}
					} else {
						for i := range n.Elts {
							ft := cclt.Fields[i].Type
							checkOrConvertType(store, last, n, &n.Elts[i].Value, ft)
						}
					}
				case *ArrayType:
					for i := range n.Elts {
						convertType(store, last, n, &n.Elts[i].Key, IntType)
						checkOrConvertType(store, last, n, &n.Elts[i].Value, cclt.Elt)
					}
				case *SliceType:
					for i := range n.Elts {
						convertType(store, last, n, &n.Elts[i].Key, IntType)
						checkOrConvertType(store, last, n, &n.Elts[i].Value, cclt.Elt)
					}
				case *MapType:
					for i := range n.Elts {
						checkOrConvertType(store, last, n, &n.Elts[i].Key, cclt.Key)
						checkOrConvertType(store, last, n, &n.Elts[i].Value, cclt.Value)
					}
				default:
					panic(fmt.Sprintf(
						"unexpected composite type %s",
						clt.String()))
				}
				// If variadic array lit, measure.
				if at, ok := clt.(*ArrayType); ok {
					if at.Vrd {
						idx := 0
						for _, elt := range n.Elts {
							if elt.Key == nil {
								idx++
							} else {
								k := int(evalConst(store, last, elt.Key).ConvertGetInt())
								if idx <= k {
									idx = k + 1
								} else {
									panic("array lit key out of order")
								}
							}
						}
						// update type
						// (dontcare)
						// at.Vrd = false
						at.Len = idx
						// update node
						cx := constInt(n, int64(idx))
						unconst(n.Type).(*ArrayTypeExpr).Len = cx
					}
				}

			// TRANS_LEAVE -----------------------
			case *KeyValueExpr:
				// NOTE: For simplicity we just
				// use the *CompositeLitExpr.
			// TRANS_LEAVE -----------------------
			case *StarExpr:
				xt := evalStaticTypeOf(store, last, n.X)
				if xt == nil {
					panic("invalid operation: cannot indirect nil")
				}
				if xt.Kind() != PointerKind && xt.Kind() != TypeKind {
					panic(fmt.Sprintf("invalid operation: cannot indirect %s (variable of type %s)", n.X.String(), xt.String()))
				}
			// TRANS_LEAVE -----------------------
			case *SelectorExpr:
				xt := evalStaticTypeOf(store, last, n.X)

				// Set selector path based on xt's type.
				switch cxt := xt.(type) {
				case *PointerType, *DeclaredType, *StructType, *InterfaceType:
					tr, _, rcvr, _, aerr := findEmbeddedFieldType(ctxpn.PkgPath, cxt, n.Sel, nil)
					if aerr {
						panic(fmt.Sprintf("cannot access %s.%s from %s",
							cxt.String(), n.Sel, ctxpn.PkgPath))
					} else if tr == nil {
						panic(fmt.Sprintf("missing field %s in %s",
							n.Sel, cxt.String()))
					}

					if len(tr) > 1 {
						// (the last vp, tr[len(tr)-1], is for n.Sel)
						if debug {
							if tr[len(tr)-1].Name != n.Sel {
								panic("should not happen")
							}
						}
						// replace n.X w/ tr[:len-1] selectors applied.
						nx2 := n.X
						for _, vp := range tr[:len(tr)-1] {
							nx2 = &SelectorExpr{
								X:    nx2,
								Path: vp,
								Sel:  vp.Name,
							}
						}
						// recursively preprocess new n.X.
						n.X = Preprocess(store, last, nx2).(Expr)
					}
					// nxt2 may not be xt anymore.
					// (even the dereferenced of xt and nxt2 may not
					// be the same, with embedded fields)
					nxt2 := evalStaticTypeOf(store, last, n.X)
					// Case 1: If receiver is pointer type but n.X is
					// not:
					if rcvr != nil &&
						rcvr.Kind() == PointerKind &&
						nxt2.Kind() != PointerKind {
						// Go spec: "If x is addressable and &x's
						// method set contains m, x.m() is shorthand
						// for (&x).m()"
						// Go spec: "As with method calls, a reference
						// to a non-interface method with a pointer
						// receiver using an addressable value will
						// automatically take the address of that
						// value: t.Mp is equivalent to (&t).Mp."
						//
						// convert to (&x).m, but leave xt as is.
						n.X = &RefExpr{X: n.X}
						setPreprocessed(n.X)
						switch tr[len(tr)-1].Type {
						case VPDerefPtrMethod:
							// When ptr method was called like x.y.z(), where x
							// is a pointer, y is an embedded struct, and z
							// takes a pointer receiver.  That becomes
							// &(x.y).z().
							// The x.y receiver wasn't originally a pointer,
							// yet the trail was
							// [VPSubrefField,VPDerefPtrMethod].
						case VPPtrMethod:
							tr[len(tr)-1].Type = VPDerefPtrMethod
						default:
							panic(fmt.Sprintf(
								"expected ultimate VPPtrMethod but got %v in trail %v",
								tr[len(tr)-1].Type,
								tr,
							))
						}
					} else if len(tr) > 0 &&
						tr[len(tr)-1].IsDerefType() &&
						nxt2.Kind() != PointerKind {
						// Case 2: If tr[0] is deref type, but xt
						// is not pointer type, replace n.X with
						// &RefExpr{X: n.X}.
						n.X = &RefExpr{X: n.X}
						setPreprocessed(n.X)
					}
					// bound method or underlying.
					// TODO check for unexported fields.
					n.Path = tr[len(tr)-1]
					// n.Path = cxt.GetPathForName(n.Sel)
				case *PackageType:
					var pv *PackageValue
					if cx, ok := n.X.(*ConstExpr); ok {
						// NOTE: *Machine.TestMemPackage() needs this
						// to pass in an imported package as *ConstExpr.
						pv = cx.V.(*PackageValue)
					} else {
						// otherwise, packages can only be referred to by
						// *NameExprs, and cannot be copied.
						pvc := evalConst(store, last, n.X)
						pv_, ok := pvc.V.(*PackageValue)
						if !ok {
							panic(fmt.Sprintf(
								"missing package in selector expr %s",
								n.String()))
						}
						pv = pv_
					}
					pn := pv.GetPackageNode(store)
					// Ensure exposed or package path match.
					if !(isUpper(string(n.Sel)) || ctxpn.PkgPath == pv.PkgPath) {
						panic(fmt.Sprintf("cannot access %s.%s from %s",
							pv.PkgPath, n.Sel, ctxpn.PkgPath))
					}
					// Ensure no modification from code declared in external realm.
					//  1. ext.Name = xyz    // illegal assignment (static & runtime)
					//  2. (ext.Foo).Bar = 1 // illegal assignment (runtime)
					// For defensiveness the runtime check remains despite static checking here.
					// See also Machine.IsReadOnly().
					if ctxpn.PkgPath != pv.PkgPath {
						if ftype == TRANS_ASSIGN_LHS {
							panic(fmt.Sprintf("cannot directly mutate %s.%s from %s",
								pv.PkgPath, n.Sel, ctxpn.PkgPath))
						}
					}
					// NOTE: this can happen with software upgrades,
					// with multiple versions of the same package path.
					n.Path = pn.GetPathForName(store, n.Sel)
					// Produce const expr if typed or untyped const.
					tt := pn.GetStaticTypeOfAt(store, n.Path)
					if isUntyped(tt) || pn.GetIsConstAt(store, n.Path) {
						cx := evalConst(store, last, n)
						return cx, TRANS_CONTINUE
					}
				case *TypeType:
					// unbound method
					xt := evalStaticType(store, last, n.X)
					switch ct := xt.(type) {
					case *PointerType:
						dt := ct.Elt.(*DeclaredType)
						n.Path = dt.GetUnboundPathForName(n.Sel)
					case *DeclaredType:
						n.Path = ct.GetUnboundPathForName(n.Sel)
					default:
						panic(fmt.Sprintf(
							"unexpected selector expression type value %s",
							xt.String()))
					}
				default:
					panic(fmt.Sprintf(
						"unexpected selector expression type %v",
						reflect.TypeOf(xt)))
				}

			// TRANS_LEAVE -----------------------
			case *FieldTypeExpr:
				// Replace const Tag with default *ConstExpr.
				convertIfConst(store, last, n, n.Tag)

			// TRANS_LEAVE -----------------------
			case *ArrayTypeExpr:
				if n.Len == nil {
					// Calculate length at *CompositeLitExpr:LEAVE
				} else {
					// Replace const Len with int *ConstExpr.
					cx := evalConst(store, last, n.Len)
					convertConst(store, last, n, cx, IntType)
					n.Len = cx
				}
				// NOTE: For all TypeExprs, the node is not replaced
				// with *constTypeExprs (as *ConstExprs are) because
				// we want to support type logic at runtime.
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *SliceTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *InterfaceTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *ChanTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *FuncTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *MapTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *StructTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *AssignStmt:
				fmt.Println("---Trans_Leave, assignStmt: ", n)
				n.AssertCompatible(store, last)
				if n.Op == ASSIGN {
					for _, lh := range n.Lhs {
						if ne, ok := lh.(*NameExpr); ok {
							if !last.GetStaticBlock().IsAssignable(store, ne.Name) {
								panic("not assignable")
							}
						}
					}
				}

				// NOTE: keep DEFINE and ASSIGN in sync.
				if n.Op == DEFINE {
					// Rhs consts become default *ConstExprs.
					for _, rx := range n.Rhs {
						// NOTE: does nothing if rx is "nil".
						convertIfConst(store, last, n, rx)
					}
					nameExprs := make([]*NameExpr, len(n.Lhs))
					for i := range len(n.Lhs) {
						nameExprs[i] = n.Lhs[i].(*NameExpr)
					}
					fmt.Println("---defineOrDecl...")
					defineOrDecl(store, last, n, false, nameExprs, nil, n.Rhs, true)
				} else { // ASSIGN, or assignment operation (+=, -=, <<=, etc.)
					// NOTE: Keep in sync with DEFINE above.
					if len(n.Lhs) > len(n.Rhs) {
						// check is done in assertCompatible where we also
						// asserted we have at lease one element in Rhs
						if cx, ok := n.Rhs[0].(*CallExpr); ok {
							// we decompose the a,b = x(...) for named and unamed
							// type value return in an assignments
							// Call case: a, b = x(...)
							ift := evalStaticTypeOf(store, last, cx.Func)
							cft := getGnoFuncTypeOf(store, ift)
							// check if we we need to decompose for named typed conversion in the function return results
							var decompose bool

							for i, rhsType := range cft.Results {
								lt := evalStaticTypeOf(store, last, n.Lhs[i])
								if lt != nil && isNamedConversion(rhsType.Type, lt) {
									decompose = true
									break
								}
							}
							if decompose {
								// only enter this section if cft.Results to be converted to Lhs type for named type conversion.
								// decompose a,b = x()
								// .decompose1, .decompose2 := x()  assignment statement expression (Op=DEFINE)
								// a,b = .decompose1, .decompose2   assignment statement expression ( Op=ASSIGN )
								// add the new statement to last.Body

								// step1:
								// create a hidden var with leading . (dot) the curBodyLen increase every time when there is a decomposition
								// because there could be multiple decomposition happens
								// we use both stmt index and return result number to differentiate the .decompose variables created in each assignment decompostion
								// ex. .decompose_3_2: this variable is created as the 3rd statement in the block, the 2nd parameter returned from x(),
								// create .decompose_1_1, .decompose_1_2 .... based on number of result from x()
								tmpExprs := make(Exprs, 0, len(cft.Results))
								for i := range cft.Results {
									rn := fmt.Sprintf(".decompose_%d_%d", index, i)
									tmpExprs = append(tmpExprs, Nx(rn))
								}
								// step2:
								// .decompose1, .decompose2 := x()
								dsx := &AssignStmt{
									Lhs: tmpExprs,
									Op:  DEFINE,
									Rhs: n.Rhs,
								}
								// dsx.SetSpan(n.GetSpan())
								dsx = Preprocess(store, last, dsx).(*AssignStmt)

								// step3:

								// a,b = .decompose1, .decompose2
								// assign stmt expression
								// The right-hand side will be converted to a call expression for named/unnamed conversion.
								// tmpExprs is a []Expr; we make a copy of tmpExprs to prevent dsx.Lhs in the previous statement (dsx) from being changed by side effects.
								// If we don't copy tmpExprs, when asx.Rhs is converted to a const call expression during the preprocessing of the AssignStmt asx,
								// dsx.Lhs will change from []NameExpr to []CallExpr.
								// This side effect would cause a panic when the machine executes the dsx statement, as it expects Lhs to be []NameExpr.

								asx := &AssignStmt{
									Lhs: n.Lhs,
									Op:  ASSIGN,
									Rhs: copyExprs(tmpExprs),
								}
								// asx.SetSpan(n.GetSpan())
								asx = Preprocess(store, last, asx).(*AssignStmt)

								// step4:
								// replace the original stmt with two new stmts
								body := last.GetBody()
								// we need to do an in-place replacement while leaving the current node
								n.Attributes = dsx.Attributes
								n.Lhs = dsx.Lhs
								n.Op = dsx.Op
								n.Rhs = dsx.Rhs

								//  insert a assignment statement a,b = .decompose1,.decompose2 AFTER the current statement in the last.Body.
								body = append(body[:index+1], append(Body{asx}, body[index+1:]...)...)
								last.SetBody(body)
							} // end of the decomposition

							// Last step: we need to insert the statements to FuncValue.body of PackageNopde.Values[i].V
							// updating FuncValue.body=FuncValue.Source.Body in updates := pn.PrepareNewValues(pv) during preprocess.
							// we updated FuncValue from source.
						}
					} else { // len(Lhs) == len(Rhs)
						switch n.Op {
						case SHL_ASSIGN, SHR_ASSIGN:
							// Special case if shift assign <<= or >>=.
							convertType(store, last, n, &n.Rhs[0], UintType)
						case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN:
							// e.g. a += b, single value for lhs and rhs,
							lt := evalStaticTypeOf(store, last, n.Lhs[0])
							checkOrConvertType(store, last, n, &n.Rhs[0], lt)
						default: // all else, like BAND_ASSIGN, etc
							// General case: a, b = x, y.
							for i, lx := range n.Lhs {
								lt := evalStaticTypeOf(store, last, lx)

								// if lt is interface, nothing will happen
								checkOrConvertType(store, last, n, &n.Rhs[i], lt)
							}
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *BranchStmt:
				switch n.Op {
				case BREAK:
					if n.Label == "" {
						findBreakableNode(last, store)
					} else {
						// Make sure that the label exists, either for a switch or a
						// BranchStmt.
						if !isSwitchLabel(ns, n.Label) {
							findBranchLabel(last, n.Label)
						}
					}
				case CONTINUE:
					if n.Label == "" {
						findContinuableNode(last, store)
					} else {
						if isSwitchLabel(ns, n.Label) {
							panic(fmt.Sprintf("invalid continue label %q\n", n.Label))
						}
						findBranchLabel(last, n.Label)
					}
				case GOTO:
					_, blockDepth, frameDepth, index := findGotoLabel(last, n.Label)
					n.BlockDepth = blockDepth
					n.FrameDepth = frameDepth
					n.BodyIndex = index
				case FALLTHROUGH:
					swchC, ok := last.(*SwitchClauseStmt)
					if !ok {
						// fallthrough is only allowed in a switch statement
						panic("fallthrough statement out of place")
					}

					if n.GetAttribute(ATTR_LAST_BLOCK_STMT) != true {
						// no more clause after the one executed, this is not allowed
						panic("fallthrough statement out of place")
					}

					// last is a switch clause, find its index in the switch and assign
					// it to the fallthrough node BodyIndex. This will be used at
					// runtime to determine the next switch clause to run.
					swch := lastSwitch(ns)

					if swch.IsTypeSwitch {
						// fallthrough is not allowed in type switches
						panic("cannot fallthrough in type switch")
					}

					for i := range swch.Clauses {
						if i == len(swch.Clauses)-1 {
							panic("cannot fallthrough final case in switch")
						}

						if &swch.Clauses[i] == swchC {
							// switch clause found
							n.BodyIndex = i
							break
						}
					}
				default:
					panic("should not happen")
				}

			// TRANS_LEAVE -----------------------
			case *IncDecStmt:
				xt := evalStaticTypeOf(store, last, n.X)
				n.AssertCompatible(xt)

			// TRANS_LEAVE -----------------------
			case *ForStmt:
				// Cond consts become bool *ConstExprs.
				checkOrConvertBoolKind(store, last, n, n.Cond)

			// TRANS_LEAVE -----------------------
			case *IfStmt:
				// Cond consts become bool *ConstExprs.
				checkOrConvertBoolKind(store, last, n, n.Cond)

			// TRANS_LEAVE -----------------------
			case *RangeStmt:
				// NOTE: k,v already defined @ TRANS_BLOCK.
				n.AssertCompatible(store, last)

			// TRANS_LEAVE -----------------------
			case *ReturnStmt:
				fnode, ft := funcOf(last)
				// Mark return statement as needing to copy
				// results to named heap items of block.
				// This is necessary because if the results
				// are unnamed, they are omitted from block.
				if ft.Results.IsNamed() && len(n.Results) != 0 {
					// NOTE: We don't know yet whether any
					// results are heap items or not, as
					// they are found after this
					// preprocessor step.  Either find heap
					// items before, or do another pass to
					// demote for speed.
					n.CopyResults = true
				}
				// Check number of return arguments.
				if len(n.Results) != len(ft.Results) {
					if len(n.Results) == 0 {
						if ft.Results.IsNamed() {
							// ok, results already named.
						} else {
							panic(fmt.Sprintf("expected %d return values; got %d",
								len(ft.Results),
								len(n.Results),
							))
						}
					} else if len(n.Results) == 1 {
						if cx, ok := n.Results[0].(*CallExpr); ok {
							ift := evalStaticTypeOf(store, last, cx.Func)
							cft := getGnoFuncTypeOf(store, ift)
							if len(cft.Results) != len(ft.Results) {
								panic(fmt.Sprintf("expected %d return values; got %d",
									len(ft.Results),
									len(cft.Results),
								))
							}
							// else, nothing more to do.
						} else {
							panic(fmt.Sprintf("expected %d return values; got %d",
								len(ft.Results),
								len(n.Results),
							))
						}
					} else {
						panic(fmt.Sprintf("expected %d return values; got %d",
							len(ft.Results),
							len(n.Results),
						))
					}
				} else {
					// Results consts become default *ConstExprs.
					for i := range n.Results {
						rtx := ft.Results[i].Type
						rt := evalStaticType(store, fnode.GetParentNode(nil), rtx)
						if isGeneric(rt) {
							// cannot convert generic result,
							// the result type depends.
							// XXX how to deal?
							panic("not yet implemented")
						} else {
							checkOrConvertType(store, last, n, &n.Results[i], rt)
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *SendStmt:
				// Value consts become default *ConstExprs.
				checkOrConvertType(store, last, n, &n.Value, nil)

			// TRANS_LEAVE -----------------------
			case *SelectCaseStmt:
				// maybe receive defines.
				// if as, ok := n.Comm.(*AssignStmt); ok {
				//     handled by case *AssignStmt.
				// }

			// TRANS_LEAVE -----------------------
			case *SwitchStmt:
				// Ensure type switch cases are unique.
				if n.IsTypeSwitch {
					types := map[string]struct{}{}
					for _, clause := range n.Clauses {
						for _, casetype := range clause.Cases {
							var ctstr string
							ctype := casetype.(*constTypeExpr).Type
							if ctype == nil {
								ctstr = nilStr
							} else {
								ctstr = casetype.(*constTypeExpr).Type.String()
							}
							if _, exists := types[ctstr]; exists {
								panic(fmt.Sprintf(
									"duplicate type %s in type switch",
									ctstr))
							}
							types[ctstr] = struct{}{}
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *ValueDecl:
				assertValidAssignRhs(store, last, n)

				// evaluate value if const expr.
				if n.Const {
					// NOTE: may or may not be a *ConstExpr,
					// but if not, make one now.
					for i, vx := range n.Values {
						assertValidConstExpr(store, last, n, vx)
						n.Values[i] = evalConst(store, last, vx)
					}
				}
				// else, value(s) may already be *ConstExpr, but
				// otherwise as far as we know the
				// expression is not a const expr, so no
				// point evaluating it further.  this makes
				// the implementation differ from
				// runDeclaration(), as this uses OpStaticTypeOf.

				nameExprs := make([]*NameExpr, len(n.NameExprs))
				for i := range n.NameExprs {
					nameExprs[i] = &n.NameExprs[i]
				}
				defineOrDecl(store, last, n, n.Const, nameExprs, n.Type, n.Values, false)

				// TODO make note of constance in static block for
				// future use, or consider "const paths".  set as
				// preprocessed.

			// TRANS_LEAVE -----------------------
			case *TypeDecl:
				// fmt.Println("---Trans_Leave, TypeDecl, n: ", n)
				if n.Name == blankIdentifier {
					n.Path = NewValuePathBlock(0, 0, blankIdentifier)
					return n, TRANS_CONTINUE
				}
				// 'tmp' is a newly constructed type value, where
				// any recursive references refer to the
				// original type constructed by predefineRecursively()
				// during during *TypeDecl:ENTER.
				//
				// For recursive definitions 'tmp' is a new
				// unnamed type that is different than the
				// original, but its elements will contain the
				// original. It is later copied back into the
				// original type, thus completing the recursive
				// definition. (thus called 'tmp')
				tmp := evalStaticType(store, last, n.Type)
				// 'dst' is the actual original type structure
				// as contructed by predefineRecursively().
				//
				// In short, the relationship between tmp and dst is:
				//   `type dst tmp`.
				dstTV := last.GetSlot(store, n.Name, true)
				dstT := dstTV.GetType()
				switch dstT := dstT.(type) {
				case *FuncType:
					*dstT = *(tmp.(*FuncType))
				case *ArrayType:
					*dstT = *(tmp.(*ArrayType))
				case *SliceType:
					*dstT = *(tmp.(*SliceType))
				case *InterfaceType:
					*dstT = *(tmp.(*InterfaceType))
				case *ChanType:
					*dstT = *(tmp.(*ChanType))
				case *MapType:
					*dstT = *(tmp.(*MapType))
				case *StructType:
					*dstT = *(tmp.(*StructType))
				case *DeclaredType:
					if n.IsAlias {
						// Nothing to do.
					} else {
						// Construct a temporary new *DeclaredType
						// and copy value to dst to keep the original pointer.
						//
						// NOTE: this is where the structured value
						// (e.g.  *ArrayType, *StructType) of declared
						// types are actually instantiated, not in
						// machine.go:runDeclaration().
						tmp2 := declareWith(ctxpn.PkgPath, last, n.Name, tmp)
						// if !n.IsAlias { // not sure why this was here.
						tmp2.Seal()
						// }
						*dstT = *tmp2
					}
				case PrimitiveType:
					dstTV.V = TypeValue{Type: tmp}
				case *PointerType:
					*dstT = *(tmp.(*PointerType))
				default:
					panic(fmt.Sprintf("unexpected type declaration type %v",
						reflect.TypeOf(dstTV)))
				}
				// We need to replace all references of the new
				// Type with old Type, including in attributes.
				n.Type.SetAttribute(ATTR_TYPE_VALUE, dstT)
				// Replace the type with *{},
				// otherwise methods would be un at runtime.
				n.Type = toConstTypeExpr(last, n.Type, dstT)
			}
			// end type switch statement
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}

		panic(fmt.Sprintf(
			"unknown stage %v", stage))
	})

	return nn
}

// defineOrDecl merges the code logic from op define (:=) and declare (var/const).
func defineOrDecl(
	store Store,
	bn BlockNode,
	n Node,
	isConst bool,
	nameExprs []*NameExpr,
	typeExpr Expr,
	valueExprs []Expr,
	isDefine bool, // if :=, nameExpr may not be the origin (if already defined)
) {
	numNames := len(nameExprs)
	numVals := len(valueExprs)

	if numVals > 1 && numNames != numVals {
		panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numVals))
	}

	sts := make([]Type, numNames) // static types
	tvs := make([]TypedValue, numNames)

	if numVals == 1 && numNames > 1 {
		parseMultipleAssignFromOneExpr(store, bn, n, sts, tvs, nameExprs, typeExpr, valueExprs[0])
	} else {
		parseAssignFromExprList(store, bn, n, sts, tvs, isConst, nameExprs, typeExpr, valueExprs)
	}

	node := skipFile(bn)

	for i, nx := range nameExprs {
		if nx.Name == blankIdentifier {
			nx.Path = NewValuePathBlock(0, 0, nx.Name)
		} else {
			_, ok := node.GetLocalIndex(nx.Name)
			if isDefine && ok {
				fmt.Println("------Define2, nx.Name: ", nx.Name)
				fmt.Println("------last: ", bn, reflect.TypeOf(bn))
				// Keep the original nx, this one is fake.
				node.Define2(isConst, nx.Name, sts[i], tvs[i], noNameSource)
			} else {
				nsType := NSValueDecl
				if isDefine {
					nsType = NSDefine
				}
				node.Define2(isConst, nx.Name, sts[i], tvs[i], NameSource{nx, n, nsType, i})
			}
			nx.Path = bn.GetPathForName(nil, nx.Name)
		}
	}
}

// parseAssignFromExprList parses assignment to multiple variables from a list of expressions.
// This function will alter the value of sts, tvs.
func parseAssignFromExprList(
	store Store,
	bn BlockNode,
	n Node,
	sts []Type,
	tvs []TypedValue,
	isConst bool,
	nameExprs []*NameExpr,
	typeExpr Expr,
	valueExprs []Expr,
) {
	numNames := len(nameExprs)

	// Ensure that function only return 1 value.
	for _, v := range valueExprs {
		if cx, ok := v.(*CallExpr); ok {
			tt, ok := evalStaticTypeOfRaw(store, bn, cx).(*tupleType)
			if ok && len(tt.Elts) > 1 {
				panic(fmt.Sprintf("multiple-value %s (value of type %s) in single-value context", cx.Func.String(), tt.Elts))
			}
		}
	}

	// Evaluate types and convert consts.
	if typeExpr != nil {
		// Only a single type can be specified.
		nt := evalStaticType(store, bn, typeExpr)
		for i := range numNames {
			sts[i] = nt
		}
		// Convert if const to nt.
		for i := range valueExprs {
			checkOrConvertType(store, bn, n, &valueExprs[i], nt)
		}
	} else if isConst {
		// Derive static type from values.
		for i, vx := range valueExprs {
			vt := evalStaticTypeOf(store, bn, vx)
			sts[i] = vt
		}
	} else { // T is nil, n not const => same as AssignStmt DEFINE
		// Convert n.Value to default type.
		for i, vx := range valueExprs {
			if cx, ok := vx.(*ConstExpr); ok {
				convertConst(store, bn, n, cx, nil)
				// convertIfConst(store, last, vx)
			} else {
				checkOrConvertType(store, bn, n, &vx, nil)
			}
			vt := evalStaticTypeOf(store, bn, vx)
			sts[i] = vt
		}
	}

	// Evaluate typed value for static definition.

	for i := range nameExprs {
		// Consider value if specified.
		if len(valueExprs) > 0 {
			vx := valueExprs[i]
			if cx, ok := vx.(*ConstExpr); ok &&
				!cx.TypedValue.IsUndefined() {
				if isConst {
					// const _ = <const_expr>: static block should contain value
					tvs[i] = cx.TypedValue
				} else {
					// var _ = <const_expr>: static block should NOT contain value
					tvs[i] = anyValue(cx.TypedValue.T)
				}
				continue
			}
		}
		// For var decls of non-const expr.
		st := sts[i]
		tvs[i] = anyValue(st)
	}
}

// parseMultipleAssignFromOneExpr parses assignment to multiple variables from a single expression.
// This function will alter the value of sts, tvs.
// Declare:
// - var a, b, c T = f()
// - var a, b = n.(T)
// - var a, b = n[i], where n is a map
// Assign:
// - a, b, c := f()
// - a, b := n.(T)
// - a, b := n[i], where n is a map
func parseMultipleAssignFromOneExpr(
	store Store,
	bn BlockNode,
	n Node,
	sts []Type,
	tvs []TypedValue,
	nameExprs []*NameExpr,
	typeExpr Expr,
	valueExpr Expr,
) {
	var tuple *tupleType
	numNames := len(nameExprs)
	switch expr := valueExpr.(type) {
	case *CallExpr:
		// Call case:
		// var a, b, c T = f()
		// a, b, c := f()
		valueType := evalStaticTypeOfRaw(store, bn, valueExpr)
		tuple = valueType.(*tupleType)
	case *TypeAssertExpr:
		// Type assert case:
		// var a, b = n.(T)
		// a, b := n.(T)
		tt := evalStaticType(store, bn, expr.Type)
		tuple = &tupleType{Elts: []Type{tt, BoolType}}
		expr.HasOK = true
	case *IndexExpr:
		// Map index case:
		// var a, b = n[i], where n is a map
		// a, b := n[i], where n is a map
		var mt *MapType
		dt := evalStaticTypeOf(store, bn, expr.X)
		mt, ok := baseOf(dt).(*MapType)
		if !ok {
			panic(fmt.Sprintf("invalid index expression on %T", dt))
		}
		tuple = &tupleType{Elts: []Type{mt.Value, BoolType}}
		expr.HasOK = true
	default:
		panic(fmt.Sprintf("unexpected value expression type %T", expr))
	}

	if numValues := len(tuple.Elts); numValues != numNames {
		panic(
			fmt.Sprintf(
				"assignment mismatch: %d variable(s) but %s returns %d value(s)",
				numNames,
				valueExpr.String(),
				numValues,
			),
		)
	}

	var st Type = nil
	if typeExpr != nil {
		// Only a single type can be specified.
		st = evalStaticType(store, bn, typeExpr)
	}

	for i := range nameExprs {
		if st != nil {
			tt := tuple.Elts[i]
			if checkAssignableTo(n, tt, st) != nil {
				panic(
					fmt.Sprintf(
						"cannot use %v (value of type %s) as %s value in assignment",
						valueExpr.String(),
						tt.String(),
						st.String(),
					),
				)
			}
			sts[i] = st
		} else {
			// Set types as return types.
			sts[i] = tuple.Elts[i]
		}
		tvs[i] = anyValue(sts[i])
	}
}

// Identifies NameExprTypeHeapDefines.
// Also finds GotoLoopStmts.
// XXX DEPRECATED but kept here in case needed in the future.
// We may still want this for optimizing heap defines;
// the current implementation of codaHeapDefinesByUse/codaHeapUsesDemoteDefines
// produces false positives.
//
//nolint:unused
func codaGotoLoopDefines(ctx BlockNode, bn BlockNode) {
	// iterate over all nodes recursively.
	_ = TranscribeB(ctx, bn, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("codaGotoLoopDefines %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_ENTER:
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_BLOCK:
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_LEAVE:
			switch n := n.(type) {
			case *ForStmt, *RangeStmt:
				Transcribe(n,
					func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
						switch stage {
						case TRANS_ENTER:
							switch n := n.(type) {
							case *FuncLitExpr:
								// inner funcs.
								return n, TRANS_SKIP
							case *FuncDecl:
								panic("unexpected inner func decl")
							case *NameExpr:
								if n.Type == NameExprTypeDefine {
									n.Type = NameExprTypeHeapDefine
								}
							}
						}
						return n, TRANS_CONTINUE
					})
			case *BranchStmt:
				switch n.Op {
				case GOTO:
					bn, _, _, _ := findGotoLabel(last, n.Label)
					// already done in Preprocess:
					// n.Depth = depth
					// n.BodyIndex = index

					// NOTE: we must not use line numbers
					// for logic, as line numbers are not
					// guaranteed (see checkNodeLinesLocations).
					// Instead we rely on the transcribe order
					// and keep track of whether we've seen
					// the label and goto stmts respectively.
					//
					// DOES NOT WORK:
					// gotoLine := n.GetLine()
					// if labelLine >= gotoLine {
					//	return n, TRANS_SKIP
					// }
					var (
						label        = n.Label
						labelReached bool
						origGoto     = n
					)

					// Recurse and mark stmts as ATTR_GOTOLOOP_STMT.
					// NOTE: ATTR_GOTOLOOP_STMT is not used.
					Transcribe(bn,
						func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
							switch stage {
							case TRANS_ENTER:
								// Check to see if label reached.
								if _, ok := n.(Stmt); ok {
									// XXX HasLabel
									if n.GetLabel() == label {
										labelReached = true
									}
									// If goto < label,
									// then not a goto loop.
									if n == origGoto && !labelReached {
										return n, TRANS_EXIT
									}
								}

								// If label not reached, continue.
								if !labelReached {
									return n, TRANS_CONTINUE
								}

								// NOTE: called redundantly
								// for many goto stmts,
								// idempotenct updates only.
								switch n := n.(type) {
								// Skip the body of these:
								case *FuncLitExpr:
									if len(ns) > 0 {
										// inner funcs.
										return n, TRANS_SKIP
									}
									return n, TRANS_CONTINUE
								case *FuncDecl:
									if len(ns) > 0 {
										panic("unexpected inner func decl")
									}
									return n, TRANS_CONTINUE
								// Otherwise mark stmt as gotoloop.
								case Stmt:
									// we're done if we
									// re-encounter origGotoStmt.
									if n == origGoto {
										return n, TRANS_EXIT // done
									}
									return n, TRANS_CONTINUE
								// Special case, maybe convert
								// NameExprTypeDefine to
								// NameExprTypeHeapDefine.
								case *NameExpr:
									if n.Type == NameExprTypeDefine {
										n.Type = NameExprTypeHeapDefine
									}
								}
								return n, TRANS_CONTINUE
							}
							return n, TRANS_CONTINUE
						})
				}
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

func RestoreNode(ctx BlockNode, n Node) {
	// Iterate over all nodes recursively.
	_ = TranscribeB(ctx, n, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("RestoreNode %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:

		// ----------------------------------------
		case TRANS_LEAVE:
			switch n.(type) {
			case *TypeDecl:
				// no re-preprocess as it's reusing PredefineFileSet result.
				// otherwise will be wiped.
			default:
				n.SetAttribute(ATTR_PREPROCESSED, false)
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

func findHeapDefinedLoopvarByUse(ctx BlockNode, bn BlockNode) (stop bool) {
	var transForBody bool
	// Iterate over all nodes recursively.
	_ = TranscribeB(ctx, bn, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("findHeapDefinedLoopvarByUse %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}
		// only rename names in body.
		if ftype == TRANS_FOR_BODY {
			transForBody = true
		}
		if ftype == TRANS_FOR_INIT || ftype == TRANS_FOR_COND || ftype == TRANS_FOR_POST {
			transForBody = false
		}

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:

		// ----------------------------------------
		case TRANS_LEAVE:
			// fmt.Println("------Trans_Leave, n: ", n, reflect.TypeOf(n))
			// fmt.Println("---is transForBody: ", transForBody)
			switch n := n.(type) {
			case *RefExpr:
				lmx := LeftmostX(n.X)
				if nx, ok := lmx.(*NameExpr); ok {

					if !transForBody {
						return n, TRANS_CONTINUE
					}
					// fmt.Println("---refX..., nx: ", nx)
					if nx.Path.Type != VPBlock {
						return n, TRANS_CONTINUE
					}
					// Ignore blank identifers
					if nx.Name == blankIdentifier {
						return n, TRANS_CONTINUE
					}
					// Ignore package names
					if nx.GetAttribute(ATTR_PACKAGE_REF) != nil {
						return n, TRANS_CONTINUE
					}
					// Ignore decls names.
					if ftype == TRANS_VAR_NAME {
						return n, TRANS_CONTINUE
					}
					// Ignore := defines, etc.
					if nx.Type != NameExprTypeLoopVarUse {
						return n, TRANS_CONTINUE
					}
					if !strings.HasPrefix(string(nx.Name), ".loopvar") {
						return n, TRANS_CONTINUE
					}
					// Find the block where name is defined
					dbn := last.GetBlockNodeForPath(nil, nx.Path)
					if _, ok := dbn.(*ForStmt); !ok {
						return n, TRANS_CONTINUE
					}
					// The leftmost name of possibly nested index
					// and selector exprs.
					// e.g. leftmost.middle[0][2].rightmost

					origName := strings.TrimPrefix(string(nx.Name), ".loopvar_")
					redefineName := fmt.Sprintf("%s%s", ".redefine_", origName)

					if _, ok := dbn.(*ForStmt); ok {
						// true if found once
						stop = true
						if _, ok := dbn.GetAttribute(ATTR_REDEFINE_NAME).(string); !ok {
							dbn.SetAttribute(ATTR_REDEFINE_NAME, redefineName)
						}

						// Reset to type normal
						nx.Type = NameExprTypeNormal
						// Redefine name
						// nx.Name = Name(redefineName)
					}
				}
			case *NameExpr:
				// fmt.Println("---findHeapDefinedLoopvarByUse, n.Name: ", n.Name)
				// NOTE: Keep in sync maybe with transpile_gno0p0.go/FindMore...
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
				if n.Type != NameExprTypeLoopVarUse {
					return n, TRANS_CONTINUE
				}
				if !strings.HasPrefix(string(n.Name), ".loopvar") {
					return n, TRANS_CONTINUE
				}

				// Find the block where name is defined.
				dbn := last.GetBlockNodeForPath(nil, n.Path)

				for {
					// If used as closure capture, mark as heap use.
					flx, _, found := findFirstClosure(stack, dbn)
					if !found {
						return n, TRANS_CONTINUE
					}

					origName := strings.TrimPrefix(string(n.Name), ".loopvar_")
					redefineName := fmt.Sprintf("%s%s", ".redefine_", origName)

					if _, ok := dbn.(*ForStmt); ok {
						stop = true

						if _, ok := dbn.GetAttribute(ATTR_REDEFINE_NAME).(string); !ok {
							dbn.SetAttribute(ATTR_REDEFINE_NAME, redefineName)
						}
					}

					// Loop again for more closures.
					dbn = flx
				}
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
	return
}

func shadowLoopvar(ctx BlockNode, bn BlockNode) (stop bool) {
	// var transForBody bool
	// Iterate over all nodes recursively.
	_ = TranscribeB(ctx, bn, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("shadowLoopvar %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		// // only rename names in body.
		// if ftype == TRANS_FOR_BODY {
		// 	transForBody = true
		// }
		// if ftype == TRANS_FOR_INIT || ftype == TRANS_FOR_COND || ftype == TRANS_FOR_POST {
		// 	transForBody = false
		// }

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:

		// ----------------------------------------
		case TRANS_LEAVE:
			if bn, ok := n.(BlockNode); ok {
				// handle continue.
				// bn is the parent block of continue.
				// where we need copy out.
				if name, ok := bn.GetAttribute(ATTR_REWRITE_CONTINUE).(string); ok && name != "" {
					fmt.Println("---bn: ", bn, reflect.TypeOf(bn))
					fmt.Println("---name: ", name)
					body := bn.GetBody()
					fmt.Println("---len of body: ", len(body))

					lhs := Nx(fmt.Sprintf("%s%s", ".loopvar_", name))
					rhs := Nx(fmt.Sprintf("%s%s", ".redefine_", name))
					// lhs.Type = NameExprTypeDefine
					lhs.Type = NameExprTypeLoopVarUse
					rhs.Type = NameExprTypeNormal

					as := A(lhs, "=", rhs)

					// 1. Find Index
					idx := -1
					for i, s := range body {
						if bs, ok := s.(*BranchStmt); ok && bs.Op == CONTINUE {
							idx = i
							break
						}
					}

					// 2. Handle Insert
					body2 := make([]Stmt, 0, len(body)+1)
					if idx != -1 {
						// Insert in middle
						body2 = append(body2, body[:idx]...)
						body2 = append(body2, as)
						body2 = append(body2, body[idx:]...)
						bn.SetBody(body2)
						// fmt.Println("---bn.Body: ", bn.GetBody())
					}
				}
			}

			switch n := n.(type) {
			case *BranchStmt:
				// find correspongding forstmt, maybe labeled.
				fmt.Println("---BranchStmt, n: ", n, n.Op)
				// fmt.Println("---Block depth: ", n.BlockDepth)
				// bn, d1, d2, idx := findGotoLabel(last, n.Label)
				// fmt.Println("---bn: ", bn, reflect.TypeOf(bn))
				// fmt.Println("---: ", d1, d2, idx)

				fmt.Println("---last: ", last)

				fs, _, _ := findFirstForStmt(stack)
				fmt.Println("---fs found: ", fs)
				if s, ok := fs.GetAttribute(ATTR_REDEFINE_NAME).(string); ok {
					fmt.Println("---s: ", s)

					// body := last.GetBody()
					// fmt.Println("---body: ", body)
					last.SetAttribute(ATTR_REWRITE_CONTINUE, strings.TrimPrefix(s, ".redefine_"))
				}

			case *ForStmt:
				// do the injection here.
				fmt.Println("------Trans_Leave, forStmt: ", n)
				// panic("!!!")
				if n2, ok := n.GetAttribute(ATTR_REDEFINE_NAME).(string); ok && n2 != "" {
					// -------------------------------------------------
					// this is the target *ForStmt
					// rename its containing names
					// all the .loopvar_i, except redeclared nested for-loop var?

					// -------------------------------------------------

					// fmt.Println("------forStmt re-written...")

					bodyBlock := n.BodyBlock
					stmts := bodyBlock.GetBody()

					origName := strings.TrimPrefix(string(n2), ".redefine_")

					lhs := Nx(n2)
					rhs := Nx(fmt.Sprintf("%s%s", ".loopvar_", origName))
					lhs.Type = NameExprTypeHeapDefine
					rhs.Type = NameExprTypeNormal
					// lhs.Type = NameExprTypeDefine

					as := A(lhs, ":=", rhs)
					stmts2 := make([]Stmt, 0, len(stmts)+2)
					stmts2 = append(stmts2, as)
					stmts2 = append(stmts2, stmts...)

					lhs2 := Nx(fmt.Sprintf("%s%s", ".loopvar_", origName))
					rhs2 := Nx(n2)
					lhs2.Type = NameExprTypeNormal
					// will promote to heap use
					rhs2.Type = NameExprTypeNormal
					as2 := A(lhs2, "=", rhs2)
					stmts2 = append(stmts2, as2)

					// Index for new injected name
					// index := len(n.GetBlockNames())
					// Set Heap type
					// Reserve for new injected name

					tv := n.GetSlot(nil, Name(".loopvar_i"), true)
					bodyBlock.Define(".redefine_i", tv.Copy(nil))

					// n.Reserve(false, lhs, n, NSDefine, index+1)

					fmt.Println("======AFTER DEFINE, names: ", n.GetBlockNames())
					fmt.Println("---n: ", n)

					bodyBlock.SetBody(stmts2)

					// fmt.Println("------after process, n: ", n)
					n.DelAttribute(ATTR_REDEFINE_NAME)
				}
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
	return
}

func renameLoopvarUse(ctx BlockNode, bn BlockNode) (stop bool) {
	// Iterate over all nodes recursively.
	_ = TranscribeB(ctx, bn, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("renameLoopvarUse %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}
		// fmt.Printf("renameLoopvarUse %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:

		// ----------------------------------------
		case TRANS_LEAVE:
			switch n := n.(type) {
			case *NameExpr:
				// fmt.Println("---starting renameLoopvarUse, nx: ", n, n.Type)
				// nn := ns[len(ns)-1]
				// fmt.Println("---1, nn: ", nn)
				// NOTE: Keep in sync maybe with transpile_gno0p0.go/FindMore...
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
				// only LoopVarUse.
				if n.Type != NameExprTypeLoopVarUse {
					return n, TRANS_CONTINUE
				}

				fmt.Println("---renameLoopvarUse, nx: ", n, n.Type)
				var origName Name
				if strings.HasPrefix(string(n.Name), ".loopvar_") {
					origName = Name(strings.TrimPrefix(string(n.Name), ".loopvar_"))
				}

				fmt.Println("---origName: ", origName)
				nn := ns[len(ns)-1]
				fmt.Println("---2, nn: ", nn)

				fmt.Println("---last: ", last)
				fmt.Println("---last.GetBlockNames: ", last.GetBlockNames())

				// "i", or ".redefine_i"

				found0, name0 := last.FindNamePrefixed(nil, origName, "")
				fmt.Println("---found0: ", found0)
				if found0 {
					n.Name = name0
					fmt.Println("===Rename to: ", name0)
				} else {
					found1, name1 := last.FindNamePrefixed(nil, origName, ".loopvar_")

					found2, name2 := last.FindNamePrefixedForPath(nil, origName, n.Path, ".redefine_")

					fmt.Printf("---found1: %t, name1: %v\n", found1, name1)
					fmt.Printf("---found2: %t, name2: %v\n", found2, name2)

					if found1 && !found2 {
						fmt.Println("---rename to name1: ", name1)
						n.Name = name1
					} else if found1 && found2 {
						fmt.Println("---rename to name2: ", name2)
						n.Name = name2
					}
				}

				fmt.Println("---after rename, n: ", n)
				nn2 := ns[len(ns)-1]
				fmt.Println("---3, nn2: ", nn2)
				n.Type = NameExprTypeNormal
				// n.Type = NameExprTypeHeapUse
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
	return
}

// Finds heap defines by their use in ref expressions or
// closures (captures). Also adjusts the name expr type,
// and sets new closure captures' path to refer to local
// capture.
// Also happens to declare all package and file names
// as heap use, so that functions added later may use them.
func codaHeapDefinesByUse(ctx BlockNode, bn BlockNode) {
	// Iterate over all nodes recursively.
	_ = TranscribeB(ctx, bn, func(
		ns []Node,
		stack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("codaHeapDefinesByUse %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:

		// ----------------------------------------
		case TRANS_LEAVE:

			switch n := n.(type) {
			case *ValueDecl:
				// Top level value decls are always heap escaped.
				// See also corresponding case in codaHeapUsesDemoteDefines.
				if !n.Const {
					switch last.(type) {
					case *PackageNode, *FileNode:
						pn := skipFile(last)
						for _, nx := range n.NameExprs {
							if nx.Name == "_" {
								continue
							}
							addAttrHeapUse(pn, nx.Name)
						}
					}
				}
			case *RefExpr:
				lmx := LeftmostX(n.X)
				if nx, ok := lmx.(*NameExpr); ok {
					// Find the block where name is defined
					dbn := last.GetBlockNodeForPath(nil, nx.Path)
					// The leftmost name of possibly nested index
					// and selector exprs.
					// e.g. leftmost.middle[0][2].rightmost
					// Mark name for heap use.
					addAttrHeapUse(dbn, nx.Name)
					// adjust NameExpr type.
					nx.Type = NameExprTypeHeapUse
				}
			case *NameExpr:
				// fmt.Println("---findHeapDefinesByUse, n.Name: ", n.Name)
				// NOTE: Keep in sync maybe with transpile_gno0p0.go/FindMore...
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
				dbn := last.GetBlockNodeForPath(nil, n.Path)
				for {
					// If used as closure capture, mark as heap use.
					flx, depth, found := findFirstClosure(stack, dbn)
					if !found {
						return n, TRANS_CONTINUE
					}
					// Ignore top level declarations.
					// This get replaced by codaPackageSelectors.
					if pn, ok := dbn.(*PackageNode); ok {
						if pn.PkgPath != ".uverse" {
							n.SetAttribute(ATTR_PACKAGE_DECL, true)
							return n, TRANS_CONTINUE
						}
					}
					// Ignore type declaration names.
					// Types cannot be passed ergo cannot be captured.
					// (revisit when types become first class objects)
					st := dbn.GetStaticTypeOf(nil, n.Name)
					if st.Kind() == TypeKind {
						return n, TRANS_CONTINUE
					}

					// Found a heap item closure capture.
					addAttrHeapUse(dbn, n.Name)
					// The path must stay same for now,
					// used later in codaHeapUsesDemoteDefines.
					idx := addHeapCapture(dbn, flx, depth, n)
					// adjust NameExpr type.
					n.Type = NameExprTypeHeapUse
					n.Path.SetDepth(uint8(depth))
					n.Path.Index = idx
					// Loop again for more closures.
					dbn = flx
				}
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

// TODO consider adding to Names type.
func addName(names []Name, name Name) []Name {
	if !slices.Contains(names, name) {
		names = append(names, name)
	}
	return names
}

func addAttrHeapUse(bn BlockNode, name Name) {
	lus, _ := bn.GetAttribute(ATTR_HEAP_USES).([]Name)
	lus = addName(lus, name)
	bn.SetAttribute(ATTR_HEAP_USES, lus)
}

func hasAttrHeapUse(bn BlockNode, name Name) bool {
	hds, _ := bn.GetAttribute(ATTR_HEAP_USES).([]Name)
	return slices.Contains(hds, name)
}

// adds ~name to func lit static block and to heap captures atomically.
func addHeapCapture(dbn BlockNode, fle *FuncLitExpr, depth int, nx *NameExpr) (idx uint16) {
	if depth <= 0 {
		panic("invalid depth")
	}
	name := nx.Name
	for _, ne := range fle.HeapCaptures {
		if ne.Name == name {
			// assert ~name also already defined.
			var ok bool
			idx, ok = fle.GetLocalIndex("~" + name)
			if !ok {
				panic("~name not added to fle atomically")
			}
			return // already exists
		}
	}

	// define ~name to fle.
	_, ok := fle.GetLocalIndex("~" + name)
	if ok {
		panic("~name already defined in fle")
	}

	tv := dbn.GetSlot(nil, name, true)
	fle.Define("~"+name, tv.Copy(nil))

	// add name to fle.HeapCaptures.
	// NOTE: this doesn't work with shadowing, see define1.gno.
	// vp := fle.GetPathForName(nil, name)
	vp := nx.Path
	vp.SetDepth(vp.Depth - uint8(depth))
	// vp.SetDepth(vp.Depth - 1) // minus 1 for fle itself.
	ne := NameExpr{
		Path: vp,
		Name: name,
		Type: NameExprTypeHeapClosure,
	}
	fle.HeapCaptures = append(fle.HeapCaptures, ne)

	// find index after define
	for i, n := range fle.GetBlockNames() {
		if n == "~"+name {
			idx = uint16(i)
			return
		}
	}

	panic("should not happen, idx not found")
}

// finds the first FuncLitExpr in the stack after (excluding) stop.  returns
// the depth of first closure, 1 if last stack item itself is a closure, or 0
// if not found.
func findFirstClosure(stack []BlockNode, stop BlockNode) (fle *FuncLitExpr, depth int, found bool) {
	faux := 0 // count faux block
	for i := len(stack) - 1; i >= 0; i-- {
		stbn := stack[i]
		switch stbn := stbn.(type) {
		case *FuncLitExpr:
			if stbn == stop { // if fle is stopBn, does not count, use last fle
				return
			}
			fle = stbn
			depth = len(stack) - 1 - faux - i + 1 // +1 since 1 is lowest.
			found = true
			// even if found, continue iteration in case
			// an earlier *FuncLitExpr is found.
		default:
			if fauxChildBlockNode(stbn) {
				faux++
			}
			if stbn == stop {
				return
			}
		}
	}
	// This can happen e.g. if stop is a package but we are
	// Preprocess()'ing an expression such as `func(){ ... }()` from
	// Machine.Eval() on an already preprocessed package.
	return
}

func findFirstForStmt(stack []BlockNode) (fst *ForStmt, depth int, found bool) {
	faux := 0 // count faux block
	for i := len(stack) - 1; i >= 0; i-- {
		stbn := stack[i]
		switch stbn := stbn.(type) {
		case *ForStmt:
			fst = stbn
			depth = len(stack) - 1 - faux - i + 1 // +1 since 1 is lowest.
			found = true
			// even if found, continue iteration in case
			// an earlier *FuncLitExpr is found.
		default:
			if fauxChildBlockNode(stbn) {
				faux++
			}
		}
	}
	// This can happen e.g. if stop is a package but we are
	// Preprocess()'ing an expression such as `func(){ ... }()` from
	// Machine.Eval() on an already preprocessed package.
	return
}

// finds the last FuncLitExpr in the stack at (including) or after stop.
// returns the depth of last function, 1 if last stack item itself is a
// closure, or 0 if not found.
func findLastFunction(last BlockNode, stop BlockNode) (fn FuncNode, depth int, found bool) {
	faux := 0   // count faux block
	depth2 := 0 // working value
	for stbn := last; stbn != stop; stbn = stbn.GetParentNode(nil) {
		depth2 += 1
		switch stbn := stbn.(type) {
		case *FuncLitExpr, *FuncDecl:
			fn = stbn.(FuncNode)
			depth = depth2 - faux
			found = true
			return
		default:
			if fauxChildBlockNode(stbn) {
				faux++
			}
			if stbn == stop {
				return
			}
		}
	}
	// This can happen e.g. if stop is a package but we are
	// Preprocess()'ing an expression such as `func(){ ... }()` from
	// Machine.Eval() on an already preprocessed package.
	return
}

// If a name is used as a heap item, Convert all other uses of such names
// for heap use. If a name of type heap define is not actually used
// as heap use, demotes them.
func codaHeapUsesDemoteDefines(ctx BlockNode, bn BlockNode) {
	// create stack of BlockNodes.
	last := ctx
	stack := append(make([]BlockNode, 0, 32), last)

	// Iterate over all nodes recursively.
	_ = Transcribe(bn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("codaHeapUsesDemoteDefines %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:
			pushInitBlock(n.(BlockNode), &last, &stack)

		// ----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {
			// type switch is a special cast that varName is
			// defined in case clauses.
			case *SwitchClauseStmt:
				// parent switch statement.
				ss := ns[len(ns)-1].(*SwitchStmt)
				if ss.IsTypeSwitch && ss.VarName != "" {
					// If the name is actually heap used:
					if hasAttrHeapUse(n, ss.VarName) {
						// Make record in static block.
						// will be populated in ExpandWith().
						n.SetIsHeapItem(ss.VarName)
					}
				}
			case *NameExpr:
				// Ignore non-block type paths
				if n.Path.Type != VPBlock {
					return n, TRANS_CONTINUE
				}
				switch n.Type {
				case NameExprTypeNormal:
					// Find the block where name is defined
					dbn := last.GetBlockNodeForPath(nil, n.Path)
					// If the name is heap used,
					if hasAttrHeapUse(dbn, n.Name) {
						// Change type to heap use.
						n.Type = NameExprTypeHeapUse
					}
				case NameExprTypeDefine, NameExprTypeHeapDefine:
					// Find the block where name is defined
					dbn := last.GetBlockNodeForPath(nil, n.Path)
					// If the name is actually heap used:
					if hasAttrHeapUse(dbn, n.Name) {
						// Promote type to heap define.
						n.Type = NameExprTypeHeapDefine
						// Make record in static block.
						dbn.SetIsHeapItem(n.Name)
					} else {
						// Demote type to regular define.
						n.Type = NameExprTypeDefine
					}
				}
			case *ValueDecl:
				// Top level var value decls are always heap escaped.
				// See also corresponding case in codaHeapDefinesByUse.
				if !n.Const {
					switch last.(type) {
					case *PackageNode, *FileNode:
						pn := skipFile(last)
						for i := range n.NameExprs {
							nx := &n.NameExprs[i]
							if nx.Name == "_" {
								continue
							}
							if !hasAttrHeapUse(pn, nx.Name) {
								panic("expected heap use for top level value decl")
							}
							nx.Type = NameExprTypeHeapDefine
							pn.SetIsHeapItem(nx.Name)
						}
					}
				}
			}
			return n, TRANS_CONTINUE

		// ----------------------------------------
		case TRANS_LEAVE:

			// Defer pop block from stack.
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
			case BlockNode:
				switch fd := n.(type) {
				case *FuncDecl:
					recv := &fd.Recv
					if hasAttrHeapUse(fd, recv.Name) {
						recv.NameExpr.Type = NameExprTypeHeapDefine
						fd.SetIsHeapItem(recv.Name)
					}
					for i := 0; i < len(fd.Type.Params); i++ {
						name := fd.Type.Params[i].Name
						if hasAttrHeapUse(fd, name) {
							fd.Type.Params[i].NameExpr.Type = NameExprTypeHeapDefine
							fd.SetIsHeapItem(name)
						}
					}
					for i := 0; i < len(fd.Type.Results); i++ {
						name := fd.Type.Results[i].Name
						if hasAttrHeapUse(fd, name) {
							fd.Type.Results[i].NameExpr.Type = NameExprTypeHeapDefine
							fd.SetIsHeapItem(name)
						}
					}
				case *FuncLitExpr:
					for i := 0; i < len(fd.Type.Params); i++ {
						name := fd.Type.Params[i].Name
						if hasAttrHeapUse(fd, name) {
							fd.Type.Params[i].NameExpr.Type = NameExprTypeHeapDefine
							fd.SetIsHeapItem(name)
						}
					}
					for i := 0; i < len(fd.Type.Results); i++ {
						name := fd.Type.Results[i].Name
						if hasAttrHeapUse(fd, name) {
							fd.Type.Results[i].NameExpr.Type = NameExprTypeHeapDefine
							fd.SetIsHeapItem(name)
						}
					}
				}

				// no need anymore
				n.DelAttribute(ATTR_HEAP_USES)
				n.DelAttribute(ATTR_HEAP_DEFINES)
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

// Replace package name with ref:
//   - `pkg`  ->                RefValue{PkgPath}
//
// Replace package declared name ref:
//   - `name` -> SelectorExpr{X:RefValue{PkgPath},Sel:name}
func codaPackageSelectors(bn BlockNode) {
	// Iterate over all nodes recursively.
	_ = Transcribe(bn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		switch stage {
		case TRANS_ENTER:
			switch n := n.(type) {
			case *NameExpr:
				// Replace a package name with RefValue{PkgPath}
				if pref, ok := n.GetAttribute(ATTR_PACKAGE_REF).(RefValue); ok {
					cx := &ConstExpr{
						Source: n,
						TypedValue: TypedValue{
							T: gPackageType,
							V: pref,
						},
					}
					return cx, TRANS_CONTINUE
				}
				// Replace a local package declared name with
				// SelectorExpr{X:RefValue{PkgPath},Sel:name}
				pdi := n.GetAttribute(ATTR_PACKAGE_DECL)
				if pdi != nil { // is true
					if n.Path.Type != VPBlock {
						panic("expected block path")
					}
					pn := packageOf(bn)
					cx := &ConstExpr{
						Source: Nx(".pkgSelector"),
						TypedValue: TypedValue{
							T: gPackageType,
							V: RefValue{PkgPath: pn.PkgPath},
						},
					}
					setPreprocessed(cx)
					sx := &SelectorExpr{
						X:    cx,
						Path: NewValuePathBlock(1, n.Path.Index, n.Name),
						Sel:  n.Name,
					}
					setPreprocessed(sx)
					return sx, TRANS_CONTINUE
				}
			}
		}
		return n, TRANS_CONTINUE
	})
}

func isSwitchLabel(ns []Node, label Name) bool {
	for {
		swch := lastSwitch(ns)
		if swch == nil {
			break
		}

		if swch.GetLabel() == label && label != "" {
			return true
		}

		ns = ns[:len(ns)-1]
	}

	return false
}

// Idempotent.
// Also makes sure the stack doesn't reach MaxUint8 in length.
func pushInitBlock(bn BlockNode, last *BlockNode, stack *[]BlockNode) {
	if !bn.IsInitialized() {
		bn.InitStaticBlock(bn, *last)
	}
	if bn.GetStaticBlock().Source != bn {
		panic("expected the source of a block node to be itself")
	}
	*last = bn
	*stack = append(*stack, bn)
	if len(*stack) >= math.MaxUint8 {
		panic("block node depth reached maximum MaxUint8")
	}
}

// like pushInitBlock(), but when the last block is a faux block,
// namely after SwitchStmt and IfStmt.
// Not idempotent, as it calls bn.Define with reference to last's TV value slot.
func pushInitBlockAndCopy(bn BlockNode, last *BlockNode, stack *[]BlockNode) {
	if _, ok := bn.(*IfCaseStmt); !ok {
		if _, ok := bn.(*SwitchClauseStmt); !ok {
			panic("should not happen")
		}
	}
	orig := *last
	pushInitBlock(bn, last, stack)
	copyFromFauxBlock(bn, orig)
}

// anything declared in orig are copied.
func copyFromFauxBlock(bn BlockNode, orig BlockNode) {
	for _, n := range orig.GetBlockNames() {
		tv := orig.GetSlot(nil, n, false)
		bn.Define(n, *tv)
	}
}

// Evaluates (constructs) the value of x which is expected to be a typeval.
// Caches the result as an attribute of x.
// To discourage mis-use, expects x to already be preprocessed.
func evalStaticType(store Store, last BlockNode, x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else if ctx, ok := x.(*constTypeExpr); ok {
		return ctx.Type // no need to set attribute.
	}
	pn := packageOf(last)
	// See comment in evalStaticTypeOfRaw.
	if store != nil && pn.PkgPath != uversePkgPath {
		// this used pn.NewPackage previously; however, that function
		// additionally calls PrepareNewValues, which is not necessary in this
		// context and incurs in very expensive allocations.
		pv := &PackageValue{
			Block: &Block{
				Source: pn,
			},
			PkgName: pn.PkgName,
			PkgPath: pn.PkgPath,
		}
		store = store.BeginTransaction(nil, nil, nil)
		store.SetCachePackage(pv)
	}
	m := NewMachine(pn.PkgPath, store)
	tv := m.EvalStatic(last, x)
	m.Release()
	if _, ok := tv.V.(TypeValue); !ok {
		panic(fmt.Sprintf("%s is not a type", x.String()))
	}
	t := tv.GetType()
	x.SetAttribute(ATTR_TYPE_VALUE, t)
	return t
}

// If it is known that the type was already evaluated,
// use this function instead of evalStaticType(store,).
func getType(x Expr) Type {
	if ctx, ok := x.(*constTypeExpr); ok {
		return ctx.Type
	} else if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else {
		panic(fmt.Sprintf(
			"getType() called on expr not yet evaluated with evalStaticType(store,): %s",
			x.String(),
		))
	}
}

// Unlike evalStaticType, x is not expected to be a typeval,
// but rather computes the type OF x.
func evalStaticTypeOf(store Store, last BlockNode, x Expr) Type {
	t := evalStaticTypeOfRaw(store, last, x)

	if tt, ok := t.(*tupleType); ok && len(tt.Elts) == 1 {
		return tt.Elts[0]
	}

	return t
}

// like evalStaticTypeOf() but returns the raw *tupleType for *CallExpr.
func evalStaticTypeOfRaw(store Store, last BlockNode, x Expr) (t Type) {
	if t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		return t
	} else if _, ok := x.(*constTypeExpr); ok {
		return gTypeType
	} else if ctx, ok := x.(*ConstExpr); ok {
		return ctx.T
	} else {
		pn := packageOf(last)
		// NOTE: do not load the package value from store,
		// because we may be preprocessing in the middle of
		// PreprocessAllFilesAndSaveBlockNodes,
		// and the preprocessor will panic when
		// package values are already there that weren't
		// yet predefined this time around.
		if store != nil && pn.PkgPath != uversePkgPath {
			pv := &PackageValue{
				// this used pn.NewPackage previously; however, that function
				// additionally calls PrepareNewValues, which is not necessary in this
				// context and incurs in very expensive allocations.
				Block: &Block{
					Source: pn,
				},
				PkgName: pn.PkgName,
				PkgPath: pn.PkgPath,
			}
			store = store.BeginTransaction(nil, nil, nil)
			store.SetCachePackage(pv)
		}
		m := NewMachine(pn.PkgPath, store)
		t = m.EvalStaticTypeOf(last, x)
		m.Release()
		x.SetAttribute(ATTR_TYPEOF_VALUE, t)
		return t
	}
}

// If it is known that the type was already evaluated,
// use this function instead of evalStaticTypeOf().
func getTypeOf(x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		if tt, ok := t.(*tupleType); ok {
			if len(tt.Elts) != 1 {
				panic(fmt.Sprintf(
					"getTypeOf() only supports *CallExpr with 1 result, got %s",
					tt.String(),
				))
			} else {
				return tt.Elts[0]
			}
		} else {
			return t
		}
	} else {
		panic(fmt.Sprintf(
			"getTypeOf() called on expr not yet evaluated with evalStaticTypeOf(): %s",
			x.String(),
		))
	}
}

// like evalStaticTypeOf() but for list of exprs, and the result
// includes the value if type is TypeKind.
func evalStaticTypedValues(store Store, last BlockNode, xs ...Expr) []TypedValue {
	res := make([]TypedValue, len(xs))
	for i, x := range xs {
		t := evalStaticTypeOf(store, last, x)
		if t != nil && t.Kind() == TypeKind {
			v := evalStaticType(store, last, x)
			res[i] = TypedValue{
				T: t,
				V: toTypeValue(v),
			}
		} else {
			res[i] = TypedValue{
				T: t,
				V: nil,
			}
		}
	}
	return res
}

// Evaluate (but do not memoize) x. May fail for a variety of reasons.
func tryEvalStatic(store Store, pn *PackageNode, last BlockNode, x Expr) (tv TypedValue, err error) {
	if cx, ok := x.(*ConstExpr); ok {
		return cx.TypedValue, nil
	}
	pv := pn.NewPackage(nilAllocator) // throwaway
	store = store.BeginTransaction(nil, nil, nil)
	store.SetCachePackage(pv)
	m := NewMachine(pn.PkgPath, store)
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
		tv = m.EvalStatic(last, x)
	}()
	return
}

func getGnoFuncTypeOf(store Store, it Type) *FuncType {
	return baseOf(it).(*FuncType)
}

func getResultTypedValues(cx *CallExpr) []TypedValue {
	if t, ok := cx.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		if tt, ok := t.(*tupleType); ok {
			res := make([]TypedValue, len(tt.Elts))
			for i, tte := range tt.Elts {
				res[i] = anyValue(tte)
			}
			return res
		} else {
			panic(fmt.Sprintf(
				"expected *tupleType of *CallExpr but got %v",
				reflect.TypeOf(t)))
		}
	} else {
		panic(fmt.Sprintf(
			"getResultTypedValues() called on call expr not yet evaluated: %s",
			cx.String(),
		))
	}
}

// Evaluate constant expressions. Assumes all operands are already defined
// consts; the machine doesn't know whether a value is const or not, so this
// function always returns a *ConstExpr, even if the operands aren't actually
// consts in the code.
//
// No type conversion is done by the machine in general -- operands of
// binary expressions should be converted to become compatible prior
// to evaluation.
//
// NOTE: Generally, conversion happens in a separate step while leaving
// composite exprs/nodes that contain constant expression nodes (e.g. const
// exprs in the rhs of AssignStmts).
//
// Array-related expressions like `len` and `cap` are manually evaluated
// as constants, even if the array itself is not a constant. This evaluation
// is handled independently of the rest of the constant evaluation process,
// bypassing machine.EvalStatic.
func evalConst(store Store, last BlockNode, x Expr) *ConstExpr {
	// TODO: some check or verification for ensuring x
	var cx *ConstExpr
	if clx, ok := x.(*CallExpr); ok {
		t := evalStaticTypeOf(store, last, clx.Args[0])
		if ar, ok := unwrapPointerType(baseOf(t)).(*ArrayType); ok {
			fv := clx.Func.(*ConstExpr).V.(*FuncValue)
			switch fv.Name {
			case "cap", "len":
				tv := TypedValue{T: IntType}
				tv.SetInt(int64(ar.Len))
				cx = &ConstExpr{
					Source:     x,
					TypedValue: tv,
				}
			default:
				panic(fmt.Sprintf("unexpected const func %s", fv.Name))
			}
		}
	}

	if cx == nil {
		// is constant?  From the machine?
		m := NewMachine(".dontcare", store)
		cv := m.EvalStatic(last, x)
		m.Release()
		cx = &ConstExpr{
			Source:     x,
			TypedValue: cv,
		}
	} else {
		if cx.Source != x {
			panic(fmt.Sprintf("source si diffenrent? %v, %v", cx.Source, x))
		}
	}
	// cx.SetSpan(x.GetSpan())
	setPreprocessed(cx)
	setConstAttrs(cx)
	return cx
}

func toConstTypeExpr(last BlockNode, source Expr, t Type) *constTypeExpr {
	source = unconst(source)
	cx := &constTypeExpr{}
	cx.Last = last
	cx.Source = source
	cx.Type = t
	cx.SetSpan(source.GetSpan())
	setPreprocessed(cx)
	return cx
}

func toConstExpr(source Expr, tv TypedValue) *ConstExpr {
	source = unconst(source)
	cx := &ConstExpr{}
	cx.Source = source
	cx.TypedValue = tv
	cx.SetSpan(source.GetSpan())
	setPreprocessed(cx)
	return cx
}

func unconst(cx Expr) Expr {
	switch cx := cx.(type) {
	case *constTypeExpr:
		return cx.Source
	case *ConstExpr:
		return cx.Source
	default:
		return cx
	}
}

func constInt(source Expr, i int64) *ConstExpr {
	source = unconst(source)
	cx := &ConstExpr{Source: source}
	cx.T = IntType
	cx.SetInt(i)
	setPreprocessed(cx)
	return cx
}

func constUntypedBigint(source Expr, i64 int64) *ConstExpr {
	source = unconst(source)
	cx := &ConstExpr{Source: source}
	cx.T = UntypedBigintType
	cx.V = BigintValue{big.NewInt(i64)}
	setPreprocessed(cx)
	return cx
}

func constNil(source Expr) *ConstExpr {
	source = unconst(source)
	cx := &ConstExpr{Source: source}
	cx.T = nil
	cx.V = nil
	setPreprocessed(cx)
	return cx
}

func setConstAttrs(cx *ConstExpr) {
	cv := &cx.TypedValue
	cx.SetAttribute(ATTR_TYPEOF_VALUE, cv.T)
	if cv.T != nil && cv.T.Kind() == TypeKind {
		if cv.GetType() == nil {
			panic("should not happen")
		}
		cx.SetAttribute(ATTR_TYPE_VALUE, cv.GetType())
	}
}

// This gets set immediately upon preprocess/TRANS_LEAVE.  This means its type
// can be evaluated during TRANS_LEVE, as the node's attribute
// ATTR_PREPROCESSED is already set. But it also makes it a bit of misnomer;
// preprocessing isn't exactly complete.
func setPreprocessed(x Node) Node {
	if x.GetAttribute(ATTR_PREPROCESS_INCOMPLETE) == nil {
		x.SetAttribute(ATTR_PREPROCESSED, true)
	}
	return x
}

func isRealm(ctx BlockNode) bool {
	pn := packageOf(ctx)
	return IsRealmPath(pn.PkgPath)
}

func packageOf(last BlockNode) *PackageNode {
	for {
		if pn, ok := last.(*PackageNode); ok {
			return pn
		}
		last = last.GetParentNode(nil)
	}
}

func fileOfSafe(last BlockNode) *FileNode {
	for {
		if last == nil {
			return nil
		}
		if fn, ok := last.(*FileNode); ok {
			return fn
		}
		last = last.GetParentNode(nil)
	}
}

func funcOf(last BlockNode) (BlockNode, *FuncTypeExpr) {
	for {
		if flx, ok := last.(*FuncLitExpr); ok {
			return flx, &flx.Type
		} else if fd, ok := last.(*FuncDecl); ok {
			return fd, &fd.Type
		}
		last = last.GetParentNode(nil)
	}
}

func findBreakableNode(last BlockNode, store Store) {
	for last != nil {
		switch last.(type) {
		case *FuncLitExpr, *FuncDecl:
			panic("break statement out of place")
		case *ForStmt:
			return
		case *RangeStmt:
			return
		case *SwitchClauseStmt:
			return
		}

		last = last.GetParentNode(store)
	}
}

func findContinuableNode(last BlockNode, store Store) {
	for last != nil {
		switch last.(type) {
		case *FuncLitExpr, *FuncDecl:
			panic("continue statement out of place")
		case *ForStmt:
			return
		case *RangeStmt:
			return
		}

		last = last.GetParentNode(store)
	}
}

func findBranchLabel(last BlockNode, label Name) (
	bn BlockNode, depth uint8, bodyIdx int,
) {
	for {
		switch cbn := last.(type) {
		case *BlockStmt, *ForStmt, *IfCaseStmt, *RangeStmt, *SelectCaseStmt, *SwitchClauseStmt:
			lbl := cbn.GetLabel()
			if label == lbl {
				bn = cbn
				return
			}
			last = skipFaux(cbn.GetParentNode(nil))
			depth += 1
		case *IfStmt, *SwitchStmt:
			// These are faux blocks -- shouldn't happen.
			panic("unexpected faux blocknode")
		case *FileNode:
			panic("unexpected file blocknode")
		case *PackageNode:
			panic("unexpected package blocknode")
		case *FuncLitExpr:
			body := cbn.GetBody()
			_, bodyIdx = body.GetLabeledStmt(label)
			if bodyIdx != -1 {
				bn = cbn
				return
			}
			panic(fmt.Sprintf(
				"cannot find branch label %q",
				label))
		case *FuncDecl:
			panic(fmt.Sprintf(
				"cannot find branch label %q",
				label))
		default:
			panic("unexpected block node")
		}
	}
}

func findGotoLabel(last BlockNode, label Name) (
	bn BlockNode, blockDepth uint8, frameDepth uint8, bodyIdx int,
) {
	for {
		switch cbn := last.(type) {
		case *IfStmt, *SwitchStmt:
			// These are faux blocks -- shouldn't happen.
			panic("unexpected faux blocknode")
		case *FileNode:
			panic("unexpected file blocknode")
		case *PackageNode:
			panic("unexpected package blocknode")
		case *FuncLitExpr, *FuncDecl:
			body := cbn.GetBody()
			_, bodyIdx = body.GetLabeledStmt(label)
			if bodyIdx != -1 {
				bn = cbn
				return
			} else {
				panic(fmt.Sprintf(
					"cannot find GOTO label %q within current function",
					label))
			}
		case *ForStmt, *RangeStmt, *SwitchClauseStmt:
			body := cbn.GetBody()
			_, bodyIdx = body.GetLabeledStmt(label)
			if bodyIdx != -1 {
				bn = cbn
				return
			} else {
				last = skipFaux(cbn.GetParentNode(nil))
				frameDepth += 1
				blockDepth = 0 // reset
			}
		case *IfCaseStmt, *SelectCaseStmt, *BlockStmt:
			body := cbn.GetBody()
			_, bodyIdx = body.GetLabeledStmt(label)
			if bodyIdx != -1 {
				bn = cbn
				return
			} else {
				last = skipFaux(cbn.GetParentNode(nil))
				// NOTE: frameDepth may be zero or positive.
				blockDepth += 1
			}
		default:
			panic("unexpected block node")
		}
	}
}

func lastDecl(ns []Node) Decl {
	for i := len(ns) - 1; 0 <= i; i-- {
		if d, ok := ns[i].(Decl); ok {
			return d
		}
	}
	return nil
}

func lastSwitch(ns []Node) *SwitchStmt {
	for i := len(ns) - 1; 0 <= i; i-- {
		if d, ok := ns[i].(*SwitchStmt); ok {
			return d
		}
	}
	return nil
}

func asValue(t Type) TypedValue {
	return TypedValue{
		T: gTypeType,
		V: toTypeValue(t),
	}
}

func anyValue(t Type) TypedValue {
	return TypedValue{
		T: t,
		V: nil,
	}
}

func isConst(x Expr) bool {
	_, ok := x.(*ConstExpr)
	return ok
}

func isConstType(x Expr) bool {
	_, ok := x.(*constTypeExpr)
	return ok
}

// check before convert type
func checkOrConvertType(store Store, last BlockNode, n Node, x *Expr, t Type) {
	if debug {
		debug.Printf("checkOrConvertType, *x: %v:, t:%v \n", *x, t)
	}
	if cx, ok := (*x).(*ConstExpr); ok {
		// e.g. int(1) == int8(1)
		assertAssignableTo(n, cx.T, t)
	} else if bx, ok := (*x).(*BinaryExpr); ok && (bx.Op == SHL || bx.Op == SHR) {
		xt := evalStaticTypeOf(store, last, *x)
		if debug {
			debug.Printf("shift, xt: %v, Op: %v, t: %v \n", xt, bx.Op, t)
		}
		if isUntyped(xt) {
			// check assignable first, see: types/shift_b6.gno
			assertAssignableTo(n, xt, t)

			if t == nil || t.Kind() == InterfaceKind {
				t = defaultTypeOf(xt)
			}

			bx.assertShiftExprCompatible2(t)
			checkOrConvertType(store, last, n, &bx.Left, t)
		} else {
			assertAssignableTo(n, xt, t)
		}
		return
	} else if *x != nil {
		xt := evalStaticTypeOf(store, last, *x)
		if t != nil {
			assertAssignableTo(n, xt, t)
		}
		if isUntyped(xt) {
			// Push type into expr if qualifying binary expr.
			if bx, ok := (*x).(*BinaryExpr); ok {
				switch bx.Op {
				case ADD, SUB, MUL, QUO, REM, BAND, BOR, XOR,
					BAND_NOT, LAND, LOR:
					lt := evalStaticTypeOf(store, last, bx.Left)
					rt := evalStaticTypeOf(store, last, bx.Right)
					if t != nil {
						// push t into bx.Left and bx.Right
						checkOrConvertType(store, last, n, &bx.Left, t)
						checkOrConvertType(store, last, n, &bx.Right, t)
						return
					} else {
						if shouldSwapOnSpecificity(lt, rt) {
							// e.g. 1.0<<s + 1
							// The expression '1.0<<s' does not trigger assertions of
							// incompatible types when evaluated alone.
							// However, when evaluating the full expression '1.0<<s + 1'
							// without a specific context type, '1.0<<s' is checked against
							// its default type, the BigDecKind, will trigger assertion failure.
							// so here in checkOrConvertType, shift expression is "finally" checked.
							checkOrConvertType(store, last, n, &bx.Left, lt)
							checkOrConvertType(store, last, n, &bx.Right, lt)
						} else {
							checkOrConvertType(store, last, n, &bx.Left, rt)
							checkOrConvertType(store, last, n, &bx.Right, rt)
						}
					}
					return
				case EQL, LSS, GTR, NEQ, LEQ, GEQ:
					lt := evalStaticTypeOf(store, last, bx.Left)
					rt := evalStaticTypeOf(store, last, bx.Right)
					if shouldSwapOnSpecificity(lt, rt) {
						checkOrConvertType(store, last, n, &bx.Left, lt)
						checkOrConvertType(store, last, n, &bx.Right, lt)
					} else {
						checkOrConvertType(store, last, n, &bx.Left, rt)
						checkOrConvertType(store, last, n, &bx.Right, rt)
					}
					// this is not a constant expression; the result here should
					// always be a BoolType. (in this scenario, we may have some
					// UntypedBoolTypes)
					t = BoolType
				default:
					// do nothing
				}
			} else if ux, ok := (*x).(*UnaryExpr); ok {
				xt := evalStaticTypeOf(store, last, *x)
				// check assignable first
				assertAssignableTo(n, xt, t)

				if t == nil || t.Kind() == InterfaceKind {
					t = defaultTypeOf(xt)
				}
				checkOrConvertType(store, last, n, &ux.X, t)
				return
			}
		}
	}
	// convert recursively
	convertType(store, last, n, x, t)
}

// 1. convert x to t if x is *ConstExpr.
// 2. otherwise, assert that x can be coerced to t.
// NOTE: also see checkOrConvertIntegerKind()
func convertType(store Store, last BlockNode, n Node, x *Expr, t Type) {
	if debug {
		debug.Printf("convertType, *x: %v:, t:%v \n", *x, t)
	}
	if cx, ok := (*x).(*ConstExpr); ok {
		convertConst(store, last, n, cx, t)
	} else if *x != nil {
		xt := evalStaticTypeOf(store, last, *x)
		if isUntyped(xt) {
			if t == nil {
				t = defaultTypeOf(xt)
			}
			// convert x to destination type t
			doConvertType(store, last, x, t)
		} else {
			// if t is interface do nothing
			if t != nil && t.Kind() == InterfaceKind {
				// do nothing
			} else if isNamedConversion(xt, t) {
				// if one side is declared name type and the other side is unnamed type
				// covert right (xt) to the type of the left (t)
				doConvertType(store, last, x, t)
			}
		}
	}
}

// convert x to destination type t
func doConvertType(store Store, last BlockNode, x *Expr, t Type) {
	cx := Expr(Call(toConstTypeExpr(last, *x, t), *x))
	cx = Preprocess(store, last, cx).(Expr)
	*x = cx
}

// isNamedConversion returns true if assigning a value of type
// xt (rhs) into a value of type t (lhs) entails an implicit type conversion.
// xt is the result of an expression type.
//
// In a few special cases, we should not perform the conversion:
//
//	case 1: the LHS is an interface, which is unnamed, so we should not
//	convert to that even if right is a named type.
//	case 2: isNamedConversion is called within evaluating make() or new()
//	(uverse functions). It returns TypType (generic) which does have IsNamed appropriate
func isNamedConversion(xt, t Type) bool {
	if t == nil {
		t = xt
	}
	// no conversion case 1: the LHS is an interface

	_, c1 := t.(*InterfaceType)

	// no conversion case2: isNamedConversion is called within evaluating make() or new()
	//   (uverse functions)

	_, oktt := t.(*TypeType)
	_, oktt2 := xt.(*TypeType)
	c2 := oktt || oktt2

	if !c1 && !c2 { // carve out above two cases
		// covert right to the type of left if one side is unnamed type and the other side is not

		if t.IsNamed() && !xt.IsNamed() ||
			!t.IsNamed() && xt.IsNamed() {
			return true
		}
	}
	return false
}

// like checkOrConvertType(last, x, nil)
func convertIfConst(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok {
		convertConst(store, last, n, cx, nil)
	}
}

func convertConst(store Store, last BlockNode, n Node, cx *ConstExpr, t Type) {
	if t != nil && t.Kind() == InterfaceKind {
		if cx.T != nil {
			assertAssignableTo(n, cx.T, t)
		}
		t = nil // signifies to convert to default type.
	}
	if isUntyped(cx.T) {
		ConvertUntypedTo(&cx.TypedValue, t)
		setConstAttrs(cx)
	} else if t != nil {
		// e.g. a named type or uint8 type to int for indexing.
		ConvertTo(nilAllocator, store, &cx.TypedValue, t, true)
		setConstAttrs(cx)
	}
}

// Returns any names not yet defined nor predefined in expr.  These happen
// upon transcribe:enter from the top, so value paths cannot be used.  If no
// names are un and x is TypeExpr, evalStaticType(store,last, x) must not
// panic.
//
// NOTE: has no side effects except for the case of composite
// type expressions, which must get preprocessed for inner
// composite type eliding to work.
//
// Args:
//   - direct: If true x must not be a *NameExpr in stack/defining (illegal direct recursion).
//   - elide: For composite type eliding.
//
// Returns:
//   - un: undefined dependency's name if any
//   - directR: if un != "", `direct` passed to final name expr.
//     NOTE: 'direct' is passed through, or becomes overridden with false and
//     passed to higher/later calls in the stack, and the `direct` argument
//     seen at the top of the stack is returned all the way back.
func findUndefinedV(store Store, last BlockNode, x Expr, stack []Name, defining map[Name]struct{}, direct bool, elide Type) (un Name, directR bool) {
	return findUndefinedAny(store, last, x, stack, defining, false, direct, false, elide)
}

func findUndefinedT(store Store, last BlockNode, x Expr, stack []Name, defining map[Name]struct{}, isalias bool, direct bool) (un Name, directR bool) {
	return findUndefinedAny(store, last, x, stack, defining, isalias, direct, true, nil)
}

func findUndefinedAny(store Store, last BlockNode, x Expr, stack []Name, defining map[Name]struct{}, isalias bool, direct bool, astype bool, elide Type) (un Name, directR bool) {
	if debugFind {
		fmt.Printf("findUndefinedAny(%v, %v, %v, isalias=%v, direct=%v, astype=%v, elide=%v\n", x, stack, defining, isalias, direct, astype, elide)
	}
	if x == nil {
		return
	}
	switch cx := x.(type) {
	case *NameExpr:
		if _, ok := UverseNode().GetLocalIndex(cx.Name); ok {
			return
		}
		/*
			if _, ok := defining[cx.Name]; !ok {
				return cx.Name
			}

				else if idx, ok := UverseNode().GetLocalIndex(tx.Name); ok {
					// uverse name
					path := NewValuePathUverse(idx, tx.Name)
					tv := Uverse().GetValueAt(nil, path)
					t = tv.GetType()
				} else {
					// yet undefined
					un = tx.Name
					directR = direct // returns along callstack.
					untype = true
					return
				}
		*/
		// XXX simplify
		if direct {
			if astype {
				if _, ok := defining[cx.Name]; ok {
					panic(fmt.Sprintf("invalid recursive type: %s -> %s",
						Names(stack).Join(" -> "), cx.Name))
				}
				if tv := last.GetSlot(store, cx.Name, true); tv != nil {
					return
				}
			} else {
				if tv := last.GetSlot(store, cx.Name, true); tv != nil {
					return
				}
				// find .loopvar_xxx
				name2 := Name(fmt.Sprintf(".loopvar_%s", cx.Name))
				if tv := last.GetSlot(store, name2, true); tv != nil {
					return
				}
				return cx.Name, direct
			}
		} else {
			if tv := last.GetSlot(store, cx.Name, true); tv != nil {
				return
			}
		}
		return cx.Name, direct
	case *BasicLitExpr:
		return
	case *BinaryExpr:
		un, directR = findUndefinedV(store, last, cx.Left, stack, defining, direct, nil)
		if un != "" {
			return
		}
		un, directR = findUndefinedV(store, last, cx.Right, stack, defining, direct, nil)
		if un != "" {
			return
		}
	case *SelectorExpr:
		return findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
	case *SliceExpr:
		un, directR = findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
		if un != "" {
			return
		}
		if cx.Low != nil {
			un, directR = findUndefinedV(store, last, cx.Low, stack, defining, direct, nil)
			if un != "" {
				return
			}
		}
		if cx.High != nil {
			un, directR = findUndefinedV(store, last, cx.High, stack, defining, direct, nil)
			if un != "" {
				return
			}
		}
		if cx.Max != nil {
			un, directR = findUndefinedV(store, last, cx.Max, stack, defining, direct, nil)
			if un != "" {
				return
			}
		}
	case *StarExpr: // POINTER & DEREF
		// NOTE: *StarExpr can either mean dereference, or a pointer type.
		// It's not only confusing for new developers, it causes complexity
		// in type checking. A *StarExpr is indirect as a type unless alias.
		if astype {
			return findUndefinedT(store, last, cx.X, stack, defining, isalias, isalias)
		} else {
			return findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
		}
	case *RefExpr:
		return findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
	case *TypeAssertExpr:
		un, directR = findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
		if un != "" {
			return
		}
		return findUndefinedT(store, last, cx.Type, stack, defining, isalias, direct)
	case *UnaryExpr:
		return findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
	case *CompositeLitExpr:
		var ct Type
		if cx.Type == nil {
			panic("should not happen")
			/*
				if elide == nil {
					panic("cannot elide unknown composite type")
				}
				ct = elide
				// When the type expr is elided, .elided is generated.
				tx := Nx(".elided")
				tx.SetSpan(Span{Pos: cx.GetPos(), End: cx.GetPos()})
				cx.Type = toConstTypeExpr(tx, elide)
			*/
		} else {
			un, directR = findUndefinedT(store, last, cx.Type, stack, defining, isalias, astype && direct)
			if un != "" {
				return
			}
			// preprocess now for eliding purposes.
			// TODO recursive preprocessing here is hacky, find a better
			// way.  This cannot be done asynchronously, cuz undefined
			// names ought to be returned immediately to let the caller
			// predefine it.
			cx.Type = Preprocess(store, last, cx.Type).(Expr) // recursive
			ct = evalStaticType(store, last, cx.Type)
			// elide composite lit element (nested) composite types.
			elideCompositeElements(last, cx, ct)
		}
		switch ct.Kind() {
		case ArrayKind, SliceKind, MapKind:
			for _, kvx := range cx.Elts {
				un, directR = findUndefinedV(store, last, kvx.Key, stack, defining, direct, nil)
				if un != "" {
					return
				}
				un, directR = findUndefinedV(store, last, kvx.Value, stack, defining, direct, ct.Elem())
				if un != "" {
					return
				}
			}
		case StructKind:
			for _, kvx := range cx.Elts {
				un, directR = findUndefinedV(store, last, kvx.Value, stack, defining, direct, nil)
				if un != "" {
					return
				}
			}
		default:
			panic(fmt.Sprintf(
				"unexpected composite lit type %s",
				ct.String()))
		}
	case *FuncLitExpr:
		un, directR = findUndefinedT(store, last, &cx.Type, stack, defining, isalias, astype && isalias)
		if un != "" {
			return
		}
		_, _, found := findLastFunction(last, nil)
		if !found {
			cx.SetAttribute(ATTR_PREPROCESS_SKIPPED, AttrPreprocessFuncLitExpr)
		}
	case *FieldTypeExpr: // FIELD
		return findUndefinedT(store, last, cx.Type, stack, defining, isalias, direct)
	case *ArrayTypeExpr:
		if cx.Len != nil {
			un, directR = findUndefinedV(store, last, cx.Len, stack, defining, direct, nil)
			if un != "" {
				return
			}
		}
		return findUndefinedT(store, last, cx.Elt, stack, defining, isalias, direct)
	case *SliceTypeExpr:
		return findUndefinedT(store, last, cx.Elt, stack, defining, isalias, astype && isalias)
	case *InterfaceTypeExpr:
		for i := range cx.Methods {
			method := &cx.Methods[i]
			direct2 := false
			if _, ok := method.Type.(*NameExpr); ok {
				direct2 = true
			}
			un, directR = findUndefinedT(store, last, &cx.Methods[i], stack, defining, isalias, direct2)
			if un != "" {
				return
			}
		}
	case *ChanTypeExpr:
		return findUndefinedT(store, last, cx.Value, stack, defining, isalias, astype && isalias)
	case *FuncTypeExpr:
		for i := range cx.Params {
			un, directR = findUndefinedT(store, last, &cx.Params[i], stack, defining, isalias, astype && isalias)
			if un != "" {
				return
			}
		}
		for i := range cx.Results {
			un, directR = findUndefinedT(store, last, &cx.Results[i], stack, defining, isalias, astype && isalias)
			if un != "" {
				return
			}
		}
	case *MapTypeExpr: // MAP
		un, directR = findUndefinedT(store, last, cx.Key, stack, defining, isalias, astype && isalias)
		if un != "" {
			return
		}
		// e.g.;
		// type Int = map[Int]IntIllegal;
		// type Int = struct{Int};
		// type Int = *Int;
		un, directR = findUndefinedT(store, last, cx.Value, stack, defining, isalias, isalias)
		if un != "" {
			return
		}
	case *StructTypeExpr: // STRUCT
		for i := range cx.Fields {
			un, directR = findUndefinedT(store, last, &cx.Fields[i], stack, defining, isalias, direct)
			if un != "" {
				return
			}
		}
	case *CallExpr:
		un, directR = findUndefinedV(store, last, cx.Func, stack, defining, direct, nil)
		if un != "" {
			return
		}
		for i := range cx.Args {
			un, directR = findUndefinedV(store, last, cx.Args[i], stack, defining, direct, nil)
			if un != "" {
				return
			}
		}
	case *IndexExpr:
		un, directR = findUndefinedV(store, last, cx.X, stack, defining, direct, nil)
		if un != "" {
			return
		}
		un, directR = findUndefinedV(store, last, cx.Index, stack, defining, direct, nil)
		if un != "" {
			return
		}
	case *constTypeExpr:
		return
	case *ConstExpr:
		return
	default:
		panic(fmt.Sprintf(
			"unexpected expr: %v (%v)",
			x, reflect.TypeOf(x)))
	}
	return
}

// like checkOrConvertType() but for any typed bool kind.
func checkOrConvertBoolKind(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok {
		convertConst(store, last, n, cx, BoolType)
	} else if x != nil {
		xt := evalStaticTypeOf(store, last, x)
		checkBoolKind(xt)
	}
}

// assert that xt is a typed bool kind.
func checkBoolKind(xt Type) {
	switch xt.Kind() {
	case BoolKind:
		return // ok
	default:
		panic(fmt.Sprintf(
			"expected typed bool kind, but got %v",
			xt.Kind()))
	}
}

// like checkOrConvertType() but for any typed integer kind.
func checkOrConvertIntegerKind(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok {
		convertConst(store, last, n, cx, IntType)
	} else if x != nil {
		xt := evalStaticTypeOf(store, last, x)
		checkIntegerKind(xt)
	}
}

// assert that xt is a typed integer kind.
func checkIntegerKind(xt Type) {
	switch xt.Kind() {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
		BigintKind:
		return // ok
	default:
		panic(fmt.Sprintf(
			"expected typed integer kind, but got %v",
			xt.Kind()))
	}
}

// predefineRecursively() recursively defines or predefines (partially defines)
// all package level declaration names. *FuncDecl and *ValueDecl declarations
// are predefined (partially defined) while *ImportDecl and *TypeDecl are fully
// defined.
//
// It assumes that initStaticBlock has been called so that names are pre-reserved.
//
// It does NOT enter function bodies.
//
// This function works for all block nodes and is called for all declarations,
// but in two stages: in the first stage at the file level, and the second
// stage while preprocessing the body of each function.
//
// Returns true if the result was also preprocessed for *ImportDecl, *TypeDecl
// and *ValueDecl. *ValueDecl values are NOT evaluated at this stage. *FuncDecl
// are only partially defined and also only partially preprocessed.
func predefineRecursively(store Store, last BlockNode, d Decl) bool {
	defer doRecover([]BlockNode{last}, d)
	stack := []Name{}
	defining := make(map[Name]struct{})
	direct := true
	return predefineRecursively2(store, last, d, stack, defining, direct)
}

// `stack` and `defining` are used for cycle detection. They hold the same data.
// NOTE: `stack` never truncates; a slice is used instead of a map to show a
// helpful message when a circular declaration is found. `defining` is also used as
// a map to ensure best time performance of circular definition detection.
func predefineRecursively2(store Store, last BlockNode, d Decl, stack []Name, defining map[Name]struct{}, direct bool) bool {
	pkg := packageOf(last)

	// NOTE: predefine fileset breaks up circular definitions like
	// `var a, b, c = 1, a, b` which is only legal at the file level.
	for _, dn := range d.GetDeclNames() {
		if isUverseName(dn) {
			panic(fmt.Sprintf(
				"builtin identifiers cannot be shadowed: %s", dn))
		}
		stack = append(stack, dn)
		defining[dn] = struct{}{}
	}
	// After definition, remove items from map.
	// The stack doesn't need to be truncated because
	// the caller's stack slice remains the same regardless,
	// but the defining map must remove predefined names,
	// otherwise later declarations will fail.
	defer func() {
		for _, dn := range d.GetDeclNames() {
			delete(defining, dn)
		}
	}()

	// recursively predefine any dependencies.
	var un Name // undefined name
	var untype bool
	var directR bool
	// var direct = true // invalid cycle detection
	for {
		un, untype, directR = tryPredefine(store, pkg, last, d, stack, defining, direct)
		if debugFind {
			fmt.Printf("tryPredefine(%v, %v, defining=%v, direct=%v)-->un=%v,untype=%v,direct2=%v\n", d, stack, defining, direct, un, untype, directR)
		}
		if un != "" {
			// `un` is undefined, so define recursively.
			// first, check circularity.
			if _, exists := defining[un]; exists {
				if untype {
					panic(fmt.Sprintf("invalid recursive type: %s -> %s",
						Names(stack).Join(" -> "), un))
				} else {
					panic(fmt.Sprintf("invalid recursive value: %s -> %s",
						Names(stack).Join(" -> "), un))
				}
			}
			// look up dependency declaration from fileset.
			file, unDecl := pkg.FileSet.GetDeclFor(un)
			// preprocess if not already preprocessed.
			if !file.IsInitialized() {
				panic("all types from files in file-set should have already been predefined")
			}
			// predefine dependency recursively.
			// `directR` is passed on.
			predefineRecursively2(store, file, *unDecl, stack, defining, directR)
		} else {
			break // predefine successfully performed.
		}
	}
	switch cd := d.(type) {
	case *FuncDecl:
		// We cannot Preprocess the body of *FuncDecl as it may
		// refer to package names in any order.
		return false
	case *ValueDecl:
		vd2 := Preprocess(store, last, cd).(*ValueDecl)
		*cd = *vd2
		return true
	case *TypeDecl:
		td2 := Preprocess(store, last, cd).(*TypeDecl)
		*cd = *td2
		return true
	case *ImportDecl:
		id2 := Preprocess(store, last, cd).(*ImportDecl)
		*cd = *id2
		return true
	default:
		panic("should not happen")
	}
}

// If a dependent name is not yet defined, that name is returned; this return
// value is used by the caller to enforce declaration order.
//
// If all dependencies are met, constructs and empty definition value (for a
// *TypeDecl is a TypeValue) and sets it on last. As an exception, *FuncDecls
// will preprocess receiver/argument/result types recursively.
func tryPredefine(store Store, pkg *PackageNode, last BlockNode, d Decl, stack []Name, defining map[Name]struct{}, direct bool) (un Name, untype bool, directR bool) {
	if d.GetAttribute(ATTR_PREDEFINED) == true {
		panic(fmt.Sprintf("decl node already predefined! %v", d))
	}

	// If un is blank, it means the predefine succeeded.
	defer func() {
		if un == "" {
			d.SetAttribute(ATTR_PREDEFINED, true)
		}
	}()

	// NOTE: These happen upon enter from the top,
	// so value paths cannot be used here.
	switch d := d.(type) {
	case *ImportDecl:
		// stdlib internal package
		if strings.HasPrefix(d.PkgPath, "internal/") && !IsStdlib(pkg.PkgPath) {
			panic("cannot import stdlib internal/ package outside of standard library")
		}

		base, isInternal := IsInternalPath(d.PkgPath)
		if isInternal &&
			pkg.PkgPath != base &&
			!strings.HasPrefix(pkg.PkgPath, base+"/") {
			panic("internal/ packages can only be imported by packages rooted at the parent of \"internal\"")
		}

		// NOTE: imports from "pure packages" are actually sometimes
		// allowed, most notably filetests.
		if IsPPackagePath(pkg.PkgPath) && IsRealmPath(d.PkgPath) && !IsTestFile(last.GetLocation().File) {
			panic(fmt.Sprintf("pure package path %q cannot import realm path %q", pkg.PkgPath, d.PkgPath))
		}

		pv := store.GetPackage(d.PkgPath, true)
		if pv == nil {
			panic(fmt.Sprintf(
				"unknown import path %s",
				d.PkgPath))
		}
		switch d.Name {
		case "": // use default
			exp, ok := expectedPkgName(d.PkgPath)
			if !ok {
				// should not happen, because the package exists in the store.
				panic(fmt.Sprintf("invalid pkg path: %q", d.PkgPath))
			}
			if exp != string(pv.PkgName) {
				panic(fmt.Sprintf(
					"package name for %q (%q) doesn't match its expected identifier %q; "+
						"the import declaration must specify an identifier", pv.PkgPath, pv.PkgName, exp))
			}
			d.Name = pv.PkgName
		case blankIdentifier: // no definition
			return
		case ".": // dot import
			panic("dot imports not allowed in Gno")
		}
		// NOTE imports usually must happen with a file,
		// and so last is usually a *FileNode, but for
		// testing convenience we allow importing
		// directly onto the package.
		last.Define(d.Name, TypedValue{
			T: gPackageType,
			V: pv,
		})
		d.Path = last.GetPathForName(store, d.Name)
	case *ValueDecl:
		// check for blank identifier in type
		// e.g., `var x _`
		if isBlankIdentifier(d.Type) {
			panic("cannot use _ as value or type")
		}
		isalias := false                                                                   // a value decl can't be.
		un, directR = findUndefinedT(store, last, d.Type, stack, defining, isalias, false) // XXX
		if un != "" {
			untype = true
			return
		}
		// NOTE: cyclic *ValueDecl at the package/file level such as
		// `var a, b, c = 1, a, b` was already split up before reaching
		// here, whereas they are illegal inside a function.
		for _, vx := range d.Values {
			un, directR = findUndefinedV(store, last, vx, stack, defining, direct, nil)
			if un != "" {
				untype = false
				return
			}
		}
		for i := range d.NameExprs {
			nx := &d.NameExprs[i]
			if nx.Name == blankIdentifier {
				nx.Path.Name = blankIdentifier
			} else {
				nx.Path = last.GetPathForName(store, nx.Name)
			}
		}
	case *TypeDecl:
		if d.Name == blankIdentifier {
			return
		}
		// before looking for dependencies, predefine empty type.
		last2 := skipFile(last)
		if !isLocallyDefined(last2, d.Name) {
			// construct empty t type
			var t Type
			switch tx := d.Type.(type) {
			case *FuncTypeExpr:
				t = &FuncType{}
			case *ArrayTypeExpr:
				t = &ArrayType{}
			case *SliceTypeExpr:
				t = &SliceType{}
			case *InterfaceTypeExpr:
				t = &InterfaceType{}
			case *ChanTypeExpr:
				t = &ChanType{}
			case *MapTypeExpr:
				t = &MapType{}
			case *StructTypeExpr:
				t = &StructType{}
			case *StarExpr:
				t = &PointerType{}
			case *NameExpr:
				// check for blank identifier in type
				// e.g., `type T _`
				if isBlankIdentifier(tx) {
					panic("cannot use _ as value or type")
				}
				// do not allow nil as type.
				if tx.Name == "nil" {
					panic("nil is not a type")
				}
				// sanity check.
				if tv := last.GetSlot(store, tx.Name, true); tv != nil {
					t = tv.GetType()
					if dt, ok := t.(*DeclaredType); ok {
						if !dt.sealed {
							// predefineRecursively should have
							// already preprocessed dependent types!
							panic("should not happen")
						}
					}
				}
				// set t for proper type.
				if idx, ok := UverseNode().GetLocalIndex(tx.Name); ok {
					// uverse name
					path := NewValuePathUverse(idx, tx.Name)
					tv := Uverse().GetValueAt(nil, path)
					t = tv.GetType()
				}
			case *SelectorExpr:
				// get package value.
				un, directR = findUndefinedV(store, last, tx.X, stack, defining, false, nil)
				if un != "" {
					untype = true
					return
				}
				pkgName := tx.X.(*NameExpr).Name
				tv := last.GetSlot(store, pkgName, true)
				pv, ok := tv.V.(*PackageValue)
				if !ok {
					panic(fmt.Sprintf(
						"unknown package name %s in %s",
						pkgName,
						tx.String(),
					))
				}
				// check package node for name.
				pn := pv.GetPackageNode(store)
				tx.Path = pn.GetPathForName(store, tx.Sel)
				ptr := pv.GetBlock(store).GetPointerTo(store, tx.Path)
				t = ptr.TV.GetType()
			default:
				panic(fmt.Sprintf(
					"unexpected type declaration type %v",
					reflect.TypeOf(d.Type)))
			}
			if d.IsAlias {
				// use t directly.
			} else {
				// create new declared type.
				pn := packageOf(last)
				dt := declareWith(pn.PkgPath, last, d.Name, t)
				t = dt
			}
			// fill in later.
			last2.Define2(true, d.Name, t, asValue(t), NameSource{&d.NameExpr, d, NSTypeDecl, -1})
			d.Path = last.GetPathForName(store, d.Name)
		} // END if !isLocallyDefined(last2, d.Name) {
		// now it is or was locally defined.

		// after predefinitions (for reasonable recursion support),
		// return any undefined dependencies.
		un, directR = findUndefinedAny(
			store, last, d.Type, stack, defining, d.IsAlias, direct, true, nil)
		if un != "" {
			untype = true
			return
		}
		// END *TypeDecl
	case *FuncDecl:
		un, directR = findUndefinedT(store, last, &d.Type, stack, defining, false, false)
		if un != "" {
			untype = true
			return
		}
		if d.IsMethod {
			un, directR = findUndefinedT(store, last, &d.Recv, stack, defining, false, false)
			if un != "" {
				untype = true
				return
			}
			if d.Recv.Name == "" || d.Recv.Name == blankIdentifier {
				panic("d.Recv.Name should have been set in initStaticBlocks")
			}
			d.Recv = *Preprocess(store, last, &d.Recv).(*FieldTypeExpr)
			d.Type = *Preprocess(store, last, &d.Type).(*FuncTypeExpr)
			rft := evalStaticType(store, last, &d.Recv).(FieldType)
			rt := rft.Type
			ft := evalStaticType(store, last, &d.Type).(*FuncType)
			ubft := ft.UnboundType(rft)
			dt := (*DeclaredType)(nil)

			// check base type of receiver type, should not be pointer type or interface type
			assertValidReceiverType := func(t Type) {
				if _, ok := t.(*PointerType); ok {
					panic(fmt.Sprintf("invalid receiver type %v (base type is pointer type)", rt))
				}
				if _, ok := t.(*InterfaceType); ok {
					panic(fmt.Sprintf("invalid receiver type %v (base type is interface type)", rt))
				}
			}

			if pt, ok := rt.(*PointerType); ok {
				assertValidReceiverType(pt.Elem())
				if ddt, ok := pt.Elem().(*DeclaredType); ok {
					assertValidReceiverType(baseOf(ddt))
					dt = ddt
				} else {
					panic("should not happen")
				}
			} else if ddt, ok := rt.(*DeclaredType); ok {
				assertValidReceiverType(baseOf(ddt))
				dt = ddt
			} else {
				panic("should not happen")
			}
			// The body may get altered during preprocessing later.
			if !dt.TryDefineMethod(&FuncValue{
				Type:       ubft,
				IsMethod:   true,
				Source:     d,
				Name:       d.Name,
				Parent:     nil, // set lazily
				FileName:   fileNameOf(last),
				PkgPath:    pkg.PkgPath,
				Crossing:   ft.IsCrossing(),
				body:       d.Body,
				nativeBody: nil,
			}) {
				// Revert to old function declarations in the package we're preprocessing.
				pkg := packageOf(last)
				pkg.StaticBlock.revertToOld()
				panic(fmt.Sprintf("redeclaration of method %s.%s",
					dt.Name, d.Name))
			}
		} else {
			if d.Name == "init" {
				panic("d.Name 'init' should have been appended with a number in initStaticBlocks")
			}
			// define package-level function.
			ft := &FuncType{}
			// define a FuncValue w/ above type as d.Name.
			// fill in later during *FuncDecl:BLOCK.
			// The body may get altered during preprocessing later.
			fv := &FuncValue{
				Type:       ft,
				IsMethod:   false,
				Source:     d,
				Name:       d.Name,
				Parent:     nil, // set lazily.
				FileName:   fileNameOf(last),
				PkgPath:    pkg.PkgPath,
				Crossing:   false, // gets set after evalStaticType(d.Type)
				body:       d.Body,
				nativeBody: nil,
			}
			// NOTE: fv.body == nil means no body (ie. not even curly braces)
			// len(fv.body) == 0 could mean also {} (ie. no statements inside)
			if fv.body == nil && store != nil {
				fv.nativeBody = store.GetNative(pkg.PkgPath, d.Name)
				if fv.nativeBody == nil {
					panic(fmt.Sprintf("function %s does not have a body but is not natively defined (did you build after pulling from the repository?)", d.Name))
				}
				fv.NativePkg = pkg.PkgPath
				fv.NativeName = d.Name
			}
			// Set placeholder ft and fv.
			pkg.Define(d.Name, TypedValue{
				T: ft,
				V: fv,
			})
			d.Type = *Preprocess(store, last, &d.Type).(*FuncTypeExpr)
			ft2 := evalStaticType(store, last, &d.Type).(*FuncType)
			if !ft.IsZero() {
				// redefining function.
				// make sure the type is the same.
				if ft.TypeID() != ft2.TypeID() {
					panic(fmt.Sprintf(
						"Redefinition (%s) cannot change .T; was %v, new %v",
						d, ft, ft2))
				}
				// keep the orig type.
			} else {
				*ft = *ft2
			}
			fv.Crossing = ft.IsCrossing()
		}
	default:
		panic(fmt.Sprintf(
			"unexpected declaration type %v",
			d.String()))
	}
	// predefine complete.
	return "", false, false // zero values
}

var reExpectedPkgName = regexp.MustCompile(`(?:^|/)([^/]+)(?:/v\d+)?$`)

// expectedPkgName returns the expected default package name from the given
// package path, given its pkgpath.
//
// This is the last part of the pkgpath, ignoring any version specifier at the
// end of the path; for instance, the expected pkg name of net/url is "url";
// the expected pkg name of math/rand/v2 is "rand".
func expectedPkgName(path string) (string, bool) {
	res := reExpectedPkgName.FindStringSubmatch(path)
	if res == nil {
		return "", false
	}
	return res[1], true
}

func skipFaux(bn BlockNode) BlockNode {
	if fauxBlockNode(bn) {
		return bn.GetParentNode(nil)
	}
	return bn
}

func fauxBlockNode(bn BlockNode) bool {
	switch bn.(type) {
	case *IfStmt, *SwitchStmt:
		return true
	}
	return false
}

func fauxChildBlockNode(bn BlockNode) bool {
	switch bn.(type) {
	case *IfCaseStmt, *SwitchClauseStmt:
		return true
	}
	return false
}

func fillNameExprPath(last BlockNode, nx *NameExpr, isDefineLHS bool) {
	// fmt.Println("---fillNameExprPath, nx: ", nx)
	// defer func() {
	// 	fmt.Println("---path filled: ", nx.Path)
	// }()

	if nx.Name == blankIdentifier {
		// Blank name has no path; caller error.
		panic("should not happen")
	}

	// If not DEFINE_LHS, yet is statically undefined, set path from parent.
	if !isDefineLHS {
		if last.GetStaticTypeOf(nil, nx.Name) == nil {
			// NOTE: We cannot simply call last.GetPathForName() as
			// below here, because .GetPathForName() doesn't
			// distinguish between undefined reserved and and
			// (pre)defined variables. See tests/files/define1.go
			// for test case.
			var (
				path      ValuePath
				i         = 0
				fauxChild = 0
			)
			for {
				i++
				if fauxChildBlockNode(last) {
					fauxChild++
				}
				last = last.GetParentNode(nil)
				if last == nil {
					if isUverseName(nx.Name) {
						idx, ok := UverseNode().GetLocalIndex(nx.Name)
						if !ok {
							panic("should not happen")
						}
						nx.Path = NewValuePathUverse(idx, nx.Name)
						return
					} else {
						panic(fmt.Sprintf(
							"name not defined: %s", nx.Name))
					}
				}
				if last.GetStaticTypeOf(nil, nx.Name) == nil {
					continue
				} else {
					path = last.GetPathForName(nil, nx.Name)
					if path.Type != VPBlock {
						panic("expected block value path type; check this is not shadowing a builtin type")
					}
					break
				}
			}
			path.SetDepth(path.Depth + uint8(i) - uint8(fauxChild))
			path.Validate()
			nx.Path = path
			return
		}
	} else if isUverseName(nx.Name) {
		panic(fmt.Sprintf(
			"builtin identifiers cannot be shadowed: %s", nx.Name))
	}
	// Otherwise, set path for name.
	// Uverse name paths get set here as well.
	nx.Path = last.GetPathForName(nil, nx.Name)
}

func isFile(n BlockNode) bool {
	if _, ok := n.(*FileNode); ok {
		return true
	} else {
		return false
	}
}

func skipFile(n BlockNode) BlockNode {
	if fn, ok := n.(*FileNode); ok {
		return packageOf(fn)
	} else {
		return n
	}
}

// returns "" if not in a file node.
func fileNameOf(n BlockNode) string {
	if n == nil {
		return ""
	}
	fn := fileOfSafe(n)
	if fn == nil {
		return ""
	}
	return fn.FileName
}

func elideCompositeElements(last BlockNode, clx *CompositeLitExpr, clt Type) {
	switch clt := baseOf(clt).(type) {
	case *ArrayType:
		et := clt.Elt
		el := len(clx.Elts)
		for i := range el {
			kvx := &clx.Elts[i]
			elideCompositeExpr(last, &kvx.Value, et)
		}
	case *SliceType:
		et := clt.Elt
		el := len(clx.Elts)
		for i := range el {
			kvx := &clx.Elts[i]
			elideCompositeExpr(last, &kvx.Value, et)
		}
	case *MapType:
		kt := clt.Key
		vt := clt.Value
		el := len(clx.Elts)
		for i := range el {
			kvx := &clx.Elts[i]
			elideCompositeExpr(last, &kvx.Key, kt)
			elideCompositeExpr(last, &kvx.Value, vt)
		}
	case *StructType:
		// Struct fields cannot be elided in Go for
		// legibility, but Gno could support them (e.g. for
		// certain tagged struct fields).
		// TODO: support eliding.
		for _, kvx := range clx.Elts {
			vx := kvx.Value
			if vclx, ok := vx.(*CompositeLitExpr); ok {
				if vclx.Type == nil {
					panic("types cannot be elided in composite literals for struct types")
				}
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected composite lit type %s",
			clt.String()))
	}
}

// if *vx is composite lit type, fill in elided type.
// if composite type is pointer type, replace composite
// expression with ref expr.
func elideCompositeExpr(last BlockNode, x *Expr, t Type) {
	if clx, ok := (*x).(*CompositeLitExpr); ok {
		if clx.Type == nil {
			// Make a placeholder name expr.
			tx := Nx(".elided")
			span := Span{Pos: (*x).GetPos(), End: (*x).GetPos()}
			tx.SetSpan(span)
			// Handle implicit &{}.
			if t.Kind() == PointerKind {
				clx.Type = toConstTypeExpr(last, tx, t.Elem())
				refx := &RefExpr{X: clx}
				refx.SetSpan(clx.GetSpan())
				*x = refx
				elideCompositeElements(last, clx, t.Elem()) // recurse
			} else {
				clx.Type = toConstTypeExpr(last, tx, t)
				elideCompositeElements(last, clx, t) // recurse
			}
		}
	}
}

// returns number of args, or if arg is a call result,
// the number of results of the return tuple type.
func countNumArgs(store Store, last BlockNode, n *CallExpr) (numArgs int) {
	if len(n.Args) != 1 {
		return len(n.Args)
	} else if cx, ok := n.Args[0].(*CallExpr); ok {
		cxift := evalStaticTypeOf(store, last, cx.Func) // cx (iface) func type
		if cxift.Kind() == TypeKind {
			return 1 // type conversion
		} else {
			cxft := getGnoFuncTypeOf(store, cxift)
			numResults := len(cxft.Results)
			return numResults
		}
	} else {
		return 1
	}
}

// This is to be run *after* preprocessing is done,
// to determine the order of var decl execution
// (which may include functions which may refer to package vars).
func findDependentNames(n Node, dst map[Name]struct{}) {
	switch cn := n.(type) {
	case *NameExpr:
		dst[cn.Name] = struct{}{}
	case *BasicLitExpr:
	case *BinaryExpr:
		findDependentNames(cn.Left, dst)
		findDependentNames(cn.Right, dst)
	case *SelectorExpr:
		findDependentNames(cn.X, dst)
	case *SliceExpr:
		findDependentNames(cn.X, dst)
		if cn.Low != nil {
			findDependentNames(cn.Low, dst)
		}
		if cn.High != nil {
			findDependentNames(cn.High, dst)
		}
		if cn.Max != nil {
			findDependentNames(cn.Max, dst)
		}
	case *StarExpr:
		findDependentNames(cn.X, dst)
	case *RefExpr:
		findDependentNames(cn.X, dst)
	case *TypeAssertExpr:
		findDependentNames(cn.X, dst)
		findDependentNames(cn.Type, dst)
	case *UnaryExpr:
		findDependentNames(cn.X, dst)
	case *CompositeLitExpr:
		findDependentNames(cn.Type, dst)
		ct := getType(cn.Type)
		switch ct.Kind() {
		case ArrayKind, SliceKind, MapKind:
			for _, kvx := range cn.Elts {
				if kvx.Key != nil {
					findDependentNames(kvx.Key, dst)
				}
				findDependentNames(kvx.Value, dst)
			}
		case StructKind:
			for _, kvx := range cn.Elts {
				findDependentNames(kvx.Value, dst)
			}
		default:
			panic(fmt.Sprintf(
				"unexpected composite lit type %s",
				ct.String()))
		}
	case *FieldTypeExpr:
		findDependentNames(cn.Type, dst)
	case *ArrayTypeExpr:
		findDependentNames(cn.Elt, dst)
		if cn.Len != nil {
			findDependentNames(cn.Len, dst)
		}
	case *SliceTypeExpr:
		findDependentNames(cn.Elt, dst)
	case *InterfaceTypeExpr:
		for i := range cn.Methods {
			findDependentNames(&cn.Methods[i], dst)
		}
	case *ChanTypeExpr:
		findDependentNames(cn.Value, dst)
	case *FuncTypeExpr:
		for i := range cn.Params {
			findDependentNames(&cn.Params[i], dst)
		}
		for i := range cn.Results {
			findDependentNames(&cn.Results[i], dst)
		}
	case *MapTypeExpr:
		findDependentNames(cn.Key, dst)
		findDependentNames(cn.Value, dst)
	case *StructTypeExpr:
		for i := range cn.Fields {
			findDependentNames(&cn.Fields[i], dst)
		}
	case *CallExpr:
		findDependentNames(cn.Func, dst)
		for i := range cn.Args {
			findDependentNames(cn.Args[i], dst)
		}
	case *IndexExpr:
		findDependentNames(cn.X, dst)
		findDependentNames(cn.Index, dst)
	case *FuncLitExpr:
		findDependentNames(&cn.Type, dst)
		for _, n := range cn.GetExternNames() {
			dst[n] = struct{}{}
		}
	case *constTypeExpr:
	case *ConstExpr:
	case *ImportDecl:
	case *ValueDecl:
		if cn.Type != nil {
			findDependentNames(cn.Type, dst)
		}
		for _, vx := range cn.Values {
			findDependentNames(vx, dst)
		}
	case *TypeDecl:
		findDependentNames(cn.Type, dst)
	case *FuncDecl:
		findDependentNames(&cn.Type, dst)
		if cn.IsMethod {
			findDependentNames(&cn.Recv, dst)
			for _, n := range cn.GetExternNames() {
				dst[n] = struct{}{}
			}
		} else {
			for _, n := range cn.GetExternNames() {
				if n == cn.Name {
					// top-level function referring to itself
				} else {
					dst[n] = struct{}{}
				}
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected node: %v (%v)",
			n, reflect.TypeOf(n)))
	}
}

// A name is locally defined on a block node
// if the type is set to anything but nil.
// A predefined name will return false.
// NOTE: the value is not necessarily set statically,
// unless it refers to a type, package, or statically declared func value.
func isLocallyDefined(bn BlockNode, n Name) bool {
	idx, ok := bn.GetLocalIndex(n)
	if !ok {
		return false
	}
	t := bn.GetStaticBlock().Types[idx]
	return t != nil
}

// r := 0
// r, ok := 1, true
func isLocallyDefined2(bn BlockNode, n Name) bool {
	_, isLocal := bn.GetLocalIndex(n)
	return isLocal
}

// ----------------------------------------
// setNodeLines & setNodeLocations

func setNodeLines(nn Node) {
	var all Span // if nn has no span defined, derive one from its contents.
	Transcribe(nn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		switch stage {
		case TRANS_ENTER:
			nspan := n.GetSpan()
			if nspan.IsZero() {
				if len(ns) == 0 {
					// Handled by TRANS_LEAVE. Case A.
					return n, TRANS_CONTINUE
				}
				lastSpan := ns[len(ns)-1].GetSpan()
				if lastSpan.IsZero() {
					// Handled by TRANS_LEAVE too. Case B.
					return n, TRANS_CONTINUE
				} else {
					// Derive from parent with .Num++.
					nspan = lastSpan
					nspan.Num += 1
					n.SetSpan(nspan)
					return n, TRANS_CONTINUE
				}
			} else {
				// Expand all with nspan.
				all = all.Union(nspan)
				return n, TRANS_CONTINUE
			}
		case TRANS_LEAVE:
			// TODO: Validate that all spans of elements are
			// strictly contained.
			nspan := n.GetSpan()
			if nspan.IsZero() {
				// Case A: len(ns) == 0
				// Case b: len(ns) > 0.
				n.SetSpan(all)
			}
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}

// Iterate over all nodes recursively and sets location information
// based on sparse expectations on block nodes, and ensures uniqueness of BlockNode.Locations.
// Ensures uniqueness of BlockNode.Locations.
func setNodeLocations(pkgPath string, fileName string, n Node) {
	if pkgPath == "" || fileName == "" {
		panic("missing package path or file name")
	}
	Transcribe(n, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		if bn, ok := n.(BlockNode); ok {
			// ensure unique location of blocknode.
			loc := Location{
				PkgPath: pkgPath,
				File:    fileName,
				Span:    bn.GetSpan(),
			}
			bn.SetLocation(loc)
		}
		return n, TRANS_CONTINUE
	})
}

// XXX check node lines, uniqueness of locations,
// and also check location pkgpath and filename.
// Even after this is implemented, locations should not be used for logic.
func checkNodeLinesLocations(pkgPath string, fileName string, n Node) {
	// TODO: XXX
}

// ----------------------------------------
// SaveBlockNodes

// Iterate over all block nodes recursively and saves them.
// Ensures uniqueness of BlockNode.Locations.
func SaveBlockNodes(store Store, fn *FileNode) {
	// First, get the package and file names.
	pn := packageOf(fn)
	store.SetBlockNode(pn)
	pkgPath := pn.PkgPath
	fileName := fn.FileName
	if pkgPath == "" || fileName == "" {
		panic("missing package path or file name")
	}
	Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		// save node to store if blocknode.
		if bn, ok := n.(BlockNode); ok {
			// Location must exist already.
			loc := bn.GetLocation()
			if loc.IsZero() {
				panic("unexpected zero block node location")
			}
			if loc.PkgPath != pkgPath {
				panic("unexpected pkg path in node location")
			}
			if loc.File != fileName {
				panic("unexpected file name in node location")
			}
			if loc.Line != bn.GetLine() {
				panic("wrong line in block node location")
			}
			if loc.Column != bn.GetColumn() {
				panic("wrong column in block node location")
			}
			// save blocknode.
			store.SetBlockNode(bn)
		}
		return n, TRANS_CONTINUE
	})
}
