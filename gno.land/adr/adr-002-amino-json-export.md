# ADR-002: Amino JSON Export for VM Queries

## Status

Accepted

## Context

Three new VM query endpoints (`qeval_json`, `qobject_json`, `qobject_binary`)
need to return structured representations of Gno values. An earlier draft
implemented this with a set of custom JSON types (`JSONField`,
`JSONStructValue`, `JSONArrayValue`, `JSONMapValue`, `JSONMapEntry`,
`JSONObjectInfo`) that duplicated much of the existing Amino encoding logic
with a different output format.

This created two problems:

1. **Maintenance burden**: ~1100 lines of custom serialization code that had to
   be kept in sync with the value types and Amino encoding.
2. **Format divergence**: The custom JSON format was different from the Amino
   JSON format used everywhere else in the system (persistence, binary
   encoding, etc.), requiring consumers to understand two formats.

## Decision

Replace the custom JSON types with standard Amino JSON encoding. The export
process is:

1. **ExportValues / ExportObject** (in `gnovm/pkg/gnolang/values_export.go`)
   walks the value tree and produces a defensive copy where:
   - Persisted (real) objects are replaced with `RefValue{ObjectID: "hash:N"}`
   - Ephemeral (unpersisted) Objects seen more than once are replaced on
     subsequent visits with `ExportRefValue{ObjectID: ":N"}`, where `N` is an
     incrementing counter assigned in the encoder's DFS traversal order (see
     "Ephemeral Reference Resolution" below)
   - Declared types in the `T` field are replaced with `RefType{ID: "pkg.Name"}`
   - All values are defensively copied to prevent accidental mutation

2. **amino.MarshalJSON** serializes the exported values using the standard
   Amino JSON encoding, which includes `@type` discriminator tags for
   polymorphic types.

### Query Endpoints

**qeval_json** returns:
```json
{
  "results": [
    {
      "T": {"@type": "/gno.RefType", "ID": "example.Item"},
      "V": {
        "@type": "/gno.StructValue",
        "ObjectInfo": {"ID": ":0", ...},
        "Fields": [
          {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "KQAAAAAAAAA="},
          {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "test"}}
        ]
      }
    }
  ],
  "@error": "optional error string"
}
```

**qobject_json** returns the Amino-JSON envelope below.
**qobject_binary** returns the same envelope Amino-binary-encoded (see
"Amino Type Registry for `qobject_binary`" below).
```json
{
  "objectid": "hash:N",
  "value": {
    "@type": "/gno.StructValue",
    "ObjectInfo": {...},
    "Fields": [...]
  }
}
```

`RefValue` fields — whether appearing inline in `qeval_json` output or as a
child reference inside a `qobject_json` body — carry `ObjectID`, `Hash`,
and (when relevant) `Escaped`:
```json
{
  "@type": "/gno.RefValue",
  "ObjectID": "68fac97482b733b2035b59fff3b5e13f0e453257:4",
  "Hash": "7bab3bb08b98414b4c8d41ecbcad91827c60a990"
}
```
`Hash` is the content hash at the time of export; clients that cache
sub-object responses should key by `(ObjectID, Hash)` pairs to detect
realm-level mutations between queries.

### Value Encoding

- **Primitives** (int, bool, float, etc.): Stored in the `N` field as base64-encoded
  8-byte values. Strings are stored in `V` as `StringValue{value: "..."}`.
- **Structs**: `StructValue` with `ObjectInfo` and positional `Fields` array.
  Field names are not included (they live in the type definition). All fields
  are emitted, including unexported (lowercase) ones — see
  "Visibility of Unexported Fields" below.
- **Pointers**: `PointerValue` with `Base` (a `RefValue` for persisted objects,
  or inline `HeapItemValue` for ephemeral).
- **Slices/Arrays**: `SliceValue` with `Base` pointing to `ArrayValue`. Byte
  arrays use `Data` (base64), others use `List`.
