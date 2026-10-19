package gnolang

import (
	"cmp"
	stderrors "errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/store"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	blankIdentifier           = "_"
	iotaIdentifier            = "iota"
	debugFind                 = false // toggle when debugging.
	AttrPreprocessFuncLitExpr = "FuncLitExpr"
	TestingBasePkgPath        = "testing"
)

// Predefine (initStaticBlocks) and partially evaluates all names
// not inside function bodies.
//
// This function must be called on *FileSets because declarations
// in file sets may be unordered.
// PredefineFileSet predefines every declaration of fset; see Preprocess.
func PredefineFileSet(store Store, pn *PackageNode, fset *FileSet) {
	predefineFileSet(store, pn, fset, nil)
}

// PreprocessFilesCollect predefines and preprocesses fset in accumulate
// mode (see PreprocessCollect) and returns every error found, in source
// order of discovery. The nodes must not be executed afterwards.
func PreprocessFilesCollect(store Store, pn *PackageNode, fset *FileSet) []*PreprocessError {
	acc := &preprocessAccumulator{poisoned: map[Name]bool{syntaxErrName: true}, referenced: map[Name]bool{}}
	predefineFileSet(store, pn, fset, acc)
	for _, fn := range fset.Files {
		preprocess(store, pn, fn, acc)
	}
	acc.errs = append(acc.errs, checkInitCycles(preprocessGasMeterOf(store), pn, fset)...)
	reportUnusedImports(store, acc, pn, fset)
	return acc.errs
}

// checkInitCycles reports every initialization cycle among the package's
// variables: a variable whose initializer, directly or through the
// functions and methods it uses, depends on itself (Go spec, "Package
// initialization"). The dependency edges are the ATTR_DECL_DEPS that
// codaInitOrderDeps recorded on each declaration. One pass of Tarjan's
// algorithm finds the strongly connected components; a component is a
// cycle when it holds a variable (functions are edges, not roots), and is
// reported once, at the variable the cycle closes on (see initCycleOf),
// in source order. Each declaration and edge visited bills
// OpCPUSlopeInitCycleVisit on gm. On chain CheckInitCycles panics with
// the first one, where runFileDecls would otherwise fail later and
// unpositioned.
func checkInitCycles(gm store.GasMeter, pn *PackageNode, fset *FileSet) []*PreprocessError {
	// Names resolve to the declaration predefine selected (the last
	// file's wins), and only those are roots: a shadowed declaration is
	// neither, or it would close a false cycle.
	index := newPredefineDeclIndex(fset)
	w := &initCycleWalk{
		gm:    gm,
		order: map[Decl]int{},
		file:  map[Decl]*FileNode{},
		nodes: map[Decl]*initCycleNode{},
	}
	w.resolve = func(name Name) (Decl, bool) {
		if e, ok := index.entries[name]; ok {
			return e.decl, true
		}
		return resolveDeclDepSafe(name, pn)
	}
	for _, fn := range fset.Files {
		for _, d := range fn.Decls {
			w.order[d] = len(w.order)
			w.file[d] = fn
		}
	}
	for _, fn := range fset.Files {
		for _, d := range fn.Decls {
			vd, ok := d.(*ValueDecl)
			if !ok || !w.live(vd, index) {
				continue
			}
			if _, seen := w.nodes[vd]; !seen {
				w.connect(vd)
			}
		}
	}
	type report struct {
		at   Decl
		path []Name
	}
	var reports []report
	for _, scc := range w.sccs {
		if at, path := w.initCycleOf(scc); at != nil {
			reports = append(reports, report{at, path})
		}
	}
	slices.SortFunc(reports, func(a, b report) int { return cmp.Compare(w.declOrder(a.at), w.declOrder(b.at)) })
	var errs []*PreprocessError
	for _, r := range reports {
		var msg strings.Builder
		msg.WriteString("initialization cycle: ")
		for i, name := range r.path {
			if i > 0 {
				msg.WriteString(" refers to ")
			}
			msg.WriteString(string(name))
		}
		loc := Location{PkgPath: pn.PkgPath, Span: r.at.GetSpan()}
		if fn := w.file[r.at]; fn != nil {
			loc.File = fn.FileName
		}
		errs = append(errs, &PreprocessError{err: fmt.Errorf("%s: %s", loc.String(), msg.String())})
	}
	return errs
}

// initCycleWalk is Tarjan's strongly-connected-components walk over the
// declaration dependency graph.
type initCycleWalk struct {
	gm      store.GasMeter
	resolve func(Name) (Decl, bool)
	order   map[Decl]int       // source order of the file set's declarations
	file    map[Decl]*FileNode // declaring file
	nodes   map[Decl]*initCycleNode
	stack   []Decl
	sccs    [][]Decl // in completion order
}

// initCycleNode is a visited declaration's walk state.
type initCycleNode struct {
	number, low int // Tarjan index, in visit order, and low link
	onStack     bool
	deps        []Name // dependency names, sorted so the walk is deterministic
}

// live reports whether vd is the declaration predefine selected for one
// of its names; a shadowed declaration is not walked.
func (w *initCycleWalk) live(vd *ValueDecl, index *predefineDeclIndex) bool {
	for _, name := range vd.GetDeclNames() {
		if index.entries[name].decl == vd {
			return true
		}
	}
	return false
}

// declOrder is the source position of d among the walked files; a
// declaration from an earlier file set sorts last.
func (w *initCycleWalk) declOrder(d Decl) int {
	if n, ok := w.order[d]; ok {
		return n
	}
	return math.MaxInt
}

func (w *initCycleWalk) connect(d Decl) {
	chargeCPUGas(w.gm, OpCPUSlopeInitCycleVisit)
	m, _ := d.GetAttribute(ATTR_DECL_DEPS).(map[Name]struct{})
	nd := &initCycleNode{number: len(w.nodes), onStack: true, deps: slices.Collect(maps.Keys(m))}
	nd.low = nd.number
	slices.Sort(nd.deps)
	w.nodes[d] = nd
	w.stack = append(w.stack, d)
	for _, name := range nd.deps {
		chargeCPUGas(w.gm, OpCPUSlopeInitCycleVisit)
		dep, ok := w.resolve(name)
		if !ok {
			continue
		}
		if nn, seen := w.nodes[dep]; !seen {
			w.connect(dep)
			nd.low = min(nd.low, w.nodes[dep].low)
		} else if nn.onStack {
			nd.low = min(nd.low, nn.number)
		}
	}
	if nd.low != nd.number {
		return
	}
	var scc []Decl
	for {
		top := w.stack[len(w.stack)-1]
		w.stack = w.stack[:len(w.stack)-1]
		w.nodes[top].onStack = false
		scc = append(scc, top)
		if top == d {
			break
		}
	}
	w.sccs = append(w.sccs, scc)
}

// initCycleOf returns the cycle to report for a component, or nil: a
// component with no variable is recursion among functions, not a cycle.
// The cycle is the first one the sorted-name walk from the component's
// earliest variable closes on a variable; it is reported at that
// variable, as its path of names closed with the variable's name.
func (w *initCycleWalk) initCycleOf(scc []Decl) (Decl, []Name) {
	members := make(map[Decl]bool, len(scc))
	var start *ValueDecl
	for _, d := range scc {
		members[d] = true
		if vd, ok := d.(*ValueDecl); ok && (start == nil || w.declOrder(vd) < w.declOrder(start)) {
			start = vd
		}
	}
	if start == nil {
		return nil, nil
	}
	if len(scc) == 1 && !slices.ContainsFunc(w.nodes[start].deps, func(name Name) bool {
		dep, ok := w.resolve(name)
		return ok && dep == start
	}) {
		return nil, nil // a lone variable with no edge to itself
	}
	type step struct {
		d    Decl
		name Name
	}
	var path []step
	onPath := map[Decl]int{}
	var walk func(d Decl) (Decl, []Name)
	walk = func(d Decl) (Decl, []Name) {
		chargeCPUGas(w.gm, OpCPUSlopeInitCycleVisit)
		for _, name := range w.nodes[d].deps {
			chargeCPUGas(w.gm, OpCPUSlopeInitCycleVisit)
			dep, ok := w.resolve(name)
			if !ok || !members[dep] {
				continue
			}
			if i, ok := onPath[dep]; ok {
				if _, isVar := dep.(*ValueDecl); !isVar {
					continue
				}
				names := make([]Name, 0, len(path)-i+1)
				for _, st := range path[i:] {
					names = append(names, st.name)
				}
				return dep, append(names, name)
			}
			onPath[dep] = len(path)
			path = append(path, step{dep, name})
			if at, names := walk(dep); at != nil {
				return at, names
			}
			path = path[:len(path)-1]
			delete(onPath, dep)
		}
		return nil, nil
	}
	onPath[start] = 0
	path = append(path, step{start, start.GetDeclNames()[0]})
	return walk(start)
}

// resolveDeclDepSafe is resolveDeclDep without the panics: a name the
// package cannot resolve (dropped by accumulate mode) is skipped.
func resolveDeclDepSafe(name Name, pn *PackageNode) (d Decl, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			d, ok = nil, false
		}
	}()
	d = resolveDeclDep(name, pn)
	return d, d != nil
}

// CheckInitCycles is the on-chain form of checkInitCycles: the first
// initialization cycle among the package's variables panics as a
// *PreprocessError, positioned at the variable that opens it.
func CheckInitCycles(store Store, pn *PackageNode, fset *FileSet) {
	if errs := checkInitCycles(preprocessGasMeterOf(store), pn, fset); len(errs) > 0 {
		panic(errs[0])
	}
}

// CheckUnusedImports is the on-chain form of the unused-import rule: once
// the package's files are preprocessed, the first import no file refers
// to panics as a *PreprocessError positioned at its import spec.
func CheckUnusedImports(store Store, pn *PackageNode, fset *FileSet) {
	reportUnusedImports(store, nil, pn, fset)
}

func predefineFileSet(store Store, pn *PackageNode, fset *FileSet, acc *preprocessAccumulator) {
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
	if acc != nil {
		acc.checkDuplicates(pn, fset)
	}
	index := newPredefineDeclIndex(pn.FileSet)
	index.acc = acc
	// In accumulate mode a declaration that fails to predefine is recorded
	// and dropped (marked preprocessed, so the walk skips it) instead of
	// stopping the whole file set.
	predefineOne := func(fn *FileNode, d Decl, do func()) {
		if acc == nil {
			do()
			return
		}
		defer func() {
			if r := recover(); r != nil {
				if isRuntimePanic(r) {
					panic(r) // a VM bug, not a checking error
				}
				acc.dropDecl(wrapPreprocessPanic([]BlockNode{pn, fn}, d, r), fn, d)
			}
		}()
		do()
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
				predefineOne(fn, d, func() { predefineRecursivelyIndexed(store, fn, d, index) })
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
				predefineOne(fn, d, func() { predefineRecursivelyIndexed(store, fn, d, index) })
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
				predefineOne(fn, d, func() { predefineRecursivelyIndexed(store, fn, d, index) })
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
				parts := make([]Decl, 0, len(vd.NameExprs))
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
					// Positioned at its name, so an error names the line;
					// the part contains the name (see Span.Union).
					span := vd.NameExprs[j].GetSpan()
					span.Num--
					part.SetSpan(span)
					parts = append(parts, part)
				}
				// Apply the split to fn.Decls BEFORE calling predefineRecursively,
				// so that GetDeclFor resolves each split name to its own individual
				// decl rather than the original multi-value decl, avoiding false
				// cycle detection.
				fn.Decls = append(fn.Decls[:i], append(parts, fn.Decls[i+1:]...)...)
				index.rebindSplitParts(fn, vd, parts)
				for j := range parts {
					if parts[j].GetAttribute(ATTR_PREDEFINED) == true {
						continue
					}
					part := parts[j]
					predefineOne(fn, part, func() { predefineRecursivelyIndexed(store, fn, part, index) })
				}
				i += len(vd.NameExprs) - 1
				continue
			} else {
				// recursively predefine dependencies.
				predefineOne(fn, d, func() { predefineRecursivelyIndexed(store, fn, d, index) })
				continue
			}
		}
	}
}

// predefineDecl identifies a package declaration and its containing file.
type predefineDecl struct {
	file *FileNode
	decl Decl
}

// lookupPredefineDecl finds the declaration binding selected by a FileSet.
func lookupPredefineDecl(fset *FileSet, name Name) predefineDecl {
	file, slot := fset.GetDeclFor(name)
	return predefineDecl{file: file, decl: *slot}
}

// lookupPredefineDeclSafe is lookupPredefineDecl reporting a miss.
func lookupPredefineDeclSafe(fset *FileSet, name Name) (predefineDecl, bool) {
	file, slot, ok := fset.GetDeclForSafe(name)
	if !ok {
		return predefineDecl{}, false
	}
	return predefineDecl{file: file, decl: *slot}, true
}

// undefinedNameNode returns the first mention of name in d, where the
// "not defined" error is positioned; nil when d never spells it.
func undefinedNameNode(gm store.GasMeter, d Decl, name Name) (found Node) {
	Transcribe(d, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
		if nx, ok := n.(*NameExpr); ok && nx.Name == name {
			found = nx
			return n, TRANS_EXIT
		}
		return n, TRANS_CONTINUE
	})
	return found
}

// predefineDeclIndex binds a FileSet to its selected package declarations.
type predefineDeclIndex struct {
	fileSet *FileSet
	entries map[Name]predefineDecl
	acc     *preprocessAccumulator // accumulate mode, else nil
}

// newPredefineDeclIndex captures the package declarations visible to PredefineFileSet.
func newPredefineDeclIndex(fset *FileSet) *predefineDeclIndex {
	index := &predefineDeclIndex{
		fileSet: fset,
		entries: make(map[Name]predefineDecl),
	}
	for i := len(fset.Files) - 1; i >= 0; i-- {
		file := fset.Files[i]
		for _, decl := range file.Decls {
			if _, isImport := decl.(*ImportDecl); isImport {
				continue
			}
			for _, name := range decl.GetDeclNames() {
				if _, exists := index.entries[name]; !exists {
					index.entries[name] = predefineDecl{file: file, decl: decl}
				}
			}
		}
	}
	return index
}

// find resolves a name within the FileSet captured by the index,
// reporting a miss.
func (index *predefineDeclIndex) find(name Name) (predefineDecl, bool) {
	indexed, ok := index.entries[name]
	return indexed, ok
}

// lookup resolves a name within the FileSet captured by the index.
func (index *predefineDeclIndex) lookup(name Name) predefineDecl {
	if indexed, ok := index.entries[name]; ok {
		if debugAssert {
			file, slot, found := index.fileSet.GetDeclForSafe(name)
			if !found || indexed.file != file || indexed.decl != *slot {
				panic(fmt.Sprintf("stale predefine declaration index for %q", name))
			}
		}
		return indexed
	}
	// The index covers the whole FileSet, so a miss can never resolve a decl.
	panic(fmt.Sprintf("name %s not defined in fileset with files %v", name, index.fileSet.FileNames()))
}

// predefineDependency is the accumulate-mode recursion into a dependency.
// A failure is recorded once, against the dependency, which is dropped
// with invalidType names; the dependent goes on against those. One
// cause, one error, at its own declaration.
func (index *predefineDeclIndex) predefineDependency(store Store, dep predefineDecl, stack []Name, defining map[Name]struct{}, direct bool) {
	defer func() {
		if r := recover(); r != nil {
			if isRuntimePanic(r) {
				panic(r)
			}
			index.acc.dropDecl(wrapPreprocessPanic([]BlockNode{dep.file}, dep.decl, r), dep.file, dep.decl)
		}
	}()
	predefineRecursively2(store, dep.file, dep.decl, stack, defining, direct, index)
}

// rebindSplitParts redirects indexed names from an original declaration in file
// to whichever part now declares them, leaving entries won elsewhere untouched.
func (index *predefineDeclIndex) rebindSplitParts(file *FileNode, original Decl, parts []Decl) {
	for _, part := range parts {
		for _, name := range part.GetDeclNames() {
			current := index.entries[name]
			if current.file == file && current.decl == original {
				index.entries[name] = predefineDecl{file: file, decl: part}
			}
		}
	}
}

func initStaticBlocks(store Store, ctx BlockNode, nn Node) {
	initStaticBlocks1(store, ctx, nn)
	initStaticBlocks2(store, ctx, nn)
}

