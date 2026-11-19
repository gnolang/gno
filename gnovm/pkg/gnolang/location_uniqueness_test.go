package gnolang

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBlockNodeLocationUniqueElseIf covers #6065: an else-if used to give the
// wrapping IfCaseStmt and the nested IfStmt identical Locations, because
// Go2Gno copied the Else *ast.IfStmt's span onto both. Locations must stay
// unique — SetBlockNode keys on them, and #6060 keys local-type TypeIDs to
// declaring-block Locations.
func TestBlockNodeLocationUniqueElseIf(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	x := 0
	if x == 0 {
		println("a")
	} else if x == 1 {
		println("b")
	} else {
		println("c")
	}
}
`
	var m *Machine
	fn, err := m.ParseFile("main.gno", src)
	require.NoError(t, err)

	setNodeLocations("test", "main.gno", fn)

	seen := map[Location]Node{}
	var dups []string
	Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		bn, ok := n.(BlockNode)
		if !ok {
			return n, TRANS_CONTINUE
		}
		loc := bn.GetLocation()
		if prev, ok := seen[loc]; ok {
			dups = append(dups, fmt.Sprintf("%s -> [%T %T]", loc, prev, bn))
		} else {
			seen[loc] = bn
		}
		return n, TRANS_CONTINUE
	})
	require.Empty(t, dups, "duplicate BlockNode.Locations:\n%v", dups)

	// And the checker that used to be a stub agrees.
	require.NotPanics(t, func() {
		checkNodeLinesLocations("test", "main.gno", fn)
	})
}

// Two independent ifs with empty else branches share a zero Span; setNodeLocations
// must still give their Else IfCaseStmts distinct Locations via Num.
func TestBlockNodeLocationUniqueEmptyElse(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	if true {
		println("a")
	}
	if false {
		println("b")
	}
}
`
	var m *Machine
	fn, err := m.ParseFile("main.gno", src)
	require.NoError(t, err)

	setNodeLocations("test", "main.gno", fn)
	require.NotPanics(t, func() {
		checkNodeLinesLocations("test", "main.gno", fn)
	})
}
