# ADR-002: Amino JSON Export for VM Queries

## Status

Accepted

## Context

The VM query endpoints (`qeval`, `qobject`) need to return structured JSON
representations of Gno values. Previously, this was implemented with a set of
custom JSON types (`JSONField`, `JSONStructValue`, `JSONArrayValue`,
`JSONMapValue`, `JSONMapEntry`, `JSONObjectInfo`) that duplicated much of the
existing Amino encoding logic with a different output format.

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
   - Ephemeral cycles are broken with `RefValue{ObjectID: ":N"}` (zero PkgID,
     synthetic incremental NewTime)
   - Declared types in the `T` field are replaced with `RefType{ID: "pkg.Name"}`
   - All values are defensively copied to prevent accidental mutation

2. **amino.MarshalJSON** serializes the exported values using the standard
   Amino JSON encoding, which includes `@type` discriminator tags for
   polymorphic types.

### Query Endpoints

**qeval** returns:
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

**qobject** returns:
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

### Value Encoding

- **Primitives** (int, bool, float, etc.): Stored in the `N` field as base64-encoded
  8-byte values. Strings are stored in `V` as `StringValue{value: "..."}`.
- **Structs**: `StructValue` with `ObjectInfo` and positional `Fields` array.
  Field names are not included (they live in the type definition).
- **Pointers**: `PointerValue` with `Base` (a `RefValue` for persisted objects,
  or inline `HeapItemValue` for ephemeral).
- **Slices/Arrays**: `SliceValue` with `Base` pointing to `ArrayValue`. Byte
  arrays use `Data` (base64), others use `List`.
- **Maps**: `MapValue` with a linked list of key-value pairs.
- **Persisted objects**: Replaced with `RefValue{ObjectID: "hash:N"}` which can
  be followed via `qobject`.
- **Nil pointers**: `V` field is omitted (Amino omitempty).
- **Declared types**: `T` field uses `RefType{ID: "pkg.TypeName"}` instead of
  the full type definition.

### Error Extraction

If the function signature's last return type implements `error`, the `@error`
field is populated by calling `.Error()` on the value. This call is panic-safe:
- Out-of-gas panics are re-panicked for proper gas accounting
- Other panics (buggy `.Error()` methods) are caught; `@error` is omitted

### Object Graph Traversal

Clients can traverse the persisted object graph by:
1. Calling `qeval` to get the root value (contains `RefValue` references)
2. Following `ObjectID` references via `qobject`
3. For pointer fields: `HeapItemValue` -> `StructValue` (alternating)

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
