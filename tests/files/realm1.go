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

func main() {
	key := "somekey"
	root = InnerNode{
		Key:   key,
		Left:  nil,
		Right: nil,
	}
}

// Realm:
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=VP[4:0FCED9A7AC6294E7FF7C3E571D1B7FB49A33A8BF]:
// - TEP[6DA88C34BA124C41F977DB66A4FC5C1A951708D2:#A8F4AE52F7B0F6059F449CF7D6003D67A6900ABA]
// - TEP(nil)
