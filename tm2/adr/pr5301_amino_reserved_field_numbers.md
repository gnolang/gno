# Amino: Reserved Field Numbers via Blank Identifier Fields

## Context

Amino is the binary codec used throughout tm2. It assigns each exported struct
field a monotonically increasing field number (starting at 1), which is used as
the key in the encoded wire format.

When a field is removed from a struct, its former field number becomes a "hole"
in the sequence. Any data encoded against the old struct still carries bytes
tagged with that field number. Without a way to signal that the number is
intentionally vacant, the decoder either errors or silently misaligns all
following fields onto the wrong field numbers.

Protobuf3 addresses this with an explicit `reserved N;` statement in the
message definition. Amino had no equivalent mechanism.

## Decision

Introduce reserved field number support via blank identifier struct fields
tagged `amino:"reserved"`:

```go
type MyStruct struct {
    A string
    _ [0]struct{} `amino:"reserved"` // field 2 is reserved
    C string                          // field 3
}
```

During struct info parsing (`parseStructInfoWLocked`), a `_` field with the
`amino:"reserved"` tag consumes the next field number and records it in
`StructInfo.Reserved`, without contributing a `FieldInfo` entry.

During binary decoding (`decodeReflectBinaryStruct`), when the wire field
number is less than the next expected field number, the decoder now consumes
and discards those wire bytes rather than returning an error. This allows data
encoded against the old struct to be decoded into the new struct transparently.

During Protobuf3 code generation, `StructInfo.Reserved` is propagated to
`P3Message.Reserved` and emitted as `reserved N;` statements, keeping the
generated `.proto` files consistent with the Go struct.

Two misuse cases are rejected at codec initialisation with a panic:

- A `_` field without `amino:"reserved"` — the blank identifier could have
  been written intending to reserve a slot, but without the tag it has no
  effect, so the mismatch is a hard error.
- An `amino:"reserved"` tag on a named field — `reserved` is only meaningful
  on blank identifier fields; applying it elsewhere would silently do nothing.

## Alternatives considered

1. **Struct-level tag** (`amino:"reserved=2,5"` on the struct itself) — the
   reserved numbers would be detached from the field ordering, making it easy
   to get the numbers wrong and harder to review. The blank identifier approach
   keeps the reservation in-line with the other field declarations.

2. **Named unexported field** (`_removed string`) — unexported fields are
   already silently skipped, so this would require a new convention to
   distinguish "intentionally reserved" from "legitimately private". Blank
   identifiers are already a Go convention for "placeholder, not used".

3. **Post-decode field number remapping** — translate incoming field numbers
   through a mapping table at decode time. More flexible but significantly more
   complex and harder to audit.

## Consequences

- Struct definitions are self-documenting about schema evolution: the presence
  and position of <code>_ struct{} `amino:"reserved"`</code> makes it clear
  which field numbers are intentionally retired.
- Backward-compatible decoding: old encoded data with a now-removed field
  round-trips correctly into a struct that marks that field as reserved.
- Generated Protobuf3 output includes `reserved N;` statements, keeping `.proto`
  files in sync with the Go definitions.
- The `[0]struct{}` type is zero-size, so reserved fields add no memory overhead
  to struct values.
- Misuse (blank field without tag, or tag on named field) fails loudly at codec
  initialisation rather than silently producing wrong field numbers.
