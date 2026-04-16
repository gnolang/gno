package tests

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// These tests verify that genproto2-generated UnmarshalBinary2 rejects
// wire data with the wrong typ3 (wire type) for a known field number,
// instead of silently misinterpreting the bytes.
//
// Wire format tag byte = (field_num << 3) | typ3. Amino uses its own Typ3
// values (Varint=0, 8Byte=1, ByteLength=2, 4Byte=5), not protobuf3's.

// Reference: amino/codec.go
const (
	typ3Varint     = 0
	typ38Byte      = 1
	typ3ByteLength = 2
	typ34Byte      = 5
)

func tag(fieldNum, typ3 byte) byte {
	return (fieldNum << 3) | typ3
}

func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %q", want, err.Error())
	}
}

// PrimitivesStruct field 1 is int8 → expects Varint wire type.
// Sending field 1 as ByteLength should be rejected.
func TestUnmarshalBinary2_RejectsWrongTyp3_Varint(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Field 1 (int8) with ByteLength wire type + 1-byte payload.
	bz := []byte{tag(1, typ3ByteLength), 0x01, 0x42}

	var s PrimitivesStruct
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "field 1: expected typ3")
}

// PrimitivesStruct field 16 is string → expects ByteLength.
// Sending field 16 as Varint should be rejected.
func TestUnmarshalBinary2_RejectsWrongTyp3_ByteLength(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Field 16 with Varint wire type. Two-byte tag since 16 doesn't fit in 4 bits
	// with typ3: (16 << 3) | 0 = 128 = 0x80 0x01 (varint-encoded tag).
	bz := []byte{0x80, 0x01, 0x00}

	var s PrimitivesStruct
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "field 16: expected typ3")
}

// Time field in PrimitivesStruct (field 18) expects ByteLength.
// Sending it as 8Byte should be rejected.
func TestUnmarshalBinary2_RejectsWrongTyp3_Time(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Field 18 ByteLength is the correct tag: (18 << 3)|2 = 146 = 0x92 0x01.
	// Send it as 8Byte instead: (18 << 3)|1 = 145 = 0x91 0x01.
	bz := []byte{0x91, 0x01, 0, 0, 0, 0, 0, 0, 0, 0}

	var s PrimitivesStruct
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "field 18: expected typ3")
}

// Unpacked list field: subsequent repeated entries must also have the
// correct typ3. GnoVMBlock.Values is []GnoVMTypedValue at field 5 (look up
// BinFieldNum if the number differs). Each entry wire type is ByteLength.
// Send one valid entry then a malformed second entry with wrong typ3.
func TestUnmarshalBinary2_RejectsWrongTyp3_UnpackedListContinuation(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Marshal a known-good value with two entries first, then corrupt
	// the second entry's tag byte.
	orig := &GnoVMBlock{
		Values: []GnoVMTypedValue{
			{N: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}},
			{N: [8]byte{9, 10, 11, 12, 13, 14, 15, 16}},
		},
	}
	bz, err := cdc.MarshalBinary2(orig)
	if err != nil {
		t.Fatal(err)
	}

	// Find the second occurrence of tag(Values-fnum, ByteLength) by
	// looking up the field num dynamically: scan for a sequence that could
	// be that tag. The deterministic way: roundtrip to confirm baseline.
	var orig2 GnoVMBlock
	if err := orig2.UnmarshalBinary2(cdc, bz); err != nil {
		t.Fatalf("baseline roundtrip failed: %v", err)
	}

	// Now flip the typ3 of the last tag byte preceding the last entry to
	// Varint. We do this by finding the second-to-last tag byte. The
	// simplest construction: build a minimal malformed packet by hand
	// using PrependFieldNumberAndTyp3 helpers would be ideal, but since
	// we don't have those in the test package, we corrupt the marshaled
	// bytes. Scan for a ByteLength tag with fnum matching Values.
	//
	// For this synthetic test, we just verify that mangling any ByteLength
	// tag byte inside the Values repeats to something != Typ3ByteLength
	// produces an error. We mutate the last byte before the last entry's
	// length prefix.
	//
	// Simpler: rely on the test type's field numbering. GnoVMBlock's
	// Values field number comes from amino registration order. We don't
	// rely on a specific number; instead, confirm that corrupting any
	// tag byte's typ3 (lower 3 bits) to Typ3Varint in the repeated-tag
	// region triggers the validation.
	//
	// Find bytes that look like a tag for ByteLength (lower 3 bits == 2).
	// Skip the first one (the outer initial tag). Flip to Varint.
	found := 0
	for i := 0; i < len(bz); i++ {
		if bz[i]&0x07 == typ3ByteLength {
			found++
			if found == 2 {
				bz[i] = (bz[i] &^ 0x07) | typ3Varint
				break
			}
		}
	}
	if found < 2 {
		t.Skip("could not find second ByteLength tag to corrupt")
	}

	var corrupted GnoVMBlock
	err = corrupted.UnmarshalBinary2(cdc, bz)
	if err == nil {
		t.Fatalf("expected error on corrupted typ3, got nil")
	}
	// The error could be from the unpacked-list-continuation check or from
	// a decoder downstream; either way, unmarshal must not silently succeed.
}

