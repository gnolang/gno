# Share interface-held values when copying arrays

## Context

`ArrayValue.Copy` deep-copied every element, including the value held by an
interface-typed element. A chain like `x = [1]any{x}` therefore copied the
whole nest on every iteration: quadratic, where Go is linear (copying an array
of interfaces copies each interface's two-word header, never the boxed
value). In the profile of the TestFiles short run this was 41% of CPU,
concentrated in `recurse1.gno` (~23s) and `gas/nested_alloc.gno`
(12,560,317,779 gas).

## Decision

### 1. Share interface-kind elements

When the array's element type has `InterfaceKind`, `ArrayValue.Copy` copies
the element `TypedValue`s as-is instead of calling `TypedValue.Copy` on each.
Non-interface element types are unchanged.

The exception is an element that is an object stamped with a different realm
(`ObjectID.PkgID`) than the new array, which is still deep-copied, whether it
is loaded or still a `RefValue`. Objects are
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

`isAddressable` (preprocess.go, from #6083) implements the Go spec rule — a
variable, a pointer indirection, a slice index, or a field selector or array
index of an addressable operand — from information the preprocessor already
has. #6083 enforces it where an address is taken: `&x` (except a composite
literal), pointer-receiver auto-ref (`x.m()` → `(&x).m()`, method values too)
and slicing an array. It left assignment targets alone, so `x.(S).F = 9`,
`x.(S).F++` and `get().F = 9` still preprocessed. This change adds the fourth
site:

| site | rejects | message |
|---|---|---|
| `assertValidAssignLhs` (`=`, `op=`, `++`/`--`, `for k = range`) | a non-addressable selector or index target; a map index stays assignable | `cannot assign to …` |

## Why sharing is sound

Values held in interface slots are sealed — immutable for as long as they are
reachable through an interface slot:

- every write into an interface slot copies on entry (`TypedValue.Assign`);
- extracting one (type assertion, type switch binding) copies again on
  assignment, parameter passing and composite-literal construction;
- they are not addressable — now enforced by the VM (#6083 and decision 2),
  not only by `go/types`.

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
gas golden changes: `gas/nested_alloc.gno` drops from 12,560,317,779 to
17,643,921. The addressability check adds no gas: it reads `ValuePath`s and
cached static types of already-preprocessed operands.

### Persisted object graph

Within a realm, a shared element that is a realm object gets a second
reference, so on finalization it escapes (`IsEscaped`, `RefCount` 2, no
`OwnerID`) rather than being persisted twice as two owned objects. This holds
whether the element was created in the same transaction
(`zrealm_iface_array_share.gno`), already stored and loaded before the copy
(`_stored.gno`), or already stored and not yet loaded (`_stored_unread.gno`).

Without sharing, `TypedValue.Copy` loads an unloaded element (a `RefValue`)
and deep-copies it like a loaded one (#6504), so every such copy persists two
owned objects. Sharing persists one escaped object whether or not the element
was read first. The `_drop_one` and `_drop_both` tests pin the reference
counting when the shared element is released.

Across realms, an array handed to another realm by value still gives the
receiver its own copy of each element stamped with another realm, as before,
whether the element was already stored and loaded (`_crossrealm.gno`),
stored and not yet loaded (`_crossrealm_unread.gno`), or allocated in the same
transaction (`_crossrealm_fresh.gno`); all three goldens pass unchanged on
master. Only an element already stamped with the receiving realm (a value
of a type that realm declares) is shared, and it is persisted under that
realm either way.

### Addressability

The VM now rejects at preprocess everything `go/types` rejects as
non-addressable as an assignment target. `TypeCheckMemPackage` already runs
in `AddPackage` and `Run`, so no transaction that previously succeeded can
hit the new errors, and no stored package can fail to re-preprocess unless
`isAddressable` were stricter than Go. `addressable_15.gno` pins the shapes
where addressability comes only through an implicit pointer dereference on a
non-addressable root (field promoted through an embedded pointer of a call
result, a named pointer type, pointer-to-array and slice indexes of call
results, `&T{}`); its output matches `go run`. `addressable_15{a..j}_err.gno`
cover each guarded path, the assignment ones (`a`–`e`, `h`, `i`) new here.

`issue_1096.gno` lost its one non-addressable stanza (`xm[1].Array[1] = 123`),
which would otherwise abort the whole file; `addressable_15i_err.gno` covers
it.

## Alternatives considered

- **Rely on `go/types` alone.** Every on-chain entry point type-checks first,
  so this is defence in depth, but the soundness of a VM-level sharing
  invariant should not depend on a check that lives outside the VM.
- **Copy-on-write for interface-held values.** Would need a write barrier on
  every lvalue path; enforcing addressability removes the writes instead.
- **Share in `copy`/`append` over `[]any` and in `StructValue.Copy` for
  interface fields too.** The same invariant applies, but these were not the
  profiled cases and each is a further gas- and persisted-graph-visible
  change needing its own realm tests. `copy`/`append` call
  `TypedValue.Copy` on each element, which loads and deep-copies it. Left as
  follow-ups.
