// PKGPATH: gno.land/r/test
package test

var root Node

type Node interface{}
type Key interface{}

type InnerNode struct {
	Key   Key
	Left  Node `gno:owned`
	Right Node `gno:owned`
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	root = InnerNode{
		Key: "new",
	}
}

// Realm:
// XXX
