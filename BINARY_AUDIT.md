# Binary Codec Audit Checklist

Walk unchecked items (`[ ]`) in order. For each, audit the generator, set
verdict ([P] PRESENT / [M] MISSING / [D] DIFFERENT / [N] N/A), and if
missing/different, append a numbered finding to BINARY_FIXES.md and
record the finding number here.

Remaining count: `grep -c '\[ \]' BINARY_AUDIT.md`

## Legend

- `MUST` — wire-format parity required; generator must emit equivalent code.
- `INIT` — codec-init-time logic; generator codegen-time decisions must match.
- `PANIC` — contract-violation panic; no generator counterpart required.
- `REFLECT` — reflect-internal guard; generator uses typed accessors, N/A.
- `META` — documented invariant/normalization, not a conditional.

---

## Dispatch layer

Items: 37. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 37.

- [x] dispatch.1 (ref amino.go:417) — MUST — Marshal dispatch to MarshalBinary2 via PBMarshaler2 + HasNativeGenproto2 — verdict: [P] — fix: n/a
- [x] dispatch.2 (ref amino.go:422) — MUST — Marshal pbbindings fallback gated by usePBBindings — verdict: [N] — fix: n/a
- [x] dispatch.3 (ref amino.go:428) — MUST — Marshal implicit fallback to MarshalReflect — verdict: [N] — fix: n/a
- [ ] dispatch.4 (ref amino.go:439) — PANIC — MarshalReflect pointer deref nil/nested-pointer panic — verdict: [N] — fix: n/a
- [x] dispatch.5 (ref amino.go:464) — MUST — MarshalReflect implicit-struct wrap when !IsStructOrUnpacked — verdict: [P] — fix: n/a
- [x] dispatch.6 (ref amino.go:488) — MUST — MarshalReflect normalize empty bytes to nil at top level — verdict: [N] — fix: n/a
- [x] dispatch.7 (ref amino.go:958) — MUST — Unmarshal ergonomic inner-pointer auto-alloc — verdict: [N] — fix: n/a
- [x] dispatch.8 (ref amino.go:967) — MUST — Unmarshal PBMessager2 fast path via HasNativeGenproto2 — verdict: [N] — fix: n/a
- [x] dispatch.9 (ref amino.go:972) — MUST — Unmarshal pbbindings fallback gated by usePBBindings — verdict: [N] — fix: n/a
- [x] dispatch.10 (ref amino.go:982) — MUST — Unmarshal fallback to UnmarshalReflect — verdict: [N] — fix: n/a
- [x] dispatch.11 (ref amino.go:988) — MUST — UnmarshalReflect require Kind()==Ptr, else ErrNoPointer — verdict: [N] — fix: n/a
- [x] dispatch.12 (ref amino.go:1002) — MUST — UnmarshalReflect accept empty input as zero for non-struct non-interface — verdict: [P] — fix: n/a
- [x] dispatch.13 (ref amino.go:1015) — MUST — UnmarshalReflect unwrap implicit-struct field 1 for non-struct — verdict: [P] — fix: n/a
- [x] dispatch.14 (ref amino.go:1027) — MUST — UnmarshalReflect reject fnum != 1 on implicit struct — verdict: [P] — fix: n/a
- [x] dispatch.15 (ref amino.go:1030) — MUST — UnmarshalReflect reject typ != typWanted on implicit struct — verdict: [P] — fix: n/a
- [ ] dispatch.16 (ref amino.go:1054) — MUST — UnmarshalReflect reject trailing bytes n != len(bz) — verdict: [M] — fix: #3
- [x] dispatch.17 (ref codec.go:832) — MUST — typeToTyp3 timeType → Typ3ByteLength — verdict: [P] — fix: n/a
- [x] dispatch.18 (ref codec.go:832) — MUST — typeToTyp3 durationType → Typ3ByteLength — verdict: [P] — fix: n/a
- [x] dispatch.19 (ref codec.go:832) — MUST — typeToTyp3 Interface/Array/Slice/String/Struct/Map → Typ3ByteLength — verdict: [P] — fix: n/a
- [x] dispatch.20 (ref codec.go:832) — MUST — typeToTyp3 Int64/Uint64 + BinFixed64 → Typ38Byte — verdict: [P] — fix: n/a
- [x] dispatch.21 (ref codec.go:832) — MUST — typeToTyp3 Int64/Uint64 else → Typ3Varint — verdict: [P] — fix: n/a
- [x] dispatch.22 (ref codec.go:832) — MUST — typeToTyp3 Int32/Uint32 + BinFixed32 → Typ34Byte — verdict: [P] — fix: n/a
- [x] dispatch.23 (ref codec.go:832) — MUST — typeToTyp3 Int32/Uint32 else → Typ3Varint — verdict: [P] — fix: n/a
- [x] dispatch.24 (ref codec.go:832) — MUST — typeToTyp3 Int/Uint + BinFixed64 → Typ38Byte — verdict: [P] — fix: n/a
- [x] dispatch.25 (ref codec.go:832) — MUST — typeToTyp3 Int/Uint else → Typ3Varint — verdict: [P] — fix: n/a
- [x] dispatch.26 (ref codec.go:832) — MUST — typeToTyp3 Int16/Int8/Uint16/Uint8/Bool → Typ3Varint — verdict: [P] — fix: n/a
- [x] dispatch.27 (ref codec.go:832) — MUST — typeToTyp3 Float64 → Typ38Byte — verdict: [P] — fix: n/a
- [x] dispatch.28 (ref codec.go:832) — MUST — typeToTyp3 Float32 → Typ34Byte — verdict: [P] — fix: n/a
- [ ] dispatch.29 (ref codec.go:832) — PANIC — typeToTyp3 default panic on unknown kind — verdict: [N] — fix: n/a
- [x] dispatch.30 (ref codec.go:146) — INIT — ValidateBasic BinFixed32 allowed only on Int32/Uint32 — verdict: [P] — fix: n/a
- [x] dispatch.31 (ref codec.go:146) — INIT — ValidateBasic BinFixed64 allowed on Int64/Uint64/Int/Uint — verdict: [P] — fix: n/a
- [x] dispatch.32 (ref codec.go:146) — INIT — ValidateBasic !Unsafe+Float anywhere → panic — verdict: [P] — fix: n/a
- [x] dispatch.33 (ref codec.go:85) — INIT — IsStructOrUnpacked Struct/Interface → true — verdict: [P] — fix: n/a
- [x] dispatch.34 (ref codec.go:85) — INIT — IsStructOrUnpacked Array/Slice + Elem typ3 ByteLength → true — verdict: [P] — fix: n/a
- [x] dispatch.35 (ref codec.go:85) — INIT — IsStructOrUnpacked else → false — verdict: [P] — fix: n/a
- [x] dispatch.36 (ref codec.go:76) — INIT — GetTyp3 always via ReprType.Type — verdict: [P] — fix: n/a
- [ ] dispatch.37 (ref codec.go:832) — META — typeToTyp3/encoder-body parity invariant — verdict: [N] — fix: n/a

---

## Binary encode

Items: 51. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 51.

