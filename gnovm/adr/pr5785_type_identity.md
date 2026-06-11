# PR5785: Type Identity Checks

## Context

Gno uses `TypeID()` in many places as a deterministic structural type key. That
key intentionally ignores struct tags, which is correct for explicit
conversions: Go permits conversions between struct types that differ only in
tags.

The same key was also used for assignment, comparison, function signatures, and
interface method matching. Those contexts require Go type identity, where
struct tags, embedded field syntax, map element types, and variadic function
signatures matter. This let Gno accept programs that Go rejects, such as:

- assigning `struct{A int "b"}` to `struct{A int "a"}`;
- assigning `map[int]int` to `map[int]string`;
- assigning or converting `func([]int)` to `func(...int)`;
- treating `struct{T}` and `struct{T T}` as the same type;
- assigning `*int` to `*interface{}` (the old pointer-element check recursed
  into assignability, which accepted interface targets).

## Decision

Add separate type identity helpers instead of changing `TypeID()`:

- `identicalTypes` implements exact Go identity.
- `identicalTypesIgnoreTags` implements the conversion rule that ignores struct
  tags (recursively) but still respects the rest of type identity.

Both are implemented as a direct recursive structural comparison
(`identical()` in `types.go`), mirroring `go/types.Identical` /
`IdenticalIgnoreTags`. No strings are built and nothing is allocated on the
comparison path (except a copy+sort for unsorted interface method lists), so
the helpers are cheap enough for the paths that use them.

### Where exact identity applies

- **Static checks** (preprocessing/type-check time): assignment of composite
  elements (pointer elements, array/slice/map elements, structs, functions),
  comparison operands, compound assignment, function redefinition, uverse
  generic unification (`specifyType`; the vararg consolidation in
  `FuncType.Specify` is deliberately not an identity context, since Go only
  requires each vararg to be assignable), and the "redundant conversion"
  elision (a tag-changing conversion must emit a real conversion so the
  resulting static type is correct).
- **Interface implementation** (`InterfaceType.VerifyImplementedBy`), used
  both statically and by runtime interface type assertions: method signatures
  must match exactly, so `M([]int)` no longer satisfies `M(...int)`.
- Conversion legality keeps the tag-ignoring rule (`identicalTypesIgnoreTags`)
  so valid struct-tag conversions continue to work.

### Where runtime value identity stays TypeID-based — and why

Runtime *value* type identity — interface equality (`isEql`), concrete type
assertions (`doOpTypeAssert1/2`), type switch concrete cases, map keys, and
the type store — intentionally remains `TypeID()`-based, i.e. struct tags
(and the other distinctions `TypeID()` conflates) are not part of a value's
dynamic type at runtime.

This is deliberate, not an oversight. The persistence model keys types by
`TypeID()` and stores interior `TypedValue` types accordingly, so tag-variant
unnamed types conflate in storage, and interior values (struct fields, slice
elements) of a converted value retain their original type representation.
Making *some* runtime operations tag-strict while the value model conflates
tags produces internal contradictions — e.g. after a tag-changing conversion,
two values of the same static type could compare unequal, and a value read
from a `[]struct{A int "a"}` could fail an assertion against its own element
type. Restricting strictness to static checks keeps the runtime internally
consistent and avoids any consensus-visible change to existing runtime
behavior. Making runtime value identity fully Go-faithful requires changes to
the persistence/type-storage model and is left for a follow-up.

Error messages for mismatches render types via `String()`, which now includes
struct tags and renders embedded fields without a redundant name, so messages
like `cannot use struct{A int "b"} as struct{A int "a"}` distinguish the two
sides.

## Alternatives Considered

1. **Change `TypeID()` to include tags and variadic signatures.** This would fix
   assignment but break valid conversions that intentionally ignore struct tags,
   and would change persisted type keys.

2. **Patch only struct assignment.** That would leave the same bug in pointers,
   maps, interfaces, function signatures, comparisons, and conversions.

3. **Keep using `TypeID()` and special-case tags.** This would not address other
   identity differences, especially embedded fields and variadic functions.

4. **Build identity strings (`typeIDForIdentity`) and compare them.** Simple,
   but it recomputes and allocates on every comparison with no memoization.
   Direct structural comparison short-circuits and does not allocate.

5. **Extend exact identity to runtime operations (`isEql`, type assertions,
   type switches).** Matches Go for those operations in isolation, but the
   `TypeID()`-keyed type storage and interior `TypedValue` types make it
   internally inconsistent (see above), and it changes consensus-visible
   runtime behavior. Rejected in favor of static-only strictness.

## Known Limitations

- Runtime value type identity ignores struct tags, embedded-field syntax, and
  variadicity: interface values whose dynamic types differ only in tags
  compare equal (Go: unequal), a concrete type assertion against a
  tag-variant type succeeds (Go: fails), and type switch cases match
  likewise. Pinned by `eql_struct_tags.gno`, `typeassert_struct_tags.gno`,
  and `typeswitch_struct_tags.gno`. Statically, such programs are rejected
  wherever Go rejects them, and `gno type check` (go/types) provides full Go
  semantics at deploy time.
- Interfaces are compared by their literal (unflattened) method lists: an
  interface that embeds another interface compares unequal to its flattened
  equivalent, even though Go treats them as identical. This matches the
  existing `TypeID()` behavior and is left for a follow-up.
- `RefType` (unresolved store references) cannot be inspected structurally
  without a store, so it is compared by `TypeID()`, preserving the previous
  behavior for persisted types.
- Interfaces embedding two same-named interfaces from different packages
  panic in `sort` on the duplicate method name, as `InterfaceType.TypeID()`
  already does; `identical()` inherits that behavior unchanged.

## Consequences

- Gno rejects more invalid programs at preprocessing/type-check time, matching
  Go identity rules more closely. This is stricter than before: code that
  relied on the looser checks (e.g. a method `M([]int)` satisfying
  `M(...int)`) will now be rejected, including runtime type assertions
  against interface types (via `VerifyImplementedBy`).
- The "redundant conversion" elision in preprocessing requires exact identity
  (tags included); a tag-changing conversion now emits a real conversion so
  the expression's static type is correct.
- Explicit struct conversions that differ only by tags remain valid.
- Runtime value-identity semantics (`==`, concrete asserts, type switches,
  map keys) are unchanged from before this PR.
- `TypeID()` remains the deterministic key for conversion/storage/runtime
  paths; exact identity checks are opt-in at call sites that need Go
  identity.
- Filetests cover struct tags, embedded fields, map element types, variadic
  function assignment/conversion, interface method matching, struct
  comparison, pointer-to-interface assignment, tag-changing conversions
  followed by assignment/equality, and the runtime value-identity rule.
  Unit tests (`TestIdenticalTypes`) cover the helpers directly.
