package main

import "gno.land/p/avl"

type String string

func main() {
	tree := avl.NewTree("", nil)
	key := "key"
	tree, _ = tree.Set(key, key)
	x, y, z := tree.Get(key)
	println(x, y, z)
}

// Output:
// 1 key true
