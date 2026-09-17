# ADR: cross-realm references are views unless the owner grants a handle

## Status

Proposed (draft, not implemented)

## Context

Gno is the only chain VM where a realm holds a live reference to another
realm's object. Direct writes through such a reference are refused: the write
gate `Machine.IsReadonly` (machine.go) compares the object's owner
(`ObjectID.PkgID`) with the active storage realm `m.Realm`.

Writes through a `/p/` method are not refused. Borrow rule #2 in
`PushFrameCall` switches `m.Realm` to the receiver's owner for the call, so
`X.GetUsers().Set(k, v)` from realm A commits to X's storage. This is the only
way a caller obtains write access to another realm's data, and it is implicit:
any exported function that returns a `/p/`-typed pointer hands it out, whether
the author meant to or not. The security guide (§5.1) and the audit harness
warn about the shape; nothing in the language marks it. Three realms already
hand-roll read-only views with `/p/nt/rotree`, which shows the default is the
wrong way round.

Survey of `examples/` (non-test realms, 2026-09-17): 19 exported functions
return a pointer, slice or map of a `/p/` type; 4 are write handles used in
production, all `*grc20.Token` (`grc20reg.Get/MustGet`,
`grc20factory.Bank/ListTokens`), consumed by 4 `RealmTeller(...)` sites.
Slices of structs are not handles (element writes hit the gate).

## Decision

A reference that crosses a realm boundary, as a return value or an argument,
is a view. The owner grants write access explicitly, and the grant is a
property of the object, so it survives being stored and reused later.

1. **`mutable(x)` uverse builtin**, shaped like `cross(x)`: takes any value,
   returns it unchanged, and sets a persisted `ObjectInfo.IsShared` flag on
   the object's first object. Only the owner may grant: the object's `PkgID`
   must equal `m.Realm.ID`, else panic. Reads of the flag are free; the write
   dirties the object once.
2. **Borrow rule #2 requires the flag.** For a `/p/` method on a foreign-owned
   real receiver, `PushFrameCall` borrows to the owner only if `IsShared` is
   set. Otherwise `m.Realm` stays the caller's and the method's writes are
   refused by the existing gate. Reads work as today.
3. **Both directions.** A passing its own `*avl.Tree` into
   `X.Register(cross(cur), t)` gives X a view unless A wrote `mutable(t)`.
4. Rules #1 (`/r/`-declared callables) and #3 (closures) are unchanged.

Persistence: `ObjectInfo.IsShared` is amino field 9 (`json:",omitempty"`,
hash-neutral when false). Field 8 is reserved for `IsAdopted` from the
adopted-stamp fix, which lands first; an adopted object is never shared, so
the two flags agree.

## Alternatives considered

- **Keep guidance only.** The dangerous code is the natural code
  (`return users`); Solidity's history with `tx.origin` says guidance alone
  fails at scale.
- **`readonly` modifier in signatures** (the spec's planned direction, with
  the default flipped). Not Go syntax, so go/parser and go/types break; only
  covers function boundaries. A `//gno:mutable` comment directive checked by
  preprocess can be added later on top of the flag.
- **Remove borrow rule #2.** Breaks the intended `avl.Tree.Set`-from-a-foreign-
  caller pattern; already rejected in the adopted-stamp ADR.
- **Per-reference grant.** Cannot survive the store; handles are kept across
  transactions.
- **Wrapper types like `rotree`.** Per-type, hand-rolled, inverse default.

## Consequences

- Breaking for realms that hand out write handles. In tree: the 5 token realms
  registering with `grc20reg.Register(...)` grant with `mutable(token)` at
  registration; `grc20reg`/`grc20factory` return the already-shared object
  unchanged. `rotree` views keep working and become optional.
- Reads through returned references are unaffected; no gas change on reads.
  Borrowed method calls read one more flag from the object info already loaded.
- Forgetting the marker now fails loudly at the first foreign write instead of
  silently granting authority. The error names `mutable(x)` as the remedy.
- Best landed before mainnet genesis. After genesis it needs a gnomod
  `gno` version gate (`0.9` writable-by-default, next version view-by-default),
  which does not exist yet.

## Tests (to write with the implementation)

- filetests: foreign `Set()` on a view is refused; `mutable(x)` by the owner
  allows it; `mutable(x)` by a non-owner panics; ingress direction; `rotree`
  unchanged; grc20 `MustGet(...).RealmTeller(0, cur).Transfer(...)` still works.
- txtar: the grant survives a store round-trip (grant in tx 1, foreign write
  in tx 2); an adopted object stays unshared.
- amino parity for field 9.
