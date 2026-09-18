# ADR: cross-realm references are views unless the owner grants a handle

## Status

Proposed

## Context

Gno is the only chain VM where a realm holds a live reference to another
realm's object. Direct writes through such a reference are refused: the write
gate `Machine.IsReadonly` (machine.go) compares the object's owner
(`ObjectID.PkgID`) with the active storage realm `m.Realm`. Nothing is stored
on the value; "readonly" is derived from ownership.

Writes through a `/p/` method were not refused. Borrow rule #2 in
`PushFrameCall` switched `m.Realm` to the receiver's owner for the call, so
`X.GetUsers().Set(k, v)` from realm A committed to X's storage. That was the
only way a caller obtained write access to another realm's data, and it was
implicit: any exported function returning a `/p/`-typed pointer handed it out
whether the author meant to or not. The security guide (§5.1) and the audit
harness warned about the shape; the language did not mark it, and three realms
hand-rolled read-only views with `/p/nt/rotree`.

Survey of `examples/` (non-test realms, 2026-09-17): 19 exported functions
return a pointer, slice or map of a `/p/` type; 4 were write handles used in
production, all `*grc20.Token` (`grc20reg.Get/MustGet`,
`grc20factory.Bank/ListTokens`), consumed by 4 `RealmTeller(...)` sites.
Slices of structs were never handles (element writes hit the gate).

## Decision

A reference that crosses a realm boundary, as a return value or an argument,
is a view. The owner grants write access explicitly, and the grant is a
property of the object, so it survives the store.

1. **`mutable(x)` uverse builtin**, shaped like `cross(x)`: returns `x`
   unchanged and sets `ObjectInfo.IsShared` on `x`'s first object. Only the
   owner may grant (`PkgID == m.Realm.ID`), else panic. Real objects are
   marked dirty; unreal ones carry the flag into their first save. The grant
   unit is the pointer's base object (`GetFirstObject`, the same object rule
   #2 inspects): `mutable(&s.field)` grants `s`.
2. **Borrow rule #2 needs the grant.** For a `/p/` method on a **real**
   foreign-owned receiver, `PushFrameCall` borrows to the owner only if
   `IsShared` is set. Otherwise `m.Realm` stays the caller's and the method's
   writes are refused by the existing gate, with a message that names
   `mutable(x)` as the remedy when the writer is library code.
3. **Two receivers still borrow without a grant.** An *unreal* foreign
   receiver: the owner's code built it in this transaction, e.g. the teller a
   token's `RealmTeller` returns, and the spec already wants a value returned
   by a foreign constructor to carry its authority. An object owned by a
   `/p/` package: its post-init immutability gate reports the write with the
   clearer message, so borrowing is harmless.
4. **Grants do not survive adoption.** `SetPkgID` clears `IsShared`, so an
   object pre-marked by its allocator (e.g. a `maketx run` script) and then
   adopted by the realm that stores it is a view for the allocator afterwards.
   This is what keeps the flag from re-opening the adopted-stamp borrow.
5. **Both directions.** A passing its own `*avl.Tree` into
   `X.Register(cross(cur), t)` gives X a view unless A wrote `mutable(t)`.
6. Rules #1 (`/r/`-declared callables) and #3 (closures) are unchanged.

Persistence: `IsShared` is amino field 8, `json:",omitempty"`, so unflagged
objects hash as before. Amino numbers fields by struct position, so the field
sits after `LastObjectSize`; a later field (e.g. an adopted-stamp mark) takes 9.

Follow-up: the refusal message is chosen by the executing package's kind
(library vs realm). Choosing it from the refused object (immutable-package
owner, ungranted foreign owner, or other) would be more precise and would let
the `/p/`-owned exemption in rule #2 go; it needs `resolvePointer` to hand the
offending value to the panic site.

Follow-up: a `//gno:mutable` directive on the declaration, checked by the
preprocessor, would move the grant into the signature without a parser
change: it inserts `mutable()` on the returned or received reference and
rejects a body that hands out a handle the declaration does not announce.
A Gno2 type qualifier (`func F() mutable *T`, `t mutable *T`) is the same
idea with syntax; both keep the flag as the mechanism.

## Alternatives considered

- **Keep guidance only.** The dangerous code (`return users`) is the natural
  code; Solidity's `tx.origin` history says guidance alone fails at scale.
- **`readonly`/`mutable` signature modifier.** Not Go syntax, so go/parser and
  go/types break; covers only function boundaries. A `//gno:mutable` comment
  directive checked against the body can be added later on top of the flag.
- **Remove borrow rule #2.** Breaks the intended `avl.Tree.Set`-from-a-
  foreign-caller pattern (GRC20 hubs).
- **Per-reference grant.** Cannot survive the store; handles are kept across
  transactions.
- **Wrapper types like `rotree`.** Per type, hand-rolled, inverse default.

## Consequences

- Breaking for realms that hand out write handles. In tree: token realms grant
  at registration, `grc20reg.Register(cross(cur), mutable(Token), "")`
  (wugnot, foo20, test20, grc20factory, treasury test); `test20` also grants
  its `PrivateLedger` as a test fixture. The `interrealm_v2` spec corpus grants
  its `Lib` containers. The 26 `zrealm_launder_*` fixtures, which pinned the
  loophole as "known-open", now pin the refusal.
- A user identifier named `mutable` is now a shadowing error (one quarantined
  test renamed).
- Reads are unchanged. A borrowed method call reads one more flag from the
  object info it already loaded.
- Forgetting the grant fails loudly at the first foreign write instead of
  silently granting authority.
- Best landed before mainnet genesis; afterwards it needs a gnomod `gno`
  version gate, which does not exist yet.

## Tests

- `zrealm_view_default.gno`: a returned view refuses `Set`, reads work.
- `zrealm_view_handle.gno`: `mutable(handle)` lets the same write commit to the
  owner.
- `zrealm_view_grant_foreign.gno`: neither side can grant an object it does
  not own.
- `zrealm_view_ingress.gno`: the argument direction, view then grant.
- `view_default_grant_persists.txtar`: a stored handle is still a handle in
  the next transaction; a stored view is still a view.
- `TestCodecParity_Gnolang/ObjectInfo/shared`: field 8 marshals identically
  in the reflect and generated codecs.
- Regenerated goldens for the launder corpus; GRC20 examples and txtars pass
  with the grants above.
