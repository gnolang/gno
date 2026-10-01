// PKGPATH: gno.land/r/decltyperec
package decltyperec

// Recursive-type companion to decltype.gno.
//
// Node contains a *Node field — a self-reference that makes the
// DeclaredType cyclic. The PR's encoding must round-trip this shape:
// copyTypeWithRefs has to emit a RefType for the self-pointer instead
// of descending infinitely, and fillType (on load, not save) has to
// re-resolve that RefType back to the concrete *DeclaredType. This
// test pins that round-trip: if copyTypeWithRefs ever regressed to
// inline a cyclic type (or fillType lost the recursion-termination
// guarded by the sealed flag in realm.go fillType, which is what
// sealed is actually for — not mutation prevention), deploy would
// stack-overflow, Types bytes would diverge, or Sum() would panic.
//
// This is a correctness pin for the recursive subdomain rather than a
// canary on Methods-mutation; Sum()'s dispatch goes through the same
// concrete-index path covered by the base test.

type Node struct {
	Value int
	Next  *Node
}

func (n *Node) Sum() int {
	if n == nil {
		return 0
	}
	return n.Value + n.Next.Sum()
}

var Head = &Node{
	Value: 1,
	Next: &Node{
		Value: 2,
		Next: &Node{
			Value: 3,
		},
	},
}

func main(cur realm) {
	println(Head.Sum())
}

// Output:
// 6

// Types:
// Node[gno.land/r/decltyperec.Node]=
//     {
//         "ID": "gno.land/r/decltyperec.Node"
//     }
