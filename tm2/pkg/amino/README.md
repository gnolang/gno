# Amino

Amino is an encoding/decoding library for structures.

In Amino, all structures are declared in a restricted set of Go. From those
declarations, Amino generates Protobuf3-compatible binary bytes.

Amino supports three encoding paths, from slowest to fastest:

1. **Amino reflect** — reflection-based encoding/decoding. No code generation
   required. This is the baseline.

2. **pbbindings** (genproto) — generates `.proto` files via `protoc`, then
   generates Go translation code between amino Go structs and protobuf-generated
   Go structs. Faster than reflect (~3x encode), but requires `protoc` and
   produces intermediate proto.Message allocations.

3. **genproto2** — generates `MarshalBinary2`/`UnmarshalBinary2` methods that
   encode directly to protobuf3 wire format without intermediate structs, proto
   files, or `protoc`. Uses backward-writing into a pre-sized buffer for
   single-allocation marshaling.

### Performance (Apple M2)

Encode (ns/op):

| Type               | genproto2 | native protobuf | amino reflect | gp2 vs proto | gp2 vs reflect |
|--------------------|-----------|-----------------|---------------|--------------|----------------|
| EmptyStruct        |       3.3 |            25.5 |          68.4 |        7.8x  |          20.7x |
| PrimitivesStruct   |       247 |             348 |         1,434 |        1.4x  |           5.8x |
| ArraysStruct       |       886 |             946 |         5,622 |        1.07x |           6.3x |
| ArraysArraysStruct |     1,350 |           1,768 |        12,028 |        1.3x  |           8.9x |
| SlicesSlicesStruct  |     5,899 |           7,468 |        35,967 |        1.27x |           6.1x |

Decode (ns/op):

| Type               | genproto2 | native protobuf | amino reflect | gp2 vs proto | gp2 vs reflect |
|--------------------|-----------|-----------------|---------------|--------------|----------------|
| EmptyStruct        |      11.5 |            49.6 |          44.6 |        4.3x  |           3.9x |
| PrimitivesStruct   |       381 |             386 |           914 |       ~1x    |           2.4x |
| ArraysStruct       |     1,548 |           2,014 |         4,027 |        1.3x  |           2.6x |
| SlicesSlicesStruct  |    31,496 |          15,802 |        60,615 |        0.5x  |           1.9x |

genproto2 encode is consistently faster than native protobuf. Decode wins on
small/medium types but native protobuf wins on deeply nested slices (fewer
allocations in the protobuf runtime). Decode allocation optimization is future
work.

Though Amino supports a subset of Protobuf3 and uses it to optimize encoding
and decoding, it is NOT intended to be a Protobuf3 library — complete support of
Protobuf3 is explicitly not its design goal.

## Getting Started

### Registering types and packages

Each package should declare in a package-local file (by convention called
`amino.go`) which should look like the following:

```go
package main

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"main", // The Go package path
		"main", // The (shorter) Proto3 package path (no slashes).
		amino.GetCallersDirname(),
	).WithDependencies(
		submodule.Package,
	).WithTypes(
		StructA{},
		StructB{},
		&StructC{}, // Pointer receiver preferred when decoding to interfaces.
	),
)
```

You can still override global registrations with local `*amino.Codec` state.
This is used by `genproto.P3Context`, which may help during migration.

## Testing and Fuzzing

See the [Makefile](Makefile) for all available targets. Key ones:

```
make test                  # run all unit tests
make fuzz                  # run all fuzz testers (see below)
make fuzz FUZZTIME=30m     # customize fuzz duration (default: 1h)
```

### Fuzz targets

