# Gno

At first, there was Bitcoin, out of entropy soup of the greater All.
Then, there was Ethereum, which was created in the likeness of Bitcoin,
but made Turing complete.

Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
Then came Gno upon Cosmos and there spring forth Gnoland,
simulated by the Gnomes of the Greater Resistance.

## Language Features

 * Like interpreted Go, but more ambitious.
 * Completely deterministic, for complete accountability.
 * Transactional persistence across data realms.
 * Designed for concurrent blockchain smart contracts systems.
 
## Status

_Update Feb 13th, 2021: Implemented Logos UI framework._

This is still a work in a progress, though much of the structure of the interpreter
and AST have taken place.  Work is ongoing now to demonstrate the Realm concept before
continuing to make the tests/files/\*.go tests pass.

Make sure you have >=[go1.15](https://golang.org/doc/install) installed, and then try this: 

```bash
> git clone git@github.com:gnolang/gno.git
> cd gno
> go mod download github.com/davecgh/go-spew
> go test tests/*.go -v -run="Test/realm.go"
```

## Ownership 

In Gno, all objects are automatically persisted to disk after every atomic
"transaction" (a function call that must return immediately). When new objects
are associated with a "ownership tree" which is maintained overlaying the
possibly cyclic object graph.  The ownership tree is composed of objects
(arrays, structs, maps, and blocks) and derivatives (pointers, slices, and so
on) with struct-tag annotations to declare the ownership tree.

If an object hangs off of the ownership tree, it becomes included in the Merkle
root, and is said to be "real".  The Merkle-ized state of reality gets updated
with state transition transactions; during such a transaction, some new
temporary objects may "become real" by becoming associated in the ownership
tree (say, assigned to a struct field or appended to a slice that was part of
the ownership tree prior to the transaction), but those that don't get garbage
collected and forgotten.

```go
type Node interface {
    ...
}

type InnerNode struct {
	Key       Key
	LeftNode  Node `gno:owned`
	RightNode Node `gno:owned`
}

type LeafNode struct {
	Key       Key  `gno:owned`
	Value     interface{}
}
```

In the above example, some fields are tagged as owned, and some are not.  An
InnerNode structure may own a LeftNode and a RightNode, and it may reference or
own a Key.  The Key is already owned by the left most LeafNode of the right
tree, so the InnerNode cannot own it.  The LeafNode can contain a reference or
own any value.  In other words, if nobody else owns a value, the LeafNode will.

We get a lack-of-owner problem when the ownership tree detaches an object
referred elsewhere (after running a statement or set of statements):

```
    A, A.B, A.C, and D are owned objects.
    D doesn't own C but refers to it.

	   (A)   (D)
	   / \   ,
	  /   \ ,
	(B)   (C)

    > A.C = nil

	   (A)       (D)
	   / \       ,
	  /   \     ,
	(B)    _   C <-- ?
```

Options:

 1. unaccounted object error (default)
   - can't detach object unless refcount is 0.
   - pushes problem to ownership tree manipulation logic.
   - various models are possible for maintaining the ownership tree, including
     reference-counting (e.g. by only deleting objects from a balanced search
tree when refcount reaches 1 using a destroy callback call); or more lazily
based on other conditions possibly including storage rent payment.
   - unaccounted object error detection is deferred until after the
     transaction, allowing objects to be temporarily unaccounted for.

 2. Invalid pointer
   - basically "weak reference pointers" -- OK if explicit and exceptional.
   - across realms it becomes a necessary construct.
   - will implement for inter-realm pointers.

 3. Auto-Inter-Realm-Ownership-Transfer (AIR-OT)
   - within a realm, refcounted garbage collection is sufficient.
   - automatic ownership transfers across realms may be desirable.
   - requires bidirectional reference tracking.

## Realms

Gno is designed with blockchain smart contract programming in mind.  A
smart-contract enabled blockchain is like a massive-multiuser-online
operating-system (MMO-OS). Each user is provided a home package, for example
"gno.land/r/username". This is not just a regular package but a "realm
package", and functions and methods declared there have special privileges.

Every "realm package" should define at least one package-level variable:

```go
// PKGPATH: gno.land/r/alice
package alice
var root interface{}

func UpdateRoot(...) error {
  root = ...
}
```

Here, the root variable can be any object, and indicates the root node in
the data realm identified by the package path "gno.land/r/alice".

Any number of package-level values may be declared in a realm; they are
all owned by the package and get merkle-hashed into a single root hash for
the package realm.

The gas cost of transactions that modify state are paid for by whoever
submits the transaction, but the storage rent is paid for by the realm.
Anyone can pay the storage upkeep of a realm to keep it alive.

## Merkle Proofs

Ultimately, there is a single root hash for all the realms.

From hash.go:

```
//----------------------------------------
// ValueHash
//
// The ValueHash of a typed value is a unique deterministic
// accountable fingerprint of that typed value, and can be used
// to prove the value or any part of its value which is
// accessible from the Gno language.
//
// `ValueHash := lh(ValueImage)`
// `ValueImage:`
//   `= 0x00` if nil value.
//   `= 0x01,varint(.) if fixed-numeric.
//   `= 0x02,sz(bytes)` if variable length bytes.
//   `= 0x03,sz(TypeID),vi(*ptr)` if non-nil ptr.
//   `= 0x04,sz(OwnerID),sz(ElemsHash),mod,ref` if object.
//   `= 0x05,vi(base),off,len,max if slice.
//   `= 0x06,sz(TypeID)` if type.
//
// `ElemsHash:`
//   `= lh(ElemImage)` if object w/ 1 elem.
//   `= ih(eh(Left),eh(Right))` if object w/ 2+ elems.
//
// `ElemImage:`
//   `= 0x10` if nil interface.
//   `= 0x11,sz(ObjectID),sz(TypeID)` if borrowed.
//   `= 0x12,sz(ObjectID),sz(TypedValueHash)` if owned.
//   `= 0x13,sz(TypeID),sz(ValueHash)` if other.
//    - other: prim/ptr/slice/type/typed-nil.
//    - ownership passed through for pointers/slices/arrays.
//
// `TypedValueHash := lh(sz(TypeID),sz(ValueHash))`
//
// * eh() are inner ElemsHashs.
// * lh() means leafHash(x) := hash(0x00,x)
// * ih() means innerHash(x,y) := hash(0x01,x,y)
// * pb() means .PrimitiveBytes().
// * sz() means (varint) size-prefixed bytes.
// * vi() means .ValueImage().Bytes().
// * off,len,max and other integers are varint encoded.
// * len(Left) is always 2^x, x=0,1,2,...
// * Right may be zero (if len(Left+Right) not 2^x)
//
// If a pointer value is owned (e.g. field tagged "owned"), the
// pointer's base if present must not already be owned.  If a
// pointer value is not owned, but refers to a value that has a
// refcount of 1, it is called "run-time" owned, and the value
// bytes include the hash of the referred value or object as if
// owned; the value bytes also include the object-id of the
// "run-time" owned object as if it were persisted separately
// from its base object, but implementations may choose to
// inline the serialization of "run-time" owned objects anyway.
//
// If an object is owned, the value hash of elements is
// included, otherwise, the value hash of elements is not
// included except for objects with refcount=1.  If owned but
// any of the elements are already owned, or if not owned but
// any of the elements have refcount>1, image derivation
// panics.
```

## Logos Browser

[Logos](/logos) is a Gno object browser.  The modern browser as well as the
modern javascript ecosystem is from a security point of view, completely fucked.
The entire paradigm of continuously updating browsers with incrementally added
features is a security nightmare.

The Logos browser is based on a new model that is vastly simpler than HTML.
The purpose of Logos is to become a fully expressive web API and implementation
standard that does most of what HTML and the World Wide Web originally intended
to do, but without becoming more complex than necessary.

## Concurrency

Initially, we don't need to implement routines because realm package functions
provide all the inter-realm functionality we need to implement rich smart
contract programming systems.  But later, for various reasons including
long-running background jobs, and parallel concurrency, Gno will implement
deterministic concurrency as well.

Determinism is supported by including a deterministic timestamp with each
channel message as well as periodic heartbeat messages even with no sends, so
that select/receive operations can behave deterministically even in the
presence of multiple channels to select from.

## Completeness

Software projects that don't become complete are projects that are forever
vulnerable.  One of the requisite goals of the Gno language and related 
software libraries like Logos is to become finished within a reasonable timeframe.

## How to become a Gnome

First, read the license.  The license doesn't take away any of your rights,
but it gives the Gno project rights to your contributions.

Contributions in the form of completed work in a pull request or issue
or comments are welcome and encouraged, especially if you are interested
in joining the project.

The biggest bottleneck in these sorts of projects is finding the right people
with the right skillset and character; and my highest priority besides coding,
is to find the right contributors.  If you can grok the complexities of this
and related projects without hand holding, and you understand the implications
of this project and are aligned with its mission, read on.

The Gno Foundation is a non-profit with missions originally stated in the [Virgo Project](https://github.com/virgo-project/virgo).
The Gno Foundation, which owns the IP to the Gno Works, proposes the following:

* The Gno Foundation permits the Gnoland chain the usage of the Gno Works.
* The Gnoland chain's staking token is given equally to three bodies:

  - 1/3 of GNOTs to the Gno Foundation.
  - 1/3 of GNOTs to the Gno Community.
  - 1/3 of GNOTs to an opinionated spoonful of ATOMs.
  
Spoonful of atoms to be weighted according to voting history, such that those
who voted in favor of good proposals and against bad proposals as judged by the
Gno Foundation, as well as those who were active in voting, are given favor.
The weighting may be such that some ATOM holders receive no GNOTs.
This is not a fork of the Cosmos Hub, but a new chain, so the distribution is
entirely at the Gno Foundation's discretion, and the foundation has strong opinions.

The Gno Community is determined by the creation and finalization of the project,
as determined by the Gno Foundation according to community contributions.

## Reading the code

Gno's code has been written with extensive comments that explain what each 
file does. Eventually, each function will be commented in the same manner. 

You can learn a great deal from reading Gnocode, and it's recommended that
both users and developers have a look.

## Contact

If you can read this, the project is evolving (fast) every day.  Check
"github.com/gnolang/gno" and @jaekwon frequently.

The best way to reach out right now is to create an issue on github, but this
will change soon.
