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

//----------------------------------------

// object has an action verb (method), which takes arguments
type Action struct {
	subject   // is this an id?
	verb      // i guess a string or Name type?
	object    // is this an id?
	arguments // are these primitive strings or any
}

// Authorization struct is only needed for Actions that require
// cryptographic authorization, where the Action's subject has
// a pubkey to verify signatures with.
//
// Presumably once Authorization is validated (signature checked)
// the Action becomes committed, and given a index number.
type Authorization struct {
	action     Action
	signatures []Signature
}

type Signature struct {
	account   number // or address with some extra data unknown
	sequence  number // or alternative to sequence
	signature []byte
}
