package gnolang

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// unusedName is a local variable or label a function declares and never
// uses (Go spec: "Declared and not used"; "Labeled statements").
type unusedName struct {
	name  Name
	msg   string
	span  Span
	label bool
}

// checkUnused reports the unused local variables and labels of fn (a
// *FuncDecl or a package-level *FuncLitExpr), which must be preprocessed
// (every name has its path), in source order. parent is fn's parent block.
// A variable is used by any occurrence except as a target of `=` or `:=`;
// a label by a goto, break or continue naming it. The variable of a type
// switch counts as used when any clause uses it.
func checkUnused(gm store.GasMeter, parent BlockNode, fn Node) []unusedName {
	// Keyed by the declaring statement: faux blocks (if, switch) copy the
	// name sources of their init statement into each case block.
	type key struct {
		origin Node
		name   Name
	}
	used := map[key]bool{}
	var blocks []BlockNode
	type label struct {
		name Name
		span Span
	}
	var labels []label // in walk order; never ranged as a map
	branched := map[Name]bool{}
	plain := func(n Name) Name { return Name(strings.TrimSuffix(string(n), ".loopvar")) }

	TranscribeB(parent, fn, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage == TRANS_ENTER {
			chargeCPUGas(gm, OpCPUSlopeCheckNode)
		}
		switch stage {
		case TRANS_BLOCK:
			blocks = append(blocks, n.(BlockNode))
		case TRANS_ENTER:
			if st, ok := n.(Stmt); ok {
				if l := st.GetLabel(); l != "" && l != blankIdentifier {
					labels = append(labels, label{l, n.GetSpan()})
				}
			}
			if bs, ok := n.(*BranchStmt); ok && bs.Label != "" {
				branched[bs.Label] = true
			}
		case TRANS_LEAVE:
			// A folded constant (`len(a)` of an array) hides the names its
			// source expression used; they count as uses.
			if cx, ok := n.(*ConstExpr); ok && cx.Source != nil {
				Transcribe(cx.Source, func(_ []Node, _ TransField, _ int, sn Node, st TransStage) (Node, TransCtrl) {
					if snx, ok := sn.(*NameExpr); ok && st == TRANS_LEAVE && snx.Path.Type == VPBlock {
						if dbn := last.GetBlockNodeForPath(nil, snx.Path); dbn != nil {
							sources := dbn.GetStaticBlock().NameSources
							if idx, ok := dbn.GetLocalIndex(snx.Name); ok && int(idx) < len(sources) {
								used[key{sources[idx].Origin, plain(snx.Name)}] = true
							}
						}
					}
					return sn, TRANS_CONTINUE
				})
				break
			}
			nx, ok := n.(*NameExpr)
			if !ok || nx.Path.Type != VPBlock || nx.Name == blankIdentifier || strings.HasPrefix(string(nx.Name), ".") {
				break
			}
			switch ftype {
			case TRANS_VAR_NAME, TRANS_RANGE_KEY, TRANS_RANGE_VALUE:
				break // a declaration site
			case TRANS_ASSIGN_LHS:
				if as, ok := ns[len(ns)-1].(*AssignStmt); ok && (as.Op == ASSIGN || as.Op == DEFINE) {
					break // a target, not a use
				}
				fallthrough
			default:
				if dbn := last.GetBlockNodeForPath(nil, nx.Path); dbn != nil {
					sources := dbn.GetStaticBlock().NameSources
					if idx, ok := dbn.GetLocalIndex(nx.Name); ok && int(idx) < len(sources) {
						used[key{sources[idx].Origin, plain(nx.Name)}] = true
					}
				}
			}
		}
		return n, TRANS_CONTINUE
	})

	var out []unusedName
	seen := map[key]bool{}
	for _, bn := range blocks {
		sb := bn.GetStaticBlock()
		for _, ns := range sb.NameSources {
			if ns.NameExpr == nil {
				continue
			}
			name := plain(ns.NameExpr.Name)
			if name == blankIdentifier || strings.HasPrefix(string(name), ".") {
				continue
			}
			switch ns.Type {
			case NSDefine, NSRangeKey, NSRangeValue, NSTypeSwitch:
			case NSValueDecl:
				if vd, ok := ns.Origin.(*ValueDecl); ok && vd.Const {
					continue // the rule is about variables; unused constants are fine
				}
			default:
				continue
			}
			k := key{ns.Origin, name}
			if seen[k] || used[k] {
				continue
			}
			if ns.Origin.GetAttribute(ATTR_PREPROCESS_DROPPED) == true {
				continue // accumulate mode dropped the statement with its uses
			}
			if idx, ok := sb.GetLocalIndex(ns.NameExpr.Name); ok && isInvalid(sb.Types[idx]) {
				continue // a dropped statement's placeholder (accumulate mode)
			}
			seen[k] = true
			span := ns.NameExpr.GetSpan()
			if span.IsZero() { // the type-switch variable's name expr is synthetic
				span = ns.Origin.GetSpan()
			}
			msg := fmt.Sprintf("declared and not used: %s", name)
			if ns.Type == NSTypeSwitch {
				msg = fmt.Sprintf("%s declared and not used", name) // gc's spelling for the guard variable
			}
			out = append(out, unusedName{name: name, msg: msg, span: span})
		}
	}
	for _, l := range labels {
		if !branched[l.name] {
			out = append(out, unusedName{name: l.name, msg: fmt.Sprintf("label %s defined and not used", l.name), span: l.span, label: true})
		}
	}
	// Source order, as gc reports: the first error is the same on every
	// node, and no map order reaches the text. A label carries its
	// statement's span, so on a tie it comes first (`L:` precedes `x := 1`).
	sort.SliceStable(out, func(i, j int) bool {
		if c := out[i].span.Pos.Compare(out[j].span.Pos); c != 0 {
			return c < 0
		}
		return out[i].label && !out[j].label
	})
	return out
}

