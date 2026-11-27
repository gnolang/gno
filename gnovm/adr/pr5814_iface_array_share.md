# Share interface-held values when copying arrays

## Context

`ArrayValue.Copy` deep-copied every element, including the value held by an
interface-typed element. A chain like `x = [1]any{x}` therefore copied the
whole nest on every iteration: quadratic, where Go is linear (copying an array
of interfaces copies each interface's two-word header, never the boxed
value). In the profile of the TestFiles short run this was 41% of CPU,
concentrated in `recurse1.gno` (~45s on current master) and
`gas/nested_alloc.gno` (12,560,317,779 gas).

## Decision

### 1. Share in-memory interface-kind elements of the executing realm

When the array's element type has `InterfaceKind`, `ArrayValue.Copy` copies
an element `TypedValue` as-is, instead of calling `TypedValue.Copy` on it,
when the element is an object that:

- is not real (exists in memory only, not yet persisted);
- is stamped (`ObjectID.PkgID`) with the executing realm
  (`Allocator.currentRealmID`); and
- the executing realm is not an ephemeral run realm (`/e/…/run`).

Every other element is deep-copied as before. Non-interface element types are
unchanged. Each exclusion closes a failure a review found in a wider rule:

- **Stored elements** (real, or still a `RefValue`). Their lifetime belongs
  to the persisted graph: a shared reference from a Go local, return value or
  later copy is not counted, so once the source drops the element, finalize
  deletes it while the copy still refers to it. `b := arr; arr[0] = nil;
  return b` across a crossing return handed the caller a ref to a deleted
  object, and a GC between the copy and the store left `arr2 = b` unreadable
  and impossible to clear (`unexpected object with id …`).
- **Elements stamped with another realm.** A realm persists an object under
  its stamp, so a shared element handed to another realm would be persisted,
  and its storage deposit charged, as the sender's object, and a private
  sender's object cannot be persisted by another realm at all
  (`iface_array_share_private_sender.txtar`). The check is against the
  executing realm, not the new array's stamp, which is the declaring realm of
  a realm-declared array type (`_declbox.gno`), and stored elements are
  excluded anyway, which covers an element whose type another realm declares
  (the `SendStored` step of `iface_array_share_private_sender.txtar`).
- **Ephemeral run realms.** Any realm may adopt an object a MsgRun script
  allocated, so two copies handed by pointer to two realms would leave one
  realm persisting the element and the other only referencing it
  (`iface_array_share_msgrun_adopt.txtar`). `IsEphemeralPath` runs a regex
  on `/e/` paths (~1.2 µs, unmetered), so it is computed once per `Realm`
  where the realm is built or decoded, and `setRealm` hands the allocator a
  plain bool.

These exclusions match where #6504 already made `TypedValue.Copy` load and
deep-copy refs, so stored and cross-realm graphs are the same as master's.

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

### 3. JSON export expands shared values at each occurrence

The export walk (`values_export.go`) emits an ephemeral object it meets again
as `ExportRefValue{":N"}`, which gnoweb and gnojs render as `<cycle :N>`. With
sharing, two copies of `[1]any{[2]int{1, 2}}` in a `qeval_json` result hold
the same in-memory `[2]int`, so the second would render as a cycle instead of
its value. An array or struct held by value that was first expanded by value
is now expanded again at each occurrence, numbered as a new inline expansion
so `:N` still names the Nth inline object in walk order. Such sharing only
references older, sealed values, so it cannot form a cycle, and the export's
size and depth budgets bound the walk. An object first emitted as a pointer
or slice base keeps its back-references, which record real aliasing; shared
values are never addressable, so they are never such a base.

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
`iface_array_share_isolation.gno` commits the extract-then-mutate shapes; its
output matches `go run`.

Immutability is not enough on its own: a shared value must also outlive every
copy that refers to it, which only in-memory objects do (decision 1, "Stored
elements").

## Consequences

### Gas

Copies that no longer happen are no longer charged. Across the filetests one
gas golden changes: `gas/nested_alloc.gno` drops from 12,560,317,779 to
17,643,921. The addressability check adds no gas: it reads `ValuePath`s and
cached static types of already-preprocessed operands.

### Persisted object graph

Within a realm, an element created in the same transaction and shared by a
copy gets a second reference, so on finalization it escapes (`IsEscaped`,
`RefCount` 2, no `OwnerID`) rather than being persisted twice as two owned
objects (`zrealm_iface_array_share.gno`); `_drop_one` and `_drop_both` pin
the reference counting when it is released.

Everywhere else the persisted graph is master's:

- an element already stored, read before the copy or not (`_stored.gno`,
  `_stored_unread.gno`, identical goldens), is deep-copied, and the copy owns
  its own object;
- a copy returned across a crossing function after its source was cleared
  owns its elements (`_return_unread.gno`, `_return_loaded.gno`,
  `iface_array_share_drain.txtar`), as does one stored after a GC
  (`_gc.gno`, `iface_array_share_gc.txtar`);
- an array handed to another realm by value gives the receiver its own copy,
  whether the element was stored and loaded (`_crossrealm.gno`), stored and
  not yet loaded (`_crossrealm_unread.gno`), allocated in the same
  transaction (`_crossrealm_fresh.gno`) or held in a realm-declared array
  type (`_declbox.gno`); these goldens pass unchanged on master;
- copies made by a MsgRun script and adopted by two realms each cost their
  realm the same storage as on master (`iface_array_share_msgrun_adopt.txtar`).

### JSON export

A value shared by decision 1 is expanded at each occurrence
(`TestExportValuesSharedIfaceElement`),
and a pointer's back-reference to a struct later held by value is kept
(`TestExportValuesPointerFirstKeepsRef`). The rule also changes one output
master could already produce: returns do not copy, so `return a, a` on an
array yields the same object twice, and master exported the second result as
`<cycle :1>`; it is now expanded too. The export runs outside consensus
(queries only), so this changes no state or gas.

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
