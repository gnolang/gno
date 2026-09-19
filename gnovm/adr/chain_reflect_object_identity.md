# chain/reflect: object identity as an injected primitive

## Context

A realm that creates something and needs a unique name for it has, today, only
names it chooses itself: a package path, a symbol, a counter it keeps, or an ID
its caller hands it. Anything a realm chooses, a realm can choose twice.

`grc20.NewToken` is the case that surfaced it. It built `Token.id` from
`<pkg-path>.<symbol>.<caller-supplied-seqid>`, with the sequence number coming
from the calling realm, so one realm could issue two independent tokens under
one identifier and their `Transfer`/`Approval` events became indistinguishable.
That is #6026.

The VM already maintains the one name a realm cannot choose. `ObjectID` is
`{PkgID, NewTime}`: the realm the object belongs to, and a tick of that realm's
storage clock, minted by the VM while persisting the object. It is unique per
object, stable across transactions, and unforgeable from inside a realm.

Two earlier attempts:

- #6101 added `runtime.NewRealmID()`, returning `<realm-path>:<realm-time>`.
  Rejected in review: a second identity scheme that no other VM concept
  consumes, and a private counter wearing an ObjectID's shape.
- #6139 derives a bech32 address from the ObjectID and exposes it as
  `chain/runtime.ObjectAddress(v interface{}) string`, together with a breaking
  `grc20` and `grc20reg` redesign, across 50 files.

## Decision

Add `chain/reflect`, an injected stdlib whose entire subject is object
identity, and expose the ObjectID itself.

```gno
type ObjectID struct { /* unexported */ }

func (oid ObjectID) String() string
func (oid ObjectID) IsZero() bool

func ObjectIDOf(v interface{}) (ObjectID, bool)
func HasIdentity(v interface{}) bool
```

Five decisions, each answering something #6139 answered differently.

### 1. An ID, not an address

`ObjectIDOf` returns the ObjectID. It does not derive an address from it.

#6139 hashes the ObjectID into a `g1…` and its own ADR then states: "The
address has no signing key. It is an identifier, not currently a payable
account." Every wallet, indexer and explorer in the ecosystem reads `g1…` as an
account, and this one can never hold or send anything. Worse, the derivation is
irreversible in practice: the preimage layout, the `PkgID` flag nibble and the
`NewTime` semantics all become consensus state the moment the first address
lands in an event, so changing any of them later is a migration.

So the irreversible half would ship now, for an "objects become potential
accounts" capability that does not exist yet and is not in that PR. Splitting
them costs nothing: `ObjectID.Address()` can be added the day the banker can
credit an object, on top of an ObjectID that is already in use, with the
preimage question answered then by the people building the accounts.

### 2. ObjectID is a type, not a string

`ObjectID` is a struct with one unexported field, so `ObjectIDOf` is the only
way to obtain a non-zero one.

#6139 returns a bare `string`. A realm can therefore fabricate an "object id",
and any consumer taking a `string` id cannot tell a VM-issued identity from one
the realm typed out. A named string type (`type ObjectID string`) has the same
hole. Given that the entire point is "an identifier the realm does not choose",
making it unforgeable by construction rather than by convention is the property,
not a detail. It stays comparable and map-key safe.

### 3. The three outcomes are separated, and none of them is a panic

| v | `HasIdentity` | `ObjectIDOf` ok |
|---|---|---|
| pointer to a standalone object, persisted | true | true |
| pointer to a standalone object, created by the running call | true | false |
| pointer into a struct field or array element | false | false |
| slice | false | false |
| scalar, nil pointer, nil interface | false | false |

`ObjectIDOf` never panics and never returns a non-zero ID it is unsure of.
`HasIdentity` splits its two false cases into "ask again after the realm
boundary" and "nothing to wait for".

#6139 has one function with three behaviours: `""` for an unpersisted object,
`""` for a pointer into a container or a slice, and a panic for a scalar. A
caller cannot tell the first two apart, which are the two that need different
handling, and the inconsistency is not derivable from the signature.

### 4. The stamping window is reported, not papered over

`PkgID` is stamped by the allocator at creation; `NewTime` only by
`Realm.assignNewObjectID`, reached from `FinalizeRealmTransaction`, which runs
when a realm frame returns. So an object created by the running call has no
complete ObjectID yet, and no primitive can conjure one: attaching to the
ownership tree *is* what mints it. `GetIsReal()` is `ID.IsFinalized()` and
`incRefCreatedDescendants` guards on the same predicate, so stamping early
makes an object look real while unsaved and skips it during finalization. That
is a lifecycle change, well past this problem.

What is a choice is how the window is reported. `ObjectIDOf` returns
`ok == false`, so the caller has to decide at the point where deciding is still
possible.