// unusedImport is an import no file of the package refers to (Go spec:
// "Import declarations"; gc: `"math" imported and not used`).
type unusedImport struct {
	name Name
	msg  string
	file *FileNode
	span Span
}

// checkUnusedImports reports the named imports of fset (dot and blank
// imports are exempt, as in Go) that no preprocessed file refers to, in
// source order: files, then import specs. A reference is a NameExpr the
// preprocessor resolved to a package (ATTR_PACKAGE_REF), in the tree or
// in the source a folded constant or type keeps. pkgName gives a path's
// package name, for gc's `imported as m` spelling of a renamed import.
func checkUnusedImports(gm store.GasMeter, fset *FileSet, pkgName func(pkgPath string) Name) []unusedImport {
	type key struct {
		file *FileNode
		name Name
	}
	used := map[key]bool{}  // never ranged
	seen := map[Node]bool{} // folded sources may refer to each other
	for _, fn := range fset.Files {
		var sources func(n Node)
		mark := func(n Node) {
			if nx, ok := n.(*NameExpr); ok && nx.GetAttribute(ATTR_PACKAGE_REF) != nil {
				used[key{fn, nx.Name}] = true
			}
		}
		// Transcribe treats a folded constant or type as a leaf; the
		// names its source used still count as uses.
		sources = func(n Node) {
			if seen[n] {
				return
			}
			seen[n] = true
			Transcribe(n, func(_ []Node, _ TransField, _ int, sn Node, st TransStage) (Node, TransCtrl) {
				if st != TRANS_ENTER {
					return sn, TRANS_CONTINUE
				}
				chargeCPUGas(gm, OpCPUSlopeCheckNode)
				mark(sn)
				switch sx := sn.(type) {
				case *ConstExpr:
					if sx.Source != nil {
						sources(sx.Source)
					}
				case *constTypeExpr:
					if sx.Source != nil {
						sources(sx.Source)
					}
				}
				return sn, TRANS_CONTINUE
			})
		}
		sources(fn)
	}
	var out []unusedImport
	for _, fn := range fset.Files {
		for _, d := range fn.Decls {
			id, ok := d.(*ImportDecl)
			if !ok || d.GetAttribute(ATTR_PREPROCESS_DROPPED) == true {
				continue
			}
			name := id.NameExpr.Name
			if name == blankIdentifier || name == "." || used[key{fn, name}] {
				continue
			}
			msg := fmt.Sprintf("%q imported and not used", id.PkgPath)
			if pkgName != nil && pkgName(id.PkgPath) != name {
				msg = fmt.Sprintf("%q imported as %s and not used", id.PkgPath, name)
			}
			out = append(out, unusedImport{name: name, msg: msg, file: fn, span: id.GetSpan()})
		}
	}
	return out
}
