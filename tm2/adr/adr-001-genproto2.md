# ADR-001: genproto2 — Direct Protobuf3 Wire Encoding

## Status

Accepted

## Context

Amino's binary encoding produces protobuf3-compatible wire bytes. The existing
`genproto` package achieves faster encoding by:

1. Generating `.proto` schema files from Go types
2. Running `protoc` to produce `.pb.go` files (Go structs implementing `proto.Message`)
3. Generating `pbbindings.go` to bridge Go structs to/from PB structs
4. At runtime: Go struct → PB struct → `proto.Marshal()` → bytes

This pipeline has several drawbacks:

- **External tooling dependency**: requires `protoc` to be installed
- **Double allocation**: every marshal/unmarshal creates an intermediate PB struct
- **Code bloat**: each type gets a `.pb.go` file, a pbbindings file, and a `.proto` file
- **Complexity**: the binding generator must handle Go↔PB struct conversion for every field type

## Decision

Introduce `genproto2`, a new code generator that produces Go methods which
directly read and write protobuf3 wire format bytes using amino's existing
`encoder.go`/`decoder.go` primitives. No intermediate PB structs, no `.pb.go`
files, no `protoc`.

### Generated interface

Each registered struct type gets three methods:

```go
type PBMessager2 interface {
    MarshalBinary2(cdc *Codec, buf []byte, offset int) (int, error)
    SizeBinary2(cdc *Codec) int
    UnmarshalBinary2(cdc *Codec, bz []byte) error
}
```

`MarshalBinary2` writes backward into a pre-allocated buffer (see
[Backward encoding](#backward-encoding) below). `SizeBinary2` computes the
exact encoded size so the buffer can be allocated once.

### Integration with Codec

`Codec.Marshal()` and `Codec.Unmarshal()` check for `PBMessager2` first,
before falling back to pbbindings or reflection:

```go
func (cdc *Codec) Marshal(o any) ([]byte, error) {
    if pbm2, ok := o.(PBMessager2); ok {
        return cdc.MarshalBinary2(pbm2)
    }
    if cdc.usePBBindings { ... }
    return cdc.MarshalReflect(o)
}

func (cdc *Codec) MarshalBinary2(pbm2 PBMessager2) ([]byte, error) {
    n := pbm2.SizeBinary2(cdc)
    buf := make([]byte, n)
    offset, err := pbm2.MarshalBinary2(cdc, buf, n)
    return buf[offset:], err
}
```

### Proto schema generation

genproto2 also generates `.proto` schema files (identical output to genproto),
so it serves as a complete replacement for genproto in production. genproto is
retained only for pbbindings fuzz testing (3-way comparison).

## Key invariant

**Generated `MarshalBinary2` must produce bytes identical to
`amino.MarshalReflect`** for all inputs. This is verified by a 3-way fuzz test
that compares amino reflect, pbbindings, and genproto2 output byte-for-byte
across 10,000 random inputs per type.

## Generator structure

| File | Purpose |
|------|---------|
| `genproto2/genproto2.go` | P3Context2, type iteration, file output, `.proto` schema delegation |
| `genproto2/gen_marshal.go` | Generate `MarshalBinary2()` method bodies |
| `genproto2/gen_unmarshal.go` | Generate `UnmarshalBinary2()` method bodies |
| `genproto2/gen_size.go` | Generate `SizeBinary2()` method bodies |

The standalone generator lives at `misc/genproto2/` and produces `pb3_gen.go`
and `.proto` files for all 17 registered amino packages.

## Backward encoding

`MarshalBinary2` writes fields **backward** into a pre-sized `[]byte` buffer,
processing struct fields in reverse order. Each field's value is written first,
then its length prefix (for ByteLength types), then its field key. This design
has two key advantages over forward writing:

1. **No temporary buffers for nested structs.** Length-prefixed fields (nested
   structs, time, duration) require knowing the encoded size before writing the
   length prefix. Forward encoding must encode into a temporary buffer to learn
   the size, then copy. Backward encoding writes the nested data first — the
   size is simply `before - offset`.

2. **No `io.Writer` interface dispatch.** All writes are direct `[]byte` slice
   operations via `Prepend*` helper functions (`encoder.go`), eliminating
   virtual method call overhead.

The buffer is pre-allocated via `SizeBinary2()`, giving a single allocation per
marshal call. The final encoded bytes are `buf[offset:]`.

## Encoding details

The generated code handles:

- All primitive types (bool, intN, uintN, float, string, []byte)
- Fixed-width encoding (`binary:"fixed32"` / `binary:"fixed64"` tags)
- Nested structs (length-prefixed, recursive)
- Slices and arrays (packed for varint/fixed elements, unpacked for ByteLength)
- Nested lists via implicit struct wrappers (proto3 limitation workaround)
- Pointer fields (nil → omit, non-nil → encode value)
- Embedded structs (flattened field iteration)
- AminoMarshaler types (MarshalAmino/UnmarshalAmino delegation)
- Interface fields (google.protobuf.Any encoding via Codec type URL lookup)
- time.Time / time.Duration (google.protobuf.Timestamp/Duration wire format)

### Default value initialization

Amino's decoder initializes missing pointer fields to default values:
- `*time.Time` → `&time.Unix(0,0).UTC()`
- `*<non-struct>` (e.g. `*int8`, `*string`, `*[]byte`) → `new(T)`
- `*<struct>` → `nil`

genproto2's unmarshal matches this behavior by initializing nil pointer fields
after the decode loop.

## Verification

1. **Unit tests**: `go test ./tm2/pkg/amino/genproto2/...`
2. **3-way fuzz**: `TestCodecStruct` and `TestCodecDef` compare amino reflect,
   pbbindings, and genproto2 byte output across 10,000 random inputs per type
3. **Native Go fuzzer**: `FuzzUnmarshalBinary2` feeds random bytes to
   UnmarshalBinary2 to ensure no panics
4. **Roundtrip tests**: handcrafted and fuzz-based encode→decode→re-encode checks

Run all fuzz testers: `make fuzz` (default 1 hour, configurable via `FUZZTIME`).

## Consequences

### Positive

- No `protoc` dependency
- No intermediate PB struct allocation — direct Go struct → wire bytes
- Simpler generated code (single `pb3_gen.go` per package vs `.pb.go` + pbbindings)
- Faster marshal/unmarshal (single allocation, no reflection, no io.Writer dispatch)
- ~2x faster encode than pbbindings, ~6x faster than reflect
- Complete replacement for genproto in production use

### Negative

- genproto must be kept around for pbbindings fuzz testing (3-way comparison)
- Generated code is tightly coupled to amino's wire format semantics

### Neutral

- `.proto` schema generation is delegated to genproto's existing code (shared logic)
