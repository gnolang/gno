package gnolang

import (
	"fmt"
	"strings"
)

// unusedName is a local variable or label a function declares and never
// uses (Go spec: "Declared and not used"; "Labeled statements").
type unusedName struct {
	name Name
	msg  string
	span Span
}

// checkUnused reports the unused local variables and labels of the
// function fn (a declaration or literal), which must be preprocessed
// (every name has its path). parent is fn's parent block. A variable is used by any occurrence except as a target of `=` or
// `:=`; a label by a goto, break or continue naming it. The variable of a
// type switch counts as used when any clause uses it.
func checkUnused(parent BlockNode, fn Node) []unusedName {
	// Keyed by the declaring statement: faux blocks (if, switch) copy the
	// name sources of their init statement into each case block.
	type key struct {
		origin Node
		name   Name
	}
	used := map[key]bool{}
	var blocks []BlockNode
	labels := map[Name]Span{}
	branched := map[Name]bool{}
	plain := func(n Name) Name { return Name(strings.TrimSuffix(string(n), ".loopvar")) }

	TranscribeB(parent, fn, func(ns []Node, stack []BlockNode, last BlockNode, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		switch stage {
		case TRANS_BLOCK:
			blocks = append(blocks, n.(BlockNode))
		case TRANS_ENTER:
			if st, ok := n.(Stmt); ok {
				if l := st.GetLabel(); l != "" {
					labels[l] = n.GetSpan()
				}
			}
			if bs, ok := n.(*BranchStmt); ok && bs.Label != "" {
				branched[bs.Label] = true
			}
		case TRANS_LEAVE:
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
		for _, ns := range bn.GetStaticBlock().NameSources {
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
			seen[k] = true
			span := ns.NameExpr.GetSpan()
			if span.IsZero() { // the type-switch variable's name expr is synthetic
				span = ns.Origin.GetSpan()
			}
			out = append(out, unusedName{name, fmt.Sprintf("declared and not used: %s", name), span})
		}
	}
	for l, span := range labels {
		if !branched[l] {
			out = append(out, unusedName{l, fmt.Sprintf("label %s defined and not used", l), span})
		}
	}
	return out
}
