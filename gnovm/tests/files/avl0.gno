package main

import "gno.land/p/nt/avl"

type String string

func main() {
	node := avl.NewNode("", nil)
	key := "key"
	node, _ = node.Set(key, key)
	x, y, z := node.Get(key)
	println(x, y, z)
}

// Output:
// 1 key true