- **Maps**: `MapValue` with a linked list of key-value pairs. Because gno maps
  can have non-string keys (ints, structs, pointers), the wire shape is a
  positional list of `{Key, Value}` tuples — not a JSON object. See
  "Map Encoding" below.
- **Persisted objects**: Replaced with `RefValue{ObjectID: "hash:N"}` which can
  be followed via `qobject_json` or `qobject_binary`.
- **Nil pointers**: `V` field is omitted (Amino omitempty).
- **Declared types**: `T` field uses `RefType{ID: "pkg.TypeName"}` instead of
  the full type definition.

### Map Encoding

Gno maps allow arbitrary key types (int, struct, pointer, interface), so the
encoded shape cannot use JSON object syntax (which requires string keys).
Instead a `MapValue` serializes as an ordered `MapList` — a linked list of
`MapListItem{Key, Value}` pairs — preserving insertion order deterministically
across nodes. Consumers that want an idiomatic JSON object for
`map[string]T` must reconstruct it client-side by walking the list and using
each `Key`'s string value.

### Visibility of Unexported Fields

The exporter emits every field of a `StructValue` including unexported
(lowercase) ones. This is intentional and matches gno's on-chain semantics:
all persisted realm state is already public — deterministically replayable
from block data — so concealing unexported fields in a read-only query would
give a false sense of privacy. Realm authors should not rely on
lowercase-naming as a confidentiality mechanism.

### Single-Hop Object Resolution

`qobject_json` / `qobject_binary` return the object identified by the
requested `ObjectID` expanded inline, but any child object reference remains
a `RefValue{ObjectID: ...}` in the output — the endpoint does not recursively
load and inline referenced objects. This is deliberate: it keeps per-query
cost proportional to a single persisted object blob (which is gas-metered at
load), and lets clients control traversal depth by issuing follow-up queries.
To walk an object graph, clients repeatedly call `qobject_*` on each
`RefValue.ObjectID` they want to expand.

### Malformed Query Input

`qeval` and `qeval_json` share `parseQueryEvalData`, which panics when the
input does not contain a `<pkgpath>.<expression>` separator (no `.` after
the first `/`). This is inherited behavior from the existing `qeval`
endpoint — `qeval_json` matches it for symmetry. The panic is caught by
BaseApp's ABCI recover and surfaced as a query error at the RPC layer. In
both endpoints a malformed input therefore produces an ABCI error response
rather than a structured JSON body. Clients must construct well-formed
query data; the endpoints are not forgiving of shape errors.

### Amino Type Registry for `qobject_binary`

`qobject_binary` returns raw Amino binary bytes (amino-encoded `Any` of the
exported value). Decoding requires the caller to have the same Amino type
registry as the node — i.e., the types declared in
`gnovm/pkg/gnolang/package.go`. Go clients that link
`github.com/gnolang/gno/gnovm/pkg/gnolang` get this automatically; other
clients must re-implement the registry or prefer `qobject_json`, which is
self-describing via `@type` discriminators.

### Error Extraction

`@error` is populated by calling `.Error()` on the last return value when it
implements the error interface. Two detection paths exist:

1. **Signature-based (preferred):** when the function signature is available,
   the ADR check is `IsErrorType(lastReturnType)` on the declared return.
   Used when the caller can supply a `FuncType` (e.g. from a resolved `Call`
   expression).

2. **Value-based (fallback):** when no signature is available — `qeval_json`
   takes this path, since it evaluates an arbitrary expression without a
   pre-resolved function type — the check is `tv.ImplError()` on the runtime
   value's type.

The two agree in almost all cases. They can differ only when the declared
return is an interface type that implements `error` but the evaluated value
has a concrete type that does not (e.g. via an explicit interface
conversion). This is a rare, mostly theoretical case.

The `.Error()` call is panic-safe in the following ways:

- **Out-of-gas in `.Error()`:** the already-computed `results` payload is
  preserved and `@error` is omitted (graceful degrade). The query does not
  fail just because the optional error-extraction step ran out of gas.
- **Other panics in `.Error()`:** same graceful degrade — e.g. a typed-nil
  pointer receiver that nil-derefs, or a user-bug panic inside `.Error()`.

