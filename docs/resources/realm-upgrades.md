# Upgrading a realm

A gno.land path is permanent: once a package is deployed, nobody can replace
its code. There are two ways around that. The **manual** way deploys the new
version at a new path and moves or proxies the state by hand; the patterns
are in `examples_ignored/r/x/manfred_upgrade_patterns/`. The **versioned
redeploy** replaces a realm's code in place and keeps its state. This page
says which fits which case, and how to use the second one.

> Status: the versioned redeploy is a proof of concept (PR #6253). It is
> consensus-breaking and ships only with a chain upgrade. Everything on this
> page is exercised by the `redeploy_*.txtar` files in
> `gno.land/pkg/integration/testdata/`; each section names its file.

## Two ways to upgrade

| | Manual: a new path per version | Versioned redeploy: the same path |
|---|---|---|
| Path and address | New (`.../v2`); callers and importers must move | Unchanged |
| State | Starts empty; copied, proxied or read through from v1 | Carried: same objects, same ObjectIDs, same deposit |
| Type identity | New (`pkgpath.Name` changes); v1 objects are foreign to v2 | Kept; v1 objects dispatch v2's methods |
| Importers | Keep calling v1 until each redeploys against v2 | Follow automatically; rebuilt in the redeploy transaction |
| Old version | Stays live, can be locked (`SetNextVersion`) | Gone, except for the func values it persisted |
| Who may do it | Anyone, at their own path | The creator (private) or the `[upgrade] authority` (public) |
| Needs | Nothing new | A chain running this VM; `version` and, for a public realm, an authority declared at the first deployment |
| Costs | Storage for two copies; a proxy or copy to write and audit | Preprocess gas for every importer, paid by the authority |

A private realm can be redeployed by its creator only, and nobody can import
it. A public realm is redeployable only if it named an `[upgrade] authority`
in its very first deployment; the authority cannot be added later, and
dropping it in a redeploy freezes the realm for good.

## Which fits: case by case

| Case | Fits | Why |
|---|---|---|
| Fix a bug or change logic in a realm nobody imports | Versioned, `private = true` | Any change to code, order or dropped globals is accepted; state kept |
| Add a feature to a realm others import | Versioned, public | Append-only keeps every importer compiling; they follow in the same transaction |
| Change a data shape (field added, type retyped) | Versioned | New type or global under a new name, converted in `migrate(from)`; manual needs a lazy copy (`upgrade_d`) and loses type identity |
| Breaking API change, importers you control | Versioned, two steps | v2 appends the new API, importers move, a later version drops the old one once nothing uses it (drop on a public realm is a follow-up) |
| Breaking API change, importers you do not control | Manual | A public redeploy is refused while any importer fails to recompile; a new path leaves them on v1 |
| Two versions live at once: A/B, a fork, independent evolution | Manual (`upgrade_a`) | A versioned redeploy has one live version per path |
| Clients should move at their own pace, then v1 locks | Manual (`upgrade_b`) | `SetNextVersion` pauses v1 when you decide; in place, every caller moves at the redeploy |
| Logic and storage in separate realms, swappable logic | Manual (`upgrade_c`, `upgrade_e`, `upgrade_f`) or versioned | Versioned removes the need for the split for one author; the split still fits pluggable implementations by different authors |
| The realm stores callbacks or closures | Either | In place they keep running the version that created them until reassigned; at a new path v1 simply stays live |
| The realm was deployed without an authority | Manual | An immutable path cannot become upgradeable |
| Upgrades decided by governance | Manual today, or a multisig as the authority | The authority must sign the `addpkg` transaction, which a realm cannot; a realm-path authority that a DAO proposal can execute is a listed follow-up |
| Today, before this ships | Manual | The versioned redeploy needs a chain upgrade |

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

Version 3 can drop `counter`, which `total` replaced. It leaves the new
code; version 1's package block still holds it for the functions that
version stored, until the old package is cleaned up:

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

## Every kind of change

What each way does with each change. "yes" means accepted and carried;
"append" means accepted only after every existing declaration; the quoted
text is the refusal the chain and `gno upgrade check` print.

| Change | Versioned, private | Versioned, public | Manual |
|---|---|---|---|
| Add a global, const, type, func | yes | append | yes, in the new path |
| Add a method | yes | append, after the type's existing methods | yes |
| Change a func or method body | yes | yes | yes |
| Change a func signature | yes | no: `declaration Count changed type from ...` | yes; importers rewrite their calls |
| Change a const value | yes | yes; importers see it after the rebuild | yes |
| Change a const type | yes | no: `declaration Limit changed type from ...` | yes |
| Reorder declarations or methods | yes | no: `declaration X moved or was removed` | yes |
| Add a new file | yes | only if it sorts after the existing files | yes |
| Rename a func or method | yes | no: a rename is a drop | yes |
| Drop a func, method or const | yes | no: `... moved or was removed` | yes |
| Drop a global | yes; it stays in the old package block until that is cleaned up | no: `declaration hits moved or was removed` | yes |
| Retype a global | no: `variable counter changed type from int to string`; new name + `migrate` | same | yes; copy by hand |
| Turn a global into a const or func | no: `counter is no longer a variable` | same | yes |
| Add a struct field or interface method | no: `type Entry changed its underlying type`; new type name + `migrate` | same | yes, as a new type; lazy copy (`upgrade_d`) |
| Drop or rename a type | no: `type Kind was removed; persisted values may still have it` | same | yes; old objects stay v1's |
| One `var (...)` mixing a carried and a new global | no: `mixes carried and new globals` | same | n/a |
| A global's initializer | skipped: the carried value wins | same | runs: the new path starts empty |
| `init()` | not run; runs on a fresh deploy of the same source | same | runs |
| `migrate(from int)` | runs once, after the carry | same | n/a; write the copy yourself |
| Add or drop an import | yes; the index follows | yes | yes |
| A stored closure or func value | keeps running the version that created it until reassigned | same | v1 stays live; nothing to do |
| Objects persisted under a type | keep their identity and get the new methods | same | foreign to the new path; convert on read |
| Importers | none (private is unimportable) | rebuilt in the transaction; one that no longer compiles refuses the redeploy | untouched; each moves when it redeploys |
| Realm address, deposit, ObjectIDs | kept | kept | new |
| The old version afterwards | code gone, archived for stored func values | same | live, lockable |

Executed by `redeploy_versioned_types.txtar`, `redeploy_public_layout.txtar`
and `redeploy_closures.txtar`.

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

## Stored functions

A func value the realm persisted, a callback in a slice, a method value, a
closure in a struct field, keeps running the code of the version that
created it: the redeploy archives the replaced source, and the stored value
resolves there, on every node alike. It reads the realm's live globals, so
old code and new code share state. To move a stored value to the new code,
reassign it, in `migrate` or later:

```go
var hooks []func() string          // carried; hooks[0] still runs v1's closure
func migrate(from int) { hooks[0] = v3hook }   // now runs v3's
```

Executed by `redeploy_closures.txtar`.

## Known gaps

- Appending a struct field is refused; it is a planned follow-up, as is a
  typed way to read a dropped global from `migrate`.
- The Constitution's rule that an immutable realm may not persist values of
  an upgradeable realm's types is not enforced.

See the ADR `gnovm/adr/pr6253_versioned_redeploy.md` for the design and the
alternatives considered.
