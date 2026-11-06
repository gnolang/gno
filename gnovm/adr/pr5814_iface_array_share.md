# Share interface-held values when copying arrays

## Context

`ArrayValue.Copy` deep-copied every element, including the value held by an
interface-typed element. A chain like `x = [1]any{x}` therefore copied the
whole nest on every iteration: quadratic, where Go is linear (copying an array
of interfaces copies each interface's two-word header, never the boxed
value). In the profile of the TestFiles short run this was 41% of CPU,
concentrated in `recurse1.gno` (~23s) and `gas/nested_alloc.gno`
(8,559,693,635 gas).

## Decision

### 1. Share interface-kind elements

When the array's element type has `InterfaceKind`, `ArrayValue.Copy` copies
the element `TypedValue`s as-is instead of calling `TypedValue.Copy` on each.
Non-interface element types are unchanged.

The exception is an element that is an object stamped with a different realm
(`ObjectID.PkgID`) than the new array, which is still deep-copied. Objects are
stamped at allocation with the executing realm (or the declaring realm of a
realm-declared type), and a realm persists an object under its stamp: a
shared element handed to another realm by value would be persisted, and its
storage deposit charged, as the sender's object, and a private sender's
object cannot be persisted by another realm at all
(`iface_array_share_private_sender.txtar` fails without this exception).
Copies within one realm, or outside any realm, still share.

### 2. The VM enforces addressability

Sharing is sound only if a value held in an interface slot can never be
written in place (see below). Go guarantees that statically, and so did
`go/types` here, but the VM did not: `x.(S).F = 9`, `x.(S).F++`,
`&x.(S).F`, `x.(S).T.PtrMethod()`, `get().F = 9` and `x.([2]int)[:]` all
preprocessed and executed, writing into (or handing out a writable view of)
the held value. With sharing, such a write lands in every array that shares
the element.

`isAddressable` (type_check.go) implements the Go spec rule — a variable, a
pointer indirection, a slice index, or a field selector or array index of an
addressable operand — from information the preprocessor already has: the
selector's `ValuePath` type (`VPField` recurses into `X`; `VPDerefField`,
`VPSubrefField` and package-qualified `VPBlock` are addressable) and the
static type of an index's operand. It is enforced at the four places that
need an address:

| site | rejects | message |
|---|---|---|
| `assertValidAssignLhs` (`=`, `op=`, `++`/`--`, `for k = range`) | a non-addressable selector or index target; a map index stays assignable | `cannot assign to …` |
| `RefExpr` TRANS_LEAVE (`&x`) | a non-addressable operand, except a composite literal | `invalid operation: cannot take address of …` |
| pointer-receiver auto-ref (`x.m()` → `(&x).m()`, method values too) | a non-addressable `x` | `cannot call pointer method … on …` |
| `SliceExpr` TRANS_LEAVE (`a[lo:hi]` on an array) | a non-addressable array operand | `cannot slice unaddressable value …` |

## Why sharing is sound

Values held in interface slots are sealed — immutable for as long as they are
reachable through an interface slot:

- every write into an interface slot copies on entry (`TypedValue.Assign`);
- extracting one (type assertion, type switch binding) copies again on
  assignment, parameter passing and composite-literal construction;
- they are not addressable — now enforced by the VM (decision 2), not only by
  `go/types`.

So two arrays sharing an element cannot observe each other's writes: a write
to either array's slot replaces that slot's `TypedValue` and leaves the other
alone. This was checked with 42 throwaway `gno run` programs: every shape Go
rejects as non-addressable aliases without decision 2 and is refused with it,
and every legal extract-then-mutate shape (define, `var`, parameter,
variadic, method receiver, composite literal, `append`, `copy`, closure
return, type switch) stays isolated with and without sharing.

## Consequences

### Gas

Copies that no longer happen are no longer charged. Across the filetests one
gas golden changes: `gas/nested_alloc.gno` drops from 8,559,693,635 to
17,013,961. The addressability check adds no gas: it reads `ValuePath`s and
cached static types of already-preprocessed operands.

### Persisted object graph

Within a realm, a shared element that is a realm object gets a second
reference, so on finalization it escapes (`IsEscaped`, `RefCount` 2, no
`OwnerID`) rather than being persisted twice as two owned objects. This holds
whether the element was created in the same transaction
(`zrealm_iface_array_share.gno`), already stored and loaded before the copy
(`_stored.gno`), or already stored and not yet loaded (`_stored_unread.gno`).

Previously the stored case depended on load state: an unloaded element is a
`RefValue`, which `TypedValue.Copy` passes through (shared, escaped), while a
loaded one was deep-copied (two owned objects). The same realm code thus
persisted different graphs depending on whether it had read the element
first; it now persists the same graph either way. The `_drop_one` and
`_drop_both` tests pin the reference counting when the shared element is
released.

Across realms, an array handed to another realm by value still gives the
receiver its own copy of each element stamped with another realm, as before,
whether the element was already stored (`_crossrealm.gno`) or allocated in
the same transaction (`_crossrealm_fresh.gno`); both goldens pass unchanged
on master. Only an element already stamped with the receiving realm (a value
of a type that realm declares) is shared, and it is persisted under that
realm either way.

### Addressability

The VM now rejects at preprocess everything `go/types` rejects as
non-addressable in these positions. `TypeCheckMemPackage` already runs in
`AddPackage` and `Run`, so no transaction that previously succeeded can hit
the new errors, and no stored package can fail to re-preprocess unless
`isAddressable` were stricter than Go. `addressable_12.gno` pins the shapes
where addressability comes only through an implicit pointer dereference on a
non-addressable root (field promoted through an embedded pointer of a call
result, a named pointer type, pointer-to-array and slice indexes of call
results, `&T{}`); its output matches `go run`.

The existing `addressable_*_err` filetests, which pinned the VM's lenient
runtime behaviour next to the `go/types` error (several printed pointers into
interface-held or map-held values), now expect the preprocess error.
`issue_1096.gno` lost its one non-addressable stanza (`xm[1].Array[1] = 123`),
which would otherwise abort the whole file; `addressable_12i_err.gno` covers
it.

A VM-side addressability check existed before and was removed in #4053,
which moved the filetests onto `go/types` (`TypeCheckError`). That version
added an `addressability()` method to every `Expr` and addressability fields
to several AST nodes. This one is a single function over existing
preprocessor state, called from the four sites above.

## Alternatives considered

- **Rely on `go/types` alone.** Every on-chain entry point type-checks first,
  so this is defence in depth, but the soundness of a VM-level sharing
  invariant should not depend on a check that lives outside the VM.
- **Copy-on-write for interface-held values.** Would need a write barrier on
  every lvalue path; enforcing addressability removes the writes instead.
- **Share in `copy`/`append` over `[]any` and in `StructValue.Copy` for
  interface fields too.** The same invariant applies, but these were not the
  profiled cases and each is a further gas- and persisted-graph-visible
  change needing its own realm tests. `copy`/`append` go through
  `unrefCopy`, which loads and deep-copies each element. Left as follow-ups.
