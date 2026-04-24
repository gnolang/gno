# Binary Codec Conditionals — Reference

An exhaustive inventory of every control-flow decision in the reflect-based
binary codec. The purpose of this document is to enable mechanical
cross-reference against the genproto2 code generator (`tm2/pkg/amino/genproto2/`)
to verify that every generator-emitted method reproduces an equivalent decision
for every registered type.

For each entry: the line number, the condition expression verbatim, what
behavior it selects, and the codec invariant that would break if the condition
is absent or wrong. A final section maps each block of conditions to the
expected place in the generator.

Scope: binary encode + decode only. JSON codec is excluded. All file paths are
relative to `tm2/pkg/amino/`.

---

## Dispatch layer (`amino.go` + `codec.go`)

### `Marshal` (`amino.go:410`)

1. **line 417**: `pbm2, ok := o.(PBMarshaler2); ok && HasNativeGenproto2(reflect.TypeOf(o))`
   → dispatches to `MarshalBinary2` when the type has genuine genproto2 methods.
   → **invariant**: promoted (embedded) methods don't satisfy — only types with their own `MarshalBinary2` take the fast path.
2. **line 422-426**: `cdc.usePBBindings && ok := o.(PBMessager); ok && HasNativePbbindings(...)` → pbbindings fallback.
3. **line 428**: implicit fallback to `MarshalReflect(o)`.
   → **invariant**: every registered (and unregistered) type must be encodable via reflection.

### `MarshalReflect` (`amino.go:436`)

1. **line 439-448**: pointer dereference with nil-panic and nested-pointer panic.
2. **line 464**: `!info.IsStructOrUnpacked(fopts)` → wrap scalar/ByteLength types in implicit struct (field 1).
3. **line 488-490**: `len(bz) == 0 { bz = nil }` → normalize empty to nil at the top level.

### `Unmarshal` (`amino.go:949`)

1. **line 958-962**: `var p *T; Unmarshal(bz, &p)` ergonomic — allocate inner pointer if nil.
2. **line 967**: `ok := ptr.(PBMessager2); ok && HasNativeGenproto2(...)` → genproto2 fast path.
3. **line 972-976**: pbbindings fallback gated by `cdc.usePBBindings`.
4. **line 982**: fallback to `UnmarshalReflect(bz, ptr)`.

### `UnmarshalReflect` (`amino.go:986`)

1. **line 988**: `rv.Kind() != reflect.Ptr { return ErrNoPointer }`.
2. **line 1002**: `len(bz) == 0 && !info.IsStructOrUnpacked(FieldOptions{}) && rv.Kind() != reflect.Interface` → accept empty input as zero value for non-struct non-interface types.
   → **invariant**: symmetric with `MarshalReflect`'s nil-normalization; empty bytes round-trip to zero value.
3. **line 1015-1017**: `!info.IsStructOrUnpacked(...) && len(bz) > 0 && rv.Kind() != reflect.Interface` → unwrap implicit struct (expect field 1).
4. **line 1027**: `fnum != 1` → error.
5. **line 1030-1034**: `typ != typWanted` (from `info.GetTyp3`) → error.
6. **line 1054**: `n != len(bz)` → error (no trailing bytes).

### `typeToTyp3` (`codec.go:832`)

Dispatch on `rt.Kind()`, with options:

| Kind                                                    | Condition        | Returns          |
|---------------------------------------------------------|------------------|------------------|
| `timeType`                                              | special          | `Typ3ByteLength` |
| `durationType`                                          | special          | `Typ3ByteLength` |
| `Interface`, `Array`, `Slice`, `String`, `Struct`, `Map`| any              | `Typ3ByteLength` |
| `Int64`, `Uint64`                                       | `BinFixed64`     | `Typ38Byte`      |
| `Int64`, `Uint64`                                       | (else)           | `Typ3Varint`     |
| `Int32`, `Uint32`                                       | `BinFixed32`     | `Typ34Byte`      |
| `Int32`, `Uint32`                                       | (else)           | `Typ3Varint`     |
| `Int`, `Uint`                                           | `BinFixed64`     | `Typ38Byte`      |
| `Int`, `Uint`                                           | (else)           | `Typ3Varint`     |
| `Int16`, `Int8`, `Uint16`, `Uint8`, `Bool`              | any              | `Typ3Varint`     |
| `Float64`                                               | any              | `Typ38Byte`      |
| `Float32`                                               | any              | `Typ34Byte`      |
| default                                                 | any              | **panic**        |

→ **invariant**: mismatch between typeToTyp3 and encoder body corrupts wire format (historical: bare `int`+`BinFixed64` returned `Typ3Varint` with 8-byte body).

### `ValidateBasic` (`codec.go:146`)

- `BinFixed32` permitted only on `Int32`/`Uint32`. Panics on `Int`/`Uint` ("not yet supported"). Panics on all other kinds.
- `BinFixed64` permitted on `Int64`/`Uint64`/`Int`/`Uint`. Panics on other kinds.
- `!Unsafe` + `Float32`/`Float64` anywhere in type → panic (floats opt-in only).

### `IsStructOrUnpacked` (`codec.go:85`)

1. `Kind() == Struct || Interface` → true.
2. `Kind() == Array || Slice` → true iff `Elem.GetTyp3(fopt) == Typ3ByteLength` (unpacked list).
3. Else → false (scalar).

### `GetTyp3` (`codec.go:76`)

1. Delegates to `typeToTyp3(info.ReprType.Type, fopts)`.
   → **invariant**: typ3 is always computed from the repr type, never the primary type.

---

## Binary encode (`binary_encode.go`)

### `encodeReflectBinary` (line 46)

1. **line 49**: `rv.Kind() == reflect.Ptr` → panic (caller must deref).
2. **line 54**: `!rv.IsValid()` → panic.
3. **line 66**: `info.IsBinaryWellKnownType` → delegate to `encodeReflectBinaryWellKnown`; terminal on success.
4. **line 75**: `info.IsAminoMarshaler` → convert via `toReprObject`, then recurse with `ReprType`.
5. **line 88** `switch info.Type.Kind()`:
   - `Interface` → `encodeReflectBinaryInterface`.
   - `Array` with `Elem().Kind() == Uint8` → `encodeReflectBinaryByteArray`; else → `encodeReflectBinaryList`.
   - `Slice` with `Elem().Kind() == Uint8` → `encodeReflectBinaryByteSlice`; else → `encodeReflectBinaryList`.
   - `Struct` → `encodeReflectBinaryStruct`.
   - `Int64` with `BinFixed64` → `EncodeInt64` (8 bytes); else → `EncodeVarint`.
   - `Int32` with `BinFixed32` → `EncodeInt32`; else → `EncodeVarint`.
   - `Int16`, `Int8` → `EncodeVarint`.
   - `Int` with `BinFixed64` → `EncodeInt64`; with `BinFixed32` → `EncodeInt32`; else → `EncodeVarint`.
   - `Uint64` with `BinFixed64` → `EncodeUint64`; else → `EncodeUvarint`.
   - `Uint32` with `BinFixed32` → `EncodeUint32`; else → `EncodeUvarint`.
   - `Uint16` → `EncodeUvarint`.
   - `Uint8` with `options&beOptionByte != 0` → `EncodeByte` (bare); else → `EncodeUvarint`.
   - `Uint` with `BinFixed64` → `EncodeUint64`; with `BinFixed32` → `EncodeUint32`; else → `EncodeUvarint`.
   - `Bool` → `EncodeBool`.
   - `Float64` requires `fopts.Unsafe`; → `EncodeFloat64`. Error otherwise.
   - `Float32` requires `fopts.Unsafe`; → `EncodeFloat32`. Error otherwise.

### `encodeReflectBinaryInterface` (line 215)

1. `rv.IsNil()` → delegate to `writeMaybeBare` with nil bytes.
2. Dereferenced value kind is Interface → panic (interface-of-interface).
3. Dereferenced nil pointer → panic (nil concrete pointer).
4. `!cinfo.Registered` → panic (concrete type must be registered).
5. `!cinfo.IsStructOrUnpacked(fopts)` → wrap value in implicit struct field 1 (Any.Value encoding).
6. After encoding: `len(bz2) == 0 || (len(bz2) == 1 && bz2[0] == 0x00)` → omit Any.Value field (only emit TypeURL).

### `encodeReflectBinaryByteArray` (line 321)

1. `ert.Kind() != Uint8` → panic (contract).
2. `rv.CanAddr()` → fast path `rv.Slice(0, length).Bytes()`; else `reflect.Copy`.

### `encodeReflectBinaryByteSlice` (line 480)

1. `ert.Kind() != Uint8` → panic (contract).

### `encodeReflectBinaryList` (line 346)

1. `ert.Kind() == Uint8` → panic (byte slice/array must route to ByteSlice/ByteArray encoder).
2. **Packed vs unpacked dispatch**: `typ3 != Typ3ByteLength || (newoptions&beOptionByte > 0)` → packed form (single length-prefix, no per-element field key); else → unpacked form (field key per element).
3. **Packed element is pointer**: `ert.Kind() == Ptr` → deref; if `erv.IsNil()`, substitute `reflect.New(ert.Elem()).Elem()` (zero value).
4. **Unpacked element**:
   - `ertIsPointer := ert.Kind() == Ptr`
   - `ertIsStruct := einfo.Type.Kind() == Struct`
   - `writeImplicit := isListType(einfo.Type) && einfo.Elem.ReprType.Type.Kind() != Uint8 && einfo.Elem.ReprType.GetTyp3(fopts) != Typ3ByteLength` → wrap nested list in implicit struct.
   - For each element: write field key (fopts.BinFieldNum, Typ3ByteLength).
   - If `isNonstructDefaultValue(erv)`: if `ertIsStruct && ertIsPointer && !fopts.NilElements` → **error** "nil struct pointers in lists not supported unless nil_elements field tag is also set"; else → write `0x00` sentinel.
   - Else (non-default): if `ertIsPointer`, deref; if `writeImplicit`, wrap inner list in implicit struct field 1; otherwise encode value with length prefix.