`@error` is best-effort: its absence means either the last return does not
implement `error`, the value was nil, or `.Error()` itself failed.

### Object Graph Traversal

Clients can traverse the persisted object graph by:
1. Calling `qeval_json` to get the root value (contains `RefValue` references)
2. Following `ObjectID` references via `qobject_json` (or `qobject_binary`)
3. For pointer fields: `HeapItemValue` -> `StructValue` (alternating)

### Ephemeral Reference Resolution

`ExportRefValue{":N"}` tags back-references to ephemeral (unpersisted) Objects
that appeared earlier in the same export — typically because the value graph
contains a shared or cyclic ephemeral Object. They are emitted by
`ExportValues` / `ExportObject` in `gnovm/pkg/gnolang/values_export.go` and
serialized with `@type`: `/gno.ExportRefValue`.

Assignment protocol: the encoder performs a DFS over the result. The first
visit to an ephemeral Object expands it inline and assigns it
`N = (count of previously-seen ephemeral Objects) + 1`. Any subsequent visit
to the same Object emits `ExportRefValue{":N"}` with the ID assigned on first
visit.

**Which value kinds count as ephemeral Objects.** The counter is bumped for
every visit to an in-gno `Object` (the internal gno interface implemented by
persisted or persistable values), specifically:

- `*HeapItemValue` — the per-allocation heap slot, wrapping any heap-held
  value. In gno, a `*T` is typically represented by a `PointerValue` whose
  `Base` is a `*HeapItemValue` that in turn holds the `*T`'s body. **Both
  the heap slot and the body get their own counter slot.**
- `*StructValue`
- `*ArrayValue` (when it holds non-byte elements)
- `*SliceValue`'s `Base` (an `*ArrayValue`)
- `*MapValue`
- `*BoundMethodValue`
- `*FuncValue` (the exporter counts these; see the code comment at the
  BoundMethodValue branch for one caveat)
- `*Block`
- `*PackageValue` (converted to `RefValue{PkgPath}`, no counter slot)

Primitives (`StringValue`, `BigintValue`, `BigdecValue`, numeric-in-N) are
NOT Objects and do not bump the counter.

**Consequence for consumers:** a user-visible `*T` pointer chain of length
K produces 2K counter slots, because each `*T` materializes as a
`HeapItemValue` + `StructValue` pair. A consumer that walks only
`StructValue` occurrences to resolve `:N` will miscount. The correct
algorithm is to count every `@type` that matches one of the Object kinds
above, in the traversal order below.

Traversal order matches the declaration order of the underlying values:
- `[]TypedValue` result slices, left to right
- `StructValue.Fields`, in declared field order
- `ArrayValue.List` / `SliceValue`, by index
- `MapValue.List`, in insertion order (the gno-level `MapList`, not Go map
  iteration)
- `Block.Values`, in order
- For pointer/slice containers, `Base` is visited before child elements
- `FuncValue.Captures`, then `FuncValue.Parent`

To resolve `:N` back to its inline expansion, a consumer walks the exported
tree in this same order, counts each inline ephemeral Object (per the kind
list above) as it is encountered, and looks up the Nth one. Persisted
`RefValue{ObjectID: "hash:N"}` references are not part of this counter —
they are resolved separately via `qobject_json` / `qobject_binary`.

## Consequences

### Positive

- Eliminates ~960 lines of custom serialization code
- Single JSON format consistent with Amino encoding used elsewhere
- `RefValue` references enable lazy traversal of large object graphs
- Cycle-safe: ephemeral cycles broken with synthetic IDs

### Negative

- Amino JSON is more verbose than the custom format (includes `@type` tags,
  `ObjectInfo`, base64-encoded primitives)
- Struct field names are not in the output — consumers need type definitions
  to label fields

### Mitigations

A follow-up PR will add:
- A `vm/qtype` query endpoint for fetching type definitions (including struct
  field names)
- A client-side JavaScript library for converting Amino JSON into
  human-readable format with resolved field names
