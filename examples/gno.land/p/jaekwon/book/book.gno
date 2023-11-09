package book

/*

GOALS:

0. Define Action, Actor, Object.

1.a Create a Book (aka List) of text,
   that is owned by a Book of Actors.

1.b Create ActionBook.

2. Allow a Book of Actors to split/fork.

3. Later, a Person is interface { Actor ...}.

*/

// TODO use this for Tiers.

import "gno.land/p/demo/avl"

//----------------------------------------
// Usage example

func main() {
	// directory of subjects/objects.
	directory := NewDirectory("test")
	// action to perform.
	action := Action{}
	// XXX
}

//----------------------------------------
// Book

// A Book is a basic data structure.
// It is like a linked list, but indexable from 0.
// XXX XXX how to do this w/ avl.Tree?
// TODO book merges?
// TODO flesh out book.
type Book struct {
	attrs   Attributes
	entries avl.Tree
}

var _ Object = &Book{}

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
// PrivActionBook

// This is a thought experiment to make books work for
// auth.  Actions are considered authorized if appended in
// PrivActionBook.
//
// An PrivActionBook is meant to be owned privately by the
// object.  This is similar to PrivKey in crypto; it is
// privileged.
//
// Actions need not necessarily be signed cryptographically
// to be authenticated in an PrivActionBook, because the
// test of authorization is merely inclusion.
//
// TODO implement:
// type CryptoActionBook struct { PubKey, PrivActionBook }
// A CryptoActionBook need not be privileged,
// perhaps anyone can append a signed action,
//
// Also, PrivActionBook.
// Also, ReadActionBook.
// All of these are backed by the same underlying "book".
type PrivActionBook struct {

	// All actions must have this object.
	object std.FullAddress

	// Maybe PrivActionBook *is* Book?
	// not sure yet.
	book *Book

	// Number of actions to keep around.
	capacity int

	// Validates sequences based on sequenceAccum,
	// which is accumulated from sequences seen.
	sequenceStrategy SequenceStrategy

	// Typically the last sequence value.
	// The type of value depends on SequenceStrategy.
	// This field allows the PrivActionBook to prune
	// all previous Actions while preserving sequencing.
	// XXX string or TextMarshaller() or Stringer() or?
	sequenceAccum string
}

func NewPrivActionBook() *PrivActionBook {
	// XXX
}

// If the action is valid, append to PrivActionBook,
// thereby making it officially authorized.
// The execution of action generally should happen
// atomically with authorization by caller.
//
//	if err := pab.Append(action); err != nil {
//	    execute(action)
//	}
func (pab *PrivActionBook) Append(action Action) error {
	// XXX copy action.

	// XXX check action.sequence against ab.last
	// XXX if good, append and return nil
	// XXX otherwise return error

	// XXX match action.object with pab.object.
	// XXX set action.object = nil for space.

	// XXX check capacity
}

func (pab *PrivActionBook) Len() int {
	return pab.book.Len()
}

func (pab *PrivActionBook) Cap() int {
	return pab.book.Cap()
}

// XXX Not sure why this would be useful,
// XXX except to show clients previous actions,
// XXX but either way developers should not rely on it
// XXX for transactional logic.
func (pab *PrivActionBook) Get(idx int) Action {
	// XXX fetch action from pab.book.Get()
	// XXX copy action
	// XXX set copy.object = pab.object
	// XXX return copy
}

// XXX SequenceStragegy.Name()?
// XXX or just make enums?
// XXX Either way need to make globally unique lookup.
func (pab *PrivActionBook) SequenceStrategy() SequenceStrategy {
	// XXX
}

// XXX clients will need this to sign,
// XXX especially after device reset.
func (pab *PrivActionBook) SequenceAccum() string {
	// XXX
}

//----------------------------------------
// misc

func hasPrefix(str, prefix string) bool {
	if len(str) <= len(prefix) {
		return false
	}
	return str[:len(prefix)] == prefix
}
