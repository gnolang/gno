# Object addresses, and GRC20 token identity

## Context

`grc20.NewToken` built `Token.id` from `<pkg-path>.<symbol>.<caller-supplied-seqid>`.
The sequence number came from the calling realm, so one realm could create two
independent tokens under one identifier. Their `Transfer`/`Approval` events are
then indistinguishable, which is issue #6026.

PR #6101 fixed this by adding `runtime.NewRealmID()`, returning
`<realm-path>:<realm-time>`. The review (moul) rejected the bespoke format: the
primitive worth building on is `ObjectID` itself — an object should be
addressable, so that Gno objects become potential accounts — rather than a
one-off ID scheme scoped to GRC20. Reserving a tick of `Realm.Time` without
attaching it to an object was rejected for the same reason: it is a private
counter wearing an ObjectID's shape.

Where identity comes from, on master:

- `PkgID` is stamped by the allocator when the object is created
  (`alloc.go`, `oi.SetPkgID(alloc.currentRealmID)`). It names the realm the
  object belongs to.
- `NewTime` is stamped by `Realm.assignNewObjectID` (`realm.go`), reached only
  from `FinalizeRealmTransaction` → `processNewCreatedMarks` →
  `incRefCreatedDescendants`. It is a tick of `rlm.Time`, minted while walking
  the ownership tree.
- Finalization runs at every realm boundary return (`doOpReturn` →
  `maybeFinalize`), not only at the end of the transaction. A foreign realm's
  finalize stamps objects it was handed too, minting their `NewTime` from the
  owning realm's clock (`assignNewObjectID`'s `touchForeignRealm` branch).

So an object's ID is complete the first time it is persisted, and the ticks
follow finalize traversal order, not allocation order.

## Decision

Give a Token two names, and let the VM own the one that has to be unique.

- `DeriveObjectIDCryptoAddr(ObjectID)` hashes the preimage
  `objectid:<pkgid>:<newtime>`, mirroring `DerivePkgCryptoAddr`'s `pkgPath:`.
  It rejects an ID with either half missing: that names no object.
- `ObjectID.DeriveAddress()` returns that address, or `""` for an ID with no
  `NewTime` yet.
- `chain/runtime.ObjectAddress(v any) string` returns it for the object behind
  `v` (`TypedValue.GetFirstObject`). It reads; it never writes. No realm clock
  advances, no counter is kept, and the object lifecycle is untouched.
- `grc20.NewToken` drops its `id seqid.ID` parameter. `Token.TokenPath()`
  keeps the readable `rlmPath.symbol` name, built from the IsCurrent-verified
  `rlm.PkgPath()`; `Token.ID()` returns `runtime.ObjectAddress(tok)`.
- Every grc20 event carries both, as `token` and `id`. `grc20reg`'s `register`
  event likewise carries `token_key` (its own rlmPath.slug key) and `token_id`.

Two tokens one realm issues under one symbol share a `token` and differ in
`id`, which is exactly the ambiguity #6026 is about. Neither name is chosen by
the realm: one is derived from its verified package path, the other from the
VM's object identity.

### The unstamped window

Until an object is first persisted its `NewTime` is zero, so there is no
address to derive and `ID()` returns `""`. Events are written into the tx event
log at `chain.Emit` time and never rewritten, so `NewToken` — and any
`Mint`/`Transfer` emitted before the token is persisted — carry an empty `id`
permanently. The `token` path is unaffected; it exists from construction.

This is accepted rather than worked around. It is inherent: an object has no
ObjectID before it is attached to the ownership tree, and attaching happens at
finalize. Two txtars pin the behaviour end to end:

- `grc20_object_id_events.txtar` — one call per step. Only `Create`'s own
  `NewToken` event has an empty `id`; mint, register and transfer all carry it.
- `grc20_object_id_events_init.txtar` — all four steps inside `init`, the worst
  case. `NewToken` and the minting `Transfer` are unstamped; `register` and the
  `Transfer` after it carry the id.

`Register` gets there by writing its entry through a same-realm `cross(cur)`
and emitting afterwards. An explicit cross is a realm boundary, so returning
from it runs grc20reg's finalize, which stamps the token the write just made
reachable. Without it, every `init()` that creates and registers in one call
would announce an empty binding. The cost is one extra finalize pass per
registration, which is rare.

## Alternatives considered

- **Keep `NewRealmID()`'s `<realm-path>:<realm-time>`.** Final from the first
  call, but a second identity scheme no other VM concept consumes, and what
  review rejected.
- **Reserve a tick (`Realm.ReserveObjectID`).** Gives a final id at
  construction and cannot collide with a stamped one, but the reserved value
  belongs to no object.
- **Stamp `NewTime` on the object early.** Would give the real object its real
  id at construction, but `GetIsReal()` *is* `ID.IsFinalized()` and
  `incRefCreatedDescendants` uses the same predicate as its already-visited
  guard, so an early-stamped object is treated as real while unsaved and is
  skipped by finalization. It needs a different reality marker and a different
  recurse guard — a lifecycle change far past this problem.
- **Drop the readable name and identify a token by address alone.** Simpler,
  but it makes every event unreadable without a lookup and gives an indexer
  nothing to fall back on during the unstamped window.

## Consequences

- `grc20.NewToken`'s signature changes, `Token.ID()` changes from a path to a
  `g1` address, and every grc20 event gains an `id` attribute while `token`
  becomes the path alone. `grc20reg` keys move from `rlmPath.symbol` to
  `rlmPath.slug` — the realm names its own entry, and a realm registering one
  token can leave the slug empty and be found under its realm path.
- `ID()` is a native call rather than a field read. A `Token` copied into a
  struct field or an array element has no id of its own and reads `""`, so a
  token is addressed through the pointer `NewToken` returns.
- Two tokens created by one call share the empty id until they are persisted,
  so unit tests can only assert id uniqueness after a crossing has returned;
  the cross-transaction case lives in `filetests/token_identity_filetest.gno`,
  which creates its tokens in `init` and reads them in `main`.
- Stdlib `.gno` source is genesis state, so declaring the native moves the app
  hash (`expectedCrossrealm38Hash` bumped).
- GRC721 has the same caller-supplied `seqid.ID` parameter and the same
  duplicate-identity exposure. It is left alone here; the same change applies
  once this lands.
- The address is fixed once it is referenced. Its inputs are the `"objectid:"`
  preimage prefix and separators, the `PkgID` produced by `PkgIDFromPkgPath`
  (whose flag nibble encodes IsStdlib/IsImmutable/IsInternal, so reclassifying a
  path or claiming the reserved bit moves every object address in that realm),
  and the object's `NewTime`. Once an address has been written into an event, a
  registry entry, or a balance, changing any of these inputs is a state break:
  the same object answers with a new address and whatever sat at the old one
  becomes unreachable. Such a change needs a migration.
