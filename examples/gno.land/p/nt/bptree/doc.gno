// Package bptree provides a mutable B+ tree implementation for storing
// key-value data in Gno realms. It implements the same ITree interface
// as the avl package but uses a B+ tree internally for better cache
// locality and fewer pointer dereferences per operation.
//
// The fanout (maximum number of children per inner node, and maximum
// number of entries per leaf node) is configurable:
//
//	tree := bptree.NewBPTree32()    // fanout 32
//	tree := bptree.NewBPTreeN(64)   // fanout 64
//
// The zero value is usable as an empty tree with fanout 32:
//
//	var tree bptree.BPTree
//	tree.Set("key", "value")
package bptree