- [ ] encode.1 (ref binary_encode.go:49) — PANIC — encodeReflectBinary rv.Kind()==Ptr panic — verdict: [N] — fix: n/a
- [ ] encode.2 (ref binary_encode.go:54) — REFLECT — encodeReflectBinary !rv.IsValid() panic — verdict: [N] — fix: n/a
- [x] encode.3 (ref binary_encode.go:66) — MUST — encodeReflectBinary IsBinaryWellKnownType → WellKnown dispatch — verdict: [P] — fix: n/a
- [x] encode.4 (ref binary_encode.go:75) — MUST — encodeReflectBinary IsAminoMarshaler → toReprObject + recurse — verdict: [P] — fix: n/a
- [x] encode.5 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Interface → encodeReflectBinaryInterface — verdict: [P] — fix: n/a
- [x] encode.6 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Array + Uint8 elem → encodeReflectBinaryByteArray — verdict: [P] — fix: n/a
- [x] encode.7 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Array else → encodeReflectBinaryList — verdict: [P] — fix: n/a
- [x] encode.8 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Slice + Uint8 elem → encodeReflectBinaryByteSlice — verdict: [P] — fix: n/a
- [x] encode.9 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Slice else → encodeReflectBinaryList — verdict: [P] — fix: n/a
- [x] encode.10 (ref binary_encode.go:88) — MUST — encodeReflectBinary kind=Struct → encodeReflectBinaryStruct — verdict: [P] — fix: n/a
- [x] encode.11 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int64 + BinFixed64 → EncodeInt64 — verdict: [P] — fix: n/a
- [x] encode.12 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int64 else → EncodeVarint — verdict: [P] — fix: n/a
- [x] encode.13 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int32 + BinFixed32 → EncodeInt32 — verdict: [P] — fix: n/a
- [x] encode.14 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int32 else → EncodeVarint — verdict: [P] — fix: n/a
- [x] encode.15 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int16/Int8 → EncodeVarint — verdict: [P] — fix: n/a
- [x] encode.16 (ref binary_encode.go:88) — MUST — encodeReflectBinary Int + BinFixed64/32/else → EncodeInt64/32/Varint — verdict: [P] — fix: n/a
- [x] encode.17 (ref binary_encode.go:88) — MUST — encodeReflectBinary Uint64 + BinFixed64 → EncodeUint64 else EncodeUvarint — verdict: [P] — fix: n/a
- [x] encode.18 (ref binary_encode.go:88) — MUST — encodeReflectBinary Uint32 + BinFixed32 → EncodeUint32 else EncodeUvarint — verdict: [P] — fix: n/a
- [x] encode.19 (ref binary_encode.go:88) — MUST — encodeReflectBinary Uint16 → EncodeUvarint — verdict: [P] — fix: n/a
- [ ] encode.20 (ref binary_encode.go:114) — MUST — encodeReflectBinary Uint8 + beOptionByte → EncodeByte else EncodeUvarint — verdict: [M] — fix: #6
- [x] encode.21 (ref binary_encode.go:88) — MUST — encodeReflectBinary Uint + BinFixed64/32/else → EncodeUint64/32/Uvarint — verdict: [P] — fix: n/a
- [x] encode.22 (ref binary_encode.go:88) — MUST — encodeReflectBinary Bool → EncodeBool — verdict: [P] — fix: n/a
- [ ] encode.23 (ref binary_encode.go:117) — MUST — encodeReflectBinary Float64 requires fopts.Unsafe → EncodeFloat64 — verdict: [M] — fix: #5
- [ ] encode.24 (ref binary_encode.go:118) — MUST — encodeReflectBinary Float32 requires fopts.Unsafe → EncodeFloat32 — verdict: [M] — fix: #5
- [x] encode.25 (ref binary_encode.go:215) — MUST — encodeReflectBinaryInterface rv.IsNil() → writeMaybeBare nil — verdict: [P] — fix: n/a
- [ ] encode.26 (ref binary_encode.go:215) — PANIC — encodeReflectBinaryInterface interface-of-interface panic — verdict: [N] — fix: n/a
- [ ] encode.27 (ref binary_encode.go:215) — PANIC — encodeReflectBinaryInterface nil concrete pointer panic — verdict: [N] — fix: n/a
- [ ] encode.28 (ref binary_encode.go:215) — PANIC — encodeReflectBinaryInterface concrete type must be registered panic — verdict: [N] — fix: n/a
- [x] encode.29 (ref binary_encode.go:215) — MUST — encodeReflectBinaryInterface !IsStructOrUnpacked → implicit struct wrap Any.Value — verdict: [P] — fix: n/a
- [x] encode.30 (ref binary_encode.go:302) — MUST — encodeReflectBinaryInterface elide Any.Value when empty/single-0x00 — verdict: [M] — fix: #4
- [ ] encode.31 (ref binary_encode.go:321) — PANIC — encodeReflectBinaryByteArray ert.Kind()!=Uint8 panic — verdict: [N] — fix: n/a
- [ ] encode.32 (ref binary_encode.go:321) — REFLECT — encodeReflectBinaryByteArray CanAddr fast path vs reflect.Copy — verdict: [N] — fix: n/a
- [ ] encode.33 (ref binary_encode.go:480) — PANIC — encodeReflectBinaryByteSlice ert.Kind()!=Uint8 panic — verdict: [N] — fix: n/a
- [ ] encode.34 (ref binary_encode.go:346) — PANIC — encodeReflectBinaryList ert.Kind()==Uint8 panic (wrong route) — verdict: [N] — fix: n/a
- [x] encode.35 (ref binary_encode.go:346) — MUST — encodeReflectBinaryList packed-vs-unpacked dispatch on typ3/beOptionByte — verdict: [P] — fix: n/a
- [x] encode.36 (ref binary_encode.go:346) — MUST — encodeReflectBinaryList packed pointer-element zero-substitution — verdict: [P] — fix: n/a
- [ ] encode.37 (ref binary_encode.go:346) — MUST — encodeReflectBinaryList unpacked ertIsPointer/ertIsStruct/writeImplicit flags — verdict: [D] — fix: #8
- [x] encode.38 (ref binary_encode.go:346) — MUST — encodeReflectBinaryList unpacked field-key per element — verdict: [P] — fix: n/a
- [ ] encode.39 (ref binary_encode.go:460) — MUST — encodeReflectBinaryList unpacked isNonstructDefaultValue sentinel vs nil_elements error — verdict: [D] — fix: #7
- [x] encode.40 (ref binary_encode.go:346) — MUST — encodeReflectBinaryList unpacked non-default deref/implicit-wrap/length-prefix — verdict: [P] — fix: n/a
- [x] encode.41 (ref binary_encode.go:500) — MUST — encodeReflectBinaryStruct skip field when !WriteEmpty && isNonstructDefaultValue — verdict: [P] — fix: n/a
- [x] encode.42 (ref binary_encode.go:500) — MUST — encodeReflectBinaryStruct UnpackedList routes through encodeReflectBinaryList(bare=true) — verdict: [P] — fix: n/a
- [x] encode.43 (ref binary_encode.go:500) — MUST — encodeReflectBinaryStruct else writeFieldIfNotEmpty with writeEmpty||ptr — verdict: [P] — fix: n/a
- [x] encode.44 (ref binary_encode.go:568) — MUST — writeFieldIfNotEmpty write field key — verdict: [P] — fix: n/a
- [x] encode.45 (ref binary_encode.go:568) — MUST — writeFieldIfNotEmpty write field value — verdict: [P] — fix: n/a
- [ ] encode.46 (ref binary_encode.go:592) — MUST — writeFieldIfNotEmpty single-0x00 rollback contract — verdict: [D] — fix: #13
- [x] encode.47 (ref binary_encode.go:602) — MUST — writeMaybeBare len(bz)==0 && bare → emit nothing — verdict: [P] — fix: n/a
- [x] encode.48 (ref binary_encode.go:602) — MUST — writeMaybeBare len(bz)==0 && !bare → emit 0x00 marker — verdict: [P] — fix: n/a
- [x] encode.49 (ref binary_encode.go:602) — MUST — writeMaybeBare len(bz)>0 && bare → emit raw — verdict: [P] — fix: n/a
- [x] encode.50 (ref binary_encode.go:602) — MUST — writeMaybeBare len(bz)>0 && !bare → emit length-prefix+bytes — verdict: [P] — fix: n/a
- [ ] encode.51 (ref binary_encode2.go:1) — META — binary_encode2.go constants/wrappers (no conditionals) — verdict: [N] — fix: n/a

