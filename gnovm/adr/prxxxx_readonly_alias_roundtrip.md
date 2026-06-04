# Preserve Readonly Provenance for Alias-Retaining Copies

## Context

PR #5747 fixed #5736 by changing `ArrayValue.Copy` and `StructValue.Copy` to
stamp copies from the destination type rather than blindly propagating the
source object's runtime `PkgID`. This allows `/p/` value types such as
`uint256.Uint` to be copied inside another realm and then mutated locally:

```go
type Uint struct {
    arr [4]uint64
}
```

That type is transitively value-only. Copying it cannot retain an alias to
foreign storage.

The same PR also removed `N_Readonly` provenance taint entirely. That created a
different behavior for value copies that retain references:

```go
// /r/victim
var Global = [1][]byte{[]byte("safe")}

func Copy(dst []byte, src string) {
    copy(dst, []byte(src))
}

// /r/attacker
x := victim.Global
victim.Copy(x[0], "pwn!")
```

The array wrapper copied into the attacker is fresh, but `x[0]` is a slice
header that still aliases `victim.Global[0]`'s backing array. Passing the alias
back into a victim-declared function borrows storage authority to the victim
realm, so the ownership gate permits the mutation.

## Decision

Reintroduce readonly provenance for foreign reads, but clear it only when a
copy is transitively alias-free.

A type is considered alias-free when it is:

- a primitive type,
- an array whose element type is alias-free,
- a struct whose fields are all alias-free.

All other types are treated conservatively as alias-retaining, including
slices, pointers, maps, interfaces, funcs, channels, packages, and unknown
internal types.

`TypedValue.Copy` therefore keeps `N_Readonly` for copied arrays and structs
unless `isDeepCopyAliasFree(tv.T)` returns true. `ArrayValue.Copy` and
`StructValue.Copy` keep the type-driven `PkgID` stamping introduced by #5747.

## Alternatives Considered

1. Keep #5747's full removal of readonly provenance.

   This keeps the implementation simple but makes retained aliases writable
   whenever they are passed back into the owning realm's helpers.

2. Revert #5747 entirely.

   This closes the alias round-trip but also reintroduces the false-positive
   panic for value-only `/p/` types such as `uint256.Uint`.

3. Track readonly provenance at precise child-reference granularity.

   This would avoid tainting a whole copied struct when only one field retains
   an alias, but it is a larger VM change. The conservative parent-level taint
   is simpler and preserves the security property.

## Consequences

- #5736 remains fixed for transitively value-only `/p/` types.
- Alias-retaining copies such as `[1][]byte` remain protected from round-trip
  mutation through borrowed owner-realm functions.
- Some local writes to copies of foreign structs/arrays with reference fields
  remain over-blocked. This is intentional until the VM has a more precise
  per-reference readonly model.