| Target | Tool | What it tests | Codec |
|--------|------|---------------|-------|
| `make fuzz` (3-way comparison) | `go test -count=999999` | Roundtrip equality: amino reflect and genproto2, and pbbindings where `make -C misc/genproto` generated them, produce identical bytes for random valid structs | All three |
| `make fuzz` (native Go fuzzer) | `go test -fuzz` (Go 1.18+) | `UnmarshalBinary2` doesn't panic on random bytes (coverage-guided) | genproto2 |
| `make gofuzz_binary` | `go-fuzz` (external tool) | amino reflect `Unmarshal` doesn't panic on random bytes | amino reflect |
| `make gofuzz_json` | `go-fuzz` (external tool) | amino reflect `JSONUnmarshal` doesn't panic on random JSON | amino reflect |

The `make fuzz` target runs the 3-way comparison and native Go fuzzer
sequentially. The `gofuzz_binary` and `gofuzz_json` targets require the
external `go-fuzz` tool and test the older amino reflect codec only.

## Field tags

| Tag | Effect |
|-----|--------|
| `json:"name"` | The field's JSON name. |
| `json:",omitempty"` | JSON leaves out an empty value: zero, or an empty string, slice or array. Binary is unaffected. |
| `json:"-"` | The field is not encoded at all, in binary or JSON. |
| `binary:"fixed64"`, `binary:"fixed32"` | An integer is written as 8 or 4 fixed bytes instead of a varint. |
| `binary:"varint"` | A signed integer is written as a plain varint instead of zigzag. |
| `amino:"unsafe"` | Allows a floating point field. |
| `amino:"write_empty"` | An empty or zero value is written instead of left out. |
| `amino:"nil_elements"` | Nil elements of a list of pointers are kept, and decode as nil. |
| `amino:"skip_zero"` | The field is left out of binary and JSON whenever its Go value is zero. |
| `amino:"reserved"` | On a blank `_` field: keeps a removed field's number from being reused. |

Several `amino` tags go in one comma-separated list, as in `amino:"unsafe,skip_zero"`.

Binary leaves out a field whose value encodes to nothing, which covers the zero
value of integers, strings, slices, nil pointers and interfaces. A zero value
that encodes to bytes is written: a zero float or fixed-size array, a struct
holding one, or a value whose `MarshalAmino` repr is not empty, such as a zero
`crypto.Address`, written as its bech32 string in 42 bytes.
`amino:"skip_zero"` leaves such a field out by testing the Go value itself, as
`reflect.Value.IsZero` does, which since Go 1.22 agrees with `==` (a float
`-0.0` is zero). A missing field decodes back to the zero value, and an
explicitly encoded zero value still decodes too, so data written before a
field took the tag stays readable. Registration refuses the tag together with
`amino:"write_empty"`, and wherever a missing field decodes to something other
than the zero value: on `time.Time` (1970-01-01) and on a pointer to anything
but a struct (a pointer to the zero value).

Adding `amino:"skip_zero"` to an existing field changes the bytes written for
its zero value, in binary and JSON. Where those bytes are hashed or signed,
that is a protocol change.

## Unsupported types

### Floating points
Floating point number types are discouraged as [they are generally
non-deterministic](https://gafferongames.com/post/floating_point_determinism/).
If you need to use them, use the field tag `amino:"unsafe"`.

### Enums
Enum types are not supported in all languages, and they're simple enough to
model as integers anyways.

### Maps
Maps are not currently supported. There is unstable experimental support for
maps in the Amino:JSON codec, but it shouldn't be relied on.

## Amino and Proto3

Amino objects are a subset of Proto3.
* Enums are not supported.
* Nested message declarations are not supported.

Amino extends Proto3's Any system with a particular concrete type
identification format (disfix bytes).

## Amino and Go

Amino objects are a subset of Go.
* Floats are nondeterministic, so aren't supported by default.
* Chans, funcs, and maps are not supported.
* Nested pointers are not allowed.
* Pointers are automatically supported in go-amino but it is an extension of
  the theoretical Amino spec.

## Limitations

* Pointer types in arrays and slices lose pointer information.
* Nested pointers are not allowed.
* Recursive ReprType not allowed.

## Links

* Amino:binary spec in [Tendermint](https://github.com/tendermint/tendermint/blob/main/spec/core/encoding.md)