---

## Binary decode

Items: 59. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 59.

- [ ] decode.1 (ref binary_decode.go:36) — REFLECT — decodeReflectBinary !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.2 (ref binary_decode.go:39) — PANIC — decodeReflectBinary Interface must not be wrapped as Ptr — verdict: [N] — fix: n/a
- [x] decode.3 (ref binary_decode.go:55) — MUST — decodeReflectBinary IsBinaryWellKnownType → WellKnown dispatch — verdict: [P] — fix: n/a
- [x] decode.4 (ref binary_decode.go:64) — MUST — decodeReflectBinary IsAminoMarshaler → repr decode + UnmarshalAmino — verdict: [P] — fix: n/a
- [x] decode.5 (ref binary_decode.go:86) — MUST — decodeReflectBinary kind=Interface → decodeReflectBinaryInterface — verdict: [P] — fix: n/a
- [x] decode.6 (ref binary_decode.go:86) — MUST — decodeReflectBinary Array+Uint8 → decodeReflectBinaryByteArray — verdict: [P] — fix: n/a
- [x] decode.7 (ref binary_decode.go:86) — MUST — decodeReflectBinary Array else → decodeReflectBinaryArray — verdict: [P] — fix: n/a
- [x] decode.8 (ref binary_decode.go:86) — MUST — decodeReflectBinary Slice+Uint8 → decodeReflectBinaryByteSlice — verdict: [P] — fix: n/a
- [x] decode.9 (ref binary_decode.go:86) — MUST — decodeReflectBinary Slice else → decodeReflectBinarySlice — verdict: [P] — fix: n/a
- [x] decode.10 (ref binary_decode.go:86) — MUST — decodeReflectBinary kind=Struct → decodeReflectBinaryStruct — verdict: [P] — fix: n/a
- [ ] decode.11 (ref binary_decode.go:86) — MUST — decodeReflectBinary Int64 + BinFixed64 → DecodeInt64 else DecodeVarint — verdict: [M] — fix: #2
- [ ] decode.12 (ref binary_decode.go:86) — MUST — decodeReflectBinary Int32 + BinFixed32 → DecodeInt32 else DecodeVarint — verdict: [M] — fix: #2
- [x] decode.13 (ref binary_decode.go:86) — MUST — decodeReflectBinary Int16/Int8 → DecodeVarint16/8 — verdict: [P] — fix: n/a
- [ ] decode.14 (ref binary_decode.go:86) — MUST — decodeReflectBinary Int + BinFixed64 → DecodeInt64 else DecodeVarint — verdict: [M] — fix: #2
- [ ] decode.15 (ref binary_decode.go:86) — MUST — decodeReflectBinary Uint64 + BinFixed64 → DecodeUint64 else DecodeUvarint — verdict: [M] — fix: #2
- [ ] decode.16 (ref binary_decode.go:86) — MUST — decodeReflectBinary Uint32 + BinFixed32 → DecodeUint32 else DecodeUvarint — verdict: [M] — fix: #2
- [x] decode.17 (ref binary_decode.go:86) — MUST — decodeReflectBinary Uint16 → DecodeUvarint16 — verdict: [P] — fix: n/a
- [x] decode.18 (ref binary_decode.go:86) — MUST — decodeReflectBinary Uint8 + bdOptionByte → DecodeByte else DecodeUvarint8 — verdict: [P] — fix: n/a
- [ ] decode.19 (ref binary_decode.go:86) — MUST — decodeReflectBinary Uint + BinFixed64 → DecodeUint64 else DecodeUvarint — verdict: [M] — fix: #2
- [x] decode.20 (ref binary_decode.go:86) — MUST — decodeReflectBinary Bool → DecodeBool — verdict: [P] — fix: n/a
- [x] decode.21 (ref binary_decode.go:86) — MUST — decodeReflectBinary Float64 requires Unsafe → DecodeFloat64 — verdict: [D] — fix: #5
- [x] decode.22 (ref binary_decode.go:86) — MUST — decodeReflectBinary Float32 requires Unsafe → DecodeFloat32 — verdict: [D] — fix: #5
- [x] decode.23 (ref binary_decode.go:86) — MUST — decodeReflectBinary String → DecodeString — verdict: [P] — fix: n/a
- [ ] decode.24 (ref binary_decode.go:86) — PANIC — decodeReflectBinary default panic on unknown kind — verdict: [N] — fix: n/a
- [x] decode.25 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface anyDepth > maxAnyDepth error — verdict: [P] — fix: n/a
- [ ] decode.26 (ref binary_decode.go:319) — REFLECT — decodeReflectBinaryInterface !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [x] decode.27 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface !rv.IsNil() error — verdict: [N] — fix: n/a
- [x] decode.28 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface empty bz → zero interface — verdict: [D] — fix: #29
- [x] decode.29 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface field 1 header checks — verdict: [P] — fix: n/a
- [x] decode.30 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface post-field1 lenbz==0 → zero concrete — verdict: [P] — fix: n/a
- [x] decode.31 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface field 2 header checks — verdict: [P] — fix: n/a
- [x] decode.32 (ref binary_decode.go:319) — MUST — decodeReflectBinaryInterface trailing bytes error — verdict: [P] — fix: n/a
- [ ] decode.33 (ref binary_decode.go:420) — MUST — decodeReflectBinaryAny IsASCIIText(typeURL) check — verdict: [M] — fix: #14
- [x] decode.34 (ref binary_decode.go:418) — MUST — decodeReflectBinaryAny len(value)==0 → zero concrete — verdict: [P] — fix: n/a
- [ ] decode.35 (ref binary_decode.go:441) — MUST — decodeReflectBinaryAny AssignableTo pre-check — verdict: [D] — fix: #15
- [x] decode.36 (ref binary_decode.go:418) — MUST — decodeReflectBinaryAny !IsStructOrUnpacked → unwrap implicit field 1 — verdict: [P] — fix: n/a
- [x] decode.37 (ref binary_decode.go:418) — MUST — decodeReflectBinaryAny field 1 fnum/typ checks — verdict: [P] — fix: n/a
- [x] decode.38 (ref binary_decode.go:418) — MUST — decodeReflectBinaryAny trailing-bytes error after decode — verdict: [P] — fix: n/a
- [ ] decode.39 (ref binary_decode.go:514) — MUST — decodeReflectBinaryAny AssignableTo post-decode — verdict: [D] — fix: #15
- [ ] decode.40 (ref binary_decode.go:525) — REFLECT — decodeReflectBinaryByteArray !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.41 (ref binary_decode.go:525) — PANIC — decodeReflectBinaryByteArray ert.Kind()!=Uint8 panic — verdict: [N] — fix: n/a
- [x] decode.42 (ref binary_decode.go:525) — MUST — decodeReflectBinaryByteArray len(bz)<length error — verdict: [P] — fix: n/a
- [ ] decode.43 (ref binary_decode.go:551) — MUST — decodeReflectBinaryByteArray len(byteslice)!=length error — verdict: [M] — fix: #9
- [ ] decode.44 (ref binary_decode.go:564) — REFLECT — decodeReflectBinaryArray !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.45 (ref binary_decode.go:564) — PANIC — decodeReflectBinaryArray ert.Kind()==Uint8 panic — verdict: [N] — fix: n/a
- [x] decode.46 (ref binary_decode.go:564) — MUST — decodeReflectBinaryArray packed-vs-unpacked loop dispatch — verdict: [P] — fix: n/a
- [x] decode.47 (ref binary_decode.go:564) — MUST — decodeReflectBinaryArray packed post-loop trailing error — verdict: [D] — fix: #10
- [ ] decode.48 (ref binary_decode.go:564) — MUST — decodeReflectBinaryArray unpacked pre-flags isErtStructPointer/writeImplicit — verdict: [D] — fix: #8
- [x] decode.49 (ref binary_decode.go:564) — MUST — decodeReflectBinaryArray unpacked element fnum/typ checks — verdict: [D] — fix: #10
- [x] decode.50 (ref binary_decode.go:651) — MUST — decodeReflectBinaryArray nil-sentinel 0x00 with NilElements/defaultValue branching — verdict: [P] — fix: n/a
- [x] decode.51 (ref binary_decode.go:564) — MUST — decodeReflectBinaryArray writeImplicit inner struct decode — verdict: [P] — fix: n/a
- [ ] decode.52 (ref binary_decode.go:625) — MUST — decodeReflectBinaryArray fnum regression / short-input error — verdict: [M] — fix: #10
- [ ] decode.53 (ref binary_decode.go:736) — REFLECT — decodeReflectBinaryByteSlice !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.54 (ref binary_decode.go:736) — PANIC — decodeReflectBinaryByteSlice ert.Kind()!=Uint8 panic — verdict: [N] — fix: n/a
- [x] decode.55 (ref binary_decode.go:736) — MUST — decodeReflectBinaryByteSlice len(bz)==0 → nil slice — verdict: [P] — fix: n/a
- [x] decode.56 (ref binary_decode.go:736) — MUST — decodeReflectBinaryByteSlice normalize empty to nil — verdict: [P] — fix: n/a
- [ ] decode.57 (ref binary_decode.go:779) — REFLECT — decodeReflectBinarySlice !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.58 (ref binary_decode.go:779) — PANIC — decodeReflectBinarySlice ert.Kind()==Uint8 panic — verdict: [N] — fix: n/a
- [x] decode.59 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice packed-vs-unpacked dispatch same as array — verdict: [P] — fix: n/a