#6139 returns `""`. Events are immutable from `chain.Emit`, so every object
created and emitted before finalization carries `id=""` in the tx event log
permanently, and `""` is a perfectly good map key: unrelated objects alias onto
one another. Its ADR's mitigation is the sentence "Callers must not use '' as an
identity or a map key". Under `ok`, the two things a realm can do are visible in
the signature: emit the ID from a later call, or `if !ok { panic(...) }` and
make the caller split the work in two.

### 5. Deliberately narrow about what has an identity

Only a pointer to a standalone heap item, which is what `new(T)` and `&T{...}`
produce.

A pointer into a struct field or an array element resolves to the container, so
lending it the container's identity makes every sibling answer the same ID. A
slice resolves to its backing array, shared by every view of it. A struct, map,
array or func arriving by value may be the persisted object or a copy of it
depending on how it got there.

The asymmetry is what drives this: teaching a kind of value to report its
identity later is compatible, while withdrawing an answer is not, because by
then that ID is in somebody's event log. So an unsettled case reports false.

## Consequences

- No change to `misc/genstd`. The native's Go parameter is `gno.TypedValue`,
  which already matches any Gno parameter type and links through with no
  Go2Gno conversion, while the Gno side declares `interface{}` and
  `GnoTypeExpression` maps it to `gno.AnyT()`. Pinned by the existing
  `misc/genstd/testdata/linkFunctions_TypedValue`.

  #6139 instead widens `mapping.isTypedValue` so a *Go* `any`/`interface{}`
  parameter is treated as a raw TypedValue. Beyond being unnecessary, that
  makes every future native declaring a Go `any` parameter silently receive a
  TypedValue rather than a converted Go value.

- The native resolves pointer bases itself and never calls
  `TypedValue.GetFirstObject`, which panics outright on a package `RefValue`
  and on a bare `*HeapItemValue`. Every shape is handled or reported as no
  identity; none aborts the transaction.

- `ObjectID.String()` is the VM's own `<realm-id>:<clock-tick>` spelling, so an
  ID in an event can be matched against a storage dump. It is documented as
  unparseable: the halves are the VM's business.

- Stdlib `.gno` source is genesis state, so adding the package moves the
  genesis app hash (`expectedCrossrealm38Hash`) whether or not anything imports
  it.

- Gas: one flat row, `chain/reflect.objectID` at 195. Measured end-to-end
  through the dispatcher: 194.6ns for the stamped path, 50.6ns unstamped,
  38.0ns with no identity, so the dearest path sets the price. Benchmarks in
  `gnovm/cmd/calibrate/reflect_bench_test.go`. Measured on an M1 where three
  shipped rows read 43.1 / 164.6 / 2821ns against table values of 45 / 148 /
  2631, so within 10% either way; a confirming run on the calibration machine
  is the last thing this needs.

- `grc20` is untouched. With the primitive in, the fix for #6026 is a small
  follow-up: drop `NewToken`'s caller-supplied `seqid.ID`, keep the readable
  `rlmPath.symbol` as the token's name, and add the ObjectID as the unique one.
  `gnovm/tests/files/zrealm_reflect_duplicate_symbol.gno` demonstrates that
  end to end on a local token type, so the primitive can be judged before the
  token redesign is agreed. GRC721 has the same exposure and the same fix.

## Alternatives considered

- **`chain/runtime.ObjectAddress`, as in #6139.** `chain/runtime` is
  `ChainID`, `ChainDomain`, `ChainHeight`, `AssertOriginCall`: information about
  the running chain. Reading which object a value is is not that, and putting
  it there leaves no home for the next question of the same kind.
- **`vm/reflect` or `gno/reflect`.** Neither `vm/` nor `gno/` exists as a
  stdlib namespace. Object identity is arguably a VM concept rather than a
  chain one, since it works under `gno test` with no chain, but it is
  inseparable from realm persistence, which is gno.land's storage model. A new
  top-level namespace for one package is a bigger commitment than a package
  under `chain/`, which is already where the non-Go-portable stdlibs live.
- **A top-level `reflect`.** Collides with Go's `reflect` for anyone reading or
  transpiling the code, and promises an API this will not have.
- **`NewRealmID()`'s `<realm-path>:<realm-time>` (#6101).** Final from the
  first call, but a bespoke scheme nothing else consumes, and what review
  rejected.
- **Reserving a tick (`Realm.ReserveObjectID`).** Gives a final ID at
  construction, but the reserved value belongs to no object.
- **Go's reflect API, or any part of it.** Out of scope on purpose. Type and
  field introspection over values a realm was handed is a capability question
  of its own; `chain/reflect` answers "which object is this" and nothing else.
