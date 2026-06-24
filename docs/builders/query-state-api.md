# Querying On-Chain State (JSON APIs)

Gno.land exposes a set of ABCI query endpoints that return structured JSON
representations of on-chain state. These are designed for programmatic access
by frontends, explorers, and developer tools — as opposed to the text-oriented
endpoints documented in [Interacting with gnokey](../users/interact-with-gnokey.md#querying-a-gnoland-network).

All endpoints are accessed via `ABCIQuery` with path `vm/<endpoint>` and a
`-data` payload. They return Amino JSON, the standard encoding used by Gno's
type system.

## Endpoints

| Endpoint | Data | Returns |
|---|---|---|
| `vm/qeval_json` | `<pkgpath>.<Expr>` | Amino JSON of the evaluated expression |
| `vm/qpkg_json` | `<pkgpath>` | Named top-level variables of a package |
| `vm/qobject_json` | `<objectid>` | Children of a persisted object |
| `vm/qtype_json` | `<typeid>` | Type definition (struct fields, etc.) |

### `vm/qeval_json`

Evaluates an expression in read-only mode and returns the result as Amino JSON
instead of a printed string. This is the JSON counterpart to `vm/qeval`.

```bash
gnokey query vm/qeval_json --data 'gno.land/r/demo/counter.GetCounter()'
```

```json
[{"T":{"@type":"/gno.PrimitiveType","value":"32"},"V":{"@type":"/gno.StringValue","value":"10"}}]
```

The response is an array of `TypedValue` objects — one per return value. See
[Amino JSON format](#amino-json-format) below.

### `vm/qpkg_json`

Returns the named top-level variables of a package as an object with `names`
and `values` arrays. This is the entry point for exploring a package's state.

```bash
gnokey query vm/qpkg_json --data 'gno.land/r/demo/counter'
```

```json
{
  "names": ["counter", "Increment", "Render"],
  "values": [
    {"T":{"@type":"/gno.PrimitiveType","value":"32"},"V":{"@type":"/gno.StringValue","value":"10"}},
    {"T":{"@type":"/gno.FuncType", ...},"V":{"@type":"/gno.FuncValue", ...}},
    {"T":{"@type":"/gno.FuncType", ...},"V":{"@type":"/gno.FuncValue", ...}}
  ]
}
```

Each entry in `values` corresponds to the same index in `names`. Variables that
hold persisted objects (structs, slices, maps, etc.) will contain a `RefValue`
with an `ObjectID` that can be drilled into with `vm/qobject_json`.

### `vm/qobject_json`

Retrieves the fields or elements of a persisted object by its ObjectID. Use
this to drill into values returned by `vm/qpkg_json` or other objects.

```bash
gnokey query vm/qobject_json --data '0186fce2acb457084a538e1c8b26f0f2b30e1d44:2'
```

```json
[
  {"N":"AQAAAAAAAAA=","T":{"@type":"/gno.PrimitiveType","value":"4"}},
  {"T":{"@type":"/gno.PointerType", ...},"V":{"@type":"/gno.PointerValue", ...}},
  {"T":{"@type":"/gno.PrimitiveType","value":"32"},"V":{"@type":"/gno.StringValue","value":"5"}}
]
```

The response is an array of `TypedValue` objects — one per field (for structs)
or element (for arrays/slices). Struct fields are returned by index; use
`vm/qtype_json` to resolve field names.

### `vm/qtype_json`

Retrieves a type definition by its TypeID. Primarily used to resolve struct
field names for objects returned by `vm/qobject_json`.

```bash
gnokey query vm/qtype_json --data 'gno.land/p/demo/avl.Node'
```

```json
{
  "typeid": "gno.land/p/demo/avl.Node",
  "type": {
    "@type": "/gno.DeclaredType",
    "PkgPath": "gno.land/p/demo/avl",
    "Name": "Node",
    "Base": {
      "@type": "/gno.StructType",
      "Fields": [
        {"Name": "key", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}},
        {"Name": "value", "Type": {"@type": "/gno.InterfaceType"}},
        {"Name": "height", "Type": {"@type": "/gno.PrimitiveType", "value": "256"}},
        {"Name": "size", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}},
        {"Name": "leftNode", "Type": {"@type": "/gno.PointerType", ...}},
        {"Name": "rightNode", "Type": {"@type": "/gno.PointerType", ...}}
      ],
      ...
    },
    ...
  }
}
```

## Amino JSON Format

All JSON endpoints use Amino encoding — Gno's native type serialization. Each
value is represented as a `TypedValue` with up to three fields:

| Field | Description |
|---|---|
| `T` | Type descriptor with an `@type` discriminator |
| `V` | Value payload (strings, structs, refs, etc.) |
| `N` | Base64-encoded 8-byte little-endian numeric value (for primitives) |

### Type discriminators (`T.@type`)

| `@type` | Kind |
|---|---|
| `/gno.PrimitiveType` | bool, int, uint, string, etc. (value is the numeric kind ID) |
| `/gno.PointerType` | Pointer to another type |
| `/gno.ArrayType` | Fixed-length array |
| `/gno.SliceType` | Slice |
| `/gno.StructType` | Struct with named fields |
| `/gno.MapType` | Map |
| `/gno.FuncType` | Function signature |
| `/gno.InterfaceType` | Interface |
| `/gno.DeclaredType` | Named type wrapping a base type |
| `/gno.RefType` | Lazy reference to a type (resolved via `qtype_json`) |

### Value discriminators (`V.@type`)

| `@type` | Description |
|---|---|
| `/gno.StringValue` | String value |
| `/gno.StructValue` | Inline struct fields |
| `/gno.ArrayValue` | Inline array elements |
| `/gno.SliceValue` | Slice with base/offset/length/maxcap |
| `/gno.PointerValue` | Pointer to a value |
| `/gno.MapValue` | Map with key-value list |
| `/gno.FuncValue` | Function closure |
| `/gno.RefValue` | Reference to a persisted object (has `ObjectID`) |
| `/gno.TypeValue` | Reified type |

### Primitive type IDs

The `PrimitiveType` value field is a numeric ID (powers of 2):

| ID | Type | ID | Type |
|---|---|---|---|
| 4 | bool | 2048 | uint |
| 16 | string | 4096 | uint8 |
| 32 | int | 8192 | uint16 |
| 64 | int8 | 32768 | uint32 |
| 128 | int16 | 65536 | uint64 |
| 512 | int32 | 1048576 | float32 |
| 1024 | int64 | 2097152 | float64 |

Primitive values are encoded in the `N` field as base64 of an 8-byte
little-endian integer (for numerics and bool), or in the `V` field as a
`StringValue` (for strings, and for int/uint which may exceed 64 bits).

### Lazy references

Large or nested values are not inlined — they are replaced with `RefValue`:

```json
{"V": {"@type": "/gno.RefValue", "ObjectID": "0186fce2...:4"}}
```

Use `vm/qobject_json` with the `ObjectID` to fetch the object's contents.

Similarly, declared types may appear as `RefType`:

```json
{"T": {"@type": "/gno.RefType", "ID": "gno.land/p/demo/avl.Node"}}
```

Use `vm/qtype_json` with the `ID` to resolve the type definition.

## Traversal Pattern

A typical client traverses the state tree as follows:

1. **`vm/qpkg_json`** — get named package variables
2. For each variable with a `RefValue`, call **`vm/qobject_json`** with its `ObjectID`
3. If the object's type is a `RefType`, call **`vm/qtype_json`** with its `ID` to get struct field names
4. Repeat step 2 recursively for nested `RefValue` references

This lazy-loading pattern avoids transferring the entire object graph upfront.

## Client Libraries

- **[@gnojs/amino](../../misc/gnojs/)** — TypeScript library that decodes Amino JSON into a navigable tree of `StateNode` objects. Handles all value types, primitive decoding, and struct field name resolution.
- **[gnoclient](https://gnolang.github.io/gno/github.com/gnolang/gno/gno.land/pkg/gnoclient.html)** — Go client (use `ABCIQuery` with the paths above)
- **[gno-js-client](https://github.com/gnolang/gno-js-client)** / **[tm2-js-client](https://github.com/gnolang/tm2-js-client)** — JavaScript/TypeScript clients for RPC access

## See Also

- [Interacting with gnokey](../users/interact-with-gnokey.md#querying-a-gnoland-network) — text-oriented query endpoints (`vm/qrender`, `vm/qfile`, `vm/qeval`, etc.)
- [Connecting Clients and Applications](connect-clients-and-apps.md) — client library overview