---

## Binary decode (slice/struct/tail)

Items: 21. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 21.

- [x] decode.60 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice packed `for len(bz)!=0` append loop — verdict: [P] — fix: n/a
- [x] decode.61 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice unpacked break on fnum > BinFieldNum — verdict: [P] — fix: n/a
- [x] decode.62 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice unpacked fnum<BinFieldNum error — verdict: [P] — fix: n/a
- [x] decode.63 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice unpacked typ != Typ3ByteLength error — verdict: [P] — fix: n/a
- [x] decode.64 (ref binary_decode.go:868) — MUST — decodeReflectBinarySlice nil-sentinel same rule as array — verdict: [P] — fix: n/a
- [ ] decode.65 (ref binary_decode.go:779) — MUST — decodeReflectBinarySlice writeImplicit handling same as array — verdict: [D] — fix: #8
- [ ] decode.66 (ref binary_decode.go:939) — REFLECT — decodeReflectBinaryStruct !rv.CanAddr() panic — verdict: [N] — fix: n/a
- [ ] decode.67 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct len(bz)==0 → remaining fields zero — verdict: [M] — fix: #11
- [x] decode.68 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct UnpackedList: BinFieldNum<fnum skip — verdict: [P] — fix: n/a
- [x] decode.69 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct UnpackedList else decode bare=true — verdict: [P] — fix: n/a
- [x] decode.70 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct packed BinFieldNum<fnum skip — verdict: [P] — fix: n/a
- [ ] decode.71 (ref binary_decode.go:1009) — MUST — decodeReflectBinaryStruct fnum<=lastFieldNum non-monotonic error — verdict: [D] — fix: #1
- [x] decode.72 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct packed BinFieldNum!=fnum unknown-field error — verdict: [P] — fix: n/a
- [x] decode.73 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct packed typ != typWanted error — verdict: [P] — fix: n/a
- [ ] decode.74 (ref binary_decode.go:971) — MUST — decodeReflectBinaryStruct absent-field reset to defaultValue — verdict: [M] — fix: #11
- [x] decode.75 (ref binary_decode.go:939) — MUST — decodeReflectBinaryStruct trailing bytes after loop error — verdict: [P] — fix: n/a
- [x] decode.76 (ref binary_decode.go:1056) — MUST — consumeAny typ3 switch dispatch — verdict: [N] — fix: n/a
- [x] decode.77 (ref binary_decode.go:1082) — MUST — decodeFieldNumberAndTyp3 num64==0 reserved error — verdict: [P] — fix: n/a
- [x] decode.78 (ref binary_decode.go:1082) — MUST — decodeFieldNumberAndTyp3 num64 > (1<<29-1) overflow error — verdict: [P] — fix: n/a
- [x] decode.79 (ref binary_decode.go:1109) — MUST — decodeMaybeBare bare/!bare length-prefix branching — verdict: [N] — fix: n/a
- [ ] decode.80 (ref binary_decode2.go:1) — META — binary_decode2.go wrappers (no new conditionals) — verdict: [N] — fix: n/a

---

## Cross-reference to genproto2

Items: 1. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 1, [ ] 0.

- [x] crossref.1 (ref codec.go:76) — META — reflect-codec↔generator section-mapping table — verdict: [N] — fix: ___

---

## TypeInfo construction

Items: 35. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 35.