### `encodeReflectBinaryStruct` (line 500)

1. For each field: if `!field.WriteEmpty && isNonstructDefaultValue(frv)` → skip field.
2. If `field.UnpackedList` → route through `encodeReflectBinaryList(..., bare=true)`.
3. Else → `writeFieldIfNotEmpty(buf, field.BinFieldNum, finfo, fopts, field.FieldOptions, dfrv, writeEmpty)`, where `writeEmpty := field.WriteEmpty || frvIsPtr`.

### `writeFieldIfNotEmpty` (line 568)

1. Write field key.
2. Write field value.
3. **line 592**: `!isWriteEmpty && lBeforeValue == lAfterValue-1 && buf.Bytes()[buf.Len()-1] == 0x00` → roll back the entire field (key + the single 0x00 byte).
   → **invariant**: this is the canonical "field is empty, omit it" rule. Every generator emission site must reproduce this decision via a `zeroCheck` guard BEFORE writing, because generator code writes backward and can't retroactively roll back.

### `writeMaybeBare` (line 602)

1. `len(bz) == 0 && bare` → emit nothing.
2. `len(bz) == 0 && !bare` → emit `0x00` (length-0 marker).
3. `len(bz) > 0 && bare` → emit bytes raw.
4. `len(bz) > 0 && !bare` → emit length-prefix + bytes.

### `binary_encode2.go`

Constants and wrappers (`EncodeFieldNumberAndTyp3`, `TimeSize`, `DurationSize`). No conditionals.

---

## Binary decode (`binary_decode.go`)

### `decodeReflectBinary` (line 33)

1. **line 36**: `!rv.CanAddr()` → panic.
2. **line 39**: `info.Type.Kind() == Interface && rv.Kind() == Ptr` → panic (interface decode must not be wrapped).
3. **line 55**: `info.IsBinaryWellKnownType` → delegate to well-known handler; terminal on success.
4. **line 64**: `info.IsAminoMarshaler` → decode into repr then call `UnmarshalAmino`.
5. **line 86** `switch info.Type.Kind()`:
   - `Interface` → `decodeReflectBinaryInterface`.
   - `Array` with `Elem().Kind() == Uint8` → `decodeReflectBinaryByteArray`; else → `decodeReflectBinaryArray`.
   - `Slice` with `Elem().Kind() == Uint8` → `decodeReflectBinaryByteSlice`; else → `decodeReflectBinarySlice`.
   - `Struct` → `decodeReflectBinaryStruct`.
   - `Int64` with `BinFixed64` → `DecodeInt64` (8 bytes); else → `DecodeVarint`.
   - `Int32` with `BinFixed32` → `DecodeInt32`; else → `DecodeVarint`.
   - `Int16` → `DecodeVarint16`. `Int8` → `DecodeVarint8`.
   - `Int` with `BinFixed64` → `DecodeInt64`; else → `DecodeVarint`.
   - `Uint64` with `BinFixed64` → `DecodeUint64`; else → `DecodeUvarint`.
   - `Uint32` with `BinFixed32` → `DecodeUint32`; else → `DecodeUvarint`.
   - `Uint16` → `DecodeUvarint16`.
   - `Uint8` with `options&bdOptionByte != 0` → `DecodeByte`; else → `DecodeUvarint8`.
   - `Uint` with `BinFixed64` → `DecodeUint64`; else → `DecodeUvarint`.
   - `Bool` → `DecodeBool`.
   - `Float64` requires `fopts.Unsafe`; → `DecodeFloat64`. Error otherwise.
   - `Float32` requires `fopts.Unsafe`; → `DecodeFloat32`. Error otherwise.
   - `String` → `DecodeString`.
   - default → panic.

### `decodeReflectBinaryInterface` (line 319)

1. `anyDepth > maxAnyDepth` (= 64) → error.
2. `!rv.CanAddr()` → panic.
3. `!rv.IsNil()` → error.
4. `len(bz) == 0` → set to zero interface; return.
5. Decode field 1 header: `fnum != 1 || typ != Typ3ByteLength` → error.
6. After field 1: `lenbz == 0` → zero-value concrete; return.
7. Decode field 2 header: `fnum != 2 || typ != Typ3ByteLength` → error.
8. After decode: `len(bz) > 0` → error (trailing bytes).

### `decodeReflectBinaryAny` (line 418)

1. `!IsASCIIText(typeURL)` → error.
2. `len(value) == 0` → zero-value concrete.
3. `!irvSet.Type().AssignableTo(rv.Type())` → error.
4. `!cinfo.IsStructOrUnpacked(fopts) && len(value) > 0` → unwrap implicit struct field 1.
5. Field 1 `fnum != 1` → error. `typ != typWanted` → error.
6. After decode: `len(value) > 0` → error (trailing).
7. `!irvSet.Type().AssignableTo(rv.Type())` → error (post-decode assignability).

### `decodeReflectBinaryByteArray` (line 525)

1. `!rv.CanAddr()` → panic.
2. `ert.Kind() != Uint8` → panic.
3. `len(bz) < length` → error.
4. `len(byteslice) != length` → error.

### `decodeReflectBinaryArray` (line 564)

