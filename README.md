# Gno: Language of Resistance.

 * Like interpreted Go, but more ambitious.
 * Completely deterministic, for complete accountability.
 * Transactional persistence across data realms.
 * Designed for concurrent blockchain smart contracts systems.
 
## Status

This is a still a work in a progress, though much of the structure of the interpreter
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
"transaction" (a function call that must return immediately.) when new objects
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

Every "realm package" should define at last one package-level variable:

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

## Contributing

Contributions in the form of completed work in a pull request or issue
or comments are welcome and encouraged, especially if you are interested
in joining the project.

The biggest bottleneck in these sorts of projects is finding the right people
with the right skillset and character; and my highest priority besides coding,
is to find the right contributors.  If you can grok the complexities of this
and related projects without hand holding, and you understand the implications
of this project and are aligned with its mission, keep in touch.

## Reading the code

Gno's code has been written with extensive comments that explain what each 
file does. Eventually, each function will be commented in the same manner. 

You can learn a great deal from reading Gnocode, and it's recommended that
both users and developers have a look. 

## Updates

If you can read this, the project is evolving (fast) every day.  Check
"github.com/gnolang/gno" and @jaekwon frequently.

The best way to reach out right now is to create an issue on github, but this
will change soon.
