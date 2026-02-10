package main

import "gno.land/p/nt/avl"

type String string

func main() {
	tree := avl.Tree{}
	key := "key"
	_ = tree.Set(key, key)
	y, z := tree.Get(key)
	println(y, z)
}

// Output:
// key true
