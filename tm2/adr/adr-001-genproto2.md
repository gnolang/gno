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

## Benchmarks

All benchmarks run on Apple M1, Go 1.22, `-benchmem -count=1`.

### Encode (ns/op)

| Type | genproto2 | reflect | pbbindings | vs reflect | vs pbbindings | allocs (g/r/p) |
|------|-----------|---------|------------|-----------|--------------|----------------|
| EmptyStruct | 3.7 | 74 | 33 | 20x | 9.0x | 0/0/0 |
| PrimitivesStruct | 288 | 1,685 | 565 | 5.9x | 2.0x | 1/50/4 |
| ShortArraysStruct | 4.5 | 295 | 34 | 65x | 7.6x | 0/0/0 |
| ArraysStruct | 1,062 | 6,914 | 2,207 | 6.5x | 2.1x | 1/158/32 |
| ArraysArraysStruct | 10,491 | 16,522 | 5,582 | 1.6x | 0.5x | 159/269/106 |
| SlicesStruct | 1,395 | 7,967 | 3,271 | 5.7x | 2.3x | 1/180/31 |
| SlicesSlicesStruct | 43,226 | 45,458 | 18,237 | 1.1x | 0.4x | 900/940/210 |
| PointersStruct | 342 | 1,942 | 643 | 5.7x | 1.9x | 1/49/5 |
| PointerSlicesStruct | 1,411 | 7,986 | 3,251 | 5.7x | 2.3x | 1/172/32 |
| ComplexSt | 3,202 | 19,931 | 7,807 | 6.2x | 2.4x | 1/438/70 |
| EmbeddedSt1 | 322 | 1,910 | 653 | 5.9x | 2.0x | 1/52/5 |
| FuzzDeepNest | 5,087 | 42,914 | 16,694 | 8.4x | 3.3x | 1/767/167 |
| FuzzPtrNest | 224 | 1,378 | 610 | 6.2x | 2.7x | 1/30/7 |

### Decode (ns/op)

| Type | genproto2 | reflect | pbbindings | vs reflect | vs pbbindings | allocs (g/r/p) |
|------|-----------|---------|------------|-----------|--------------|----------------|
| EmptyStruct | 14 | 52 | 77 | 3.8x | 5.7x | 0/0/1 |
| PrimitivesStruct | 468 | 1,153 | 587 | 2.5x | 1.3x | 5/8/6 |
| ArraysStruct | 1,983 | 4,956 | 3,372 | 2.5x | 1.7x | 32/42/52 |
| SlicesStruct | 4,018 | 13,464 | 4,754 | 3.4x | 1.2x | 75/246/69 |
| SlicesSlicesStruct | 39,088 | 77,212 | 27,154 | 2.0x | 0.7x | 697/1372/446 |
| ComplexSt | 6,795 | 19,057 | 8,020 | 2.8x | 1.2x | 136/320/150 |
| FuzzDeepNest | 12,847 | 24,597 | 14,821 | 1.9x | 1.2x | 258/319/236 |

genproto2 encode typically uses **1 alloc/op** (the pre-sized output buffer).
The two regressions (ArraysArraysStruct, SlicesSlicesStruct) are nested-list
types where proto3's implicit struct wrapper requires per-element allocation.

## Consequences

### Positive

- No `protoc` dependency for encoding/decoding
- No intermediate PB struct allocation — direct Go struct → wire bytes
- Simpler generated code (single `pb3_gen.go` per package vs `.pb.go` + pbbindings)
- Faster marshal/unmarshal (single allocation, no reflection, no io.Writer dispatch)
- ~2–3x faster encode than pbbindings, ~6x faster than reflect (typical types)
- Complete replacement for genproto in production use

### Negative

- Generated code is tightly coupled to amino's wire format semantics

### Why genproto and protobuf are retained

- **Interoperability**: the protobuf dependency (`google.golang.org/protobuf`) is
  kept so that users who want to use protobuf for other purposes (gRPC, external
  APIs, cross-language communication) can do so without adding a separate dependency.
- **3-way fuzz testing**: genproto's pbbindings provide an independent encoding
  path for byte-exact comparison against genproto2 and amino reflect, catching
  bugs that roundtrip tests alone would miss.
- **`.proto` schema generation**: genproto2 delegates `.proto` file generation to
  genproto's existing code, avoiding duplication.

### Neutral

- `.proto` schema generation is delegated to genproto's existing code (shared logic)