- [x] typeinfo.1 (ref codec.go:594) — INIT — newTypeInfoUnregistered Ptr/Map/Func kind panic — verdict: [N] — fix: n/a
- [x] typeinfo.2 (ref codec.go:602) — INIT — newTypeInfoUnregistered double existence-check for typeInfos[rt] — verdict: [N] — fix: n/a
- [x] typeinfo.3 (ref codec.go:622) — INIT — newTypeInfoUnregistered MarshalAmino method detection → isAminoMarshaler — verdict: [P] — fix: n/a
- [x] typeinfo.4 (ref codec.go:626) — INIT — newTypeInfoUnregistered UnmarshalAmino on *T without MarshalAmino → panic — verdict: [P] — fix: n/a
- [x] typeinfo.5 (ref codec.go:630) — INIT — newTypeInfoUnregistered repr-type mismatch panic — verdict: [P] — fix: n/a
- [x] typeinfo.6 (ref codec.go:651) — INIT — newTypeInfoUnregistered ReprType set to repr or self — verdict: [P] — fix: n/a
- [x] typeinfo.7 (ref codec.go:662) — INIT — newTypeInfoUnregistered IsBinary/JSON-WellKnownType flags from wellknown.go — verdict: [P] — fix: n/a
- [x] typeinfo.8 (ref codec.go:665) — INIT — newTypeInfoUnregistered Array/Slice → populate info.Elem + ElemIsPtr — verdict: [P] — fix: n/a
- [x] typeinfo.9 (ref codec.go:673) — INIT — newTypeInfoUnregistered Struct → StructInfo via parseStructInfoWLocked — verdict: [P] — fix: n/a
- [x] typeinfo.10 (ref codec.go:683) — INIT — parseStructInfoWLocked defer recover → "panic parsing struct" rewrap — verdict: [P] — fix: n/a
- [x] typeinfo.11 (ref codec.go:689) — INIT — parseStructInfoWLocked rt.Kind()!=Struct defensive panic — verdict: [P] — fix: n/a
- [x] typeinfo.12 (ref codec.go:694) — INIT — parseStructInfoWLocked skip unexported fields — verdict: [P] — fix: n/a
- [x] typeinfo.13 (ref codec.go:700) — INIT — parseStructInfoWLocked parseFieldOptions skip flag — verdict: [P] — fix: n/a
- [x] typeinfo.14 (ref codec.go:706) — INIT — parseStructInfoWLocked BinFieldNum = position among exported — verdict: [P] — fix: n/a
- [x] typeinfo.15 (ref codec.go:707) — INIT — parseStructInfoWLocked recursive getTypeInfoWLocked on field type — verdict: [P] — fix: n/a
- [x] typeinfo.16 (ref codec.go:711) — INIT — parseStructInfoWLocked UnpackedList determination table — verdict: [P] — fix: n/a
- [x] typeinfo.17 (ref codec.go:730) — INIT — parseStructInfoWLocked FieldInfo construction + ValidateBasic — verdict: [P] — fix: n/a
- [x] typeinfo.18 (ref codec.go:753) — INIT — parseFieldOptions jsonTag=="-" skip — verdict: [P] — fix: n/a
- [x] typeinfo.19 (ref codec.go:759) — INIT — parseFieldOptions JSONName + fallback to field.Name — verdict: [P] — fix: n/a
- [x] typeinfo.20 (ref codec.go:767) — INIT — parseFieldOptions JSONOmitEmpty — verdict: [P] — fix: n/a
- [x] typeinfo.21 (ref codec.go:775) — INIT — parseFieldOptions binTag fixed64/fixed32 switch — verdict: [P] — fix: n/a
- [x] typeinfo.22 (ref codec.go:783) — INIT — parseFieldOptions aminoTag unsafe/write_empty/nil_elements — verdict: [P] — fix: n/a
- [x] typeinfo.23 (ref codec.go:147) — INIT — ValidateBasic BinFixed32: OK on Int32/Uint32; Int/Uint panic; else panic — verdict: [P] — fix: n/a
- [x] typeinfo.24 (ref codec.go:158) — INIT — ValidateBasic BinFixed64: OK on 64-bit + Int/Uint; else panic — verdict: [P] — fix: n/a
- [x] typeinfo.25 (ref codec.go:166) — INIT — ValidateBasic direct Float32/64 without Unsafe panic — verdict: [P] — fix: n/a
- [x] typeinfo.26 (ref codec.go:171) — INIT — ValidateBasic nested Float32/64 repr without Unsafe panic — verdict: [P] — fix: n/a
- [x] typeinfo.27 (ref codec.go:146) — INIT — ValidateBasic recursive validation via sub-struct types — verdict: [P] — fix: n/a
- [x] typeinfo.28 (ref codec.go:543) — INIT — getTypeInfoFromFullnameRLock Timestamp redirect to timeType unless UseGoogleTypes — verdict: [P] — fix: n/a
- [x] typeinfo.29 (ref codec.go:548) — INIT — getTypeInfoFromFullnameRLock Duration redirect to durationType unless UseGoogleTypes — verdict: [P] — fix: n/a
- [x] typeinfo.30 (ref codec.go:554) — INIT — getTypeInfoFromFullnameRLock fullnameToTypeInfo miss → error — verdict: [N] — fix: n/a
- [x] typeinfo.31 (ref codec.go:76) — INIT — GetTyp3 expression via ReprType — verdict: [P] — fix: n/a
- [ ] typeinfo.32 (ref codec.go:85) — INIT — IsStructOrUnpacked branch table (struct/interface/list/other) — verdict: [D] — fix: #18
- [x] typeinfo.33 (ref reflect.go:216) — INIT — marshalAminoReprType signature panics (arity/return/pointer) — verdict: [N] — fix: n/a
- [x] typeinfo.34 (ref amino.go:60) — INIT — HasNativeGenproto2 Ptr normalize before map lookup — verdict: [P] — fix: n/a
- [x] typeinfo.35 (ref amino.go:65) — INIT — HasNativeGenproto2 map lookup in genproto2Types — verdict: [P] — fix: n/a

---

## Value semantics

Items: 31. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 31.

