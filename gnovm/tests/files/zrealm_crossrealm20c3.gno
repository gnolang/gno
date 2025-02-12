// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

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
