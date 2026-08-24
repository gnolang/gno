package main

import "gno.land/p/nt/avl/v0"

type String string

func main() {
	tree := avl.Tree{}
	key := "key"
	_ = tree.Set(key, key)
	y := tree.Get(key)
	println(y)
}

// Output:
// key