- [x] value.1 (ref reflect.go:116) — MUST — defaultValue Ptr→Ptr nested pointer panic — verdict: [N] — fix: n/a
- [x] value.2 (ref reflect.go:116) — MUST — defaultValue *time.Time → &emptyTime — verdict: [P] — fix: n/a
- [x] value.3 (ref reflect.go:116) — MUST — defaultValue *Struct → typed nil — verdict: [P] — fix: n/a
- [x] value.4 (ref reflect.go:116) — MUST — defaultValue *other → reflect.New (allocated zero) — verdict: [P] — fix: n/a
- [x] value.5 (ref reflect.go:116) — MUST — defaultValue Struct == timeType → emptyTime — verdict: [P] — fix: n/a
- [x] value.6 (ref reflect.go:116) — MUST — defaultValue other struct → reflect.Zero — verdict: [P] — fix: n/a
- [x] value.7 (ref reflect.go:116) — MUST — defaultValue other kinds → reflect.Zero — verdict: [P] — fix: n/a
- [x] value.8 (ref reflect.go:74) — MUST — isNonstructDefaultValue durationType short-circuit false — verdict: [P] — fix: n/a
- [x] value.9 (ref reflect.go:79) — MUST — isNonstructDefaultValue Ptr nil OR recurse on Elem — verdict: [P] — fix: n/a
- [x] value.10 (ref reflect.go:79) — MUST — isNonstructDefaultValue Bool == false — verdict: [P] — fix: n/a
- [x] value.11 (ref reflect.go:79) — MUST — isNonstructDefaultValue Int kinds == 0 — verdict: [P] — fix: n/a
- [x] value.12 (ref reflect.go:79) — MUST — isNonstructDefaultValue Uint kinds == 0 — verdict: [P] — fix: n/a
- [x] value.13 (ref reflect.go:79) — MUST — isNonstructDefaultValue String rv.Len()==0 — verdict: [P] — fix: n/a
- [x] value.14 (ref reflect.go:79) — MUST — isNonstructDefaultValue Chan/Map/Slice IsNil||Len==0 — verdict: [P] — fix: n/a
- [x] value.15 (ref reflect.go:79) — MUST — isNonstructDefaultValue Func/Interface IsNil — verdict: [P] — fix: n/a
- [x] value.16 (ref reflect.go:79) — MUST — isNonstructDefaultValue Struct → false — verdict: [P] — fix: n/a
- [x] value.17 (ref reflect.go:79) — MUST — isNonstructDefaultValue Float/Array/Complex default → false — verdict: [D] — fix: #30
- [ ] value.18 (ref reflect.go:40) — REFLECT — maybeDerefValue Ptr/nil/elem sequence — verdict: [N] — fix: n/a
- [x] value.19 (ref reflect.go:54) — MUST — maybeDerefAndConstruct auto-alloc nil pointer — verdict: [P] — fix: n/a
- [ ] value.20 (ref reflect.go:54) — PANIC — maybeDerefAndConstruct nested pointer panic — verdict: [N] — fix: n/a
- [x] value.21 (ref decoder.go:335) — MUST — DecodeByteSlice uvarint length read — verdict: [P] — fix: n/a
- [x] value.22 (ref decoder.go:341) — MUST — DecodeByteSlice count > len(bz) error — verdict: [P] — fix: n/a
- [x] value.23 (ref decoder.go:345) — MUST — DecodeByteSlice allocate+copy — verdict: [P] — fix: n/a
- [x] value.24 (ref decoder.go:40) — MUST — DecodeVarint/Uvarint stdlib outcome propagation — verdict: [P] — fix: n/a
- [x] value.25 (ref decoder.go:14) — MUST — DecodeVarint8 bounds error on out-of-range — verdict: [P] — fix: n/a
- [x] value.26 (ref decoder.go:14) — MUST — DecodeVarint16 bounds error — verdict: [P] — fix: n/a
- [x] value.27 (ref decoder.go:14) — MUST — DecodeUvarint8/16/32 bounds error — verdict: [P] — fix: n/a
- [x] value.28 (ref decoder.go:54) — MUST — Fixed-size decoders len(bz)<size error — verdict: [P] — fix: n/a
- [x] value.29 (ref decoder.go:54) — MUST — Fixed-size decoders little-endian read — verdict: [P] — fix: n/a
- [x] value.30 (ref decoder.go:54) — MUST — DecodeBool bz[0] not in {0,1} error — verdict: [P] — fix: n/a
- [x] value.31 (ref decoder.go:79) — MUST — DecodeByte len(bz)==0 error — verdict: [N] — fix: n/a

---

## Value semantics (consumeAny/SkipField/field-key)

Items: 10. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 10.

- [x] value.32 (ref binary_decode.go:1056) — MUST — consumeAny Typ3Varint → DecodeVarint — verdict: [N] — fix: n/a
- [x] value.33 (ref binary_decode.go:1056) — MUST — consumeAny Typ38Byte → DecodeInt64 (8 bytes) — verdict: [N] — fix: n/a
- [x] value.34 (ref binary_decode.go:1056) — MUST — consumeAny Typ3ByteLength → DecodeByteSlice — verdict: [N] — fix: n/a
- [x] value.35 (ref binary_decode.go:1056) — MUST — consumeAny Typ34Byte → DecodeInt32 (4 bytes) — verdict: [N] — fix: n/a
- [x] value.36 (ref binary_decode.go:1056) — MUST — consumeAny default invalid typ3 error — verdict: [N] — fix: n/a
- [x] value.37 (ref binary_decode.go:1056) — MUST — consumeAny err propagation without sliding — verdict: [N] — fix: n/a
- [x] value.38 (ref binary_decode.go:1056) — MUST — consumeAny slide on success — verdict: [N] — fix: n/a
- [x] value.39 (ref binary_decode2.go:12) — MUST — SkipField wrapper delegates to consumeAny — verdict: [N] — fix: n/a
- [ ] value.40 (ref binary_decode.go:1082) — META — decodeFieldNumberAndTyp3 typ = value64 & 0x07 accepts all 8 — verdict: [N] — fix: n/a
- [x] value.41 (ref binary_decode.go:1082) — MUST — decodeFieldNumberAndTyp3 num64 bounds (both guards) — verdict: [P] — fix: n/a

---

## Well-known types

Items: 38. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 38.

