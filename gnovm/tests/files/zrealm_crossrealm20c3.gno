// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type Node struct {
	*Node
}

var root interface{}

func main() {
	node := &Node{}
	node.Node = node

	root = node
	println("ok")
}

// Output:
// ok