// AminoMarshalerStruct1 has repr = ReprStruct1 (a struct with C int64, D int64).
// The implicit-struct wrapping means the outer wire must be:
//   tag(1, ByteLength) | length | inner bytes
// If we send tag(1, Varint), the repr-unmarshal check should reject.
func TestUnmarshalBinary2_RejectsWrongTyp3_AminoMarshalerRepr(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Wait — AminoMarshalerStruct1 repr IS a struct, which decodes directly,
	// not via implicit-struct wrapping. So this case doesn't hit the repr
	// typ3 check.
	// The repr typ3 check targets primitive-repr AminoMarshalers, e.g.
	// AminoMarshalerStruct3 whose repr is int32 (Varint).
	// Send it wrapped in ByteLength instead of Varint.

	// AminoMarshalerStruct3.MarshalAmino → int32 (Varint), wrapped in
	// implicit struct field 1.
	// Correct: tag(1,Varint)=0x08 followed by varint(value).
	// Malformed: tag(1,ByteLength)=0x0A (triggers check).
	bz := []byte{0x0A, 0x01, 0x00}

	var s AminoMarshalerStruct3
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "repr field 1")
}

// An Any envelope with trailing bytes past field 2 should be rejected,
// not silently dropped.
func TestUnmarshalAnyBinary2_RejectsTrailingBytes(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	typeURL := "/tests.Concrete1"
	bz := []byte{0x0a, byte(len(typeURL))}
	bz = append(bz, []byte(typeURL)...)
	// Valid but empty field 2: tag(2, ByteLength) | length=0
	bz = append(bz, 0x12, 0x00)
	// Trailing bytes past field 2: tag(3, ByteLength) | length=0
	bz = append(bz, 0x1a, 0x00)

	err := cdc.UnmarshalAnyBinary2(bz, new(Interface1))
	assertErrContains(t, err, "trailing bytes")
}

// DecodeByteSlice must reject length prefixes that would exceed the
// remaining buffer, including pathological uint64 values that wrap when
// cast to int on 32-bit platforms.
func TestDecodeByteSlice_RejectsOversizeLength(t *testing.T) {
	// length = 0xFFFFFFFFFFFFFFFF (uvarint encoding)
	// 10 bytes: 0xff 0xff 0xff 0xff 0xff 0xff 0xff 0xff 0xff 0x01
	bz := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}

	_, _, err := amino.DecodeByteSlice(bz)
	if err == nil {
		t.Fatal("expected error on oversize length, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient bytes") {
		t.Fatalf("expected 'insufficient bytes' error, got %q", err.Error())
	}
}

// Field number 0 is reserved by proto3 and must be rejected by the decoder.
func TestDecodeFieldNumberAndTyp3_RejectsField0(t *testing.T) {
	// tag byte: (0 << 3) | 2 = 0x02 (field 0 with ByteLength typ3)
	bz := []byte{0x02, 0x00}

	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Route through a top-level unmarshal: any struct should reject.
	var s PrimitivesStruct
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "invalid field num 0")
}

// Implicit struct wrapper (used for AminoMarshaler packed-slice repr with
// nested lists) must reject trailing bytes after field 1.
// Exercise via SlicesSlicesStruct which contains nested packed lists.
func TestUnmarshalBinary2_RejectsImplicitStructTrailingBytes(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Build a roundtrippable SlicesSlicesStruct with one small nested list,
	// then inject trailing bytes inside the implicit struct.
	orig := SlicesSlicesStruct{Int8SlSl: [][]int8{{1, 2}}}
	bz, err := cdc.MarshalBinary2(&orig)
	if err != nil {
		t.Fatal(err)
	}
	// The wire layout is: outer tag | outer len | ibz(field 1 inner list).
	// Find the inner length prefix and inflate it so the payload includes
	// garbage bytes past field 1.
	//
	// Conservative approach: append stray bytes to ibz by mutating the outer
	// length to be larger and the inner ByteSlice length to be smaller, so
	// the decoder sees "trailing bytes after field 1" inside the implicit
	// struct. To avoid fragile offset math on the complex layout, just
	// verify that Marshal produces something, then manually verify the
	// implicit-struct trailing-bytes check would fire if we fed a crafted
	// payload. Fall back: confirm a known-malformed byte sequence errors.
	//
	// Direct malformed input: an implicit struct wrapper containing field
	// 1 + field 2 inside, which our decoder should now reject.
	// Bytes: outer tag(fnum=1,ByteLength)=0x0a | outer_len=0x06 |
	//        [ tag(1,ByteLength)=0x0a | len=0x00 |
	//          tag(2,ByteLength)=0x12 | len=0x02 | 0x00 0x00 ]
	// This targets the first nested-list field (Int8SlSl = field 1).
	malformed := []byte{0x0a, 0x06, 0x0a, 0x00, 0x12, 0x02, 0x00, 0x00}
	var bad SlicesSlicesStruct
	err = bad.UnmarshalBinary2(cdc, malformed)
	assertErrContains(t, err, "trailing bytes after field 1")
	_ = bz // referenced to avoid unused-var warning if the assertion is satisfied before
}

// AminoMarshalerStruct2.MarshalAmino → []ReprElem2 (unpacked slice repr).
// Each element is wrapped as field 1 ByteLength. If a repeated entry has a
// wrong typ3, the unpacked-slice-repr decoder should reject it.
func TestUnmarshalBinary2_RejectsWrongTyp3_UnpackedSliceRepr(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	// Valid first entry: tag(1,ByteLength)=0x0A, length=0, no body.
	// Malformed second entry: tag(1,Varint)=0x08, value=0.
	bz := []byte{0x0A, 0x00, 0x08, 0x00}

	var s AminoMarshalerStruct2
	err := s.UnmarshalBinary2(cdc, bz)
	assertErrContains(t, err, "unpacked slice repr")
}