1. `!rv.CanAddr()` → panic.
2. `ert.Kind() == Uint8` → panic (should've routed to ByteArray decoder).
3. **Packed vs unpacked**: `typ3 != Typ3ByteLength || (newoptions&beOptionByte > 0)` → packed loop `for i := range length`; else → unpacked loop `for i := range length`.
4. Packed: after loop, `len(bz) > 0` → error.
5. Unpacked pre-flags: `isErtStructPointer := ert.Kind() == Ptr && einfo.Type.Kind() == Struct`; `writeImplicit := isListType(einfo.Type) && einfo.Elem.ReprType.Type.Kind() != Uint8 && einfo.Elem.ReprType.GetTyp3(fopts) != Typ3ByteLength`.
6. Unpacked element: `fnum != fopts.BinFieldNum` → error. `typ != Typ3ByteLength` → error.
7. **nil sentinel** (line 651): `(len(bz) > 0 && bz[0] == 0x00) && (!isErtStructPointer || fopts.NilElements)` → consume 1 byte and `erv.Set(...)`:
   - `fopts.NilElements` → `reflect.Zero(erv.Type())` (nil for pointer kinds).
   - else → `defaultValue(erv.Type())` (nil for `*Struct`; `new(T)` for other non-struct pointers; `emptyTime` for `*time.Time`).
8. `writeImplicit` → decode implicit struct: field 1 key checks (`fnum != 1` → error, `ityp != Typ3ByteLength` → error), payload decode, trailing-bytes check `len(ibz) > 0` → error.
9. After all unpacked elements: `len(bz) > 0` and `fnum <= fopts.BinFieldNum` → field-number-regression error.

### `decodeReflectBinaryByteSlice` (line 736)

1. `!rv.CanAddr()` → panic.
2. `ert.Kind() != Uint8` → panic.
3. `len(bz) == 0` → set to nil slice; return.
4. After decode: `len(byteslice) == 0` → normalize to nil.

### `decodeReflectBinarySlice` (line 779)

1. `!rv.CanAddr()` → panic.
2. `ert.Kind() == Uint8` → panic (should've routed to ByteSlice decoder).
3. Packed vs unpacked dispatch: same as array.
4. Packed: `for len(bz) != 0` (unbounded) → append elements until exhausted.
5. Unpacked: `for len(bz) != 0` + break when `fnum > fopts.BinFieldNum` (next field belongs to parent).
6. Unpacked: `fnum < fopts.BinFieldNum` → error (field regression).
7. Unpacked: `typ != Typ3ByteLength` → error.
8. **nil sentinel** (line 868): same rule as array (`!isErtStructPointer || fopts.NilElements`; `reflect.Zero` vs `defaultValue`).
9. `writeImplicit` handling: same as array.

### `decodeReflectBinaryStruct` (line 939)

1. `!rv.CanAddr()` → panic.
2. Per-field loop:
   - `len(bz) == 0` → remaining fields stay at zero value; return.
   - `field.UnpackedList`:
     - `field.BinFieldNum < fnum` → skip (field absent → zero).
     - else → decode unpacked list with `bare=true`.
   - Else (packed field):
     - `field.BinFieldNum < fnum` → skip.
     - `fnum <= lastFieldNum` → error (non-monotonic).
     - `field.BinFieldNum != fnum` → error (unknown field / skipped-but-present).
     - `typ != typWanted` → error.
3. After loop: `len(bz) > 0` → error (trailing / unknown fields rejected).

### `consumeAny` (line 1056)

Switch on `typ3`; dispatches to varint/fixed32/fixed64/ByteLength skip; invalid typ3 → error.

### `decodeFieldNumberAndTyp3` (line 1082)

1. `num64 == 0` → error (field 0 reserved).
2. `num64 > (1<<29 - 1)` → error (field number overflow).

### `decodeMaybeBare` (line 1109)

1. `bare` → pass bytes as-is.
2. `!bare` → unwrap length-prefixed bytes.

### `binary_decode2.go`

`DecodeFieldNumberAndTyp3` and `SkipField` wrappers. No new conditionals.

---

## Cross-reference to genproto2

Every condition above must have an equivalent decision in the generated
`MarshalBinary2` / `SizeBinary2` / `UnmarshalBinary2` methods. The generator
lives at `tm2/pkg/amino/genproto2/`:

| Reflect-codec section                  | Generator counterpart (where to look)                                              |
|----------------------------------------|-------------------------------------------------------------------------------------|
| `Marshal` dispatch                     | `amino.go:410` (hand-written dispatch, not generator-emitted)                       |
| `MarshalReflect` implicit-struct wrap  | `gen_marshal.go:58` `writeReprMarshal` (for non-struct top-level types)             |
| `MarshalReflect` empty-bytes normalize | `gen_marshal.go` generator wrappers; see `TestCodecParity_AminoFixtures/EmptyReprOnZero/zero` |
| `UnmarshalReflect` empty-bytes         | `amino.go:1002` only; generator has its own per-type UnmarshalBinary2               |
| `UnmarshalReflect` implicit-struct     | `gen_unmarshal.go` top-level generator for non-struct registered types              |
| `typeToTyp3`                           | Used by generator via `einfo.GetTyp3(fopts)` at emission time — any fix must regen  |
| `ValidateBasic`                        | Generator need not reproduce (panic-only); relies on this at codec-init time        |
| `encodeReflectBinary` kind dispatch    | `gen_marshal.go:~300+` `writeFieldValueMarshal` + `writePrimitiveEncode`            |
| `encodeReflectBinary` AminoMarshaler   | `gen_marshal.go:232-290` AminoMarshaler branches (3 shapes: non-struct, struct+scalar-repr, struct+struct-repr) |
| `encodeReflectBinaryInterface` Any     | `gen_marshal.go` `writeInterfaceFieldMarshal` + `MarshalAnyBinary2` helper          |
| `encodeReflectBinaryList` packed/unpacked | `gen_marshal.go:~420+` `writeUnpackedListMarshal` + `writePackedListMarshal`     |
| `encodeReflectBinaryList` nil-in-list + NilElements | `gen_marshal.go:~460+` unpacked-list element emission                    |
| `encodeReflectBinaryStruct` WriteEmpty + skip-default | `gen_marshal.go:~200+` `writeFieldMarshal` per-field dispatch             |
| `writeFieldIfNotEmpty` single-0x00 rollback | `gen_marshal.go` emission sites use `zeroCheck(repr/accessor)` guards BEFORE emission; for struct-repr fields, `writeLengthPrefixedField`'s `dataLen > 0` gate |
| `writeMaybeBare`                       | `gen_marshal.go` inline; emit `0x00` or length-prefixed bytes based on context      |
| `decodeReflectBinary` kind dispatch    | `gen_unmarshal.go` per-field case statements in UnmarshalBinary2                    |
| `decodeReflectBinaryInterface` Any     | `gen_unmarshal.go` `UnmarshalAnyBinary2` + interface-field decode branches          |
| `decodeReflectBinaryArray/Slice` nil sentinel + NilElements | `gen_unmarshal.go:~585-625` unpacked-list pointer-element branches |
| `decodeReflectBinaryStruct` field-num monotonic + unknown-field rejection | `gen_unmarshal.go` per-field switch with `lastFieldNum` tracking |
| `decodeFieldNumberAndTyp3` fnum bounds | Generator calls `amino.DecodeFieldNumberAndTyp3(bz)` directly                       |
| `decodeMaybeBare`                      | Generator inlines bare vs length-prefixed logic per call site                       |

The next audit step is to walk each row of this table and verify, for at least
one representative concrete type per shape, that the generator emits Go code
whose control flow matches the reflect codec's conditions. Fixtures in the
per-package `parity_test.go` files exercise this automatically at test time —
but the audit should confirm that EVERY conditional has at least one fixture
that would fail under a generator divergence.

**Known gaps at the time of writing** (all fixed in this branch):

- `typeToTyp3` for `Int`/`Uint` + `BinFixed64` returning `Typ3Varint` instead of `Typ38Byte` (pre-fix wire was varint-key + fixed64-body).
- Generator's `dataLen > 0` emission gate not handling the single-0x00-byte case like `writeFieldIfNotEmpty` does (pre-fix: top-level StringValue emitted `0x0a 0x00` where reflect emitted nil bytes).
- Generator's AminoMarshaler zero-check using `zeroCheckOriginal` (Go-value zeroness) instead of `zeroCheck("repr", ...)` (repr-value zeroness). Affected three branches.
- Generated `UnmarshalBinary2` for `[]*Struct amino:"nil_elements"` decoding `0x00` as `&Struct{}` instead of `nil` (consensus-wedging).
- Generator's `reflect.Int`/`reflect.Uint` decode body not honoring `BinFixed64`.
- `UnmarshalReflect` rejecting empty input for non-struct top-level types — fixed to accept as zero value, matching the encoder's nil-normalization.

---

## TypeInfo construction (codec init time)

TypeInfo is the codec's cached reflective description of a Go type. It is
built once per `(*Codec, reflect.Type)` pair, either lazily on first use
(`getTypeInfoWLocked`) or eagerly at `RegisterTypeFrom` time. Decisions made
here are load-bearing for every subsequent encode/decode — the kind dispatch
in `encodeReflectBinary`/`decodeReflectBinary` reads `info.ReprType.Type.Kind()`
and `info.Fields[i].FieldOptions`, never the raw `reflect.StructField`.

### `newTypeInfoUnregisteredWLocked` (`codec.go:593`)

1. **line 594-601**: `switch rt.Kind()`:
   - `Ptr` → panic "unexpected pointer type" (getTypeInfoWLocked must deref first).
   - `Map` → panic "map type not supported".
   - `Func` → panic "func type not supported".
   → **invariant**: TypeInfo is never constructed for pointer/map/func kinds. The codec supports pointers only via `getTypeInfoWLocked`'s auto-deref (line 514-519) and the struct-field pointer-flattening in `parseStructInfoWLocked`.
2. **line 602-604, 612-614**: double existence check for `cdc.typeInfos[rt]` — first guard is a contract check, second is a re-entrance check after the zero-value `*TypeInfo` is inserted on line 615.
   → **invariant**: the stub `info` is inserted into `cdc.typeInfos` BEFORE `parseStructInfoWLocked` runs, so mutually-recursive struct types resolve to the same `*TypeInfo` pointer rather than infinitely recursing.
3. **line 622**: `rm, ok := rt.MethodByName("MarshalAmino"); ok` → `isAminoMarshaler = true`, repr type extracted from method signature.
4. **line 626**: `rm, ok := reflect.PointerTo(rt).MethodByName("UnmarshalAmino"); ok` (checked on `*T`, not `T`) combined with `!isAminoMarshaler` → panic "Must implement both (o).MarshalAmino and (*o).UnmarshalAmino".
   → **invariant**: AminoMarshaler is all-or-nothing. A type with only `MarshalAmino` is a silent programmer error; with only `UnmarshalAmino`, this check fires.
5. **line 630-633**: `reprType != reprType2` → panic "Must match MarshalAmino and UnmarshalAmino repr types".
6. **line 651-661**: if `isAminoMarshaler`, `info.ReprType = <TypeInfo of reprType>`; else `info.ReprType = info` (self-reference, so `ReprType` is always non-nil).
   → **invariant**: every typ3 computation, `IsStructOrUnpacked` decision, and kind dispatch in the binary codec reads `info.ReprType.Type`, so a Go struct wrapping an `int64` repr behaves at the wire level like `int64`.
7. **line 662-664**: `IsBinaryWellKnownType`/`IsJSONWellKnownType`/`IsJSONAnyValueType` set from `wellknown.go` lookups. `isBinaryWellKnownType` returns true only for `timeType` and `durationType`.
8. **line 665-672**: `rt.Kind() == Array || Slice` → populate `info.Elem` (recursive TypeInfo) and `info.ElemIsPtr`.
   → **invariant**: `info.Elem` being prepopulated is documented precondition for `IsStructOrUnpacked` and for the nil-sentinel logic in list decoders.
9. **line 673-675**: `rt.Kind() == Struct` → `info.StructInfo = parseStructInfoWLocked(rt)`.

### `parseStructInfoWLocked` (`codec.go:682`)

Builds `FieldInfo` list for a Go struct. The crucial fact is that the output
list contains ONE `FieldInfo` per *exported, non-skipped* struct field, with
`BinFieldNum` assigned by position in the output list, not by position in the
Go struct. This is the canonical mapping that the decoder's field-monotonic
check (`decodeReflectBinaryStruct`) relies on.

1. **line 683-688**: `defer recover()` wraps the entire body — any panic in
   `getTypeInfoWLocked` (e.g. unsupported field kind) gets rewrapped as
   "panic parsing struct %v". The original panic value is discarded.
2. **line 689**: `rt.Kind() != reflect.Struct` → panic "should not happen" (defensive).
3. **line 694-699**: `for i := range rt.NumField()`; `!isExported(field) → continue`.
   → **invariant**: unexported fields (first rune not upper, or `PkgPath != ""`) are never assigned a `BinFieldNum` and never participate in encoding. See `isExported` (`reflect.go:199`).
4. **line 700-703**: `skip, fopts := parseFieldOptions(field); if skip { continue }`.
   → **invariant**: `json:"-"` suppresses BOTH JSON and binary encoding. A field tagged this way consumes a Go struct index but contributes no `BinFieldNum`, so subsequent fields shift down.
5. **line 706**: `fopts.BinFieldNum = uint32(len(infos) + 1)`.
   → **invariant**: `BinFieldNum` is assigned by POSITION IN THE EXPORTED, NON-SKIPPED SUBSET, starting at 1. An unexported or `json:"-"` field in the middle of a struct does NOT leave a gap in `BinFieldNum`. The generator MUST emit field numbers using this same rule, or wire-format diverges from the reflect codec.
6. **line 707-710**: `cdc.getTypeInfoWLocked(ftype)` — recursive TypeInfo construction on field type. Pointer field types are dereferenced here (by getTypeInfoWLocked's top-of-function loop at line 514-519), so `fieldTypeInfo.Type` is ALWAYS non-pointer; the original pointer-ness is preserved only in `FieldInfo.Type` (set at line 731) and exposed via `FieldInfo.IsPtr()`.
7. **line 711-728**: `UnpackedList` determination. Table:

   | `frepr.Kind()`   | `frepr.Elem().Kind()` (after Ptr-deref) | typ3 of deref'd elem | `UnpackedList` |
   |------------------|------------------------------------------|----------------------|----------------|
   | `Array`/`Slice`  | `Uint8`                                  | (skipped)            | `false`        |
   | `Array`/`Slice`  | other                                    | `Typ3ByteLength`     | **`true`**     |
   | `Array`/`Slice`  | other                                    | not `Typ3ByteLength` | `false`        |
   | anything else    | N/A                                      | N/A                  | `false`        |

   The element type is dereferenced in a loop (`for etype.Kind() == Ptr { etype = etype.Elem() }`) before `typeToTyp3` is called.
   → **invariant**: `UnpackedList` is true iff elements are themselves struct/interface/list/string (the ByteLength-carried kinds). `[]byte` is NEVER unpacked (routed to byteslice encoder). `[]int32` has `Typ3Varint` elements → packed. `[]time.Duration` has `Typ3ByteLength` repr elements → unpacked, even though Go-kind is `Int64`.
8. **line 730-740**: `FieldInfo{...}` constructed, `ValidateBasic()` called (may panic). `Index: i` preserves the Go struct field index for the decoder to `rv.Field(Index)`.
   → **invariant**: `Index` is the GO struct index; `BinFieldNum` is the PROTOBUF field number. They differ whenever there is an unexported or `json:"-"` field earlier in the struct.

Pointer-field behavior: when `ftype.Kind() == Ptr`, `fieldTypeInfo.Type` is the
dereferenced element type (see step 6). This means `FieldInfo.TypeInfo.Type.Kind()`
for a `*MyStruct` field is `reflect.Struct`, not `reflect.Ptr`. Kind dispatch in
the encoder/decoder uses `finfo.TypeInfo.Type.Kind()` for the body and
`finfo.IsPtr()` (which reads `finfo.Type.Kind()`) for pointer-specific branches.

### `parseFieldOptions` (`codec.go:746`)

Parses three struct tags (`binary`, `amino`, `json`) into a `FieldOptions`.

1. **line 753-756**: `jsonTag == "-"` → `skip = true`, return immediately (no other fields populated).
   → **invariant**: `json:"-"` suppresses binary too; there is no separate `binary:"-"` tag.
2. **line 759-764**: `jsonTagParts[0]` → `fopts.JSONName` (empty string → fall back to `field.Name`).
3. **line 767-771**: `jsonTagParts[1] == "omitempty"` → `fopts.JSONOmitEmpty = true`. Not consulted by the binary codec, but carried on every field.
4. **line 775-780**: `switch binTag`:
   - `"fixed64"` → `fopts.BinFixed64 = true`
   - `"fixed32"` → `fopts.BinFixed32 = true`
   - other/empty → neither set
   → **invariant**: `ValidateBasic` (at line 146) will later panic if the tag is inconsistent with the field's kind. Parse here is unconditional; validation is deferred.
5. **line 783-794**: `switch aminoTag` per comma-separated tag. Multiple amino flags can coexist:
   - `"unsafe"` → `fopts.Unsafe = true` (permits float)
   - `"write_empty"` → `fopts.WriteEmpty = true` (bypass skip-default)
   - `"nil_elements"` → `fopts.NilElements = true` (permit nil struct pointers in lists)
   - unrecognized tokens (including the empty string from a trailing comma) are silently ignored.
   → **invariant**: the three flags are independent bits. A field tagged `amino:"unsafe,write_empty"` sets both.

Note: `UseGoogleTypes` is NOT parsed from struct tags here. It is a
runtime `FieldOptions` field set by top-level callers (e.g.
`MarshalAny`); struct-field options always have `UseGoogleTypes == false`.

### `ValidateBasic` (`codec.go:146`)

Called once per `FieldInfo` at the end of `parseStructInfoWLocked` (line 739).
All panics here surface at codec-init time, not at encode/decode time.

1. **line 147-157**: `finfo.BinFixed32`:
   - `Int32`, `Uint32` (via `GetUltimateElem`) → OK.
   - `Int`, `Uint` → panic "fixed32 not yet supported for int/uint".
   - default → panic "unexpected tag fixed32 for non-32bit type".
2. **line 158-165**: `finfo.BinFixed64`:
   - `Int64`, `Uint64`, `Int`, `Uint` → OK.
   - default → panic "unexpected tag fixed64 for non-64bit type".
   → **asymmetry with BinFixed32**: `fixed64` IS accepted on bare `Int`/`Uint` (encoder writes 8 bytes). This asymmetry is deliberate — `typeToTyp3` has a matching branch at `codec.go:861-867`. See the pre-fix wire corruption noted at the bottom of this document.
3. **line 166-170**: `!finfo.Unsafe && finfo.TypeInfo.Type.Kind() in {Float32, Float64}` → panic "floating point types are unsafe for go-amino".
4. **line 171-174**: `!finfo.Unsafe && finfo.TypeInfo.GetUltimateElem().Type.Kind() in {Float32, Float64}` → panic "floating point types are unsafe for go-amino, even for repr types".
   → **two distinct panic sites**: site 3 catches direct float fields (`F float32`). Site 4 catches floats nested inside slices/arrays/AminoMarshaler repr types (`[]float32`, AminoMarshaler wrapping a `float64` repr). A field that trips site 3 will never reach site 4; the two are non-overlapping.
5. Recursive validation is implicit: when the struct's fields include sub-structs, `getTypeInfoWLocked` triggers `parseStructInfoWLocked` on each sub-struct, which calls `ValidateBasic` on its own fields. Full struct-tree validation completes before the top-level TypeInfo is returned.

### `getTypeInfoFromFullnameRLock` (`codec.go:536`)

Maps a proto fullname to a `*TypeInfo`. Used by the Any decoder.

1. **line 543**: `fullname == "google.protobuf.Timestamp" && !fopts.UseGoogleTypes` → redirect to `timeType` (`time.Time`).
2. **line 548**: `fullname == "google.protobuf.Duration" && !fopts.UseGoogleTypes` → redirect to `durationType` (`time.Duration`).
3. **line 554-558**: else `cdc.fullnameToTypeInfo[fullname]`; miss → error "unrecognized concrete type full name %s".
   → **invariant**: `UseGoogleTypes` is propagated from the top-level caller into every nested decode. With `UseGoogleTypes=true`, Any-wrapped Timestamp/Duration decode into the registered `gTimestampType`/`gDurationType` (protobuf-struct shape); with false (default), they decode into `time.Time` / `time.Duration`.

### `GetTyp3` (`codec.go:76`) / `IsStructOrUnpacked` (`codec.go:85`)

1. **`GetTyp3`**: single expression — `typeToTyp3(info.ReprType.Type, fopts)`.
   → **invariant**: typ3 is ALWAYS computed from the repr type, never the primary Go type. For a Go struct wrapping an `int64` repr, `GetTyp3` returns `Typ3Varint`, not `Typ3ByteLength`. This is why AminoMarshaler types with scalar repr are wire-compatible with the repr scalar.
2. **`IsStructOrUnpacked`** branch table:

   | `rinfo.Type.Kind()` | `rinfo.Elem.GetTyp3(fopt)` | Result |
   |---------------------|-----------------------------|--------|
   | `Struct`            | N/A                         | `true` |
   | `Interface`         | N/A                         | `true` |
   | `Array` or `Slice`  | `Typ3ByteLength`            | `true` (unpacked list) |
   | `Array` or `Slice`  | other                       | `false` (packed list → scalar stream) |
   | other kinds         | N/A                         | `false` |

   → **invariant**: `IsStructOrUnpacked` is the single predicate that decides whether a top-level value needs implicit-struct wrapping (`MarshalReflect`/`UnmarshalReflect` at `amino.go`). An unpacked list of struct pointers is struct-like at the wire level because each element is length-prefixed and field-keyed. A packed list of `int32` is scalar-like.

### Repr-type discovery

At registration (`newTypeInfoUnregisteredWLocked` lines 620-661), repr is
determined from the `MarshalAmino`/`UnmarshalAmino` method signatures via
`marshalAminoReprType`/`unmarshalAminoReprType` (`reflect.go:216`, `:234`):

| Go kind | Repr kind | `IsAminoMarshaler` | `ReprType` | Encode path |
|---------|-----------|--------------------|-----------|--------------|
| Struct  | Struct    | `true`             | TypeInfo of repr struct | `encodeReflectBinaryStruct` via `toReprObject` |
| Struct  | non-struct (e.g. `int64`, `string`, `[]byte`) | `true` | TypeInfo of repr scalar | scalar encoder; `IsStructOrUnpacked` delegates to repr |
| non-struct (e.g. `enum alias`) | non-struct | `true` | TypeInfo of repr scalar | scalar encoder via `toReprObject` |
| anything | (no MarshalAmino) | `false` | `info` (self) | direct encode |

Panics in `marshalAminoReprType`/`unmarshalAminoReprType`:
- `MarshalAmino` with ≠ 1 input or ≠ 2 output → panic.
- `MarshalAmino` second return not `error` → panic.
- Repr type is pointer → panic "Representative objects cannot be pointers".
- `UnmarshalAmino` first input not pointer (receiver must be `*T`) → panic.
- `UnmarshalAmino` returns ≠ 1, or not `error` → panic.
- `MarshalAmino.Out(0) != UnmarshalAmino.In(1)` → panic "Must match MarshalAmino and UnmarshalAmino repr types".

### `HasNativeGenproto2` (`amino.go:59`)

1. **line 60-62**: `rt.Kind() == Ptr` → `rt = rt.Elem()` (normalize before map lookup).
2. **line 65**: `genproto2Types[rt]` — map lookup in global (non-codec-scoped) registry populated by generator `init()` functions.
   → **invariant**: only types with their OWN genproto2 methods appear in the map. A struct that merely EMBEDS a genproto2 type inherits its `MarshalBinary2` via Go method promotion, satisfies the `PBMarshaler2` interface assertion at `amino.go:417`, but would emit the inner type's wire format — which is wrong for the outer type. The `HasNativeGenproto2` check (ANDed with the interface assertion at lines 417, 551, 633, 679, 779, 967, 1205) blocks this.
   → **invariant**: `HasNativePbbindings` (`amino.go:91`) has identical structure for the pbbindings fast path.

---

## Value semantics

Helpers that define Amino's notion of "default value" and "zero value", which
drive skip-default on encode and nil-sentinel interpretation on decode. Also:
the low-level primitive decoders and their overflow/bounds guards.

### `defaultValue` (`reflect.go:116`)

Returns the canonical default `reflect.Value` for a type. The caller uses this
to initialize a struct field or decoded list element when the wire shows
"default" (empty bytes or `0x00` sentinel).

Branch table:

| `rt.Kind()` | Sub-condition | Returned value |
|-------------|---------------|----------------|
| `Ptr`       | `rt.Elem().Kind() == Ptr` | **panic** "nested pointers not allowed" |
| `Ptr`       | `rt.Elem() == timeType` | allocated `*time.Time` set to `emptyTime` (1970-01-01 UTC) |
| `Ptr`       | `rt.Elem().Kind() == Struct` | `reflect.Zero(rt)` → typed nil pointer |
| `Ptr`       | other (e.g. `*int`, `*string`) | `reflect.New(rt.Elem())` → allocated zero non-struct (`new(int)`, `new(string)`, etc.) |
| `Struct`    | `rt == timeType` | value-type `time.Time` set to `emptyTime` |
| `Struct`    | other | `reflect.Zero(rt)` |
| (other)     | any | `reflect.Zero(rt)` |

→ **invariant**: the struct/non-struct asymmetry is a deliberate proto3 wart.
For `*Struct`, nil is distinguishable from empty-but-present in the outer
struct (field key absent vs field key present with `0x00`). For `*int`, the
wire cannot distinguish `nil *int` from `*int` pointing to `0`, so the codec
CANNOT round-trip nil and must allocate on decode.
→ **invariant**: `emptyTime` (Unix epoch) is the default for time values,
not Go's `time.Time{}` (year 0001). `time.Time{}` cannot be represented in
the protobuf Timestamp wire format (seconds value overflows negative).
Generator code emitting time-field decoders must initialize with `emptyTime`
for wire parity.
→ Interaction with `nil_elements`: in `decodeReflectBinaryArray/Slice` nil
sentinel handling, if `fopts.NilElements`, the codec uses `reflect.Zero(ert)`
(true nil for pointer elements); otherwise `defaultValue(ert)` — which for
`*Struct` still yields nil, but for `*time.Time` yields `&emptyTime`. The
two are NOT equivalent for time-pointer elements.

### `isNonstructDefaultValue` (`reflect.go:71`)

The skip-default predicate used by `encodeReflectBinaryStruct` (to decide
whether to emit a field) and by `encodeReflectBinaryList` (to decide whether
to emit `0x00` vs the element's bytes).

1. **line 74-77**: `rv.Type() == durationType` → `false`.
   → **invariant**: `time.Duration` is a Go `int64`-kind type, so the default
   kind switch would treat `Duration(0)` as default and skip. But Duration is
   encoded as a two-field struct (seconds + nanos), and its "default" is a
   zero-struct — which on the wire is an EMPTY length-prefixed struct, not
   an absent field. Returning false here routes Duration through the
   struct/length-prefix path and lets `writeFieldIfNotEmpty`'s single-0x00
   rollback handle actual emptiness.
2. **line 79-103** `switch rv.Kind()`:

   | Kind                         | Default iff           |
   |------------------------------|-----------------------|
   | `Ptr`                        | `rv.IsNil()` → `true`; else recurse on `rv.Elem()` |
   | `Bool`                       | `rv.Bool() == false`  |
   | `Int`/`Int8`/`Int16`/`Int32`/`Int64` | `rv.Int() == 0` |
   | `Uint`/`Uint8`/`Uint16`/`Uint32`/`Uint64` | `rv.Uint() == 0` |
   | `String`                     | `rv.Len() == 0`       |
   | `Chan`/`Map`/`Slice`         | `rv.IsNil() \|\| rv.Len() == 0` |
   | `Func`/`Interface`           | `rv.IsNil()`          |
   | `Struct`                     | `false` (never default) |
   | default (Float32/Float64/Array/Complex...) | `false` |

   → **invariant**: a non-nil `*int` pointing to value `0` IS considered
   default (recursion into the element). So a field of type `*int` with value
   `new(int)` round-trips as absent-field → nil-pointer → reallocated to
   `new(int)` via `defaultValue`. The nil/zero-pointer distinction cannot
   cross the wire.
   → **invariant**: struct values are NEVER skipped via this predicate. A
   struct field that holds `reflect.Zero(StructT)` is still emitted with
   its field key; skip-default for struct fields relies entirely on the
   downstream single-0x00 rollback in `writeFieldIfNotEmpty`. This is why
   the generator needs the `dataLen > 0` gate for struct-repr fields.
   → **invariant**: Float32/Float64 fall into the `default` branch and
   return `false`. A `float32` field with value `0.0` is therefore
   NOT skipped; it emits its field key plus 4 zero bytes. (Floats are
   opt-in via `amino:"unsafe"` anyway.)

### `maybeDerefValue` (`reflect.go:40`)

1. `rv.Kind() == Ptr` → `rvIsPtr = true`.
2. `rv.IsNil()` → `rvIsNilPtr = true`, return early (drv stays zero-value — the caller must check `rvIsNilPtr` before using `drv`).
3. Non-nil pointer → `rv = rv.Elem()`.
4. Return `drv = rv` (the final non-pointer value).

### `maybeDerefAndConstruct` (`reflect.go:54`)

Auto-allocates nil pointers during decode.

1. `rv.Kind() == Ptr` and `rv.IsNil()` → `reflect.New(rv.Type().Elem())` and `rv.Set(newPtr)`.
2. Then `rv = rv.Elem()`.
3. `rv.Kind() == Ptr` (after one deref) → panic "unexpected pointer pointer".
   → **invariant**: a single layer of pointer auto-construct. Nested pointers (`**T`) are not supported anywhere in the codec, and are blocked earlier by `getTypeInfoWLocked` (`codec.go:515`) and `defaultValue` (`reflect.go:122`).

### `DecodeByteSlice` (`decoder.go:332`)

Single guard against unbounded allocation from a crafted length prefix.

1. **line 335**: `DecodeUvarint(bz)` → `count`, advance by `_n`.
2. **line 341**: `count > uint64(len(bz))` → error "insufficient bytes decoding []byte of length %v: have %d".
   → **invariant**: count is compared as UNSIGNED against `len(bz)`. A signed cast to `int` on 32-bit platforms would wrap negative for `count > math.MaxInt32` and silently pass the check; the uint64 comparison prevents the wrap. This is the single gate that prevents a malicious varint from forcing `make([]byte, HUGE)` before checking whether the bytes are actually available.
3. **line 345**: `make([]byte, count)` then `copy(bz2, bz[0:count])`.

### `DecodeVarint` / `DecodeUvarint` (`decoder.go:40`, `:128`)

Thin wrappers around `encoding/binary.Varint`/`Uvarint`. Three outcomes from
the stdlib:

| `n` returned by stdlib | Meaning | Error emitted |
|------------------------|---------|----------------|
| `n > 0`                | OK      | nil |
| `n == 0`               | buffer too small | "buffer too small" |
| `n < 0`                | value overflows 64 bits; `-n` bytes were consumed | "EOF decoding [u]varint" (varint variant says "EOF decoding varint"); returned `n` is negated back to positive |

→ **invariant**: the overflow case returns a non-zero `n` (the number of
bytes that would have been consumed), so callers can still `slide` past the
malformed varint before reporting the error. Buffer-too-small returns
`n == 0` and the caller must not slide.

### `DecodeVarint8` / `DecodeVarint16` / `DecodeUvarint8/16/32` (`decoder.go:14–126`)

Decode a full varint, then bounds-check against the narrow type's range.

1. `DecodeVarint8`: `i64 < math.MinInt8 || i64 > math.MaxInt8` → error "EOF decoding int8".
2. `DecodeVarint16`: `i64 < math.MinInt16 || i64 > math.MaxInt16` → error "EOF decoding int16".
3. `DecodeUvarint8`: `u64 > math.MaxUint8` → error.
4. `DecodeUvarint16`: `u64 > math.MaxUint16` → error.
5. `DecodeUvarint32`: `u64 > math.MaxUint32` → error.

→ **invariant**: bounds-checked decoders ERROR on out-of-range values; they
do NOT silently truncate. A `uint8` field decoded from a varint carrying
`256` returns an error rather than writing `0` to the field. The error
string "EOF decoding intN" is historical and misleading — the condition is
an out-of-range value, not an EOF.

### Fixed-size decoders (`decoder.go:54–209`)

`DecodeInt32`/`DecodeInt64`/`DecodeUint32`/`DecodeUint64`/`DecodeFloat32`/`DecodeFloat64`/`DecodeBool` share a shape:

1. `len(bz) < size` → error "EOF decoding <type>". `size` is 4 for 32-bit types, 8 for 64-bit, 1 for bool.
2. Little-endian read of `size` bytes via `binary.LittleEndian.UintNN`.
3. `DecodeBool` additionally: `bz[0] not in {0, 1}` → error "invalid bool". Any other byte value is rejected rather than coerced to true.
4. `DecodeByte` (`decoder.go:79`): `len(bz) == 0` → error "EOF decoding byte".

### `consumeAny` (`binary_decode.go:1056`)

Skip a single field value given its typ3. Used by `SkipField` (binary_decode2.go:12) via the exported wrapper, and directly by the struct decoder to skip unknown fields.

| `typ3`           | Action                              | Error path |
|------------------|-------------------------------------|------------|
| `Typ3Varint`     | `DecodeVarint(bz)` (discards value, returns `_n`) | propagates |
| `Typ38Byte`      | `DecodeInt64(bz)` (fixed 8 bytes)   | `len(bz) < 8` → "EOF decoding int64" |
| `Typ3ByteLength` | `DecodeByteSlice(bz)` (length-prefixed skip) | varint overflow, insufficient bytes |
| `Typ34Byte`      | `DecodeInt32(bz)` (fixed 4 bytes)   | `len(bz) < 4` → "EOF decoding int32" |
| default          | — | error "invalid typ3 bytes %v" |

1. `err != nil` → return without sliding (caller keeps its position).
2. `err == nil` → `slide(&bz, &n, _n)` — advance `n` by bytes consumed; `bz` advancement is local to this function.

→ **invariant**: `consumeAny` does NOT validate the varint value range, does
NOT validate that `Typ3ByteLength` bytes are valid UTF-8 or valid inner
protobuf. It is a pure "skip this many bytes" operation. Sensible, because
the current codec (`decodeReflectBinaryStruct` line 1039-1048) REJECTS
unknown fields entirely, so `consumeAny` is only reached via generator-
emitted code that explicitly tolerates unknown fields. If the policy
changes to accept unknown fields, the monotonic-fnum check and trailing-
bytes rejection must change in lockstep.

### `SkipField` (`binary_decode2.go:12`)

One-line wrapper: `return consumeAny(typ, bz)`. No additional conditions. Exported for use by generated `UnmarshalBinary2` code that opts to skip unknown fields.

### `decodeFieldNumberAndTyp3` (`binary_decode.go:1082`) — (already in main doc)

Cross-referenced here because the helpers above feed it:

1. `DecodeUvarint(bz)` → `value64`. Errors propagate.
2. `typ = Typ3(value64 & 0x07)` (low 3 bits).
3. `num64 = value64 >> 3`; `num64 == 0` → "invalid field num 0 (reserved)".
4. `num64 > (1<<29 - 1)` → "invalid field num %v".

→ **invariant**: the `& 0x07` mask accepts ALL 8 possible typ3 values on the
wire, including unused/reserved ones (Typ3_3, Typ3_5, Typ3_6, Typ3_7).
Rejection of unused typ3 happens only downstream in `consumeAny` (default
branch) or via the `typ != typWanted` check in struct field decode. So a
reserved typ3 on a known field number gets a "expected field type X for #Y,
got Z" error; on an unknown field number gets "unknown field number N" (the
error comes from the decoder's unknown-field rejection, not from the
field-key parse).

---

## Well-known types

WKTs are special in that their reflect encode/decode paths call hand-written
helpers rather than the generic struct walker. All file paths are relative
to `tm2/pkg/amino/`. Only `time.Time` and `time.Duration` are binary WKTs
(`isBinaryWellKnownType` rejects all others). `emptypb.Empty` and the Google
`*pb.Timestamp`/`*pb.Duration`/`wrapperspb.*` types fall through to the
generic struct path even when tagged as JSON WKTs.

### `isBinaryWellKnownType` (`wellknown.go:176`)

The dispatch predicate that sets `TypeInfo.ConcreteInfo.IsBinaryWellKnownType`.

1. **line 179**: `case timeType, durationType: return true`
   → selects `time.Time` / `time.Duration` for the WKT encode/decode fast path.
   → **invariant**: every other type (including `timestamppb.Timestamp`, `durationpb.Duration`, `emptypb.Empty`, wrapper types) returns `false` and flows through the generic struct walker. Adding a type here without providing matching branches in `encodeReflectBinaryWellKnown` / `decodeReflectBinaryWellKnown` produces an `ok=false` no-op and silently mis-encodes.

Call sites that read the flag:
- `binary_encode.go:66` — `if info.IsBinaryWellKnownType` → call `encodeReflectBinaryWellKnown(...)`; if `ok || err != nil`, short-circuit the rest of `encodeReflectBinary`.
- `binary_decode.go:55` — symmetric on the decode side.

→ **invariant**: the WKT branch runs *before* `IsAminoMarshaler` and before the `Kind()` switch. `time.Time` has kind `Struct`, so without this early exit it would be encoded as a struct with all its unexported fields.

### `encodeReflectBinaryWellKnown` (`wellknown.go:337`)

1. **line 339**: `rv.Kind() == reflect.Interface` → `panic("expected a concrete type to decode to")`.
   → **invariant**: caller (`encodeReflectBinary`) must have resolved the interface value to a concrete before dispatching.
2. **line 343**: `if !bare` → length-prefixed path: recurse with `bare=true` into a pooled buffer, then write `EncodeByteSlice(w, buf.Bytes())`.
   → selects the wire shape when the WKT appears as a struct field value (framed by its parent as `Typ3ByteLength`).
   → **invariant**: symmetric with `decodeReflectBinaryWellKnown`'s `decodeMaybeBare`. Emitting bare bytes where a length prefix is expected corrupts the parent's field layout.
   → **invariant**: the pooled buffer (`poolBytesBuffer.Get`/`Put`) must be returned on every path — the `defer` at line 345 is load-bearing.
3. **line 357**: `switch info.Type`.
   - **line 359** `case timeType`: extract `rv.Interface().(time.Time)`, call `EncodeTime(w, t)`.
   - **line 367** `case durationType`: extract `rv.Interface().(time.Duration)`, call `EncodeDuration(w, d)`.
   - default (fall-through): `return false, nil` (ok=false → caller continues with the generic path).
   → **invariant**: the type assertion must match the reflect type exactly; a named type whose underlying type is `time.Time` will *not* match `timeType` and will fall through.

### `decodeReflectBinaryWellKnown` (`wellknown.go:380`)

1. **line 382**: `rv.Kind() == reflect.Interface` → panic.
2. **line 386**: `bz, err = decodeMaybeBare(bz, &n, bare)` — when `!bare`, consumes a length-prefixed byteslice before dispatching. On error, returns early.
   → **invariant**: the `n` counter counts *header bytes* (the uvarint length) here; the body-length bytes get added as `DecodeTime`/`DecodeDuration` slide. Mismatched accounting will either over-consume or leave tail bytes for the caller's outer struct loop.
3. **line 390**: `switch info.Type`.
   - **line 392** `case timeType`: call `DecodeTime(bz)`; slide `bz`/`n` by `n_`.
   - **line 401** `case durationType`: symmetric with `DecodeDuration(bz)`.
4. **line 411** (default): `return false, 0, nil` → ok=false no-op for non-time/duration types.

### `decodeSecondsAndNanos` (`decoder.go:271`)

Inner state machine parsing field 1 (seconds, varint) and field 2 (nanos, varint) of a `Timestamp`/`Duration` submessage body. Shared by `DecodeTimeValue` (`decoder.go:217`) and `DecodeDurationValue` (`decoder.go:240`).

State: `sawSec`, `sawNs` booleans (both initially false).

Loop condition **line 273**: `for len(bz) > 0` — iterate until bytes exhausted.

1. **line 280**: `case fieldNum == 1 && typ == Typ3Varint` — seconds field.
   - **line 281**: `if sawSec` → error `"duplicate field 1 (seconds)"`.
     → **invariant**: protobuf allows repeated scalar fields on the wire (later wins), but Amino rejects duplicates here as a stricter shape contract.
   - **line 285**: `if sawNs` → error `"seconds (field 1) after nanos (field 2): out of order"`.
     → **invariant**: strict ascending field-number ordering inside Timestamp/Duration. Generator must emit seconds before nanos (the encoder does — `EncodeTimeValue` lines 199/210 — so decoder parity is guaranteed for generated output).
   - On success: decode uvarint into `s`, set `sawSec = true`.
2. **line 300**: `case fieldNum == 2 && typ == Typ3Varint` — nanos field.
   - **line 301**: `if sawNs` → error `"duplicate field 2 (nanos)"`. No `sawSec`-before-nanos check — seconds-missing is legal, defaults to 0.
   - Decode uvarint into a wider `nv := int64(nsec)`.
   - **line 315**: `if 1e9 <= nv || nv <= -1e9` → `InvalidTimeError("nanoseconds not in interval [-999999999, 999999999] %v", nv)`.
     → **invariant**: bounds check is on the raw wire value (pre-assignment to `int32`). Widening to `int64` first is required — assigning directly to `int32` would truncate and smuggle out-of-range values.
   - Set `ns = int32(nv)`, `sawNs = true`.
3. **line 321** (default): `err = "unexpected field in Timestamp/Duration: num=%v typ=%v"`.
   → **invariant**: unknown fields, wrong wire types, reserved field 0 all reject here. No forward-compat field skipping for WKTs.

Post-loop (line 325): `return` with implicitly zero `s` / `ns` if `sawSec` / `sawNs` never flipped.
→ **invariant**: missing seconds or nanos default to zero. Round-tripping `(0,0)` produces zero body bytes.

### `EncodeTimeValue` / `EncodeTime` (`encoder.go:192`, `224`)

`EncodeTime(w, t)` at line 224: `EncodeTimeValue(w, t.Unix(), int32(t.Nanosecond()))`.

1. **line 194**: `err = validateTimeValue(s, ns); if err != nil { return }` — validation returns an `InvalidTimeError` (not a panic).
2. **line 199**: `if s != 0` → emit field 1 header + `EncodeUvarint(w, uint64(s))`. Skipped entirely when `s == 0`.
   → **invariant**: this is the proto3 zero-skip for the seconds subfield. `TimeSize` and genproto2's `SizeBinary2` must mirror this exactly.
3. **line 210**: `if ns != 0` → emit field 2 header + body. Skipped when `ns == 0`.
   → **invariant**: a zero time (Unix epoch) encodes to zero bytes.
4. Negative `s` gets bit-casted to a full-width uvarint (10 bytes on the wire).

### `validateTimeValue` (`encoder.go:228`)

1. **line 229**: `s < minTimeSeconds || s >= maxTimeSeconds` (constants: `minTimeSeconds=-62135596800`, `maxTimeSeconds=253402300800` exclusive) → `InvalidTimeError`.
2. **line 233**: `ns < 0 || ns > maxTimeNanos` (`999999999`, inclusive) → error.

### `EncodeDurationValue` / `EncodeDuration` (`encoder.go:261`, `293`)

`EncodeDuration(w, d)` splits `d.Nanoseconds()` into `(s, ns)` with `sns/1e9` and `int32(sns%1e9)`, calls `validateDurationValue`, then delegates.

1. **line 263**: `err = validateDurationValue(s, ns); if err != nil { return err }`.
2. **line 268**: `if s != 0` → emit field 1.
3. **line 279**: `if ns != 0` → emit field 2.
   → **invariants** mirror `EncodeTimeValue`.

### `validateDurationValue` (`encoder.go:303`)

1. **line 304**: `(s > 0 && ns < 0) || (s < 0 && ns > 0)` → `InvalidDurationError("signs of seconds and nanos do not match")`.
2. **line 308**: `s < minDurationSeconds || s > maxDurationSeconds` (`±315576000000`) → error. Max is *inclusive*.
3. **line 312**: `ns < minDurationNanos || ns > maxDurationNanos` (`±999999999`, both inclusive) → error.

### `validateDurationValueGo` (`encoder.go:328`)

Applied by `DecodeDuration` (not the encoder). Tighter bound than `validateDurationValue` because Go's `time.Duration` is an `int64` nanoseconds.

1. **line 329**: delegate to `validateDurationValue` first.
2. **line 333**: `s < minDurationSecondsGo || s > maxDurationSecondsGo` (`±math.MaxInt64/1e9`) → error.
3. **line 337**: compute `sns := s*1e9 + int64(ns)`; sign-mismatch → error.
   → **invariant**: protects the `time.Duration(s*1e9 + int64(ns))` construction from silent wraparound.

### `DecodeTime` (`decoder.go:226`)

1. **line 228**: `t = emptyTime` (defensive default, `1970-01-01 00:00:00 UTC`).
   → **invariant**: on error, caller sees epoch-1970, not Go's zero time (year 0001). Matches the proto3 "zero Timestamp" convention.
2. **line 229**: `s, ns, n, err := DecodeTimeValue(bz)`. Early return on error.
3. **line 234**: `t = time.Unix(s, int64(ns))`.
4. **line 236**: `t = t.UTC().Truncate(0)` — strip timezone and monotonic clock.
   → **invariant**: required for `reflect.DeepEqual` round-trip parity. Generators must match this normalization.

### `DecodeDuration` (`decoder.go:249`)

1. **line 251**: `s, ns, n, err := DecodeDurationValue(bz)`. Early return on error.
2. **line 256**: `err = validateDurationValueGo(s, ns)` — Go-specific tighter-bounds check. Early return on error.
3. **line 261**: `d = time.Duration(s*1e9 + int64(ns))`.

### `TimeSize` (`binary_encode2.go:17`)

Precomputed body-byte count that must exactly match `EncodeTime` output. Used by genproto2 `SizeBinary2`.

1. **line 18**: `s := t.Unix()`, `ns := int32(t.Nanosecond())`.
2. **line 21**: `if s != 0` → add `UvarintSize(fieldKey(1, Typ3Varint)) + UvarintSize(uint64(s))`.
3. **line 24**: `if ns != 0` → add `UvarintSize(fieldKey(2, Typ3Varint)) + UvarintSize(uint64(ns))`.
4. **Total**: `0` when both `s` and `ns` are zero.
   → **invariant**: subfield zero-skip arithmetic must mirror `EncodeTimeValue` exactly. If `TimeSize` adds a field when the encoder omits it (or vice versa), the generator's `SizeBinary2` disagrees with `MarshalBinary2`.

### `DurationSize` (`binary_encode2.go:32`)

Symmetric to `TimeSize`.

1. **line 33**: `sns := d.Nanoseconds()`, `sec, nsec := sns/1e9, int32(sns%1e9)`.
2. **line 36**: `if sec != 0` → add field-1 key + varint size.
3. **line 39**: `if nsec != 0` → add field-2 key + varint size.
4. **Total**: `0` when both components are zero.

### Empty WKT (structural)

`emptypb.Empty` is *not* a binary WKT — `isBinaryWellKnownType` returns false for `gEmptyType`. Its binary codec behavior is emergent from the generic struct path:

- `Empty{}` has no exported proto fields → generic encoder's field loop emits zero bytes.
- **Marshal**: `MarshalBinary2` of `Empty` emits zero bytes. Wrapped as a field, only the field header+length-prefix is written, length 0.
- **Unmarshal**: accepts zero bytes; any non-zero body bytes error via `binary_decode.go:271-279` (monotonic-field-number check + unknown-field branch).
- **Registration**: `wellknown.go:164` registers `gEmptyType` with `typeURL=/google.protobuf.Empty` for Any-typeURL resolution, not binary dispatch.

→ **invariant**: generators must not emit a special-cased Empty handler; the generic zero-field struct treatment is correct.

### `google.protobuf.Timestamp` / `Duration` Any typeURL redirection (`codec.go:543, 548`)

`getTypeInfoFromFullnameRLock` resolves an `Any` wrapper's `TypeUrl` to a concrete `TypeInfo`.

1. **line 543**: `if fullname == "google.protobuf.Timestamp" && !fopts.UseGoogleTypes` → return `getTypeInfoWLock(timeType)`.
2. **line 548**: `if fullname == "google.protobuf.Duration" && !fopts.UseGoogleTypes` → return `getTypeInfoWLock(durationType)`.

→ **invariant**: by default (`UseGoogleTypes == false`), an Any whose `TypeUrl` declares Google Timestamp/Duration decodes into Go-native `time.Time`/`time.Duration`, not `timestamppb.Timestamp` / `durationpb.Duration`.
→ **invariant**: only Timestamp and Duration get this fullname-redirect. Empty, wrapper types, and Any itself resolve through normal `fullnameToTypeInfo` lookup.

---

## Any envelope helpers

Hand-written Any encode/decode paths in `amino.go`. The reflect-side
equivalents (`encodeReflectBinaryInterface` / `decodeReflectBinaryAny`) are
covered earlier in this document; this section documents the genproto2
fast-path variants and their Prepend/Depth siblings.

Constants referenced: `maxAnyDepth = 64` (`binary_decode.go:15`).

### `MarshalAny` (`amino.go:532`)

Public Any encoder. Non-Binary2, buffer-based. Dispatches to
`marshalAnyBinary2` (genproto2 fast path) or reflect path
`encodeReflectBinaryInterface`.

1. **line 536**: `o == nil` → `errors.New("MarshalAny() requires non-nil argument")`.
   → **invariant**: MarshalAny cannot encode the zero interface. No registered type means no TypeURL.
2. **line 542**: `rv, _, _ = maybeDerefValue(reflect.ValueOf(o))` — pointer peeled before interface/PBMarshaler2 checks.
3. **line 546**: `rv.Kind() == reflect.Interface` → `errors.New("MarshalAny() requires registered concrete type")`.
4. **line 551**: `pbm2, ok := rv.Interface().(PBMarshaler2); ok && HasNativeGenproto2(rt)` → fast path: `marshalAnyBinary2(pbm2)`.
   → **invariant**: `HasNativeGenproto2` guards against promoted methods on embedded structs.
5. **line 572**: reflect fallback: `encodeReflectBinaryInterface(buf, iinfo, ..., bare=true)`.
   → **invariant**: the top-level Any envelope is bare — no outer length prefix is added.

### `marshalAnyBinary2` (`amino.go:583`)

Buffer-based genproto2 Any encoder. Called only from `MarshalAny`. Uses forward encoding.

1. **line 586**: `rv.Kind() == reflect.Ptr` → dereference.
2. **line 593**: `!cinfo.Registered` → `"cannot encode unregistered concrete type %v"`.
3. **line 608**: `encodeFieldNumberAndTyp3(buf, 1, Typ3ByteLength)` — field 1 (TypeURL) always emitted.
4. **line 616**: `len(valueBz) > 0` → elides field 2 when inner value encodes to empty.
   → **audit note**: does NOT handle `len(valueBz) == 1 && valueBz[0] == 0x00` (reflect-side elides this too at `binary_encode.go:302`).

### `MarshalAnyBinary2` (`amino.go:631`)

Prepend-based Any encoder for use inside generated `MarshalBinary2` methods. The production hot path.

1. **line 632-640**: `pbm2, ok := o.(PBMarshaler2); !ok || !HasNativeGenproto2(reflect.TypeOf(o))` → fallback: `cdc.MarshalAny(o)` then `PrependBytes`.
2. **line 643**: `rv.Kind() == reflect.Ptr` → dereference.
3. **line 650**: `!cinfo.Registered` → `"MarshalAnyBinary2: cannot encode unregistered concrete type %v"`.
4. **line 657-662**: `before := offset; offset, err = pbm2.MarshalBinary2(cdc, buf, offset); innerLen := before - offset`.
5. **line 663**: `innerLen > 0` → emit field 2 header.
   → **invariant**: matches `marshalAnyBinary2`'s `len(valueBz) > 0` and `SizeAnyBinary2`'s `innerSize > 0`. Tri-partite coupling: all three must fire on the same types.
   → **audit note**: does NOT handle `innerLen == 1 && buf[offset] == 0x00`. See BINARY_FIXES.md #4.
6. **line 669-670**: field 1 (TypeURL) always prepended.

### `SizeAnyBinary2` (`amino.go:677`)

Arithmetic size computation. Must match `MarshalAnyBinary2` byte count exactly.

1. **line 679**: same `ok && HasNativeGenproto2` gate. Fallback: `len(cdc.MarshalAny(o))`.
2. **line 698**: field 1 (TypeURL) size always added.
3. **line 700-706**: `innerSize > 0` → field 2 header + inner size.
   → **audit note**: arithmetic-only — cannot inspect whether inner output would be `[0x00]`. See BINARY_FIXES.md #21.

### `UnmarshalAny` (`amino.go:1103`)

Public, non-Depth Any decoder.

1. **line 1108**: `rv.Kind() != reflect.Ptr` → `ErrNoPointer`.
2. **line 1114**: `ok, err2 := cdc.unmarshalAnyBinary2(bz, rv); ok` → return fast-path result (even on error).
   → **invariant**: `ok` is the commitment flag, not the success flag. A parse error after byte-level dispatch must not fall through to reflect.
3. **line 1131**: reflect fallback `decodeReflectBinaryInterface(bz, iinfo, rv, FieldOptions{}, true, 0)` — `bare=true`, `anyDepth=0`.

### `unmarshalAnyBinary2` (`amino.go:1141`)

Internal genproto2 Any decoder called by `UnmarshalAny`. Returns `(committed, err)`. Entry `anyDepth` implicitly 1.

1. **line 1143**: `len(bz) == 0` → returns `(false, nil)` (not committed).
2. **line 1149-1154**: `fnum != 1 || typ != Typ3ByteLength` → committed error.
3. **line 1157-1161**: `DecodeString(bz)` — TypeURL.
   → **audit note**: **no** `IsASCIIText(typeURL)` check here. Sibling gap #23.
4. **line 1165**: `len(bz) > 0` → decode field 2 only if bytes remain.
5. **line 1170**: `fnum != 2 || typ != Typ3ByteLength` → committed error.
6. **line 1183**: `len(bz) > 0` after field 2 → committed trailing-bytes error.
7. **line 1204-1206**: `!ok || !HasNativeGenproto2(cinfo.Type)` → returns `(false, nil)` (fall through to reflect).
8. **line 1210**: `!irvSet.Type().AssignableTo(rv.Type())` → committed error.
9. **line 1215**: `len(value) > 0` → only recurse into `pbm2.UnmarshalBinary2(cdc, value, 1)` when non-empty.
10. **line 1216**: `anyDepth=1` hardcoded at entry.
    → **audit note**: no `maxAnyDepth` guard at entry. BINARY_FIXES.md #24.

### `UnmarshalAnyBinary2` + `unmarshalAnyBinary2Depth` (`amino.go:717`, `:721`)

Depth-aware variant. `UnmarshalAnyBinary2` is a one-line wrapper that increments `anyDepth`; `unmarshalAnyBinary2Depth` is the body. Always commits (returns plain `error`).

1. **line 722**: `anyDepth > maxAnyDepth` (= 64) → `"exceeded max Any nesting depth %d"`.
   → **invariant**: the sole increment site for binary-path depth tracking.
2. **line 725**: `len(bz) == 0` → returns nil (no-op). Differs from sibling: no fallback.
3. **line 730-736**: field 1 header checks — same contract as sibling.
4. **line 738-742**: TypeURL decoded. Same audit note: no `IsASCIIText` check (gap #14).
5. **line 746-761**: field 2 present iff `len(bz) > 0`; same gates.
6. **line 763-765**: trailing-bytes rejection.
7. **line 778-783**: `!ok || !HasNativeGenproto2(cinfo.Type)` → delegates to `cdc.unmarshalAny2Depth(typeURL, value, ptr, anyDepth)`.
   → **invariant**: `anyDepth` is **propagated**, not reset. Before this fix, switching from genproto2 to reflect mid-recursion zeroed the counter.
8. **line 791**: `!irvSet.Type().AssignableTo(rv.Type())` → committed error.
9. **line 796**: `len(value) > 0` → recurse `pbm2.UnmarshalBinary2(cdc, value, anyDepth)` — depth passed without further increment.
   → **invariant**: single increment per Any layer.

### `UnmarshalAny2` + `unmarshalAny2Depth` (`amino.go:1226`, `:1234`)

Destructured variant: takes `(typeURL, value)` directly.

1. **line 1238**: `typeURL == ""` → returns nil.
   → **invariant**: empty TypeURL encodes a nil interface in proto3. Must match encoder's `IsEmpty` treatment.
2. **line 1243**: `rv.Kind() != reflect.Ptr` → `ErrNoPointer`.
3. **line 1247**: delegates to `decodeReflectBinaryAny(typeURL, value, rv, FieldOptions{}, anyDepth)` — this path performs `IsASCIIText` validation (unlike envelope-parsing entries).

### Fallback / `consumeAny` interaction

When struct decode encounters an unknown field number, `consumeAny(typ3, bz)` skips the body. For an Any-typed unknown field this consumes the full ByteLength payload (`DecodeByteSlice`) without attempting registry resolution.

→ **invariant**: unknown *fields* are silently skipped (forward-compat); unknown *typeURLs* inside an expected Any *error* (strict-Any semantics). Asymmetry is deliberate.

---

## Framing helpers

Length-prefix framing. All four variants wrap a `Marshal`/`Unmarshal` call
with a uvarint byte-length prefix. Empty inner (`nil` or zero-length) still
emits a one-byte `0x00` length prefix.

### `MarshalSized` (`amino.go:300`)

1. **line 308**: inner `cdc.Marshal(o)` error propagated.
2. **line 314**: `EncodeUvarint(buf, uint64(len(bz)))` — always writes length prefix even when `len(bz) == 0`.
   → **invariant**: a zero-length framed blob is **one byte** (`0x00`), not zero bytes. `UnmarshalSized` line 828 rejects zero-byte input — so `MarshalSized(empty) → [0x00]` round-trips.
3. **line 325**: `copyBytes(buf.Bytes())` before returning pooled buffer.
   → **invariant**: returned slice must not alias the pool.

### `MarshalSizedWriter` (`amino.go:330`)

1. **line 335**: `cdc.MarshalSized(o)` error propagated.
2. **line 339**: `w.Write(bz)` — partial-write returns `(n, err)`.
   → **audit note**: no retry loop on short writes.

### `MarshalAnySized` / `MarshalAnySizedWriter` (`amino.go:353`, `:388`)

Structurally identical to the non-Any pair, but inner call is `cdc.MarshalAny(o)`.

### `UnmarshalSized` (`amino.go:827`)

1. **line 828**: `len(bz) == 0` → `"unmarshalSized cannot decode empty bytes"`.
   → **invariant**: input must at least contain the length byte.
2. **line 834**: `n < 0` → `"Error reading msg byte-length prefix: got code %v"`.
3. **line 837**: `u64 > uint64(len(bz) - n)` → truncation error.
4. **line 840**: `u64 < uint64(len(bz) - n)` → trailing-bytes error.
   → **invariant**: exact-length match is mandatory.
5. **line 847**: delegates to `cdc.Unmarshal(bz, ptr)`.

### `UnmarshalSizedReader` (`amino.go:853`)

Streaming variant. Two distinct overflow guards.

1. **line 856**: `maxSize < 0` → `panic`.
2. **line 863-878**: byte-at-a-time uvarint read. `buf[i]&0x80 == 0` is terminator.
3. **line 872**: `n >= maxSize` during prefix read → sets err.
   → **audit note**: the error is assigned but loop does not `break`. Next iteration overwrites. Guard is load-bearing only if reader terminates naturally within `maxSize`.
4. **line 883-894**: `maxSize > 0` branch — *two* separate checks:
   - **line 884**: `uint64(maxSize) < u64` — payload alone exceeds maxSize.
   - **line 888**: `(maxSize - n) < int64(u64)` — payload + prefix exceeds maxSize.
   → **invariant**: either check alone is insufficient; second accounts for prefix bytes.
5. **line 904-905**: `io.ReadFull(r, bz)` → errors on truncated stream.

`maxSize == 0` means "unlimited" (documented). Bypasses *both* overflow checks.

### Shared helpers

- `EncodeUvarint` / `binary.Uvarint`: forward uvarint emission / parsing. Amino does not use `DecodeUvarintWithMax`; Max check is inlined in `UnmarshalSizedReader`.
- `copyBytes` (`amino.go:806`): defensive copy out of pooled buffer.
- `poolBytesBuffer`: shared `sync.Pool` of `*bytes.Buffer`.