// initStaticBlocks1 renames ForStmt-init declared names with a ".loopvar"
// suffix so they don't collide with same-named decls inside the body of the
// same block.
//
// Algorithm: a single TranscribeB walk maintains a stack of "active loopvar"
// frames (map[Name]bool). The top frame is the set of names currently in
// scope as renameable. Frames are pushed at block-introducing nodes and
// popped at TRANS_LEAVE, so in-place mutations (deleting a shadowed name
// from the top frame on encountering an inner DEFINE/var/type decl) are
// scoped to the enclosing block. A NameExpr is renamed (Name += ".loopvar")
// iff its name is in the top frame and its position is not a declaration
// site (composite-literal key, var-name, range key/value, or assign-LHS in
// a DEFINE).
//
// O(N) total, replacing the previous O(N + Σ K_i × M_i) per-loopvar walk.
func initStaticBlocks1(store Store, ctx BlockNode, nn Node) {
	// Fast path for the common case (no for-loop with multi-name DEFINE
	// init in scope): if the parent frame is empty and `modify` doesn't
	// add anything, skip the push entirely. `didPush` records whether each
	// scope actually pushed a frame so pop can match.
	var stack []map[Name]bool
	var didPush []bool
	top := func() map[Name]bool {
		if len(stack) == 0 {
			return nil
		}
		return stack[len(stack)-1]
	}
	pushClone := func(modify func(map[Name]bool)) {
		parent := top()
		// Fast path: no parent state and no contribution -> skip alloc.
		// We must NOT skip when parent is non-empty even if modify trims
		// the result to empty: the empty frame still shadows parent.
		if modify == nil && len(parent) == 0 {
			didPush = append(didPush, false)
			return
		}
		f := map[Name]bool{}
		maps.Copy(f, parent)
		if modify != nil {
			modify(f)
		}
		stack = append(stack, f)
		didPush = append(didPush, true)
	}
	pop := func() {
		if didPush[len(didPush)-1] {
			stack = stack[:len(stack)-1]
		}
		didPush = didPush[:len(didPush)-1]
	}

	_ = TranscribeB(ctx, nn, func(
		ns []Node,
		bstack []BlockNode,
		last BlockNode,
		ftype TransField,
		index int,
		n Node,
		stage TransStage,
	) (Node, TransCtrl) {
		defer doRecover(bstack, n)
		if debug {
			debug.Printf("initStaticBlocks1 %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		case TRANS_ENTER:
			// FuncLit/FuncDecl push at ENTER (not BLOCK) so param/result/recv
			// masking takes effect before the FuncTypeExpr children are walked.
			switch n := n.(type) {
			case *FuncLitExpr:
				pushClone(func(f map[Name]bool) {
					for i := range n.Type.Params {
						delete(f, n.Type.Params[i].NameExpr.Name)
					}
					for i := range n.Type.Results {
						delete(f, n.Type.Results[i].NameExpr.Name)
					}
				})
			case *FuncDecl:
				pushClone(func(f map[Name]bool) {
					if n.IsMethod {
						delete(f, n.Recv.NameExpr.Name)
					}
					for i := range n.Type.Params {
						delete(f, n.Type.Params[i].NameExpr.Name)
					}
					for i := range n.Type.Results {
						delete(f, n.Type.Results[i].NameExpr.Name)
					}
				})
			}

		case TRANS_BLOCK:
			// Other block-introducing nodes push here. Pairs naturally with
			// TranscribeB's BLOCK→LEAVE flow.
			switch n := n.(type) {
			case *ForStmt:
				pushClone(func(f map[Name]bool) {
					fsinit, ok := n.Init.(*AssignStmt)
					if !ok || fsinit.Op != DEFINE {
						return
					}
					for _, lx := range fsinit.Lhs {
						nx := lx.(*NameExpr)
						ln := nx.Name
						if ln == blankIdentifier {
							continue
						}
						if strings.HasSuffix(string(ln), ".loopvar") {
							continue
						}
						nx.Name += ".loopvar"
						f[ln] = true
					}
				})
			case *RangeStmt:
				pushClone(func(f map[Name]bool) {
					if n.Op != DEFINE {
						return
					}
					if kx, ok := n.Key.(*NameExpr); ok {
						delete(f, kx.Name)
					}
					if vx, ok := n.Value.(*NameExpr); ok {
						delete(f, vx.Name)
					}
				})
			case *SwitchStmt:
				pushClone(func(f map[Name]bool) {
					if n.IsTypeSwitch && n.VarName != "" {
						delete(f, n.VarName)
					}
				})
			case *BlockStmt, *IfStmt, *IfCaseStmt,
				*SwitchClauseStmt, *SelectCaseStmt:
				pushClone(nil)
			}

		case TRANS_LEAVE:
			switch n := n.(type) {
			case *ForStmt, *FuncLitExpr, *FuncDecl, *RangeStmt,
				*BlockStmt, *IfStmt, *IfCaseStmt, *SwitchClauseStmt,
				*SwitchStmt, *SelectCaseStmt:
				pop()
			case *AssignStmt:
				// An inner DEFINE shadows matching loopvar names from this
				// point onward in the enclosing block (until pop).
				if n.Op == DEFINE {
					f := top()
					for _, lx := range n.Lhs {
						if nx, ok := lx.(*NameExpr); ok {
							delete(f, nx.Name)
						}
					}
				}
			case *ValueDecl:
				f := top()
				for i := range n.NameExprs {
					delete(f, n.NameExprs[i].Name)
				}
			case *TypeDecl:
				delete(top(), n.NameExpr.Name)
			case *NameExpr:
				f := top()
				if f == nil || !f[n.Name] {
					return n, TRANS_CONTINUE
				}
				switch ftype {
				case TRANS_VAR_NAME,
					TRANS_RANGE_KEY,
					TRANS_RANGE_VALUE:
					return n, TRANS_CONTINUE
				case TRANS_ASSIGN_LHS:
					if as, ok := ns[len(ns)-1].(*AssignStmt); ok && as.Op == DEFINE {
						return n, TRANS_CONTINUE
					}
				}
				n.Name += ".loopvar"
			}
		}
		return n, TRANS_CONTINUE
	})
}

// Initialize static blocks, and also reserves all names.
// TODO: ensure and keep idempotent.
// PrpedefineFileSet may precede Preprocess.
func initStaticBlocks2(store Store, ctx BlockNode, nn Node) {
	// iterate over all nodes recursively.
	_ = TranscribeB(ctx, nn, func(
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
			debug.Printf("initStaticBlocks2 %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {
			case *AssignStmt:
				if n.Op == DEFINE {
					for i, lx := range n.Lhs {
						nx := lx.(*NameExpr)
						ln := nx.Name
						if ln == blankIdentifier {
							continue
						}
						if !isLocallyReserved(last, ln) {
							// if loopvar, will promote to
							// NameExprTypeHeapDefine later.
							nx.Type = NameExprTypeDefine
							last.Reserve(false, nx, n, NSDefine, i)
						} else {
							nx.Type = NameExprTypeDefine
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
		panic(wrapPreprocessPanic(stack, n, r))
	}
}

// wrapPreprocessPanic turns a panic raised while preprocessing n into a
// *PreprocessError carrying the location and the block stack; a
// *PreprocessError is returned as is.
func wrapPreprocessPanic(stack []BlockNode, n Node, r any) *PreprocessError {
	if pe, ok := r.(*PreprocessError); ok {
		return pe
	}
	last := stack[len(stack)-1]
	loc := last.GetLocation()
	if !n.GetSpan().IsZero() {
		loc.SetSpanOverride(n.GetSpan())
	}
	if sp, ok := r.(*spanError); ok {
		if !sp.span.IsZero() {
			loc.SetSpanOverride(sp.span)
		}
		r = sp.msg
	}
	var err error
	if rerr, ok := r.(error); ok {
		err = errors.Wrap(rerr, loc.String())
	} else {
		err = fmt.Errorf("%s: %v", loc.String(), r)
	}
	return &PreprocessError{err: err, stack: stack}
}

// spanError is a checking error positioned at a node other than the one
// being transcribed; see panicAt.
type spanError struct {
	span Span
	msg  string
}

func (p *spanError) Error() string { return p.msg }

// panicAt raises a preprocess error positioned at the first of at with
// a span; a folded constant borrows its source's.
func panicAt(msg string, at ...Node) {
	for _, x := range at {
		for x != nil && x.GetSpan().IsZero() {
			switch cx := x.(type) {
			case *ConstExpr:
				x = cx.Source
			case *constTypeExpr:
				x = cx.Source
			default:
				x = nil
			}
		}
		if x != nil {
			panic(&spanError{span: x.GetSpan(), msg: msg})
		}
	}
	panic(msg)
}

// preprocessAccumulator collects the errors of one PreprocessCollect run.
// A statement or declaration that fails is dropped from the walk; the
// names it declares get invalidType where their slot is reachable, else
// they are poisoned, so "not declared" follow-ons about them are not
// reported (go/types' invalid-object suppression).
type preprocessAccumulator struct {
	errs     []*PreprocessError
	poisoned map[Name]bool // declared by a dropped statement
	// referenced is keyed by bare name, package-wide: a dropped statement
	// mentioning x hides "declared and not used: x" in every function.
	// Over-suppression; an unused report is never a false positive.
	referenced map[Name]bool // mentioned inside a dropped statement
	// testHook, when set, runs inside the guarded transform before each
	// node's stage; tests use it to raise a panic from within the walk.
	testHook func(n Node, stage TransStage)
}

// Both spellings the preprocessor uses for an unknown name.
var reUndeclaredName = regexp.MustCompile(`name (\S+) not (?:declared|defined)|name not defined: (\S+)`)

// record keeps pe unless it is a follow-on: a dependent of a dropped
// declaration, an unknown poisoned name, a rule tripping over an
// invalidType operand (already reported where it arose), or a message
// already recorded through another path.
func (acc *preprocessAccumulator) record(pe *PreprocessError) {
	for _, name := range pe.dropped {
		acc.poisoned[name] = true
	}
	var dropped *droppedDeclError
	if stderrors.As(pe.err, &dropped) {
		return
	}
	msg := pe.err.Error()
	if m := reUndeclaredName.FindStringSubmatch(msg); m != nil && acc.poisoned[Name(m[1]+m[2])] {
		return
	}
	if strings.Contains(msg, invalidType{}.String()) {
		return
	}
	for _, e := range acc.errs {
		if e.err.Error() == msg {
			return
		}
	}
	acc.errs = append(acc.errs, pe)
}

// dropStmt records pe for the failed statement n, dropped in place in
// block last: each name it declares in last becomes invalidType, so later
// uses are typed and checked; a name out of reach is poisoned instead,
// so a suppressed cascade still covers what it declares.
func (acc *preprocessAccumulator) dropStmt(pe *PreprocessError, n Node, last BlockNode) {
	n.SetAttribute(ATTR_PREPROCESS_DROPPED, true) // later rules skip what it declares
	acc.markReferenced(n)
	for _, name := range declaredNamesOf(n) {
		if !defineInvalid(last.GetStaticBlock(), name) {
			acc.poisoned[name] = true
		}
	}
	acc.record(pe)
}

// dropDecl records pe for the failed file-level declaration d of fn and
// drops it (markDropped). What it declares becomes invalidType in its
// block, so dependents are checked against an invalid object rather than
// failing on a missing name; a declared type keeps its identity with an
// invalid base.
func (acc *preprocessAccumulator) dropDecl(pe *PreprocessError, fn BlockNode, d Decl) {
	acc.markReferenced(d)
	acc.record(pe)
	markDropped(d)
	switch d := d.(type) {
	case *ImportDecl:
		name := d.Name
		if name == "" {
			if exp, ok := expectedPkgName(d.PkgPath); ok {
				name = Name(exp)
			}
		}
		defineInvalid(fn.GetStaticBlock(), name)
	case *TypeDecl:
		sb := skipFile(fn).GetStaticBlock()
		if idx, ok := sb.GetLocalIndex(d.Name); ok {
			if dt, ok := sb.Types[idx].(*DeclaredType); ok && !dt.sealed {
				dt.Base = invalidType{}
			} else if tv, ok := sb.Block.Values[idx].V.(TypeValue); !ok || tv.Type == nil {
				// a type slot never filled (an alias that failed to
				// define): a type value, so a receiver or a field of
				// it evaluates to invalidType instead of crashing.
				sb.Types[idx] = gTypeType
				sb.Block.Values[idx] = asValue(invalidType{})
			}
		}
	default:
		sb := skipFile(fn).GetStaticBlock()
		for _, name := range d.GetDeclNames() {
			if !defineInvalid(sb, name) {
				acc.poisoned[name] = true
			}
		}
	}
}

// checkDuplicates is checkDuplicates (the deploy path's package-level
// rule) for accumulate mode: each redeclaration is reported at the
// declaration that loses and that one is dropped. A method loses to the
// one predefined first (files in order); any other name resolves as
// GetDeclFor does (last file, first declaration), so that one survives.
func (acc *preprocessAccumulator) checkDuplicates(pn *PackageNode, fset *FileSet) {
	type site struct {
		fn *FileNode
		d  Decl
	}
	first := map[Name]site{}
	visit := func(fn *FileNode, d Decl) {
		for _, key := range declDupKeys(d) {
			prev, seen := first[key]
			if !seen {
				first[key] = site{fn, d}
				continue
			}
			shown := key
			if fd, ok := d.(*FuncDecl); ok && fd.IsMethod {
				if nx, ok := destar(fd.Recv.Type).(*NameExpr); ok {
					shown = nx.Name + "." + fd.Name // the key prints its value path
				}
			}
			at := prev.fn.GetLocation()
			at.SetSpanOverride(prev.d.GetSpan())
			msg := fmt.Sprintf("%s redeclared in this block; other declaration at %s", shown, at.String())
			acc.record(wrapPreprocessPanic([]BlockNode{pn, fn}, d, msg))
			markDropped(d)
		}
	}
	isMethod := func(d Decl) bool {
		fd, ok := d.(*FuncDecl)
		return ok && fd.IsMethod
	}
	for _, fn := range fset.Files {
		for _, d := range fn.Decls {
			if isMethod(d) {
				visit(fn, d)
			}
		}
	}
	for i := len(fset.Files) - 1; i >= 0; i-- {
		fn := fset.Files[i]
		for _, d := range fn.Decls {
			if !isMethod(d) {
				visit(fn, d)
			}
		}
	}
}

// defineInvalid gives name, reserved in sb, the type invalidType and
// reports whether it could: a slot already holding a sealed type or a
// value of another type is left alone.
func defineInvalid(sb *StaticBlock, name Name) bool {
	if name == "" || name == blankIdentifier {
		return false
	}
	idx, ok := sb.GetLocalIndex(name)
	if !ok {
		return false
	}
	if old := sb.Types[idx]; old != nil && !isInvalid(old) {
		if _, isFunc := old.(*FuncType); !isFunc {
			return false
		}
		// a function placeholder: callers would check the arity of a
		// signature that never completed.
	}
	sb.Types[idx] = invalidType{}
	sb.Block.Values[idx] = anyValue(invalidType{})
	return true
}

// invalidConstExpr is the placeholder for an expression already reported:
// typed invalidType, so every rule passes it through.
func invalidConstExpr(source Expr) *ConstExpr {
	return toConstExpr(source, anyValue(invalidType{}))
}

// isInvalidExpr reports whether x was typed invalidType, without
// evaluating it: a placeholder, a constant of (or naming) an invalid
// type, or an expression propagated at TRANS_LEAVE.
func isInvalidExpr(x Expr) bool {
	switch x := x.(type) {
	case nil:
		return false
	case *constTypeExpr:
		return isInvalid(x.Type)
	case *ConstExpr:
		if x.T == nil {
			return false
		}
		if tv, ok := x.V.(TypeValue); ok {
			return isInvalid(tv.Type)
		}
		return isInvalid(x.T)
	}
	t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type)
	return ok && isInvalid(t)
}

// isInvalidOperand reports whether a preprocessed operand without a type
// attribute (a selector, an index) has an invalid static type, e.g. a
// field declared with an undefined type.
func isInvalidOperand(store Store, last BlockNode, x Expr) (invalid bool) {
	if x == nil || x.GetAttribute(ATTR_PREPROCESSED) != true {
		return false
	}
	switch x.(type) {
	case *SelectorExpr, *IndexExpr, *StarExpr, *CallExpr:
	default:
		return false
	}
	defer func() {
		if recover() != nil {
			invalid = false
		}
	}()
	return isInvalid(evalStaticTypeOf(store, last, x))
}

// anyInvalidTV reports whether an argument was typed invalidType.
func anyInvalidTV(tvs []TypedValue) bool {
	for _, tv := range tvs {
		if isInvalid(tv.T) {
			return true
		}
	}
	return false
}

// operandsOf lists the sub-expressions whose invalidity makes x invalid
// (go/types: an invalid operand silences the operation).
func operandsOf(x Expr) []Expr {
	switch x := x.(type) {
	case *BinaryExpr:
		return []Expr{x.Left, x.Right}
	case *UnaryExpr:
		return []Expr{x.X}
	case *StarExpr:
		return []Expr{x.X}
	case *RefExpr:
		return []Expr{x.X}
	case *SelectorExpr:
		return []Expr{x.X}
	case *IndexExpr:
		return []Expr{x.X, x.Index}
	case *SliceExpr:
		return []Expr{x.X, x.Low, x.High, x.Max}
	case *TypeAssertExpr:
		return []Expr{x.X, x.Type}
	case *CallExpr:
		return []Expr{x.Func}
	case *CompositeLitExpr:
		return []Expr{x.Type}
	}
	return nil
}

// undefinedName records, at each use in d, that un is declared nowhere
// and replaces those uses by invalidType placeholders, so d keeps being
// predefined (its signature, its dependents). Statements are left to the
// walk, which resolves their names itself. Reports whether a use was found.
func (acc *preprocessAccumulator) undefinedName(store Store, last BlockNode, d Decl, un Name) bool {
	found := false
	typePos := map[Node]bool{} // nodes in type position (a *T of a type is one)
	Transcribe(d, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		if _, ok := n.(Stmt); ok && n != Node(d) {
			return n, TRANS_SKIP
		}
		asType := isTypeField(ftype) || (ftype == TRANS_STAR_X && typePos[ns[len(ns)-1]])
		if asType {
			typePos[n] = true
		}
		// A qualified name whose package is un: the whole selector is
		// the use (tryPredefine reads the package through it).
		if sx, ok := n.(*SelectorExpr); ok {
			if x, ok := sx.X.(*NameExpr); ok && x.Name == un {
				found = true
				acc.record(wrapPreprocessPanic([]BlockNode{last}, sx, fmt.Sprintf("name not defined: %s", un)))
				if asType {
					return toConstTypeExpr(last, sx, invalidType{}), TRANS_CONTINUE
				}
				return invalidConstExpr(sx), TRANS_CONTINUE
			}
		}
		nx, ok := n.(*NameExpr)
		if !ok || nx.Name != un {
			return n, TRANS_CONTINUE
		}
		switch ftype {
		case TRANS_VAR_NAME, TRANS_FIELDTYPE_NAME:
			return n, TRANS_CONTINUE
		case TRANS_COMPOSITE_KEY:
			// a struct key names a field, not a value.
			if isStructLitType(store, last, ns[len(ns)-1].(*CompositeLitExpr).Type) {
				return n, TRANS_CONTINUE
			}
		}
		found = true
		at := Node(nx)
		if nx.GetSpan().IsZero() { // a split part's copied value
			at = d
		}
		acc.record(wrapPreprocessPanic([]BlockNode{last}, at, fmt.Sprintf("name not defined: %s", un)))
		if asType {
			return toConstTypeExpr(last, nx, invalidType{}), TRANS_CONTINUE
		}
		return invalidConstExpr(nx), TRANS_CONTINUE
	})
	return found
}

// isStructLitType reports whether a composite literal's type expression,
// preprocessed or not, denotes a struct (its keys are then field names).
func isStructLitType(store Store, last BlockNode, tx Expr) (isStruct bool) {
	switch tx := tx.(type) {
	case *StructTypeExpr:
		return true
	case *constTypeExpr:
		return tx.Type.Kind() == StructKind
	case *ArrayTypeExpr, *SliceTypeExpr, *MapTypeExpr:
		return false
	}
	defer func() { // the type may not be predefined yet: then not a struct
		if recover() != nil {
			isStruct = false
		}
	}()
	return evalStaticType(store, last, tx).Kind() == StructKind
}

// isTypeField reports whether a child at ftype is a type expression.
func isTypeField(ftype TransField) bool {
	switch ftype {
	case TRANS_COMPOSITE_TYPE, TRANS_TYPE_TYPE, TRANS_VAR_TYPE, TRANS_CONST_TYPE,
		TRANS_TYPEASSERT_TYPE, TRANS_FIELDTYPE_TYPE, TRANS_ARRAYTYPE_ELT,
		TRANS_SLICETYPE_ELT, TRANS_CHANTYPE_VALUE, TRANS_MAPTYPE_KEY, TRANS_MAPTYPE_VALUE:
		return true
	}
	return false
}

// markReferenced notes every name a dropped node mentions: a use the walk
// never sees, so the unused-name rules stay quiet about it.
func (acc *preprocessAccumulator) markReferenced(n Node) {
	if acc.referenced == nil {
		acc.referenced = map[Name]bool{}
	}
	Transcribe(n, func(_ []Node, _ TransField, _ int, sn Node, stage TransStage) (Node, TransCtrl) {
		if nx, ok := sn.(*NameExpr); ok && stage == TRANS_ENTER {
			acc.referenced[nx.Name] = true
		}
		return sn, TRANS_CONTINUE
	})
}

// droppedDeclError is raised, in accumulate mode, by a declaration that
// depends on one dropped without a definition; record drops it quietly.
type droppedDeclError struct{ name Name }

func (e *droppedDeclError) Error() string {
	return fmt.Sprintf("name %s not declared (declaration dropped)", e.name)
}

// markDropped marks a declaration accumulate mode dropped: predefined and
// preprocessed so every loop skips it, dropped so a dependent reaching it
// in tryPredefine is dropped quietly instead of failing.
func markDropped(d Decl) {
	d.SetAttribute(ATTR_PREDEFINED, true)
	d.SetAttribute(ATTR_PREPROCESSED, true)
	d.SetAttribute(ATTR_PREPROCESS_DROPPED, true)
}

func firstDeclName(d Decl) Name {
	if names := d.GetDeclNames(); len(names) > 0 {
		return names[0]
	}
	return ""
}

// isRuntimePanic reports whether r is a Go runtime error, raw or already
// wrapped by doRecover: a VM bug, which accumulate mode never records.
func isRuntimePanic(r any) bool {
	switch r := r.(type) {
	case runtime.Error:
		return true
	case *PreprocessError:
		var re runtime.Error
		return stderrors.As(r.err, &re)
	}
	return false
}

// declaredNamesOf lists the names a statement or declaration introduces.
func declaredNamesOf(n Node) []Name {
	switch n := n.(type) {
	case *AssignStmt:
		if n.Op != DEFINE {
			return nil
		}
		var names []Name
		for _, lx := range n.Lhs {
			if nx, ok := lx.(*NameExpr); ok {
				names = append(names, nx.Name)
			}
		}
		return names
	case *ValueDecl:
		return n.GetDeclNames()
	case *TypeDecl:
		return []Name{n.Name}
	case *FuncDecl:
		if n.IsMethod {
			return nil
		}
		return []Name{n.Name}
	case *DeclStmt:
		var names []Name
		for _, d := range n.Body {
			names = append(names, declaredNamesOf(d)...)
		}
		return names
	}
	return nil
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
	return preprocess(store, ctx, n, nil)
}

// PreprocessCollect is Preprocess for static checking off-chain: instead
// of stopping at the first error, it drops each failing statement or
// file declaration, records its error, and continues. The returned node
// must not be executed. The on-chain path uses Preprocess, which keeps
// the first-error semantics, gas and error text unchanged.
func PreprocessCollect(store Store, ctx BlockNode, n Node) (Node, []*PreprocessError) {
	acc := &preprocessAccumulator{poisoned: map[Name]bool{syntaxErrName: true}, referenced: map[Name]bool{}}
	nn := preprocess(store, ctx, n, acc)
	return nn, acc.errs
}

func preprocess(store Store, ctx BlockNode, n Node, acc *preprocessAccumulator) Node {
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
				if n.HasAttribute(ATTR_PREPROCESS_SKIPPED) {
					n.DelAttribute(ATTR_PREPROCESS_SKIPPED)
				}
				if n.HasAttribute(ATTR_PREPROCESS_INCOMPLETE) {
					n.DelAttribute(ATTR_PREPROCESS_INCOMPLETE)
				}
				return n, TRANS_CONTINUE
			})
	}()

	// Bulk of the preprocessor function
	n = preprocess1(store, ctx, n, acc)

	// XXX check node lines and locations
	checkNodeLinesLocations("XXXpkgPath", "XXXfileName", n)

	// Record package-level initialization order dependencies.
	// Must run before codaPackageSelectors replaces NameExprs.
	if fn, ok := n.(*FileNode); ok {
		codaInitOrderDeps(packageOf(ctx), fn)
	}

	// "coda" means "conclusion".
	// NOTE: need to use Transcribe() here instead of `bn, ok := n.(BlockNode)`
	// because say n may be a *CallExpr containing an anonymous function.
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
				// codaGotoLoopDefines(ctx, bn)
				codaHeapDefinesByUse(ctx, bn)
				codaHeapUsesDemoteDefines(ctx, bn)
				codaPackageSelectors(bn)
				return n, TRANS_SKIP
			}
			return n, TRANS_CONTINUE
		})

	return n
}