- [x] wkt.1 (ref wellknown.go:179) — INIT — isBinaryWellKnownType time/duration only — verdict: [P] — fix: n/a
- [ ] wkt.2 (ref wellknown.go:337) — PANIC — encodeReflectBinaryWellKnown Interface kind panic — verdict: [N] — fix: n/a
- [x] wkt.3 (ref wellknown.go:343) — MUST — encodeReflectBinaryWellKnown !bare → length-prefix with pooled buf — verdict: [P] — fix: n/a
- [x] wkt.4 (ref wellknown.go:357) — MUST — encodeReflectBinaryWellKnown switch info.Type — verdict: [P] — fix: n/a
- [x] wkt.5 (ref wellknown.go:359) — MUST — encodeReflectBinaryWellKnown timeType → EncodeTime — verdict: [P] — fix: n/a
- [x] wkt.6 (ref wellknown.go:367) — MUST — encodeReflectBinaryWellKnown durationType → EncodeDuration — verdict: [P] — fix: n/a
- [x] wkt.7 (ref wellknown.go:357) — MUST — encodeReflectBinaryWellKnown default ok=false fall-through — verdict: [P] — fix: n/a
- [ ] wkt.8 (ref wellknown.go:382) — PANIC — decodeReflectBinaryWellKnown Interface kind panic — verdict: [N] — fix: n/a
- [x] wkt.9 (ref wellknown.go:386) — MUST — decodeReflectBinaryWellKnown decodeMaybeBare when !bare — verdict: [P] — fix: n/a
- [x] wkt.10 (ref wellknown.go:390) — MUST — decodeReflectBinaryWellKnown switch info.Type — verdict: [P] — fix: n/a
- [x] wkt.11 (ref wellknown.go:392) — MUST — decodeReflectBinaryWellKnown timeType → DecodeTime — verdict: [P] — fix: n/a
- [x] wkt.12 (ref wellknown.go:401) — MUST — decodeReflectBinaryWellKnown durationType → DecodeDuration — verdict: [P] — fix: n/a
- [x] wkt.13 (ref wellknown.go:411) — MUST — decodeReflectBinaryWellKnown default ok=false no-op — verdict: [P] — fix: n/a
- [x] wkt.14 (ref decoder.go:273) — MUST — decodeSecondsAndNanos loop condition `for len(bz)>0` — verdict: [P] — fix: n/a
- [x] wkt.15 (ref decoder.go:281) — MUST — decodeSecondsAndNanos duplicate field 1 error — verdict: [P] — fix: n/a
- [x] wkt.16 (ref decoder.go:285) — MUST — decodeSecondsAndNanos seconds-after-nanos out-of-order error — verdict: [P] — fix: n/a
- [x] wkt.17 (ref decoder.go:280) — MUST — decodeSecondsAndNanos seconds decode into s — verdict: [P] — fix: n/a
- [x] wkt.18 (ref decoder.go:301) — MUST — decodeSecondsAndNanos duplicate field 2 error — verdict: [P] — fix: n/a
- [x] wkt.19 (ref decoder.go:315) — MUST — decodeSecondsAndNanos nanos range |nv|<1e9 — verdict: [P] — fix: n/a
- [x] wkt.20 (ref decoder.go:300) — MUST — decodeSecondsAndNanos widening nsec→int64 then to int32 — verdict: [P] — fix: n/a
- [x] wkt.21 (ref decoder.go:321) — MUST — decodeSecondsAndNanos default unknown-field error — verdict: [P] — fix: n/a
- [ ] wkt.22 (ref decoder.go:325) — META — decodeSecondsAndNanos post-loop defaults zero — verdict: [N] — fix: n/a
- [x] wkt.23 (ref encoder.go:194) — MUST — EncodeTimeValue validateTimeValue error propagation — verdict: [P] — fix: n/a
- [x] wkt.24 (ref encoder.go:199) — MUST — EncodeTimeValue s != 0 → emit field 1 — verdict: [P] — fix: n/a
- [x] wkt.25 (ref encoder.go:210) — MUST — EncodeTimeValue ns != 0 → emit field 2 — verdict: [P] — fix: n/a
- [x] wkt.26 (ref encoder.go:229) — MUST — validateTimeValue s range — verdict: [P] — fix: n/a
- [x] wkt.27 (ref encoder.go:233) — MUST — validateTimeValue ns range — verdict: [P] — fix: n/a
- [x] wkt.28 (ref encoder.go:263) — MUST — EncodeDurationValue validateDurationValue error propagation — verdict: [P] — fix: n/a
- [x] wkt.29 (ref encoder.go:268) — MUST — EncodeDurationValue s != 0 → emit field 1 — verdict: [P] — fix: n/a
- [x] wkt.30 (ref encoder.go:279) — MUST — EncodeDurationValue ns != 0 → emit field 2 — verdict: [P] — fix: n/a
- [x] wkt.31 (ref encoder.go:304) — MUST — validateDurationValue sign mismatch s/ns — verdict: [P] — fix: n/a
- [x] wkt.32 (ref encoder.go:308) — MUST — validateDurationValue s range ±315576000000 — verdict: [P] — fix: n/a
- [x] wkt.33 (ref encoder.go:312) — MUST — validateDurationValue ns range ±999999999 — verdict: [P] — fix: n/a
- [x] wkt.34 (ref encoder.go:329) — MUST — validateDurationValueGo s range Go-int64 tighter — verdict: [P] — fix: n/a
- [x] wkt.35 (ref encoder.go:337) — MUST — validateDurationValueGo sns overflow sign-mismatch — verdict: [P] — fix: n/a
- [ ] wkt.36 (ref decoder.go:228) — META — DecodeTime defensive t = emptyTime — verdict: [N] — fix: n/a
- [ ] wkt.37 (ref decoder.go:236) — META — DecodeTime t = t.UTC().Truncate(0) normalization — verdict: [N] — fix: n/a
- [x] wkt.38 (ref decoder.go:249) — MUST — DecodeDuration validateDurationValueGo + compute d — verdict: [P] — fix: n/a

---

## Well-known types (sizes / Empty / Any-WKT redirect)

Items: 11. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 11.

- [x] wkt.39 (ref binary_encode2.go:18) — MUST — TimeSize t.Unix()/t.Nanosecond() — verdict: [P] — fix: n/a
- [x] wkt.40 (ref binary_encode2.go:21) — MUST — TimeSize s != 0 add field-1 key+varint size — verdict: [P] — fix: n/a
- [x] wkt.41 (ref binary_encode2.go:24) — MUST — TimeSize ns != 0 add field-2 key+varint size — verdict: [P] — fix: n/a
- [ ] wkt.42 (ref binary_encode2.go:17) — META — TimeSize total 0 when both zero — verdict: [N] — fix: n/a
- [x] wkt.43 (ref binary_encode2.go:33) — MUST — DurationSize split sns into sec/nsec — verdict: [P] — fix: n/a
- [x] wkt.44 (ref binary_encode2.go:36) — MUST — DurationSize sec != 0 add field-1 size — verdict: [P] — fix: n/a
- [x] wkt.45 (ref binary_encode2.go:39) — MUST — DurationSize nsec != 0 add field-2 size — verdict: [P] — fix: n/a
- [ ] wkt.46 (ref binary_encode2.go:32) — META — DurationSize total 0 when both zero — verdict: [N] — fix: n/a
- [ ] wkt.47 (ref wellknown.go:164) — META — Empty WKT emergent from generic struct path (no special case) — verdict: [N] — fix: n/a
- [x] wkt.48 (ref codec.go:543) — INIT — Any fullname Timestamp redirect unless UseGoogleTypes — verdict: [P] — fix: n/a
- [x] wkt.49 (ref codec.go:548) — INIT — Any fullname Duration redirect unless UseGoogleTypes — verdict: [P] — fix: n/a

---

## Any envelope helpers

Items: 35. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 35.

