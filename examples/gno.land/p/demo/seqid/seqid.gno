package main

import (
	"gno.land/p/demo/avl"
)

var tree *avl.Tree

func Set(key string, value int) {
	// tree.Set takes in a string key, and an interface{} value
	tree.Set(key, value)
}

func Get(key string) {
	// tree.Get returns the  interface{} value at given key,
	// and a bool to signify the existence of the key-value pair
	rawValue, exists := tree.Get(key)
	if !exists {
		panic("value at given key does not exist")
	}
	// Get returns the
	value := rawValue.(int)

}
