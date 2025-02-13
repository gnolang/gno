// https://github.com/gnolang/gno/issues/3731

package main

type node struct {
	r []int
}

func (n *node) foo(targ, wndex int) {
	_ = n.r[targ, wndex]
}

func main() {}

// ERROR:
// files/parse_indexListExpr.gno:10:6: invalid operation: more than one index
