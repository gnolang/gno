# Upgrading a realm

A gno.land path is permanent: once a package is deployed, nobody can replace
its code. This page is about the two exceptions, and how to use the second
one: a **versioned redeploy**, which replaces a realm's code in place and
keeps its state.

> Status: the versioned redeploy is a proof of concept (PR #6253). It is
> consensus-breaking and ships only with a chain upgrade. Everything on this
> page is exercised by the `redeploy_*.txtar` files in
> `gno.land/pkg/integration/testdata/`; each section names its file.

## Which upgrade do you need

| You want | Use | State |
|---|---|---|
| A new version beside the old one | A new path (`.../v2`), as `r/docs/minisocial` does | Starts empty; copy or proxy by hand |
| To restart a realm of your own | `private = true`, redeploy without a `version` | Wiped: `init()` runs again over fresh globals |
| To change the code and keep the state, realm only you use | `private = true` and `version = N` | Carried |
| To change the code and keep the state, realm others import | `version = N` and `[upgrade] authority` | Carried; importers rebuilt |
| To close an upgradeable realm for good | Redeploy without the `[upgrade]` section | Carried once, then permanent |

A private realm can be redeployed by its creator only, and nobody can import
it. A public realm is redeployable only if it named an `[upgrade] authority`
in its very first deployment; the authority cannot be added later.

## Walkthrough: a private ledger

The first deployment of `gno.land/r/demo/ledger`:

```toml
# gnomod.toml
module = "gno.land/r/demo/ledger"
gno = "0.9"
private = true
version = 1
```

```go
package ledger

import "strings"

type Entry struct {
	Name string
	N    int
}

var (
	counter int
	names   []string
	last    *Entry
)

func Add(cur realm, name string) {
	counter++
	names = append(names, name)
	last = &Entry{Name: name, N: counter}
}

func Names() string { return strings.Join(names, ",") }
```

Deploy and use it as usual:

```sh
gnokey maketx addpkg -pkgdir ./v1 -pkgpath gno.land/r/demo/ledger ... creator
gnokey maketx call -pkgpath gno.land/r/demo/ledger -func Add -args alice ... creator
gnokey maketx call -pkgpath gno.land/r/demo/ledger -func Add -args bob ... creator
```

Version 2 keeps every global, adds one, adds a method to the carried type,
and derives the new global from the old ones:

```toml
# gnomod.toml
module = "gno.land/r/demo/ledger"
gno = "0.9"
private = true
version = 2
```

```go
package ledger

import (
	"strconv"
	"strings"
)

type Entry struct { // same name, same fields: keeps its identity
	Name string
	N    int
}

// New method: the Entry objects version 1 persisted answer with it.
func (e *Entry) Describe() string {
	return e.Name + "#" + strconv.Itoa(e.N) + " via v2"
}

var (
	counter int      // same name, same type: carried; an initializer here is skipped
	names   []string // carried
	last    *Entry   // carried
	total   int64    // new: starts at zero
)

// migrate runs in place of init, after the carry, handed the version it
// replaces. Every carried global is already here under its old name.
func migrate(from int) {
	if from == 1 {
		total = int64(counter) * 1000
	}
}

func Add(cur realm, name string) {
	counter++
	total += 1000
	names = append(names, name)
	last = &Entry{Name: name, N: counter}
}

func Names() string         { return strings.Join(names, ",") }
func Total() int64          { return total }
func LastDescribed() string { return last.Describe() }
```

Redeploy with the same command, from the same creator:

```sh
gnokey maketx addpkg -pkgdir ./v2 -pkgpath gno.land/r/demo/ledger ... creator
gnokey query vm/qeval --data 'gno.land/r/demo/ledger.Names()'          # "alice,bob"
gnokey query vm/qeval --data 'gno.land/r/demo/ledger.Total()'          # 2000
gnokey query vm/qeval --data 'gno.land/r/demo/ledger.LastDescribed()'  # "bob#2 via v2"
```

The keeper accepted it because the live package is version 1, the new one
is version 2, and the signer is the creator. `init()` did not run;
`migrate(1)` did. The objects are the same objects, with the same ObjectIDs
and the same storage deposit; nothing was copied.

Version 3 can drop `counter`, which `total` replaced. Its heap item is
released, and whatever it owned with it:

```go
var (
	names []string
	last  *Entry
	total int64
)
```

Executed by `redeploy_versioned_state.txtar`.

## Walkthrough: a public vault with importers

A public realm names who may redeploy it:

```toml
module = "gno.land/r/demo/vault"
gno = "0.9"
version = 1

[upgrade]
authority = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
```

```go
package vault

import "strconv"

type Receipt struct{ Amount int64 }

func (r *Receipt) String() string { return "receipt for " + strconv.FormatInt(r.Amount, 10) + " (v1)" }

var total int64

func Deposit(cur realm, n int64) *Receipt {
	total += n
	return &Receipt{Amount: n}
}

func Total() int64 { return total }
```

A client realm imports it and keeps the receipts it is handed. The client
must itself be upgradeable or private: an immutable realm may not import an
upgradeable one, because code it depends on could change under it
(CONSTITUTION, Realm Upgrading).

```go
package client

import "gno.land/r/demo/vault"

var receipts []*vault.Receipt

func Deposit(cur realm, n int64) {
	receipts = append(receipts, vault.Deposit(cross(cur), n))
}

func First() string { return receipts[0].String() }
```

Importers compile `vault.Deposit` to a slot in the vault's package block and
`r.String()` to an index in `Receipt`'s method list. So a public redeploy
may only **append**: new declarations go after the existing ones, new
methods after the existing ones, and nothing moves, disappears or changes
type. Version 2:

```go
package vault

import "strconv"

type Receipt struct{ Amount int64 }

func (r *Receipt) String() string { return "receipt for " + strconv.FormatInt(r.Amount, 10) + " (v2)" }
func (r *Receipt) Doubled() int64 { return r.Amount * 2 } // appended

var total int64

func Deposit(cur realm, n int64) *Receipt { // body changed, signature kept
	n *= 2
	total += n
	count++
	return &Receipt{Amount: n}
}

func Total() int64 { return total }

var count int // appended

func migrate(from int) { count = 1 } // one deposit happened under v1

func Count() int { return count } // appended
```

Only the authority may send it:

```sh
gnokey maketx addpkg -pkgdir ./v2 -pkgpath gno.land/r/demo/vault ... authority
gnokey query vm/qeval --data 'gno.land/r/demo/client.First()'   # "receipt for 5 (v2)"
```

The receipt version 1 handed out now answers with version 2's method. The
client was not redeployed: the redeploy re-preprocessed every importer of
the vault, direct or through other importers, in the same transaction, so
every node runs the same code whether or not it has restarted since. The
authority pays preprocessing gas for each importer, as for a deploy; a
widely imported realm costs more to upgrade.

To freeze the realm, redeploy once more without the `[upgrade]` section.
The state is carried that last time, and the path is permanent afterwards.

Executed by `redeploy_public_authority.txtar` (authority, append rule,
freezing) and `redeploy_public_importers.txtar` (importers rebuilt, before
and after a restart).

## What a redeploy may change

The carry keys globals by name and static type, and declared types by name
and underlying type, because the realm's own persisted objects still carry
both. Each refusal leaves the realm as the previous version wrote it, and
names the offender.

| Change | Private | Public | Refusal |
|---|---|---|---|
| Add a global, type, func, method | yes | yes, appended | |
| Change a func or method body | yes | yes | |
| Change or drop a method | yes | change yes, drop no | `method Len moved or was removed` |
| Drop a global | yes | no | `declaration hits moved or was removed` |
| Reorder declarations | yes | no | `declaration Receipt moved or was removed` |
| Change a func signature or const type | yes | no | `declaration Count changed type from ...` |
| Retype a global | no | no | `variable counter changed type from int to string` |
| Turn a global into a const or func | no | no | `counter is no longer a variable` |
| Drop or rename a declared type | no | no | `type Kind was removed; persisted values may still have it` |
| Change a type's fields or interface methods | no | no | `type Entry changed its underlying type` |
| One `var (...)` naming a carried and a new global | no | no | `mixes carried and new globals` |

Executed by `redeploy_versioned_types.txtar` and `redeploy_public_layout.txtar`.

A global is carried whole: its initializer in the new source is skipped.
Give a changed shape a new name and fill it in `migrate`, then drop the old
one in the version after.

## The gnomod.toml rules

| Rule | Refusal |
|---|---|
| `version` must step by exactly one over the live realm | `redeploying ... needs version = 3 in gnomod.toml, got 9` |
| `version` needs `private = true` or an authority | `version in gnomod.toml requires private = true or an [upgrade] authority` |
| An authority needs a `version` and a valid address | `an [upgrade] authority in gnomod.toml requires version` |
| Only realm paths may be private or upgradeable | `private or upgradeable packages must be realm packages` |
| A private realm stays private | `a private package cannot be overridden by a public package` |
| An upgradeable realm cannot turn private | `an upgradeable package cannot become private` |
| Only the authority, or a private realm's creator, may redeploy | `only the upgrade authority g1... may redeploy ...` |
| A frozen or immutable realm is permanent | `package already exists` |
| An immutable realm may not import an upgradeable one | `is upgradeable and cannot be imported by an immutable realm` |

Version unset on both sides is the old private redeploy, which starts the
globals over. `version = 1` over an unversioned private realm opts in and
carries them, with `migrate(0)`.

Executed by `redeploy_gnomod_rules.txtar`.

## Check before you deploy

The rules above are the chain's; two ways to run them first:

```sh
gno upgrade check ./v1 ./v2            # no node: the two directories, same module path
gnokey maketx addpkg ... -simulate only  # a node: the redeploy against the live version
```

`gno upgrade check` type-checks both versions, preprocesses them and runs the
same comparison the chain runs, printing the same refusal. `-simulate only`
runs the real transaction on the node without broadcasting it, so it also
covers the authority, the live version number and the importers.

## `migrate(from int)`

- Declared as `migrate(from int)` or `migrate(cur realm, from int)`; any
  other signature is refused at type check. Like `init`, it may be declared
  more than once and cannot be called.
- Runs in place of `init()` on every versioned redeploy that declares it,
  after the carry, before anything else. `from` is the version replaced.
- A migration left in the source runs again on the next version. Guard it
  on `from`, or remove it when it has done its job.

## Known gaps

- A persisted func value, method value or closure still runs the code of
  the version that created it.
- Appending a struct field is refused; it is a planned follow-up, as is a
  typed way to read a dropped global from `migrate`.
- The Constitution's rule that an immutable realm may not persist values of
  an upgradeable realm's types is not enforced.

See the ADR `gnovm/adr/pr6253_versioned_redeploy.md` for the design and the
alternatives considered.
