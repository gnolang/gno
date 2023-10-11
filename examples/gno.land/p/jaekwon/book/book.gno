/*

This is a concept for a book (basically a list).
A Book has a title, owner, etc.
A Book can be forked.
Books can be components of a virtual Person (where a Person is an interface).

*/

import "gno.land/p/demo/avl"

type Attributes map[string]string

type Book struct {
	Attributes
	entries avl.Tree
}

type Attributes struct {
	meta   avl.Tree // catchall
	author Author
	// TODO
}