- [x] any.1 (ref amino.go:536) — MUST — MarshalAny nil-arg error — verdict: [N] — fix: n/a
- [ ] any.2 (ref amino.go:542) — REFLECT — MarshalAny maybeDerefValue pointer peel — verdict: [N] — fix: n/a
- [x] any.3 (ref amino.go:546) — MUST — MarshalAny Interface kind error — verdict: [N] — fix: n/a
- [x] any.4 (ref amino.go:551) — MUST — MarshalAny PBMarshaler2+HasNativeGenproto2 fast path — verdict: [P] — fix: n/a
- [x] any.5 (ref amino.go:572) — MUST — MarshalAny reflect fallback with bare=true — verdict: [N] — fix: n/a
- [x] any.6 (ref amino.go:586) — MUST — marshalAnyBinary2 pointer dereference — verdict: [N] — fix: n/a
- [x] any.7 (ref amino.go:593) — MUST — marshalAnyBinary2 !cinfo.Registered error — verdict: [N] — fix: n/a
- [ ] any.8 (ref amino.go:608) — META — marshalAnyBinary2 field 1 TypeURL always emitted — verdict: [N] — fix: n/a
- [ ] any.9 (ref amino.go:616) — MUST — marshalAnyBinary2 len(valueBz) > 0 field-2 emission — verdict: [M] — fix: #22
- [x] any.10 (ref amino.go:632) — MUST — MarshalAnyBinary2 PBMarshaler2/HasNativeGenproto2 fallback via MarshalAny+PrependBytes — verdict: [P] — fix: n/a
- [x] any.11 (ref amino.go:643) — MUST — MarshalAnyBinary2 pointer dereference — verdict: [P] — fix: n/a
- [x] any.12 (ref amino.go:650) — MUST — MarshalAnyBinary2 !cinfo.Registered error — verdict: [P] — fix: n/a
- [ ] any.13 (ref amino.go:657) — MUST — MarshalAnyBinary2 inner marshal + innerLen computation — verdict: [D] — fix: #25
- [ ] any.14 (ref amino.go:663) — MUST — MarshalAnyBinary2 innerLen > 0 field-2 emission — verdict: [M] — fix: #4
- [ ] any.15 (ref amino.go:669) — META — MarshalAnyBinary2 field 1 TypeURL always prepended — verdict: [N] — fix: n/a
- [x] any.16 (ref amino.go:679) — MUST — SizeAnyBinary2 PBMarshaler2 fallback via len(MarshalAny) — verdict: [P] — fix: n/a
- [ ] any.17 (ref amino.go:698) — META — SizeAnyBinary2 field 1 size always added — verdict: [N] — fix: n/a
- [ ] any.18 (ref amino.go:700) — MUST — SizeAnyBinary2 innerSize > 0 field-2 size — verdict: [M] — fix: #21
- [x] any.19 (ref amino.go:1108) — MUST — UnmarshalAny Kind()!=Ptr ErrNoPointer — verdict: [P] — fix: n/a
- [x] any.20 (ref amino.go:1114) — MUST — UnmarshalAny unmarshalAnyBinary2 committed return — verdict: [P] — fix: n/a
- [x] any.21 (ref amino.go:1131) — MUST — UnmarshalAny reflect fallback decodeReflectBinaryInterface — verdict: [N] — fix: n/a
- [x] any.22 (ref amino.go:1143) — MUST — unmarshalAnyBinary2 len(bz)==0 → (false, nil) — verdict: [N] — fix: n/a
- [ ] any.23 (ref amino.go:1143) — MUST — unmarshalAnyBinary2 anyDepth > maxAnyDepth guard — verdict: [M] — fix: #24
- [x] any.24 (ref amino.go:1149) — MUST — unmarshalAnyBinary2 field-1 fnum/typ committed error — verdict: [N] — fix: n/a
- [ ] any.25 (ref amino.go:1157) — MUST — unmarshalAnyBinary2 IsASCIIText(typeURL) check — verdict: [M] — fix: #23
- [x] any.26 (ref amino.go:1165) — MUST — unmarshalAnyBinary2 field-2 decoded iff len(bz)>0 — verdict: [N] — fix: n/a
- [x] any.27 (ref amino.go:1170) — MUST — unmarshalAnyBinary2 field-2 fnum/typ committed error — verdict: [N] — fix: n/a
- [x] any.28 (ref amino.go:1183) — MUST — unmarshalAnyBinary2 trailing bytes committed error — verdict: [N] — fix: n/a
- [x] any.29 (ref amino.go:1204) — MUST — unmarshalAnyBinary2 !ok||!HasNativeGenproto2 fall-through to reflect — verdict: [N] — fix: n/a
- [x] any.30 (ref amino.go:1210) — MUST — unmarshalAnyBinary2 AssignableTo check — verdict: [N] — fix: n/a
- [x] any.31 (ref amino.go:1215) — MUST — unmarshalAnyBinary2 recurse only when len(value)>0 — verdict: [N] — fix: n/a
- [x] any.32 (ref amino.go:722) — MUST — unmarshalAnyBinary2Depth anyDepth>maxAnyDepth error — verdict: [P] — fix: n/a
- [ ] any.33 (ref amino.go:738) — MUST — unmarshalAnyBinary2Depth IsASCIIText(typeURL) check — verdict: [M] — fix: #14
- [x] any.34 (ref amino.go:778) — MUST — unmarshalAnyBinary2Depth !ok||!HasNativeGenproto2 delegate with anyDepth propagated — verdict: [P] — fix: n/a
- [ ] any.35 (ref amino.go:791) — MUST — unmarshalAnyBinary2Depth AssignableTo check (single) — verdict: [D] — fix: #15

---

## Any envelope helpers (Any2 + consumeAny interaction)

Items: 4. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 4.

- [ ] any.36 (ref amino.go:796) — META — unmarshalAnyBinary2Depth single increment per Any layer — verdict: [N] — fix: n/a
- [x] any.37 (ref amino.go:1238) — MUST — UnmarshalAny2/unmarshalAny2Depth empty typeURL nil interface — verdict: [P] — fix: n/a
- [x] any.38 (ref amino.go:1243) — MUST — UnmarshalAny2 Kind()!=Ptr ErrNoPointer — verdict: [N] — fix: n/a
- [x] any.39 (ref amino.go:1247) — MUST — UnmarshalAny2 delegates to decodeReflectBinaryAny (IsASCIIText path) — verdict: [N] — fix: n/a

---

## Framing helpers

Items: 16. Verdict distribution: [P] 0, [M] 0, [D] 0, [N] 0, [ ] 16.

- [x] framing.1 (ref amino.go:308) — MUST — MarshalSized inner Marshal error propagated — verdict: [N] — fix: n/a
- [x] framing.2 (ref amino.go:314) — MUST — MarshalSized EncodeUvarint always writes length prefix (even 0) — verdict: [N] — fix: n/a
- [x] framing.3 (ref amino.go:325) — MUST — MarshalSized copyBytes before returning pool buf — verdict: [N] — fix: n/a
- [x] framing.4 (ref amino.go:335) — MUST — MarshalSizedWriter inner MarshalSized error propagated — verdict: [N] — fix: n/a
- [x] framing.5 (ref amino.go:339) — MUST — MarshalSizedWriter w.Write(bz) partial write — verdict: [N] — fix: n/a
- [x] framing.6 (ref amino.go:353) — MUST — MarshalAnySized structurally identical with MarshalAny inner — verdict: [N] — fix: n/a
- [x] framing.7 (ref amino.go:388) — MUST — MarshalAnySizedWriter structurally identical — verdict: [N] — fix: n/a
- [x] framing.8 (ref amino.go:828) — MUST — UnmarshalSized len(bz)==0 error — verdict: [N] — fix: n/a
- [x] framing.9 (ref amino.go:834) — MUST — UnmarshalSized n<0 varint-header error — verdict: [N] — fix: n/a
- [x] framing.10 (ref amino.go:837) — MUST — UnmarshalSized u64 > remaining truncation error — verdict: [N] — fix: n/a
- [x] framing.11 (ref amino.go:840) — MUST — UnmarshalSized u64 < remaining trailing-bytes error — verdict: [N] — fix: n/a
- [x] framing.12 (ref amino.go:847) — MUST — UnmarshalSized delegates to Unmarshal — verdict: [N] — fix: n/a
- [ ] framing.13 (ref amino.go:856) — PANIC — UnmarshalSizedReader maxSize<0 panic — verdict: [N] — fix: n/a
- [x] framing.14 (ref amino.go:872) — MUST — UnmarshalSizedReader n>=maxSize during prefix read — verdict: [N] — fix: n/a
- [x] framing.15 (ref amino.go:884) — MUST — UnmarshalSizedReader maxSize<payload and maxSize-n<payload checks — verdict: [N] — fix: n/a
- [x] framing.16 (ref amino.go:904) — MUST — UnmarshalSizedReader io.ReadFull truncated-stream error — verdict: [N] — fix: n/a

---
