package book

/*

This is a concept for a book (basically a list).
A Book has a title, owner, etc.
A Book can be forked.
Books can be components of a virtual Person (where a Person is an interface).

*/

import "gno.land/p/demo/avl"

type Book struct {
	attrs   Attributes
	entries avl.Tree
}

func (bk *Book) Append(XXX) XXX {
	XXX
}

func (bk *Book) Size() int {
	XXX
}

func (bk *Book) Get(n int) XXX {
	XXX
}

//----------------------------------------
// Attributes

type Attributes struct {
	meta   avl.Tree // catchall
	author Author
	// TODO
}

//----------------------------------------
// ActionBook
// XXX thought experiment

// XXX this doesn't make sense.
func (ab *ActionBook) Authorize(action Action) error {
	// XXX check action.sequence against ab.last
}

//----------------------------------------
// Action

// A subject acts upon an object with a verb and arguments.
// As if subject is calling object.verb(args...).
//
// The sequence is usually an incrementing number,
// but could be something else tracked by an action book.
//
// The subject and object are denoted by their respective
// paths (or URIs?).
type Action struct {
	sequence string   // typically an incrementing number
	subject  string   // subject path
	object   string   // object path
	verb     string   // verb name
	args     []string // legible args
}

func (a Action) Sequence() string { return a.sequence }
func (a Action) Subject() string  { return a.subject }
func (a Action) Object() string   { return a.object }
func (a Action) Verb() string     { return a.verb }
func (a Action) NumArgs() int     { return len(a.args) }
func (a action) Arg(n int) string { return a.args[n] }

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