func preprocess1(store Store, ctx BlockNode, n Node, acc *preprocessAccumulator) Node {
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
	// In accumulate mode a failed statement is dropped: record the error,
	// rewind the block stack to the statement's enclosing block, go on.
	var onPanic PanicHandler
	if acc != nil {
		onPanic = func(ns []Node, ftype TransField, index int, n Node, r any) (Node, bool) {
			pe, ok := r.(*PreprocessError)
			if !ok || isRuntimePanic(pe) {
				return nil, false // not a checking error: a VM bug, let it surface
			}
			rewound := false
			for i := len(ns) - 1; i >= 0 && !rewound; i-- {
				if bn, ok := ns[i].(BlockNode); ok {
					for k := len(stack) - 1; k > 0; k-- {
						if stack[k] == bn {
							stack = stack[:k+1]
							break
						}
					}
					rewound = true
				}
			}
			if !rewound {
				stack = stack[:1] // every block above ctx was pushed below n
			}
			last = stack[len(stack)-1]
			switch ftype {
			case TRANS_COMPOSITE_KEY, TRANS_COMPOSITE_VALUE:
				// the element alone drops: typed invalid, the literal's
				// rule passes it (a key so typed is removed there).
				acc.record(pe)
				return invalidConstExpr(n.(Expr)), true
			case TRANS_SWITCH_CASE:
				n.SetAttribute(ATTR_PREPROCESS_DROPPED, true) // the switch skips it
			}
			acc.dropStmt(pe, n, last)
			return n, true
		}
	}

	// In accumulate mode declarations predefine through an index carrying
	// the accumulator, so undefined names and failing dependencies resolve
	// to invalidType instead of dropping the dependent; built on first use.
	var accIndex *predefineDeclIndex
	predefineAcc := func(last BlockNode, d Decl) bool {
		if acc == nil {
			return predefineRecursively(store, last, d)
		}
		if accIndex == nil {
			if ctxpn.FileSet == nil {
				return predefineRecursively(store, last, d)
			}
			accIndex = newPredefineDeclIndex(ctxpn.FileSet)
			accIndex.acc = acc
		}
		return predefineRecursivelyIndexed(store, last, d, accIndex)
	}

	// predefineGuarded is predefineRecursively with the same drop-and-record
	// rule for file-level declarations in accumulate mode; a dropped
	// declaration is marked preprocessed so the walk below skips it.
	predefineGuarded := func(last BlockNode, d Decl) {
		if acc == nil {
			predefineRecursively(store, last, d)
			return
		}
		defer func() {
			if r := recover(); r != nil {
				if isRuntimePanic(r) {
					panic(r) // a VM bug, not a checking error
				}
				acc.dropDecl(wrapPreprocessPanic(stack, d, r), last, d)
			}
		}()
		predefineAcc(last, d)
	}

	// report raises a checking error. In accumulate mode it is recorded
	// and the walk goes on, so a node can report more than one.
	report := func(msg string, at ...Node) {
		if acc == nil {
			panicAt(msg, at...)
		}
		func() {
			defer func() {
				acc.record(wrapPreprocessPanic(stack, stack[len(stack)-1], recover()))
			}()
			panicAt(msg, at...)
		}()
	}

	t := func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
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
					checkShortVarDecl(ns, last, n)
				}
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
					preprocessed := predefineAcc(last, d)
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
					// The `cur` naming rules are enforced at TRANS_LEAVE below,
					// on the RESOLVED parameter type. They used to live here and
					// test the SPELLING (`p.Type.(*NameExpr).Name == "realm"`),
					// which an alias walked straight past — see there.
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
				pushInitBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *ForStmt:
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
				case InvalidKind:
					// accumulate mode: reported already; key and value
					// are invalid too.
					if n.Op == DEFINE {
						for _, kv := range []Expr{n.Key, n.Value} {
							if nx, ok := kv.(*NameExpr); ok {
								last.Define(nx.Name, anyValue(xt))
							}
						}
					}
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
				case ArrayKind, SliceKind:
				default:
					panic(fmt.Sprintf(
						"range iteration requires map, string, array, slice, or pointer to array; got %s",
						xt.Kind().String(),
					))
				}
				// key value if define.
				if n.Op == DEFINE && n.Key != nil && n.Value != nil {
					if kn, vn := n.Key.(*NameExpr).Name, n.Value.(*NameExpr).Name; kn == vn && kn != blankIdentifier {
						panic(fmt.Sprintf("%s repeated on left side of :=", kn))
					}
				}
				if n.Op == DEFINE && xt.Kind() != InvalidKind {
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
				// a crossing function can only be declared in a realm
				// (or, as a carve-out, in any *_test.gno file so p/ tests
				// can declare `TestXxx(cur realm, t *testing.T)` to drive
				// migrated methods).
				if ft.IsCrossing() && !crossingAllowed(ctx, n) {
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
									ct = nil
								} else if _, ok := cxx.T.(*TypeType); ok {
									ct = cxx.GetType()
								} else {
									panic(fmt.Sprintf("%s is not a type", constString(cxx)))
								}
							} else {
								ct = evalStaticType(store, last, cx)
							}
							n.Cases[i] = toConstTypeExpr(last, cx, ct)
							// maybe type-switch def.
							if ss.VarName != "" {
								if len(n.Cases) == 1 && ct != nil {
									// If there is only 1 case, the
									// define applies with type.
									// (re-definition).
									last.Define(
										ss.VarName, anyValue(ct))
								} else {
									// If there are 2 or more
									// cases, or the sole case is nil,
									// the type is the tag type.
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
						// Each case is compared to the tag with ==; checked
						// here, before the conversion types an untyped nil.
						assertComparableTo(tt, evalStaticTypeOf(store, last, cx), isNilLiteral(ss.X), isNilLiteral(cx))
						checkOrConvertType(store, last, n, &cx, tt) // #nosec G601
						n.Cases[i] = cx
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncDecl:
				// retrieve cached function type.
				// the type and receiver are already set in predefineRecursively.
				ft := getType(&n.Type).(*FuncType)
				if !n.IsMethod && strings.HasPrefix(string(n.Name), "init.") {
					// The only parameter allowed is the crossing one, `cur realm`.
					if len(n.Type.Results) > 0 ||
						(len(n.Type.Params) > 0 && !(len(n.Type.Params) == 1 && ft.IsCrossing())) {
						panic("func init must have no arguments and no return values")
					}
				}
				// main, like init, may take only the crossing parameter.
				if !n.IsMethod && n.Name == "main" && packageOf(last).PkgName == "main" &&
					(len(n.Type.Results) > 0 ||
						(len(n.Type.Params) > 0 && !(len(n.Type.Params) == 1 && ft.IsCrossing()))) {
					panic("func main must have no arguments and no return values")
				}
				// a crossing function can only be declared in a realm
				// (or, as a carve-out, in any *_test.gno file so p/ tests
				// can declare `TestXxx(cur realm, t *testing.T)` to drive
				// migrated methods).
				if ft.IsCrossing() && !crossingAllowed(ctx, n) {
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
					// Supports out-of-order declarations.
					//
					// Machine entry points call PredefineFileSet
					// first, so anything tryPredefine handled
					// already has ATTR_PREDEFINED set and these
					// loops just skip it. The predefine fallback
					// stays for callers that invoke Preprocess
					// directly.
					//
					// This must run after pushInitBlock above,
					// otherwise it would happen at
					// *FileNode:ENTER.

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
								predefineGuarded(n, d)
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
								predefineGuarded(n, d)
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
								predefineGuarded(n, d)
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
							predefineGuarded(n, d)
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
				if !n.IsTypeSwitch && n.X != nil {
					// The tag must be comparable even with no case to
					// compare it to; checked before the clauses, which
					// would otherwise fail on the comparison first.
					assertSwitchTag(n.X, evalStaticTypeOf(store, last, n.X))
				}
				if n.IsTypeSwitch {
					tt := evalStaticTypeOf(store, last, n.X)
					if _, ok := baseOf(tt).(*InterfaceType); !ok {
						panic(fmt.Sprintf("cannot type switch on non-interface value %s (type %s)",
							exprString(n.X), typeString(tt)))
					}
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

			// Accumulate mode: an expression with an operand already
			// reported is invalid itself, and its rule does not run.
			if acc != nil {
				if x, ok := n.(Expr); ok {
					for _, op := range operandsOf(x) {
						if isInvalidExpr(op) || isInvalidOperand(store, last, op) {
							x.SetAttribute(ATTR_TYPEOF_VALUE, invalidType{})
							return n, TRANS_CONTINUE
						}
					}
				}
			}

			// The main TRANS_LEAVE switch.
			switch n := n.(type) {
			// TRANS_LEAVE -----------------------
			case *NameExpr:
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
						// Struct keys are field names, not variable
						// references, so undo any ".loopvar" rename applied
						// by the earlier loop-var pass (which cannot yet
						// distinguish struct keys from map/array/slice keys
						// referencing an enclosing loop variable).
						fname := strings.TrimSuffix(string(n.Name), ".loopvar")
						n.Name = Name(fname)
						n.Path = bt.GetPathForName(n.Name)
						// Check for unexported fields from external packages.
						if !isUpper(fname) && bt.PkgPath != ctxpn.PkgPath {
							panic(fmt.Sprintf(
								"cannot refer to unexported field %s in struct literal of type %s",
								fname, clt.String()))
						}
						return n, TRANS_CONTINUE
					case *ArrayType, *SliceType:
						fillNameExprPath(last, n, false)
						if last.GetIsConst(store, n.Name) {
							cx := evalConst(store, last, n)
							return cx, TRANS_CONTINUE
						}
						panic("slice/array literals may not contain non-const keys")
					}
				}
				// specific and general cases
				switch n.Name {
				case blankIdentifier:
					n.Path = NewValuePathBlock(0, 0, blankIdentifier)
					return n, TRANS_CONTINUE
				case iotaIdentifier:
					pd := lastDecl(ns)
					valueDecl, ok := pd.(*ValueDecl)
					if !ok || !valueDecl.Const {
						panic("cannot use iota outside constant declaration")
					}

					io := pd.GetAttribute(ATTR_IOTA).(int)
					cx := constUntypedBigint(n, int64(io))
					return cx, TRANS_CONTINUE
				case ".cur":
					// special name that is defined in uverse as undefined,
					// but should not be statically evaluated.
					// (will be replaced TRANS_LEAVE *CallExpr).
					return n, TRANS_CONTINUE
				case ".origin":
					// special name for compiler-synthesized chain-root cross
					// calls (keeper.go's MsgCall dispatch). Only valid as the
					// first arg of a crossing call; at TRANS_LEAVE the outer
					// call substitutes the with-cross AST shape (Args[0]=nil,
					// WithCross=true) so runtime routes through
					// callingCurOrOrigin → buildOriginRealm to mint an
					// EOA-origin cur. The dot prefix makes `.origin`
					// unparseable from user source; only post-parse AST
					// injection (gno.Nx(".origin")) can introduce it.
					if ftype != TRANS_CALL_ARG || index != 0 {
						panic(".origin can only be used as the first argument to a crossing function")
					}
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
						as := ns[len(ns)-1].(*AssignStmt)
						fillNameExprPath(last, n, as.Op == DEFINE)
						return n, TRANS_CONTINUE
					case TRANS_VAR_NAME:
						fillNameExprPath(last, n, true)
						return n, TRANS_CONTINUE
					default:
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
					debug.Printf("TRANS_LEAVE, BinaryExpr, OP: %v, lx: %v, rx: %v, lt: %v, rt: %v, isLeftConstExpr: %v, isRightConstExpr %v, isLeftUntyped: %v, isRightUntyped: %v \n", n.Op, n.Left, n.Right, lt, rt, lic, ric, isUntyped(lt), isUntyped(rt))
				}
				// Shift case.
				if n.Op == SHL || n.Op == SHR {
					n.assertShiftExprCompatible1(store, last, lt, rt)
					if baseOf(rt) != UintType {
						// convert n.Right to (gno) uint type,
						rn := Expr(Call("uint", n.Right))
						// reset/create n2 to preprocess right child.
						n2 := &BinaryExpr{
							Left:  n.Left,
							Op:    n.Op,
							Right: rn,
						}
						// Mark the uint() conversion as a shift RHS so
						// doOpCall can assert non-negative at runtime.
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

				// General cases.
				n.AssertCompatible(store, lt, rt) // check compatibility against binaryExprs other than shift expr
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
								checkOrConvertType(store, last, n, &n.Left, rt)
							} else {
								// convert n.Right to left type.
								checkOrConvertType(store, last, n, &n.Right, lt)
							}
						}
						// Then, evaluate the expression.
						cx := evalConst(store, last, n)
						return cx, TRANS_CONTINUE
					} else if isUntyped(lcx.T) {
						if !isUntyped(rt) { // right is typed
							checkOrConvertType(store, last, n, &n.Left, rt)
						} else if n.Op == EQL || n.Op == NEQ {
							// both untyped.
							// no push context type to == != expr.
							// fall through to default type of left or right.
							dt := defaultTypeOf(rt)
							if shouldSwapOnSpecificity(lt, rt) {
								dt = defaultTypeOf(lt)
							}
							checkOrConvertType(store, last, n, &n.Right, dt)
							checkOrConvertType(store, last, n, &n.Left, dt)
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
						if !isUntyped(lt) { // left is typed non-const, right is untyped const.
							checkOrConvertType(store, last, n, &n.Right, lt)
						} else if n.Op == EQL || n.Op == NEQ { // both untyped.
							// both untyped.
							// no push context type to == != expr.
							// fall through to default type of left or right.
							dt := defaultTypeOf(rt)
							if shouldSwapOnSpecificity(lt, rt) {
								dt = defaultTypeOf(lt)
							}
							checkOrConvertType(store, last, n, &n.Right, dt)
							checkOrConvertType(store, last, n, &n.Left, dt)
						}
					} else if rcx.T == nil { // RHS is nil
						// refer to tests/files/types/eql_0f20.gno
						checkOrConvertType(store, last, n, &n.Right, lt)
					} else if isUntyped(lt) { // left is not const, right is typed const
						checkOrConvertType(store, last, n, &n.Left, rt)
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
						if isWhole(ct) {
							if bd, ok := arg0.TypedValue.V.(BigdecValue); ok {
								if !bd.IsInt() {
									panic(fmt.Sprintf(
										"cannot convert %s to integer type",
										arg0))
								}
							}
							if isNumeric(at) {
								convertConst(store, last, n, arg0, ct)
								constConverted = true
							}
						} else if ct.Kind() == Float32Kind || ct.Kind() == Float64Kind {
							// Convert float-valued and untyped-int constants
							// directly to the target type: constant conversion
							// has no signed zero, so an underflowing constant
							// rounds to +0, unlike the machine's runtime
							// narrowing; and an untyped int may exceed int64
							// (e.g. float64(1<<100)) yet be representable in
							// the float target, so it must not take the
							// default-type (int) path below.
							switch at.Kind() {
							case BigintKind, BigdecKind, Float32Kind, Float64Kind:
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
							case SliceKind, PointerKind, FuncKind, MapKind, InterfaceKind:
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
								mustAssignableTo(store, n, at, ct)
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
						mustAssignableTo(store, n, at, ct)
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
					var fv *FuncValue
					if cx, ok := n.Func.(*ConstExpr); ok {
						fv = cx.GetFunc()
					}
					if fv != nil && fv.PkgPath == uversePkgPath {
						switch fv.Name {
						case "len", "cap":
							if len(n.Args) == 1 {
								if msg := lenCapArgError(fv.Name, evalStaticTypeOf(store, last, n.Args[0])); msg != "" {
									reportOrPanic(acc, stack, n, msg)
								}
							}
						}
						switch fv.Name {
						case "append":
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
						case "copy":
							if len(n.Args) == 2 {
								assertCopyArgs(
									evalStaticTypeOf(store, last, n.Args[0]),
									evalStaticTypeOf(store, last, n.Args[1]))
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
						case "make":
							// Self-contained handling for make() builtin.
							// Validate argument count based on target type,
							// check size arguments are integers, resolve generics,
							// then skip the general case via TRANS_CONTINUE.
							// NOTE: If the general *FuncType call-expr path below changes
							// (e.g. embedded-call expansion, generic Specify semantics, or
							// checkOrConvertType behaviour), this block may need updating.
							ft := bnft
							n.NumArgs = countNumArgs(store, last, n)

							if n.Varg {
								panic("make does not accept variadic spread (...)")
							}
							if len(n.Args) == 0 {
								panic("missing argument to make")
							}

							// Validate arg count per target type.
							tt := evalStaticType(store, last, n.Args[0])
							switch baseOf(tt).(type) {
							case *SliceType:
								if len(n.Args) < 2 || len(n.Args) > 3 {
									panic(fmt.Sprintf(
										"invalid operation: make(%s) expects 2 or 3 arguments; found %d",
										tt, len(n.Args)))
								}
							case *MapType:
								if len(n.Args) > 2 {
									panic(fmt.Sprintf(
										"invalid operation: make(%s) expects 1 or 2 arguments; found %d",
										tt, len(n.Args)))
								}
							case *ChanType:
								panic("channel type is not yet supported")
							default:
								panic(fmt.Sprintf(
									"invalid argument: cannot make %s; type must be slice, map", tt))
							}

							// Reject negative constant size arguments (len, cap, hint).
							// Skip n.Args[0] which is the type argument.
							for _, arg := range n.Args[1:] {
								if cx, ok := arg.(*ConstExpr); ok {
									tv := cx.TypedValue
									if tv.T == nil || !isNumeric(tv.T) {
										panic(fmt.Sprintf(
											"cannot use %v as type int in argument to make", tv))
									}
									if tv.Sign() < 0 {
										panic(fmt.Sprintf(
											"invalid argument: index %v must not be negative", tv))
									}
								}
							}

							// Specify function param/result generics.
							argTVs := evalStaticTypedValues(store, last, n.Args...)
							isVarg := n.Varg
							if acc != nil && anyInvalidTV(argTVs) {
								n.SetAttribute(ATTR_TYPEOF_VALUE, invalidType{})
								return n, TRANS_CONTINUE
							}
							sft := ft.Specify(store, n, argTVs, isVarg)
							spts := sft.Params
							srts := FieldTypeList(sft.Results).Types()

							// Update func attributes with specified types.
							n.Func.SetAttribute(ATTR_TYPEOF_VALUE, sft)
							cx := n.Func.(*ConstExpr)
							fv2 := cx.V.(*FuncValue).Copy(store.GetAllocator())
							fv2.Type = sft
							cx.T = sft
							cx.V = fv2
							n.SetAttribute(ATTR_TYPEOF_VALUE, &tupleType{Elts: srts})

							// Type-check arguments.
							// First arg is the type -- check against resolved param type.
							checkOrConvertType(store, last, n, &n.Args[0], spts[0].Type)

							// make's variadic params are declared as Vrd(AnyT()), so untyped
							// constants won't be automatically coerced to int; enforce it here.
							for i := 1; i < len(n.Args); i++ {
								expectedType := spts[len(spts)-1].Type.Elem()
								at := evalStaticTypeOf(store, last, n.Args[i])
								switch {
								case isUntyped(at):
									expectedType = IntType
								case !isInteger(at):
									panic(fmt.Sprintf(
										"invalid argument: index %v (variable of type %v) must be integer",
										n.Args[i], at))
								}
								checkOrConvertType(store, last, n, &n.Args[i], expectedType)
							}

							// For slices with 3 args, check len <= cap when both are constants.
							if _, ok := baseOf(tt).(*SliceType); ok && len(n.Args) == 3 {
								lcx, lOk := n.Args[1].(*ConstExpr)
								ccx, cOk := n.Args[2].(*ConstExpr)
								if lOk && cOk {
									if lcx.TypedValue.GetInt() > ccx.TypedValue.GetInt() {
										panic(fmt.Sprintf(
											"invalid argument: len larger than cap in make(%s)", tt))
									}
								}
							}

							return n, TRANS_CONTINUE
						case "cross":
							// cross(rlm) — the explicit cross-call form.
							// Three constraints, all enforced here:
							//
							//  (a) The argument must be a bare NameExpr (a
							//      realm-typed identifier in scope), not an
							//      arbitrary expression. The Go typechecker
							//      shim narrows to realm type via the
							//      gnobuiltins signature.
							//
							//  (b) cross must appear at Args[0] of a parent
							//      CallExpr (ftype == TRANS_CALL_ARG, index==0).
							//      Statement-level (`x := cross(rlm)`),
							//      non-Args[0] positions (`f(rlm, cross(rlm))`),
							//      and ExprStmt usage are all rejected.
							//
							//  (c) The parent CallExpr's Func must resolve to
							//      a crossing function. cross at Args[0] of
							//      a non-crossing call is rejected.
							//
							// Together these guarantee cross only appears where
							// the outer crossing-call's LEAVE handler will pick
							// it up — preventing the "virtual function" panic
							// or the silent-identity-return paths.
							if len(n.Args) != 1 {
								panic("cross takes exactly one argument: the in-scope realm")
							}
							if _, ok := n.Args[0].(*NameExpr); !ok {
								panic("cross argument must be a bare realm-typed identifier (a name, not an expression)")
							}
							if ftype != TRANS_CALL_ARG || index != 0 {
								panic("cross(rlm) can only be used as the first argument to a crossing-function call")
							}
							pc, ok := ns[len(ns)-1].(*CallExpr)
							if !ok {
								// ftype guarantees parent is a CallExpr; sanity.
								panic("cross(rlm): internal — parent is not a CallExpr")
							}
							// baseOf unwraps DeclaredType to its underlying *FuncType,
							// so named function types (`type Fn func(realm) error`)
							// are recognized as crossing-shape calls.
							pft, _ := baseOf(evalStaticTypeOf(store, last, pc.Func)).(*FuncType)
							if pft == nil || !pft.IsCrossing() {
								panic("cross(rlm) can only be used as the first argument to a crossing-function call")
							}
						case "crossing":
							panic("crossing() is reserved and deprecated")
						case "attach":
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
								panic(fmt.Sprintf("missing realm argument in calling crossing function call %v (expected cur or cross(rlm))", n))
							}

							// cross(rlm) appears here as a *CallExpr at Args[0]:
							// the inner CallExpr's TRANS_LEAVE (above) has already
							// validated the shape (exactly one NameExpr arg).
							// Leave Args[0] in place — at runtime the inner cross
							// native body validates IsCurrent-strict on rlm and
							// pushes it back unchanged. installCrossingCur peeks
							// that value and uses it as the new cur's prev.
							if inner, ok := n.Args[0].(*CallExpr); ok {
								innerFunc, fnOK := inner.Func.(*ConstExpr)
								if !fnOK || innerFunc.GetFunc() == nil ||
									innerFunc.GetFunc().PkgPath != uversePkgPath ||
									innerFunc.GetFunc().Name != "cross" {
									panic(fmt.Sprintf("only `cur` or `cross(rlm)` allowed as first arg to a crossing function; got call to %v", inner.Func))
								}
								// cross(rlm) form. The inner CallExpr's shape
								// was already validated by the inner TRANS_LEAVE
								// (one bare NameExpr arg). No path resolution
								// needed; the normal eval machinery handles
								// whatever scope rlm comes from (function param,
								// closure capture, threaded helper, etc.).
								n.SetWithCross()
								goto LEAVE_CALL_EXPR_END_CHECK_CROSSING
							}

							nx, ok := n.Args[0].(*NameExpr)
							if !ok || nx.Name != Name("cur") && nx.Name != Name(".cur") && nx.Name != Name(".origin") {
								panic(fmt.Sprintf("only `cur` or `cross(rlm)` are allowed as the first argument to a crossing function but got %s", n.Args[0]))
							}
							switch nx.Name {
							case Name(".origin"):
								// compiler-internal sentinel for chain-root MsgCall
								// synthesis. Lowers to with-cross AST shape so the
								// runtime path (installCrossingCur →
								// callingCurOrOrigin → buildOriginRealm) mints an
								// EOA-origin cur.
								n.SetWithCross()
								n.Args[0] = constNil(nx)
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
								// PackageNode top — no cross ancestor.
								crlm := NewConcreteRealm(store.GetAllocator(), ctxpn.PkgPath, gOriginRealmTV)
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
								} else if fv := ftv.GetUnboundFunc(); fv != nil {
									// fv == nil: typed-nil crossing func (e.g.
									// `var f func(cur realm); f(cur)`) or a
									// lazy interface bind (no concrete func
									// until call time); fall through, the
									// runtime check covers both.
									if fv.PkgPath != ctxpn.PkgPath {
										panic(fmt.Sprintf("cannot cur-call to external realm function %s.%v from %v",
											fv.PkgPath, n.Func, ctxpn.PkgPath))
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
								// NOTE: closure-captured cur is allowed here. With pointer-
								// receiver .grealm methods preserving HIV through dispatch
								// (see DefineNativePtrMethod), and persistence refusing realm
								// values (refusePersistRealmHIV), a captured cur cannot
								// outlive its originating frame across txs. Within a single
								// tx, the captured cur is a valid handle to its frame's
								// authority — runtime semantics of pkg.Fn(cur, ...) flow
								// the captured value through installCrossingCur as the
								// new cur's prev.
							default:
								panic("should not happen")
							} // END Check validity of crossing arg n.Args[0].(*NameExpr).
						}
					LEAVE_CALL_EXPR_END_CHECK_CROSSING:
						// Reject crossing calls inside `defer`. doOpReturnCallDefers
						// (op_call.go) doesn't go through doOpPrecall, so
						// installCrossingCur never mints a fresh cur for the
						// deferred callee — the Args[0] placeholder stays nil and
						// the callee crashes on first use of cur. Wrapping the
						// cross-call in a closure deferred at the closure call
						// site is the workaround:
						//   defer func() { f(cross, args...) }()
						// The closure call is non-crossing; inside the closure
						// body, the cross-call goes through the normal precall
						// path and mints the fresh cur correctly.
						if n.WithCross && ftype == TRANS_DEFER_CALL {
							panic("crossing call `cross(rlm)` cannot be deferred directly — the deferred-call dispatch path doesn't mint a fresh cur. Wrap in a closure: `defer func() { fn(cross(rlm), args...) }()`")
						}
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
					if acc != nil && anyInvalidTV(argTVs) {
						// arity was checked; the arguments were reported.
						n.SetAttribute(ATTR_TYPEOF_VALUE, invalidType{})
						return n, TRANS_CONTINUE
					}
					sft := ft.Specify(store, n, argTVs, isVarg)
					spts := sft.Params
					srts := FieldTypeList(sft.Results).Types()
					// If generics were specified, override attr
					// and constexpr with specified types.  Also
					// copy the function value with updated type.
					n.Func.SetAttribute(ATTR_TYPEOF_VALUE, sft)
					if cx, ok := n.Func.(*ConstExpr); ok {
						switch fv := cx.V.(type) {
						case nil:
							// typed-nil func: nothing to specialize;
							// runtime nil-panics on call.
						case *FuncValue:
							fv2 := fv.Copy(store.GetAllocator())
							fv2.Type = sft
							cx.T = sft
							cx.V = fv2
						default:
							panic(fmt.Sprintf("unexpected const func value %T", cx.V))
						}
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
									mustAssignableTo(store, n, tv.T, spts[len(spts)-1].Type.Elem())
								} else {
									mustAssignableTo(store, n, tv.T, spts[i].Type)
								}
							} else {
								mustAssignableTo(store, n, tv.T, spts[i].Type)
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
										checkOrConvertType(store, last, n, &n.Args[i], spts[len(spts)-1].Type.Elem())
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
				assertIndexable(n.X, dt)
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
					checkOrConvertIndexKind(store, last, n, n.Index)
					assertConstIndex(n.Index, staticLengthOf(dt, n.X))
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
				checkOrConvertIndexKind(store, last, n, n.Low)
				checkOrConvertIndexKind(store, last, n, n.High)
				checkOrConvertIndexKind(store, last, n, n.Max)

				t := evalStaticTypeOf(store, last, n.X)
				switch t.Kind() {
				case PointerKind:
					if t.Elem().Kind() != ArrayKind {
						panic(fmt.Sprintf("cannot slice variable of type %v", t))
					}
				case StringKind:
					if n.Max != nil {
						panic("invalid operation: 3-index slice of string")
					}
				case SliceKind, ArrayKind:
					// good.
				default:
					panic(fmt.Sprintf("cannot slice variable of type %v", t))
				}
				assertConstSliceBounds(n, staticLengthOf(t, n.X))

				// if n.X is untyped, convert to corresponding type
				if isUntyped(t) {
					dt := defaultTypeOf(t)
					checkOrConvertType(store, last, n, &n.X, dt)
				}

			// TRANS_LEAVE -----------------------
			case *TypeAssertExpr:
				if n.Type == nil {
					// x.(type) is only a type switch guard; the parser
					// accepts it anywhere (go/types reports it here too).
					panic("use of .(type) outside type switch")
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
				if at := evalStaticType(store, last, n.Type); at != nil {
					if err := dynamicTypeError(preprocessGasMeterOf(store), t, at); err != nil {
						panic(fmt.Sprintf("impossible type assertion: %v does not implement %v (%v)", at, t, err))
					}
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
				dropInvalidKeys := func() {
					// accumulate mode: an element whose key dropped goes
					// (a dropped value stays, typed invalid).
					n.Elts = slices.DeleteFunc(n.Elts, func(kv KeyValueExpr) bool {
						return isInvalidExpr(kv.Key)
					})
				}
				if acc != nil {
					dropInvalidKeys()
				}
				// checkElt is checkOrConvertType (convertType with convert)
				// for one element: in accumulate mode its failure is
				// recorded at the element, which becomes invalid, so the
				// siblings are still checked.
				checkElt := func(x *Expr, t Type, convert bool) {
					check := checkOrConvertType
					if convert {
						check = convertType
					}
					defer func() {
						r := recover()
						if r == nil {
							return
						}
						if isRuntimePanic(r) {
							panic(r)
						}
						if acc == nil {
							switch r.(type) {
							case *PreprocessError, *spanError: // positioned already
								panic(r)
							}
							panicAt(fmt.Sprint(r), unconst(*x)) // gc reports at the element
						}
						acc.record(wrapPreprocessPanic(stack, unconst(*x), r)) // a const keeps its source's span
						*x = invalidConstExpr(*x)
					}()
					check(store, last, n, x, t)
				}
				// Get or evaluate composite type.
				clt := evalStaticType(store, last, n.Type)
				n.Type = toConstTypeExpr(last, n.Type, clt)
				// Replace const Elts with default *ConstExpr. An element
				// that fails is reported at the element (checkElt).
				switch cclt := baseOf(clt).(type) {
				case *StructType:
					if n.IsKeyed() {
						// NOTE: unexported field check for keyed literals
						// is handled in TRANS_COMPOSITE_KEY (*NameExpr).
						seen := map[Name]bool{}
						for i := range n.Elts {
							key := n.Elts[i].Key.(*NameExpr).Name
							if seen[key] {
								panicAt(fmt.Sprintf("duplicate field name %s in struct literal", key), n.Elts[i].Key)
							}
							seen[key] = true
							path := cclt.GetPathForName(key)
							ft := cclt.GetStaticTypeOfAt(path)
							checkElt(&n.Elts[i].Value, ft, false)
						}
					} else {
						// Check for unexported fields in unkeyed struct literals
						// from external packages (implicit assignment).
						// Empty struct literals (T{}) are always allowed, even
						// from external packages, as they are zero-value initializations.
						if len(n.Elts) > 0 &&
							cclt.hasInaccessibleUnexportedFields(ctxpn.PkgPath) {
							panic(fmt.Sprintf(
								"implicit assignment to unexported field in struct literal of type %s",
								clt.String()))
						}
						for i := range n.Elts {
							ft := cclt.Fields[i].Type
							checkElt(&n.Elts[i].Value, ft, false)
						}
					}
				case *ArrayType:
					for i := range n.Elts {
						if cx, ok := n.Elts[i].Key.(*ConstExpr); ok && cx.TypedValue.Sign() < 0 {
							panicAt(fmt.Sprintf("invalid argument: index must not be negative: %v", cx.TypedValue), cx)
						}
						checkElt(&n.Elts[i].Key, IntType, true)
						checkElt(&n.Elts[i].Value, cclt.Elt, false)
					}
					length := int64(-1) // [...]T is measured below
					if !cclt.Vrd {
						length = int64(cclt.Len)
					}
					assertLiteralIndices(preprocessGasMeterOf(store), n, length)
				case *SliceType:
					for i := range n.Elts {
						if cx, ok := n.Elts[i].Key.(*ConstExpr); ok && cx.TypedValue.Sign() < 0 {
							panicAt(fmt.Sprintf("invalid argument: index must not be negative: %v", cx.TypedValue), cx)
						}
						checkElt(&n.Elts[i].Key, IntType, true)
						checkElt(&n.Elts[i].Value, cclt.Elt, false)
					}
					assertLiteralIndices(preprocessGasMeterOf(store), n, -1)
				case *MapType:
					// kset tracks seen const keys for duplicate detection.
					// checkOrConvertType must be called before the check so that
					// values are stored in N (not V), making TypedValue comparable.
					kset := make(map[TypedValue]struct{})
					for i := range n.Elts {
						if n.Elts[i].Key == nil {
							report("missing key in map literal", n.Elts[i].Value)
							n.Elts[i].Key = invalidConstExpr(n.Elts[i].Value) // dropped below
							continue
						}
						checkElt(&n.Elts[i].Key, cclt.Key, false)
						checkElt(&n.Elts[i].Value, cclt.Value, false)
						elt := n.Elts[i]
						if cx, ok := elt.Key.(*ConstExpr); ok && !cx.TypedValue.IsUndefined() && !isInvalidExpr(cx) && !isInterfaceConversion(cx) {
							if _, ok := kset[cx.TypedValue]; ok {
								report(fmt.Sprintf("duplicate key %s in map literal", constString(cx)), cx)
							}
							kset[cx.TypedValue] = struct{}{}
						}
					}
				default:
					panic(fmt.Sprintf(
						"unexpected composite type %s",
						clt.String()))
				}
				if acc != nil {
					dropInvalidKeys() // a key that failed its own check
				}
				// If variadic array lit, measure.
				if at, ok := clt.(*ArrayType); ok {
					if at.Vrd {
						idx := int64(0)
						for _, elt := range n.Elts {
							// The implied length is the largest index + 1; an index of
							// MaxInt64 overflows that to MaxInt64+1. Reject it as out of
							// bounds (matching Go, whose largest valid index here is
							// MaxInt64-1) so idx can't wrap negative past
							// checkArrayAllocFits below.
							if elt.Key == nil {
								if idx == math.MaxInt64 {
									panic(fmt.Sprintf("array index %d out of bounds [0:0]", idx))
								}
								idx++
							} else {
								k := evalConst(store, last, elt.Key).ConvertGetInt()
								if idx > k {
									panic("array lit key out of order")
								}
								if k == math.MaxInt64 {
									panic(fmt.Sprintf("array index %d out of bounds [0:0]", k))
								}
								idx = k + 1
							}
						}
						// update type
						// (dontcare)
						// at.Vrd = false
						// Reject an oversized [...]T{idx: v} array, whose length
						// is only known here, at compile time (see checkArrayAllocFits).
						checkArrayAllocFits(at.Elt, idx)
						at.Len = int(idx)
						// Mutating Len invalidates the cached typeid.
						at.typeid = ""
						// update node
						cx := constInt(n, idx)
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
			case *RefExpr:
				// Cache the static type of X as ATTR_REF_ELEM_TYPE
				// on the RefExpr so doOpRef can use it (instead of
				// the runtime type, which is wrong for interface variables).
				xt := evalStaticTypeOf(store, last, n.X)
				// Refuse a pointer to a realm-typed `cur`.
				//
				// Rejecting `cur = ...` (AssignStmt, below) only matches a
				// *NameExpr on the left. `p := &cur; *p = cur.Previous()`
				// writes the same slot through a StarExpr and does not match,
				// so it preprocesses cleanly and reaches the callee's frame
				// Cur, leaving only doOpCall's runtime check to notice — at
				// call time rather than at compile time. Enumerating
				// assignment shapes is the losing end of this: a slot that
				// cannot be addressed cannot be written through any of them,
				// so the address is what gets refused. This also covers `&cur`
				// escaping into a helper that assigns through it, which no LHS
				// rule would see.
				if ne, ok := n.X.(*NameExpr); ok && ne.Name == "cur" {
					if xt == gRealmType {
						panic("cannot take the address of a realm-typed `cur`: the binding is fixed for the life of the frame, and a pointer to it is a way to rebind it. `realm` is an interface, so a *realm is never needed — pass `cur` by value")
					}
				}
				if tt, ok := xt.(*tupleType); ok {
					panic(fmt.Sprintf(
						"cannot take address of multi-value call (results: %s)",
						tt.String()))
				}
				n.SetAttribute(ATTR_REF_ELEM_TYPE, xt)
			// TRANS_LEAVE -----------------------
			case *SelectorExpr:
				xt := evalStaticTypeOf(store, last, n.X)

				// Set selector path based on xt's type.
				switch cxt := xt.(type) {
				case *PointerType, *DeclaredType, *StructType, *InterfaceType:
					tr, _, rcvr, _, status := findEmbeddedFieldType(preprocessGasMeterOf(store), ctxpn.PkgPath, cxt, n.Sel)
					switch status {
					case embedLookupAccessError:
						panic(fmt.Sprintf("cannot access %s.%s from %s",
							cxt.String(), n.Sel, ctxpn.PkgPath))
					case embedLookupAmbiguous:
						panic(fmt.Sprintf("ambiguous selector %s in %s",
							n.Sel, cxt.String()))
					case embedLookupNone:
						panic(fmt.Sprintf("missing field %s in %s",
							n.Sel, cxt.String()))
					}
					// embedLookupFound guarantees tr != nil below.

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
						rx := &RefExpr{X: n.X}
						rx.SetAttribute(ATTR_REF_ELEM_TYPE, nxt2)
						n.X = rx
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
						rx := &RefExpr{X: n.X}
						rx.SetAttribute(ATTR_REF_ELEM_TYPE, nxt2)
						n.X = rx
						setPreprocessed(n.X)
					}
					// bound method or underlying.
					// NOTE: unexported field access is already checked
					// by findEmbeddedFieldType above (status).
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
					var dt *DeclaredType
					switch ct := xt.(type) {
					case *PointerType:
						dt = ct.Elt.(*DeclaredType)
					case *DeclaredType:
						dt = ct
					default:
						panic(fmt.Sprintf(
							"unexpected selector expression type value %s",
							xt.String()))
					}
					// Ensure exposed or package path match. Without this gate,
					// `(*pkg.Type).unexportedMethod` from outside pkg resolves
					// successfully and yields a callable *FuncValue, since
					// GetUnboundPathForName does not enforce visibility (cf.
					// FindEmbeddedFieldType which does, at types.go:1685).
					// A confused-deputy callback in pkg can then invoke the
					// unexported method on pkg's own state.
					if !(isUpper(string(n.Sel)) || ctxpn.PkgPath == dt.PkgPath) {
						panic(fmt.Sprintf("cannot access unexported method %s.%s from %s",
							dt.PkgPath, n.Sel, ctxpn.PkgPath))
					}
					n.Path = dt.GetUnboundPathForName(n.Sel)
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
					// A root node is a literal's type preprocessed on its own.
					if len(ns) > 0 {
						if _, ok := ns[len(ns)-1].(*CompositeLitExpr); !ok {
							panic("invalid use of [...] array (outside a composite literal)")
						}
					}
				} else {
					// Replace const Len with int *ConstExpr.
					cx := evalConst(store, last, n.Len)
					convertConst(store, last, n, cx, IntType)
					assertArrayLength(cx.GetInt())
					n.Len = cx
					// Reject an oversized fixed-size array at compile time (see
					// checkArrayAllocFits). This covers [N]T and [N]T{...}; the
					// [...]T{idx: v} form is measured and checked elsewhere.
					checkArrayAllocFits(evalStaticType(store, last, n.Elt), cx.GetInt())
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
				checkInterfaceElems(store, last, n)
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *FuncTypeExpr:
				ftype := evalStaticType(store, last, n)
				// The `cur` naming rules decide on the RESOLVED type, because
				// that is what decides whether the function is crossing.
				//
				// They used to run at TRANS_ENTER against the SPELLING of the
				// type expression — `p.Type.(*NameExpr).Name == "realm"` — while
				// IsCrossing() compares the resolved type to gRealmType. A type
				// alias fits exactly in that gap: `type R = realm` leaves the
				// function crossing while neither naming rule fires, so it could
				// take a SECOND realm parameter, ordinary and caller-supplied,
				// literally named `cur`. That parameter then passes the cur-call
				// provenance check (it is declared by a crossing FuncDecl) and is
				// forwarded by bare name into a callee's frame identity, with no
				// assignment and no address taken anywhere for the write-shape
				// rules to catch.
				//
				// Deciding on the resolved type is what keeps the two answers
				// from drifting apart again: this asks the same question
				// IsCrossing() asks, of the same value.
				if ft, ok := ftype.(*FuncType); ok {
					for i := range ft.Params {
						if ft.Params[i].Type != gRealmType {
							continue
						}
						pn := ft.Params[i].Name
						if i == 0 {
							// XXX refactor .arg stuff; see isUnnamedResult.
							if pn != "cur" && !strings.HasPrefix(string(pn), ".arg") {
								panic("a crossing function's first realm argument must have name `cur`")
							}
						} else if pn == "cur" {
							panic("only the first realm type argument of a crossing function may have name `cur`")
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *MapTypeExpr:
				if msg := mapKeyError(preprocessGasMeterOf(store), evalStaticType(store, last, n).(*MapType).Key); msg != "" {
					panicAt(msg, n.Key, n)
				}

			// TRANS_LEAVE -----------------------
			case *StructTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *AssignStmt:
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

				// The crossing `cur` parameter is a fixed binding, not a
				// rebindable local. It names the realm the frame is executing
				// as, established at entry and constant for the life of the
				// frame; rebinding the name makes it describe a realm the frame
				// is not in.
				//
				// `cur` is reserved as a PARAMETER name — the preprocessor lets
				// only a crossing function's first realm parameter use it (see
				// FuncTypeExpr above) — but a local may still shadow it in an
				// inner block (`cur := cur.Previous()`), so a realm-typed `cur`
				// on the left of `=` is USUALLY, not always, that parameter.
				// Rejecting both is deliberate: a shadowing local cannot be
				// cur-called anyway (the provenance check below resolves the
				// name's declaring block node and requires a crossing
				// FuncDecl/FuncLitExpr), so refusing it costs nothing and avoids
				// reasoning about which one this is.
				//
				// It matters because the parameter is the one realm value a
				// no-cross crossing call can forward (the cur-call check below
				// requires the bare name `cur`): a rebound `cur` passed down
				// becomes the callee's frame Cur, and from there IsCurrent() and
				// cross(rlm) both agree with it. Storing a realm value into some
				// other variable is unaffected — it cannot be cur-called and
				// keeps its own persist-time guard — so only `cur` is rejected
				// here. doOpCall carries the matching runtime check.
				if n.Op != DEFINE {
					for _, lh := range n.Lhs {
						ne, ok := lh.(*NameExpr)
						if !ok || ne.Name != "cur" {
							continue
						}
						if evalStaticTypeOf(store, last, ne) == gRealmType {
							panic("cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame")
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
			case *FuncDecl:
				// Unused locals and labels, once the body has its paths.
				if n.Body != nil {
					reportUnused(store, acc, stack, n)
				}

			// TRANS_LEAVE -----------------------
			case *FuncLitExpr:
				// A package-level literal (`var f = func() {...}`) has no
				// FuncDecl to walk it; one nested in a function is walked
				// by that function's check.
				if !enclosedByFunc(stack[:len(stack)-1]) {
					reportUnused(store, acc, stack, n)
				}

			case *IncDecStmt:
				n.AssertCompatible(store, last)

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
				// `for _, cur = range rs` writes the same parameter slot as
				// `cur = rs[0]` but is a RangeStmt, not an AssignStmt, so the
				// rejection in the AssignStmt handler never sees it and no
				// address is taken for the RefExpr handler to refuse. It
				// preprocessed cleanly and reached the callee's frame Cur,
				// noticed only by doOpCall's runtime check at call time. Found
				// by review AFTER the RefExpr rejection was added, which is the
				// whole argument for keeping that runtime check: this is the
				// third write shape into one slot, and the second one the
				// syntactic rules missed.
				// DEFINE (`for _, cur := range`) is not a write to the
				// parameter — it binds a new name, which cannot be cur-called
				// (the provenance check below refuses a name whose declaring
				// block is not a crossing FuncDecl/FuncLitExpr).
				if n.Op != DEFINE {
					for _, lh := range []Expr{n.Key, n.Value} {
						ne, ok := lh.(*NameExpr)
						if !ok || ne.Name != "cur" {
							continue
						}
						if evalStaticTypeOf(store, last, ne) == gRealmType {
							panic("cannot assign to a realm-typed `cur` in a range clause: it names the realm this frame is executing as, and that binding is fixed for the life of the frame")
						}
					}
				}

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
						if cx, ok := n.Results[0].(*CallExpr); ok && isInvalidExpr(cx) {
							// already reported.
						} else if ok {
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
			case *SwitchStmt:
				assertSingleDefault(n.Clauses)
				if !n.IsTypeSwitch {
					for _, dup := range duplicateSwitchCases(store, n.Clauses) {
						report(dup.msg, dup.at...)
					}
				}
				// Ensure type switch cases are unique and possible.
				if n.IsTypeSwitch {
					live := n.Clauses
					if acc != nil { // a dropped clause has no checked cases
						live = slices.DeleteFunc(slices.Clone(n.Clauses), func(c SwitchClauseStmt) bool {
							return c.GetAttribute(ATTR_PREPROCESS_DROPPED) == true
						})
					}
					for _, dup := range duplicateTypeSwitchCases(preprocessGasMeterOf(store), live) {
						report(dup.msg, dup.at...)
					}
					tt := evalStaticTypeOf(store, last, n.X)
					for i := range n.Clauses {
						if acc != nil && n.Clauses[i].GetAttribute(ATTR_PREPROCESS_DROPPED) == true {
							continue
						}
						for _, casetype := range n.Clauses[i].Cases {
							ctype := casetype.(*constTypeExpr).Type
							if ctype == nil {
								continue
							}
							if err := dynamicTypeError(preprocessGasMeterOf(store), tt, ctype); err != nil {
								panicAt(fmt.Sprintf(
									"impossible type switch case: %s cannot have dynamic type %v (%v)",
									exprString(n.X), ctype, err), casetype, &n.Clauses[i])
							}
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
						reflect.TypeFor[*TypedValue]()))
				}
				// We need to replace all references of the new
				// Type with old Type, including in attributes.
				n.Type.SetAttribute(ATTR_TYPE_VALUE, dstT)
				// Replace the type with *{},
				// otherwise methods would be un at runtime.
				if dt, ok := dstT.(*DeclaredType); ok && !n.IsAlias {
					checkTypeExprShape(preprocessGasMeterOf(store), n.Type)
					checkTypeShape(preprocessGasMeterOf(store), last, dt)
				}
				n.Type = toConstTypeExpr(last, n.Type, dstT)
			}
			// end type switch statement
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}

		panic(fmt.Sprintf(
			"unknown stage %v", stage))
	}
	if acc != nil && acc.testHook != nil {
		inner := t
		t = func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			defer doRecover(stack, n)
			acc.testHook(n, stage)
			return inner(ns, ftype, index, n, stage)
		}
	}
	nn := transcribeWith(n, t, onPanic)

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
	if isConst && numVals == 0 && numNames > 0 && typeExpr == nil {
		// This occurs when a const group line inherits values from the previous
		// line but the number of names doesn't match (go2gno sets Values to nil
		// in that case). Report a proper error instead of a confusing internal panic.
		panic(fmt.Sprintf("assignment mismatch: %d variable(s) but 0 value(s)", numNames))
	}

	sts := make([]Type, numNames) // static types
	tvs := make([]TypedValue, numNames)

	if numVals == 1 && numNames > 1 && isInvalidExpr(valueExprs[0]) {
		// accumulate mode: the value was reported; the names are invalid.
		for i := range sts {
			sts[i] = invalidType{}
			tvs[i] = anyValue(invalidType{})
		}
	} else if numVals == 1 && numNames > 1 {
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
			if err := checkAssignableTo(store, n, tt, st); err != nil {
				if debug {
					debug.Printf("checkAssignableTo fail: %v\n", err)
				}
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

// addAttrHeapUse adds name to bn's heap-use set. ATTR_HEAP_USES holds a
// map[Name]struct{} (was []Name pre-fix); membership and insertion are
// O(1). DEGEN3 DeepClosureCapture: K refs × D closures × O(1) = O(K*D).
func addAttrHeapUse(bn BlockNode, name Name) {
	set, _ := bn.GetAttribute(ATTR_HEAP_USES).(map[Name]struct{})
	if _, ok := set[name]; ok {
		return
	}
	if set == nil {
		set = map[Name]struct{}{}
		bn.SetAttribute(ATTR_HEAP_USES, set)
	}
	set[name] = struct{}{}
}

func hasAttrHeapUse(bn BlockNode, name Name) bool {
	set, _ := bn.GetAttribute(ATTR_HEAP_USES).(map[Name]struct{})
	_, ok := set[name]
	return ok
}

// adds ~name to func lit static block and to heap captures atomically.
func addHeapCapture(dbn BlockNode, fle *FuncLitExpr, depth int, nx *NameExpr) (idx uint16) {
	if depth <= 0 {
		panic("invalid depth")
	}
	name := nx.Name
	if _, exists := fle.heapCapturesIdx[name]; exists {
		// assert ~name also already defined.
		var ok bool
		idx, ok = fle.GetLocalIndex("~" + name)
		if !ok {
			panic("~name not added to fle atomically")
		}
		return // already exists
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
	if fle.heapCapturesIdx == nil {
		fle.heapCapturesIdx = map[Name]uint16{}
	}
	fle.heapCapturesIdx[name] = uint16(len(fle.HeapCaptures) - 1)

	// find index after define (O(1) via StaticBlock.nameIndex once
	// past threshold; linear scan only on small blocks).
	var found bool
	idx, found = fle.GetLocalIndex("~" + name)
	if !found {
		panic("should not happen, idx not found")
	}
	return
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
	// Fast path: build compositional types directly from the AST,
	// reusing ATTR_TYPE_VALUE on inner subexpressions. Avoids spinning up
	// a Machine per call, which would re-run all inner doOp* even when
	// inner subexpression types are already cached. Without this, a chain
	// like [][]…[]int costs O(N²) doOp invocations (one full sub-walk per
	// TRANS_LEAVE level).
	if t, ok := staticTypeFromAST(store, last, x); ok {
		validateTypeDepth(t, x)
		x.SetAttribute(ATTR_TYPE_VALUE, t)
		return t
	}
	return evalStaticTypeMachine(store, last, x)
}

// evalStaticTypeMachine is the original Machine-backed implementation of
// evalStaticType, retained for AST shapes the fast path doesn't cover
// (NameExpr, SelectorExpr, etc.). Sets ATTR_TYPE_VALUE before returning.
func evalStaticTypeMachine(store Store, last BlockNode, x Expr) Type {
	pn := packageOf(last)
	// See comment in evalStaticTypeOfRaw.
	if store != nil && pn.PkgPath != uversePkgPath {
		// this used pn.NewPackage previously; however, that function
		// additionally calls PrepareNewValues, which is not necessary in this
		// context and incurs in very expensive allocations.
		// Stamp PkgID on both the PackageValue and its inner Block.
		pid := pn.GetPkgID()
		blk := &Block{
			Source: pn,
		}
		blk.ObjectInfo.SetPkgID(pid)
		pv := &PackageValue{
			Block:   blk,
			PkgName: pn.PkgName,
			PkgPath: pn.PkgPath,
			PkgID:   pid,
		}
		pv.ObjectInfo.SetPkgID(pid)
		store = store.BeginTransaction(nil, nil, nil, nil)
		store.SetCachePackage(pv)
	}
	m := NewMachineWithOptions(MachineOptions{
		PkgPath:            pn.PkgPath,
		Store:              store,
		BoundedPanicRender: true,
	})
	tv := m.EvalStatic(last, x)
	m.Release()
	if _, ok := tv.V.(TypeValue); !ok {
		panic(fmt.Sprintf("%s is not a type", x.String()))
	}
	t := tv.GetType()
	validateTypeDepth(t, x)
	x.SetAttribute(ATTR_TYPE_VALUE, t)
	return t
}

// staticTypeFromAST constructs a Type directly from an AST type expression,
// mirroring doOp{Slice,Array,Pointer,Map,Chan,Func,Struct,Interface}Type in
// op_types.go field-for-field. Inner type expressions are resolved via
// evalStaticType, which hits ATTR_TYPE_VALUE on cache hits — making the
// fast path O(1) per node when called bottom-up via TRANS_LEAVE.
//
// Returns (nil, false) for AST shapes not handled here (NameExpr,
// SelectorExpr, *ConstExpr, etc.); the caller must fall back to the
// Machine path.
func staticTypeFromAST(store Store, last BlockNode, x Expr) (Type, bool) {
	switch x := x.(type) {
	case *SliceTypeExpr:
		return &SliceType{
			Elt: evalStaticType(store, last, x.Elt),
			Vrd: x.Vrd,
		}, true
	case *ArrayTypeExpr:
		if x.Len == nil { // [...]T variadic; Len is patched at preprocess.go:2346
			return &ArrayType{
				Elt: evalStaticType(store, last, x.Elt),
				Vrd: true,
			}, true
		}
		// evalConst requires the length to already be a *ConstExpr;
		// fall back to Machine if it hasn't been folded yet.
		if _, ok := x.Len.(*ConstExpr); !ok {
			return nil, false
		}
		return &ArrayType{
			Elt: evalStaticType(store, last, x.Elt),
			Len: int(evalConst(store, last, x.Len).GetInt()),
		}, true
	case *StarExpr:
		// *T as a type expression. Mirrors doOpStar's TypeType case
		// (op_expressions.go:191): wraps inner type in PointerType.
		return &PointerType{Elt: evalStaticType(store, last, x.X)}, true
	case *MapTypeExpr:
		return &MapType{
			Key:   evalStaticType(store, last, x.Key),
			Value: evalStaticType(store, last, x.Value),
		}, true
	case *ChanTypeExpr:
		return &ChanType{
			Dir: x.Dir,
			Elt: evalStaticType(store, last, x.Value),
		}, true
	case *FuncTypeExpr:
		// doOpFuncType does NOT call fillEmbeddedName on params/results.
		return &FuncType{
			Params:  buildFieldTypesAST(store, last, x.Params, false),
			Results: buildFieldTypesAST(store, last, x.Results, false),
		}, true
	case *StructTypeExpr:
		st := &StructType{
			PkgPath: packageOf(last).PkgPath,
			Fields:  buildFieldTypesAST(store, last, x.Fields, true),
		}
		validateEmbedDepth(st, "<anonymous struct>")
		validateStructFields(st, "<anonymous struct>")
		return st, true
	case *InterfaceTypeExpr:
		pkgPath := packageOf(last).PkgPath
		it := &InterfaceType{
			PkgPath: pkgPath,
			// Build without embed naming (embed=false): flattenInterfaceMethods
			// expands embedded interfaces into their method set, making identity
			// the method set rather than the embedded-interface (alias) spelling.
			Methods: flattenInterfaceMethods(buildFieldTypesAST(store, last, x.Methods, false), pkgPath),
			Generic: x.Generic,
		}
		validateEmbedDepth(it, "<anonymous interface>")
		validateInterfaceMethods(it, "<anonymous interface>")
		return it, true
	}
	return nil, false
}

// buildFieldTypesAST constructs a []FieldType directly from FieldTypeExprs,
// mirroring doOpFieldType + doOp{Struct,Interface,Func}Type aggregation.
// embed=true mirrors doOpStructType, which names embedded fields per field;
// embed=false mirrors doOpFuncType and the interface path (which leaves embeds
// unnamed and flattens them via flattenInterfaceMethods).
func buildFieldTypesAST(store Store, last BlockNode, fxs FieldTypeExprs, embed bool) []FieldType {
	fts := make([]FieldType, len(fxs))
	for i := range fxs {
		fx := &fxs[i]
		ft := FieldType{
			Name: fx.Name,
			Type: evalStaticType(store, last, fx.Type),
		}
		if fx.Tag != nil {
			ft.Tag = Tag(evalConst(store, last, fx.Tag).GetString())
		}
		if embed {
			fillEmbeddedName(&ft, fx.Type)
			if msg := embeddedFieldError(ft); ft.Embedded && msg != "" {
				panicAt(msg, fx)
			}
		}
		fts[i] = ft
	}
	return fts
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
			// Stamp PkgID on PackageValue + Block.
			pid := pn.GetPkgID()
			blk := &Block{
				Source: pn,
			}
			blk.ObjectInfo.SetPkgID(pid)
			pv := &PackageValue{
				// this used pn.NewPackage previously; however, that function
				// additionally calls PrepareNewValues, which is not necessary in this
				// context and incurs in very expensive allocations.
				Block:   blk,
				PkgName: pn.PkgName,
				PkgPath: pn.PkgPath,
				PkgID:   pid,
			}
			pv.ObjectInfo.SetPkgID(pid)
			store = store.BeginTransaction(nil, nil, nil, nil)
			store.SetCachePackage(pv)
		}
		m := NewMachineWithOptions(MachineOptions{
			PkgPath:            pn.PkgPath,
			Store:              store,
			BoundedPanicRender: true,
		})
		t = m.EvalStaticTypeOf(last, x)
		m.Release()
		validateTypeDepth(t, x)
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
	// store.GetAllocator() may be nil here — store.BeginTransaction
	// below forks alloc which propagates nil — and that's fine: this
	// PackageValue is a throwaway, never persisted, so a nil (no-op)
	// allocator suffices.
	pv := pn.NewPackage(nil) // throwaway
	store = store.BeginTransaction(nil, nil, nil, nil)
	store.SetCachePackage(pv)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath:            pn.PkgPath,
		Store:              store,
		BoundedPanicRender: true,
	})
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
		m := NewMachineWithOptions(MachineOptions{
			PkgPath:            ".dontcare",
			Store:              store,
			BoundedPanicRender: true,
		})
		cv := m.EvalStatic(last, x)
		m.Release()
		// The machine computes with runtime (IEEE) semantics, but a folded
		// expression is a constant, and constants have no signed zero: a
		// float -0 result (e.g. from negation) becomes +0.
		if cv.T != nil {
			switch cv.T.Kind() {
			case Float32Kind:
				if cv.GetFloat32() == softfloat.NegZero32 {
					cv.SetFloat32(0)
				}
			case Float64Kind:
				if cv.GetFloat64() == softfloat.NegZero64 {
					cv.SetFloat64(0)
				}
			}
		}
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

// crossingAllowed reports whether n (a FuncDecl or FuncLitExpr with a
// realm-first-arg signature) is allowed to be a crossing function in ctx.
// Crossing functions are allowed in realm packages OR in any *_test.gno
// file (regardless of package). The test-file carve-out lets `p/` package
// tests declare `func TestXxx(cur realm, t *testing.T)` and pass `cur`
// into newly-migrated methods — production code in `p/` still can't
// define crossing functions.
//
// MsgRun ephemeral `/e/` packages may also declare a crossing top-level
// `main` — the single entry point — so a run script can opt into
// `func main(cur realm)` + `cross(cur)`. Only the FuncDecl named `main`
// is allowed (no helper functions, no function literals), matching the
// ephemeral one-shot model.
func crossingAllowed(ctx BlockNode, n BlockNode) bool {
	if isRealm(ctx) {
		return true
	}
	file := n.GetLocation().File
	if strings.HasSuffix(file, "_test.gno") {
		return true
	}
	if fd, ok := n.(*FuncDecl); ok && fd.Name == "main" {
		if IsEphemeralPath(packageOf(ctx).PkgPath) {
			return true
		}
	}
	return false
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
	if isInvalid(t) || isInvalidExpr(*x) {
		return // accumulate mode: already reported.
	}
	if cx, ok := (*x).(*ConstExpr); ok {
		// A nil t means "no destination type, just default-convert below"
		// (the typeless `var x = <expr>` path and recursive untyped-operand
		// calls). It must be guarded here: checkAssignableTo now panics on a
		// nil dt rather than treating it as a no-op.
		if t != nil {
			// e.g. int(1) == int8(1)
			mustAssignableTo(store, n, cx.T, t)
		}
	} else if bx, ok := (*x).(*BinaryExpr); ok && (bx.Op == SHL || bx.Op == SHR) {
		xt := evalStaticTypeOf(store, last, *x)
		if debug {
			debug.Printf("shift, xt: %v, Op: %v, t: %v \n", xt, bx.Op, t)
		}

		if isUntyped(xt) {
			if t == nil || t.Kind() == InterfaceKind {
				if t != nil {
					// An untyped shift assigned to an interface target takes
					// its default type below, which drops the target on the
					// floor — so assert satisfaction before doing that, or
					// `var r R = 1 << 2` for a non-empty R is only caught at
					// runtime. Checked against xt, not defaultTypeOf(xt), so
					// the diagnostic names the untyped operand.
					mustAssignableTo(store, n, xt, t)
				}
				t = defaultTypeOf(xt)
			}
			// t is the type from context or default.
			bx.assertShiftExprCompatible2(t)

			// Convert untyped to typed.
			checkOrConvertType(store, last, n, &bx.Left, t)
			bx.SetAttribute(ATTR_TYPEOF_VALUE, t) // propagate converted type from left operand to shift expr.
		} else if t != nil {
			mustAssignableTo(store, n, xt, t)
		}
		return
	} else if *x != nil {
		xt := evalStaticTypeOf(store, last, *x)
		if t != nil {
			mustAssignableTo(store, n, xt, t)
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
						// An interface target gives the whole expression
						// its default type, which both operands take (the
						// shift in `1.0 + 1<<s` is then a float64 shift).
						if t.Kind() == InterfaceKind {
							t = defaultTypeOf(xt)
						}
						// push t into bx.Left and bx.Right
						checkOrConvertType(store, last, n, &bx.Left, t)
						checkOrConvertType(store, last, n, &bx.Right, t)
						return
					} else { // t is nil, and bx is untyped binary expr.
						if shouldSwapOnSpecificity(lt, rt) {
							// e.g. 1.0<<s + 1
							// The expression '1.0<<s' does not trigger assertions of
							// incompatible types when evaluated alone.
							// However, when evaluating the full expression '1.0<<s + 1'
							// without a specific context type, '1.0<<s' is checked against
							// its default type, the BigDecKind, will trigger assertion failure.
							// so here in checkOrConvertType, shift expression is "finally" checked.
							checkOrConvertType(store, last, n, &bx.Left, nil)
							checkOrConvertType(store, last, n, &bx.Right, lt)
						} else {
							checkOrConvertType(store, last, n, &bx.Left, rt)
							checkOrConvertType(store, last, n, &bx.Right, nil)
						}
					}
					return
				case EQL, LSS, GTR, NEQ, LEQ, GEQ:
					t = BoolType
				default:
					// do nothing
				}
			} else if ux, ok := (*x).(*UnaryExpr); ok {
				xt := evalStaticTypeOf(store, last, *x)
				// check assignable first
				if t != nil {
					mustAssignableTo(store, n, xt, t)
				}

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
	if isInvalid(t) || isInvalidExpr(*x) {
		return // accumulate mode: already reported.
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

// checkArrayAllocFits rejects a fixed-size array of element type et and the given
// length whose backing allocation would overflow int64, mirroring the runtime
// allocator's allocMustFit guard (AllocateListArray/AllocateDataArray). Go's
// compiler rejects such an array at compile time ("larger than address space");
// go/types does not, so without this it would only fail at allocation time.
// Called for every fixed-size array form: [N]T, [N]T{...}, and [...]T{idx: v}.
func checkArrayAllocFits(et Type, length int64) {
	if length <= 0 {
		return // negative lengths are rejected earlier (go/types).
	}
	per := arrayItemAllocSize(et)
	if length > (math.MaxInt64-allocArray)/per {
		panic(fmt.Sprintf("type [%d]%s larger than address space", length, et.String()))
	}
}

// like checkOrConvertType(last, x, nil)
func convertIfConst(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok {
		convertConst(store, last, n, cx, nil)
	}
}

func convertConst(store Store, last BlockNode, n Node, cx *ConstExpr, t Type) {
	if isInvalid(t) || isInvalid(cx.T) {
		return // accumulate mode: already reported.
	}
	if t != nil && t.Kind() == InterfaceKind {
		if cx.T != nil {
			mustAssignableTo(store, n, cx.T, t)
		}
		t = nil // signifies to convert to default type.
	}
	if isUntyped(cx.T) {
		ConvertUntypedTo(&cx.TypedValue, t)
		setConstAttrs(cx)
	} else if t != nil {
		// e.g. a named type or uint8 type to int for indexing.
		// convertConst's store may be nil (e.g. the Preprocess(nil, ...)
		// special-case in append-iface-nil at line ~1804). The
		// conversion lands in a ConstExpr that isn't persisted directly,
		// so use a nil (no-op) allocator: no stamping needed.
		ConvertTo(nil, store, &cx.TypedValue, t, true)
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
					panicAt(fmt.Sprintf("invalid recursive type: %s", cycleString(stack, cx.Name)),
						packageDeclOf(last, cx.Name))
				}
				if tv := last.GetSlot(store, cx.Name, true); tv != nil {
					return
				}
			} else {
				if tv := last.GetSlot(store, cx.Name, true); tv != nil {
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
	case BoolKind, InvalidKind:
		return // ok, or already reported (accumulate mode)
	default:
		panic(fmt.Sprintf(
			"expected typed bool kind, but got %v",
			xt.Kind()))
	}
}

// checkOrConvertIndexKind ensures the expression evaluates to an integer
// and, if it is a constant, ensures it is non-negative.
func checkOrConvertIndexKind(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok && cx.TypedValue.Sign() < 0 {
		panic(fmt.Sprintf("invalid argument: index must not be negative: %v", cx.TypedValue))
	}
	checkOrConvertIntegerKind(store, last, n, x)
}

// like checkOrConvertType() but for any typed integer kind.
func checkOrConvertIntegerKind(store Store, last BlockNode, n Node, x Expr) {
	if cx, ok := x.(*ConstExpr); ok {
		convertConst(store, last, n, cx, IntType)
	} else if x != nil {
		xt := evalStaticTypeOf(store, last, x)
		if isUntyped(xt) {
			dt := IntType // Push IntType for untyped.
			checkOrConvertType(store, last, n, &x, dt)
		} else {
			checkIntegerKind(xt)
		}
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
	return predefineRecursivelyIndexed(store, last, d, nil)
}

// predefineRecursivelyIndexed is the indexed entry point for PredefineFileSet.
func predefineRecursivelyIndexed(store Store, last BlockNode, d Decl, index *predefineDeclIndex) bool {
	defer doRecover([]BlockNode{last}, d)
	stack := []Name{}
	defining := make(map[Name]struct{})
	direct := true
	return predefineRecursively2(store, last, d, stack, defining, direct, index)
}

// `stack` and `defining` are used for cycle detection. They hold the same data.
// NOTE: `stack` never truncates; a slice is used instead of a map to show a
// helpful message when a circular declaration is found. `defining` is also used as
// a map to ensure best time performance of circular definition detection.
func predefineRecursively2(store Store, last BlockNode, d Decl, stack []Name, defining map[Name]struct{}, direct bool, index *predefineDeclIndex) bool {
	pkg := packageOf(last)
	// A plain panic while predefining d is about d: position it there,
	// not at the declaration that depended on it.
	defer func() {
		if r := recover(); r != nil {
			if msg, ok := r.(string); ok {
				panic(&spanError{span: d.GetSpan(), msg: msg})
			}
			panic(r)
		}
	}()

	// NOTE: PredefineFileSet splits multi-value decls like `var a, b = c, d`
	// into individual decls before calling this function, so that GetDeclFor
	// resolves each name to its own standalone decl. Without the split,
	// all names in the original multi-value decl would be added to `defining`,
	// causing false cycle detection for valid DAG dependencies.
	for _, dn := range d.GetDeclNames() {
		if isUverseName(dn) {
			panic(fmt.Sprintf(
				"builtin identifiers cannot be shadowed: %s", dn))
		}
		if _, isFunc := d.(*FuncDecl); !isFunc {
			if dn == "init" {
				panic("cannot declare init - must be func")
			}
			if dn == "main" && pkg.PkgName == "main" {
				panic("cannot declare main - must be func")
			}
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
			// first, check circularity. The error names the cycle only and
			// sits at the declaration that starts it, as in Go.
			if _, exists := defining[un]; exists {
				what := "value"
				if untype {
					what = "type"
				}
				msg := fmt.Sprintf("invalid recursive %s: %s", what, cycleString(stack, un))
				if !untype && index != nil && index.acc != nil { // gc's wording off chain; on chain unchanged
					msg = "initialization cycle: " + strings.ReplaceAll(cycleString(stack, un), " -> ", " refers to ")
				}
				var at Decl
				if index == nil {
					at = lookupPredefineDecl(pkg.FileSet, un).decl
				} else {
					at = index.lookup(un).decl
				}
				panicAt(msg, at)
			}
			// Look up the dependency declaration in the fileset.
			//
			// index is non-nil only when called from PredefineFileSet, the
			// package-level pass where lookups can chain N deep
			// (v0 -> v1 -> ... -> vN) and a live scan per hop is O(N^2).
			// Calls from Preprocess (function-body DeclStmts and the
			// per-file fallback loops) pass nil: on machine paths they run
			// after PredefineFileSet has already defined every package-level
			// name, so this branch is only reached for a name defined
			// nowhere, which is a single scan followed by GetDeclFor's panic.
			if index != nil && index.acc != nil {
				// Accumulate mode: a name declared nowhere becomes an
				// invalidType placeholder at each use; a failing
				// dependency is dropped with invalidType names. Either
				// way d goes on. A dependency that defined nothing
				// (or un recurring) drops d quietly.
				dependency, declared := index.entries[un]
				if !declared && index.acc.undefinedName(store, last, d, un) {
					continue
				}
				if declared && dependency.decl.GetAttribute(ATTR_PREDEFINED) != true {
					if !dependency.file.IsInitialized() {
						panic("all types from files in file-set should have already been predefined")
					}
					index.predefineDependency(store, dependency, stack, defining, directR)
				}
				if last.GetSlot(store, un, true) == nil {
					panic(&droppedDeclError{name: un})
				}
				continue
			}
			var dependency predefineDecl
			var found bool
			if index == nil {
				dependency, found = lookupPredefineDeclSafe(pkg.FileSet, un)
			} else {
				dependency, found = index.find(un)
			}
			if !found {
				panicAt(fmt.Sprintf("name %s not defined in fileset with files %v", un, pkg.FileSet.FileNames()),
					undefinedNameNode(preprocessGasMeterOf(store), d, un), d)
			}
			// preprocess if not already preprocessed.
			if !dependency.file.IsInitialized() {
				panic("all types from files in file-set should have already been predefined")
			}
			if dependency.decl.GetAttribute(ATTR_PREDEFINED) == true {
				// Dropped by accumulate mode after it was predefined.
				panic(fmt.Sprintf("name %s not declared", un))
			}
			// predefine dependency recursively.
			// `directR` is passed on.
			predefineRecursively2(store, dependency.file, dependency.decl, stack, defining, directR, index)
		} else {
			break // predefine successfully performed.
		}
	}
	var acc *preprocessAccumulator // nested walk inherits accumulate mode
	if index != nil {
		acc = index.acc
	}
	switch cd := d.(type) {
	case *FuncDecl:
		// We cannot Preprocess the body of *FuncDecl as it may
		// refer to package names in any order.
		return false
	case *ValueDecl:
		vd2 := preprocessDependency(store, last, cd, index).(*ValueDecl)
		*cd = *vd2
		return true
	case *TypeDecl:
		td2 := preprocessDependency(store, last, cd, index).(*TypeDecl)
		*cd = *td2
		return true
	case *ImportDecl:
		id2 := preprocess(store, last, cd, acc).(*ImportDecl)
		*cd = *id2
		return true
	default:
		panic("should not happen")
	}
}

// preprocessDependency preprocesses a declaration reached through a
// dependent one, with the accumulator of index (if any) so a node in
// it can report more than one error. If it fails it is marked done
// first, so accumulate mode does not visit it again and report it twice.
func preprocessDependency(store Store, last BlockNode, d Decl, index *predefineDeclIndex) Node {
	var acc *preprocessAccumulator
	if index != nil {
		acc = index.acc
	}
	defer func() {
		if r := recover(); r != nil {
			if _, isRuntime := r.(runtime.Error); isRuntime {
				panic(r)
			}
			d.SetAttribute(ATTR_PREPROCESSED, true)
			pe := wrapPreprocessPanic([]BlockNode{last}, d, r)
			pe.dropped = append(pe.dropped, d.GetDeclNames()...)
			panic(pe)
		}
	}()
	return preprocess(store, last, d, acc)
}

// packageDeclOf returns the package-level declaration of name, or nil
// for a local or unknown one.
func packageDeclOf(last BlockNode, name Name) (d Decl) {
	defer func() {
		if recover() != nil {
			d = nil
		}
	}()
	return lookupPredefineDecl(packageOf(last).FileSet, name).decl
}

// cycleString prints the declaration cycle closed by name: the part of
// stack from name's first occurrence, then name again.
func cycleString(stack []Name, name Name) string {
	for i, n := range stack {
		if n == name {
			stack = stack[i:]
			break
		}
	}
	return Names(append(append([]Name{}, stack...), name)).Join(" -> ")
}

// If a dependent name is not yet defined, that name is returned; this return
// value is used by the caller to enforce declaration order.
//
// If all dependencies are met, constructs and empty definition value (for a
// *TypeDecl is a TypeValue) and sets it on last. As an exception, *FuncDecls
// will preprocess receiver/argument/result types recursively.
func tryPredefine(store Store, pkg *PackageNode, last BlockNode, d Decl, stack []Name, defining map[Name]struct{}, direct bool) (un Name, untype bool, directR bool) {
	if d.GetAttribute(ATTR_PREDEFINED) == true {
		if d.GetAttribute(ATTR_PREPROCESS_DROPPED) == true {
			// accumulate mode: a dependent of a dropped declaration.
			panic(&droppedDeclError{name: firstDeclName(d)})
		}
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
			case *constTypeExpr:
				// accumulate mode's invalidType placeholder for an
				// undefined name (undefinedName).
				t = tx.Type
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
			if isInvalid(unwrapPointerType(rt)) {
				// accumulate mode: the receiver type was reported.
				panic(&droppedDeclError{name: d.Name})
			}
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
			methodFV := &FuncValue{
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
			}
			// Method FuncValue belongs to the declaring package's
			// realm.
			methodFV.ObjectInfo.SetPkgID(PkgIDFromPkgPath(pkg.PkgPath))
			if st, ok := baseOf(dt).(*StructType); ok {
				for _, f := range st.Fields {
					if f.Name == d.Name {
						packageOf(last).StaticBlock.revertToOld()
						panic(fmt.Sprintf("field and method with the same name %s", d.Name))
					}
				}
			}
			if !dt.TryDefineMethod(methodFV) {
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
			// Top-level function FuncValue belongs to the declaring
			// package's realm.
			fv.ObjectInfo.SetPkgID(PkgIDFromPkgPath(pkg.PkgPath))
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
					panicAt("missing type in composite literal", vclx)
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
				// Cache static type for doOpRef; see TRANS_LEAVE *RefExpr.
				clx.SetAttribute(ATTR_TYPEOF_VALUE, t.Elem())
				refx := &RefExpr{X: clx}
				refx.SetAttribute(ATTR_REF_ELEM_TYPE, t.Elem())
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
		if cxift.Kind() == TypeKind || isInvalid(cxift) {
			return 1 // type conversion, or already reported
		} else {
			cxft := getGnoFuncTypeOf(store, cxift)
			numResults := len(cxft.Results)
			return numResults
		}
	} else {
		return 1
	}
}

// codaInitOrderDeps records ATTR_DECL_DEPS on package-level *ValueDecl and
// *FuncDecl nodes by scanning for all references to other package-level names.
// It must run after preprocess1 (so all NameExpr paths are filled) and before
// codaPackageSelectors (which replaces NameExprs with SelectorExprs).
//
// This is implemented as a separate post-preprocess pass (rather than inline
// within preprocess1) because preprocess1 recursively calls Preprocess for
// inner function literal bodies with a fresh context, losing the enclosing
// declaration. By iterating fn.Decls and transcribing each one, the target
// declaration is always known from the outer loop.
func codaInitOrderDeps(pn *PackageNode, fn *FileNode) {
	if pn.PkgPath == ".uverse" {
		return
	}
	for _, decl := range fn.Decls {
		switch decl.(type) {
		case *FuncDecl, *ValueDecl:
		default:
			continue
		}

		var deps map[Name]struct{}
		addDep := func(name Name) {
			if deps == nil {
				deps = make(map[Name]struct{})
			}
			deps[name] = struct{}{}
		}

		_ = TranscribeB(fn, decl, func(
			ns []Node,
			stack []BlockNode,
			last BlockNode,
			ftype TransField,
			index int,
			n Node,
			stage TransStage,
		) (Node, TransCtrl) {
			switch stage {
			case TRANS_ENTER:
				// Skip function literal bodies that were not preprocessed;
				// same guard as the main coda Transcribe in Preprocess.
				if flx, ok := n.(*FuncLitExpr); ok {
					if flx.GetAttribute(ATTR_PREPROCESS_SKIPPED) == AttrPreprocessFuncLitExpr {
						return n, TRANS_SKIP
					}
				}
			case TRANS_LEAVE:
				switch n := n.(type) {
				case *NameExpr:
					// Only track names resolved via the block hierarchy.
					if n.Path.Type != VPBlock {
						return n, TRANS_CONTINUE
					}
					// Ignore blank identifiers.
					if n.Name == blankIdentifier {
						return n, TRANS_CONTINUE
					}
					// Ignore package-name references (e.g. `fmt` in `fmt.Println`).
					if n.GetAttribute(ATTR_PACKAGE_REF) != nil {
						return n, TRANS_CONTINUE
					}
					// Ignore the declaration name itself (LHS of var/const decl).
					if ftype == TRANS_VAR_NAME {
						return n, TRANS_CONTINUE
					}
					// Check that the name is defined at package level.
					dbn := last.GetBlockNodeForPath(nil, n.Path)
					if dbn != pn {
						return n, TRANS_CONTINUE
					}
					// Ignore type declarations; they have no runtime
					// initialization order.
					if li, ok := pn.GetLocalIndex(n.Name); ok {
						if pn.NameSources[li].Type == NSTypeDecl {
							return n, TRANS_CONTINUE
						}
					}
					addDep(n.Name)
				case *SelectorExpr:
					// Track same-package method calls so that resolveEffectiveDeps
					// can transitively discover vars referenced in method bodies.
					// e.g. `A = T{}.GetB()` records "T.GetB" as a dep of A;
					// resolveEffectiveDeps then walks GetB's body to find B.
					switch n.Path.Type {
					case VPValMethod, VPPtrMethod, VPDerefValMethod, VPDerefPtrMethod:
						// Get the receiver type from ATTR_REF_ELEM_TYPE
						// on the RefExpr, or ATTR_TYPEOF_VALUE on n.X.
						//
						// For auto-addressed receivers (n.X is a synthetic *RefExpr),
						// read from ATTR_REF_ELEM_TYPE. For already-pointer receivers
						// (VPDerefValMethod/VPDerefPtrMethod), n.X is not a RefExpr;
						// fall back to ATTR_TYPEOF_VALUE on the expression itself.
						var xt Type
						var ok bool
						if rx, ok2 := n.X.(*RefExpr); ok2 {
							xt, ok = rx.GetAttribute(ATTR_REF_ELEM_TYPE).(Type)
						}
						if !ok {
							xt, ok = n.X.GetAttribute(ATTR_TYPEOF_VALUE).(Type)
						}
						if !ok {
							break
						}
						// A call result (`f().m`) is typed as a one-element tuple.
						if tt, ok2 := xt.(*tupleType); ok2 && len(tt.Elts) == 1 {
							xt = tt.Elts[0]
						}
						// Dereference pointer receiver types.
						if pt, ok2 := xt.(*PointerType); ok2 {
							xt = pt.Elt
						}
						dt, ok := xt.(*DeclaredType)
						if !ok || dt.PkgPath != pn.PkgPath {
							break
						}
						addDep(dt.Name + "." + n.Sel)
					case VPField:
						// VPField is struct-field access (s.F) or an unbound
						// method expression (T.M, (*T).M). Only the latter
						// evaluates n.X as a type, so ATTR_TYPE_VALUE holds
						// the receiver type; struct fields lack it and break.
						rt, ok := n.X.GetAttribute(ATTR_TYPE_VALUE).(Type)
						if !ok {
							break
						}
						// Unwrap a pointer receiver, covering both (*T).M and
						// the pointer-alias form (`type P = *T; P.M`).
						if pt, ok := rt.(*PointerType); ok {
							rt = pt.Elt
						}
						dt, ok := rt.(*DeclaredType)
						if !ok || dt.PkgPath != pn.PkgPath {
							break
						}
						addDep(dt.Name + "." + n.Sel)
					}
				}
			}
			return n, TRANS_CONTINUE
		})

		if deps != nil {
			decl.SetAttribute(ATTR_DECL_DEPS, deps)
		}
	}
}

// resolveDeclDep resolves a dependency name (as stored in ATTR_DECL_DEPS) to
// the corresponding Decl in pn. Method dependencies are encoded as "Type.Method".
func resolveDeclDep(name Name, pn *PackageNode) Decl {
	id, sel, isMethod := strings.Cut(string(name), ".")
	if isMethod {
		li, found := pn.GetLocalIndex(Name(id))
		if !found || pn.NameSources[li].Type != NSTypeDecl {
			panic(fmt.Sprintf("type %s not found in package %s", id, pn.PkgName))
		}
		dt, ok := pn.Types[li].(*DeclaredType)
		if !ok {
			panic(fmt.Sprintf("type %s is not a *DeclaredType in package %s", id, pn.PkgName))
		}
		idx, ok := dt.lookupMethod(Name(sel))
		if !ok {
			panic(fmt.Sprintf("method %s not found in type %s", sel, id))
		}
		return dt.Methods[idx].V.(*FuncValue).Source.(*FuncDecl)
	}
	li, found := pn.GetLocalIndex(name)
	if !found {
		panic(fmt.Sprintf("name %s not found in package %s", name, pn.PkgName))
	}
	return pn.NameSources[li].Origin.(Decl)
}

// resolveEffectiveDeps computes, for every Decl reachable from the given
// declarations, the set of *ValueDecl dependencies obtained by collapsing
// FuncDecl edges (FuncDecls are transparent pass-throughs).
//
// The result is a shared cache: cache[d] = list of *ValueDecl that d
// (transitively through FuncDecls) depends on. Each Decl is visited at most
// once across all calls, so total work is O(V+E).
//
// Circular variable dependencies are detected and cause a panic with a
// descriptive chain.
func resolveEffectiveDeps(decls []Decl, pn *PackageNode, fdeclared map[Name]struct{}) map[Decl][]*ValueDecl {
	cache := map[Decl][]*ValueDecl{} // fully resolved
	onStack := map[Decl]bool{}       // grey: currently in DFS path

	// inFDeclared reports whether all names declared by d are already in
	// fdeclared, meaning d has already been initialized.
	inFDeclared := func(d *ValueDecl) bool {
		for _, n := range d.GetDeclNames() {
			if _, ok := fdeclared[n]; !ok {
				return false
			}
		}
		return true
	}

	var walk func(d Decl, path []Name) []*ValueDecl
	walk = func(d Decl, path []Name) []*ValueDecl {
		if res, ok := cache[d]; ok {
			return res
		}
		onStack[d] = true
		m, _ := d.GetAttribute(ATTR_DECL_DEPS).(map[Name]struct{})
		// Sort dependency names for deterministic DFS traversal order.
		names := slices.Collect(maps.Keys(m))
		slices.Sort(names)

		var result []*ValueDecl
		for _, name := range names {
			dep := resolveDeclDep(name, pn)
			switch dep := dep.(type) {
			case *FuncDecl:
				if onStack[dep] {
					// Mutually recursive functions are fine; skip.
					continue
				}
				// Collapse: inherit effective deps from the FuncDecl.
				for _, vd := range walk(dep, append(path, name)) {
					if !slices.Contains(result, vd) {
						result = append(result, vd)
					}
				}
			case *ValueDecl:
				if onStack[dep] {
					bld := strings.Builder{}
					if fn, sourceDecl, ok := pn.FileSet.GetDeclForSafe(path[0]); ok {
						fmt.Fprintf(&bld, "%s/%s:%s: ", pn.PkgPath, fn.FileName, (*sourceDecl).GetSpan().Pos.String())
					}
					bld.WriteString("circular dependency: ")
					for _, n := range path {
						bld.WriteString(string(n))
						bld.WriteString(" -> ")
					}
					bld.WriteString(string(name))
					panic(bld.String())
				}
				// Skip already-initialized decls entirely.
				if len(fdeclared) > 0 && inFDeclared(dep) {
					continue
				}
				if !slices.Contains(result, dep) {
					result = append(result, dep)
				}
				// Recurse into ValueDecl deps to detect cycles and to
				// discover transitive deps through FuncDecl chains rooted
				// in this ValueDecl. Kahn's handles direct ValueDecl→ValueDecl
				// transitivity, but we still need to walk through for cycle
				// detection and for FuncDecl collapse.
				walk(dep, append(path, name))
			default:
				panic(fmt.Sprintf("unexpected gnolang.Decl: %#v", dep))
			}
		}
		delete(onStack, d)
		cache[d] = result
		return result
	}

	for _, d := range decls {
		if _, ok := cache[d]; ok {
			continue
		}
		rootNames := d.GetDeclNames()
		rootName := Name("_")
		if len(rootNames) > 0 {
			rootName = rootNames[0]
		}
		walk(d, []Name{rootName})
	}
	return cache
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

// if name is is reserved.
// checkShortVarDecl enforces the two rules of `:=` the resolver does not
// see after name resolution: no name twice on the left, and at least one
// new variable in the current Go scope. Gno folds a control statement's
// header (range variables, `for`/`if`/`switch` init) into the same block
// as its body, where Go opens a new scope, so a name that header declared
// may be shadowed; a name declared by a sibling statement, a parameter,
// the type-switch guard or a select case's receive (Go declares those in
// the clause's own block) may not. initStaticBlocks reserved the names on
// behalf of this statement, so "new" also means reserved by this very node.
func checkShortVarDecl(ns []Node, last BlockNode, n *AssignStmt) {
	seen := map[Name]bool{}
	anyNew := false
	for _, lx := range n.Lhs {
		ln := lx.(*NameExpr).Name
		if ln == blankIdentifier {
			continue
		}
		if strings.HasPrefix(string(ln), ".") {
			return // synthesized by the preprocessor
		}
		if seen[ln] {
			panic(fmt.Sprintf("%s repeated on left side of :=", ln))
		}
		seen[ln] = true
		src := definedBy(last, ln)
		if src == nil || src.Origin == n || declaredByHeader(ns, last, src) {
			anyNew = true
		}
	}
	if !anyNew && len(seen) > 0 {
		panic("no new variables on left side of :=")
	}
}

// reportUnused runs checkUnused over fn, the function node on top of
// stack, and raises each unused name: the first panics on chain, all are
// reported in accumulate mode (minus names a dropped statement declared
// or mentioned, which the walk could never see used).
func reportUnused(store Store, acc *preprocessAccumulator, stack []BlockNode, fn BlockNode) {
	if len(stack) < 2 {
		return
	}
	for _, u := range checkUnused(preprocessGasMeterOf(store), stack[len(stack)-2], fn) {
		if acc != nil && (acc.poisoned[u.name] || acc.referenced[u.name]) {
			continue
		}
		loc := fn.GetLocation()
		loc.SetSpanOverride(u.span)
		pe := &PreprocessError{err: fmt.Errorf("%s: %s", loc.String(), u.msg), stack: stack}
		if acc == nil {
			panic(pe)
		}
		acc.record(pe)
	}
}

// reportUnusedImports runs checkUnusedImports over the package's files and
// raises each unused import: the first panics on chain, all are reported
// in accumulate mode (minus names a dropped statement mentioned, which the
// walk could never see used).
func reportUnusedImports(store Store, acc *preprocessAccumulator, pn *PackageNode, fset *FileSet) {
	pkgName := func(pkgPath string) Name { // cached since predefine
		if pv := store.GetPackage(pkgPath, false); pv != nil {
			return pv.PkgName
		}
		return ""
	}
	for _, u := range checkUnusedImports(preprocessGasMeterOf(store), fset, pkgName) {
		if acc != nil && (acc.poisoned[u.name] || acc.referenced[u.name]) {
			continue
		}
		loc := Location{PkgPath: pn.PkgPath, File: u.file.FileName, Span: u.span}
		pe := &PreprocessError{err: fmt.Errorf("%s: %s", loc.String(), u.msg)}
		if acc == nil {
			panic(pe)
		}
		acc.record(pe)
	}
}

// reportOrPanic raises msg at n: the first error panics on chain; in
// accumulate mode it is recorded and the walk goes on, for a rule whose
// failure leaves the node's type intact (gc reports every such site).
func reportOrPanic(acc *preprocessAccumulator, stack []BlockNode, n Node, msg string) {
	pe := wrapPreprocessPanic(stack, n, msg)
	if acc == nil {
		panic(pe)
	}
	acc.record(pe)
}

// enclosedByFunc reports whether stack holds a function node.
func enclosedByFunc(stack []BlockNode) bool {
	for _, bn := range stack {
		switch bn.(type) {
		case *FuncDecl, *FuncLitExpr:
			return true
		}
	}
	return false
}

// declaredByHeader reports whether src was declared by the header of the
// control statement whose body last holds (see checkShortVarDecl).
func declaredByHeader(ns []Node, last BlockNode, src *NameSource) bool {
	switch src.Type {
	case NSRangeKey, NSRangeValue:
		return true
	}
	parentOf := func(bn BlockNode) Node {
		for i := len(ns) - 1; i >= 0; i-- {
			if ns[i] == bn && i > 0 {
				return ns[i-1]
			}
		}
		return nil
	}
	switch b := last.(type) {
	case *ForStmt:
		return b.Init != nil && src.Origin == b.Init
	case *IfCaseStmt:
		if p, ok := parentOf(b).(*IfStmt); ok {
			return p.Init != nil && src.Origin == p.Init
		}
	case *SwitchClauseStmt:
		if p, ok := parentOf(b).(*SwitchStmt); ok {
			return p.Init != nil && src.Origin == p.Init
		}
	}
	return false
}

// definedBy returns the name source that reserved name in bn's own static
// block, nil when the name is not local to bn.
func definedBy(bn BlockNode, name Name) *NameSource {
	idx, ok := bn.GetLocalIndex(name)
	if !ok {
		return nil
	}
	sources := bn.GetStaticBlock().NameSources
	if int(idx) >= len(sources) {
		return nil
	}
	return &sources[idx]
}

func isLocallyReserved(bn BlockNode, n Name) bool {
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
	pkgPath := pn.PkgPath
	fileName := fn.FileName
	if pkgPath == "" || fileName == "" {
		panic("missing package path or file name")
	}
	// Collect first and publish as one batch: the store seals a batch under a
	// single sealer, and sealing node by node would re-walk the package's type
	// graph once per node.
	// The same walk seals the types held on expressions rather than in a static
	// block; see sealExprTypes for why it belongs here and not at publication.
	sl := newSealer()
	bns := []BlockNode{pn}
	Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		sl.sealExprTypes(n)
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
			bns = append(bns, bn)
		}
		return n, TRANS_CONTINUE
	})
	store.SetBlockNodes(bns)
}
