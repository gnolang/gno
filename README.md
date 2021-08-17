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

_Update Aug 16th, 2021: basic file tests pass_

Basic Go file tests now pass.  Working on realm/ownership logic under tests/files/zrealm\*.go.

_Update Jul 22nd, 2021: create pkgs/crypto/keys/client as crypto wallet._

The new wallet will be used for signed communications.

_Update Jul ?, 2021: Public invited to contribute to Gnolang/files tests.

_Update Feb 13th, 2021: Implemented Logos UI framework._

This is a still a work in a progress, though much of the structure of the interpreter
and AST have taken place.  Work is ongoing now to demonstrate the Realm concept before
continuing to make the tests/files/\*.go tests pass.

Make sure you have >=[go1.15](https://golang.org/doc/install) installed, and then try this: 

```bash
> git clone git@github.com:gnolang/gno.git
> cd gno
> make test
```

## Ownership 

_TODO: update documentation on ownership, which is being worked on now_

In Gno, all objects are automatically persisted to disk after every atomic
"transaction" (a function call that must return immediately.) when new objects
are associated with a "ownership tree" which is maintained overlaying the
possibly cyclic object graph (NOTE: cyclic references for persistence not
supported at this stage).  The ownership tree is composed of objects (arrays,
structs, maps, and blocks) and derivatives (pointers, slices, and so on) with
optional struct-tag annotations to define the ownership tree.

If an object hangs off of the ownership tree, it becomes included in the Merkle
root, and is said to be "real".  The Merkle-ized state of reality gets updated
with state transition transactions; during such a transaction, some new
temporary objects may "become real" by becoming associated in the ownership
tree (say, assigned to a struct field or appended to a slice that was part of
the ownership tree prior to the transaction), but those that don't get garbage
collected and forgotten.

In the first release of Gno, all fields are owned in the same realm, and no
cyclic dependencies are allowed outside the bounds of a realm transaction (this
will change in phase 2, where ref-counted references and weak references will
be supported).

We get a lack-of-owner problem when the ownership tree detaches an object
referred elsewhere (after running a statement or set of statements):

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

The gas cost of transactions that modify state are paid for by whoever
submits the transaction, but the storage rent is paid for by the realm.
Anyone can pay the storage upkeep of a realm to keep it alive.

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

## Logos Browser

[Logos](/logos) is a Gno object browser.  The modern browser as well as the
modern javascript ecosystem is from a security point of view, completely fucked.
The entire paradigm of continuously updating browsers with incrementally added
features is a security nightmare.

The Logos browser is based on a new model that is vastly simpler than HTML.
The purpose of Logos is to become a fully expressive web API and implementation
standard that does most of what HTML and the World Wide Web originally intended
to do, but without becoming more complex than necessary.

## Completeness

Software projects that don't become complete are projects that are forever
vulnerable.  One of the requisite goals of the Gno language and related 
software libraries like Logos is to become finished within a reasonable timeframe.

## How to become a Gnome

First, read the license.  The license doesn't take away any of your rights, but
it gives the Gno project rights to your contributions.

Contributions in the form of completed work in a pull request or issue or
comments are welcome and encouraged, especially if you are interested in
joining the project.

The biggest bottleneck in these sorts of projects is finding the right people
with the right skillset and character; and my highest priority besides coding,
is to find the right contributors.  If you can grok the complexities of this
and related projects without hand holding, and you understand the implications
of this project and are aligned with its mission, read on.

The Gno Foundation is a non-profit with missions originally stated in the
[Virgo Project](https://github.com/virgo-project/virgo).  The Gno Foundation,
which owns the IP to the Gno Works, proposes the plan as laid out in the plan
file.

## Contact

If you can read this, the project is evolving (fast) every day.  Check
"github.com/gnolang/gno" and @jaekwon frequently.

The best way to reach out right now is to create an issue on github, but this
will change soon.
