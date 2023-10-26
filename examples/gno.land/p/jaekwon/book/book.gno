package book

/*

GOALS:

1. Create a Book (aka List) of text,
   that is owned by a Book of Actors.

2. Allow a Book of Actors to split/fork.

3. Later, a Person is interface { Actor ...}.

*/

// TODO use this for Tiers.
// XXX open questions
// * Do we want a directory?
// * Should action memoize object/subject/args?
// * Should args be interface{}?
//   - NOTE: Amino *can* encode almost anything registered.
//   - XXX: could one pass in a pointer ref as arg?
//     - not without a directory, conflicts with GC system.
//     - with directory, any ObjectID() can be looked up.
//   - PRO: less duplicate string conversion work
//   - PRO: more flexible
//   - CON: args are mutable, potentially confusing.
//     - object/subject *should* be mutable.
//   - CON: malleability becomes user problem.
//     - so how do you sign an object? SignBytes()?
//   - CON: requires some reflection call system.
//

import "gno.land/p/demo/avl"

//----------------------------------------
// Usage example

func main() {
	// directory of subjects/objects.
	directory := NewDirectory("test")
	// action to perform.
	action := Action{}

	//----------------------------------------
	// CASES
	// Here are some usage examples.
	// For simplicity all methods panic on error.

	// CASE 1A: dummy "priv" account.
	// Access to privAccount is like access to priv keys.
	privAccount.AppendAction(action)
	directory.Get(action.object).Receive(action)

	// CASE 1B: ActionReceiveer with no signing
	// This is just a wrapper around CASE 1A.
	ar := ActionReceiveer(directory)
	sar.ReceiveBy(action, privAccount)
	// sar.ReceiveBy() impl:
	//   privAccount.AppendAction(action)
	//   ar.dir.GetObject(action.object).Receive(action)

	// CASE 2A: public crypto key account.
	// NOTE: This is NOT secure because
	// anyone can call account.Authenticate(),
	// but not anyone would complete the tx.
	privKey := _ // imagine we have the key
	signed := SignAction(action, privKey)
	account.Authenticate(signed)
	directory.Get(action.object).Receive(action)

	// CASE 2B: SignedActionReceiveer
	// SignedActionReceiveer embeds an ActionReceiveer.
	sar := SignedActionReceiveer()
	sar.Receive(signed)
	// sar.Receive() impl:
	//   signed.subject.Authenticate(signed)
	//   sar.dir.GetObject(action.object).Receive(action)
}

//----------------------------------------
// Book

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
// PrivActionBook

// This is a thought experiment to make books work for auth.
// Actions are considered authorized if appended in PrivActionBook.
//
// An PrivActionBook is meant to be owned privately by the object.
// This is similar to PrivKey in crypto; it is privileged.
//
// Actions need not necessarily be signed cryptographically
// to be authenticated in an PrivActionBook, because the test of
// authorization is merely inclusion.
//
// XXX consider:
// type CryptoActionBook struct { PubKey, PrivActionBook }
// A CryptoActionBook need not be privileged,
// perhaps anyone can append a signed action,
// XXX but how to guarantee execution?
// XXX without something like `Everything.Execute(Action)`??
//
// XXX consider:
// type ReadActionBook, a readonly ActionBook??
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
