// PKGPATH: gno.land/r/test
package test

// A self-cycle the save guard escaped is read, written into, wrapped,
// unwrapped and dropped by later finalizes in the same transaction.

type Node struct {
	Next *Node
	V    int
}

var (
	Root *Node
	Keep *Node
)

func init() { Root = &Node{} }

func build(cur realm) *Node {
	x := &Node{}
	x.Next = x
	Root.Next = x
	Root = nil
	return x
}

func keep(cur realm, x *Node) { Keep = x }
func bump(cur realm)          { Keep.V++; Keep.Next.V++ }
func wrap(cur realm)          { Keep = &Node{Next: Keep, V: 9} }
func unwrap(cur realm)        { Keep = Keep.Next }
func drop(cur realm)          { Keep.Next = nil; Keep = nil }

func main(cur realm) {
	x := build(cross(cur))
	println(x.Next == x)
	x.V = 40
	keep(cross(cur), x)
	bump(cross(cur))
	println(Keep.V)
	wrap(cross(cur))
	println(Keep.V, Keep.Next.V)
	unwrap(cross(cur))
	drop(cross(cur))
	println(Keep == nil)
}

// Output:
// true
// 42
// 9 42
// true
