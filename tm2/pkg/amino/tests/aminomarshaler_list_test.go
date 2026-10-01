package tests

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests/crosspkg"
)

// These tests exercise genproto2 code-gen fixes for AminoMarshaler list
// elements (fixes in gen_marshal.go, gen_unmarshal.go, gen_size.go). The
// motivating case is []crypto.Address where crypto.Address = [20]byte has
// MarshalAmino() (string, error). Before the fix, generated code emitted
// `string(elem)` for a [20]byte which doesn't compile.

func newAminoListsCodec() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(amino.RegisterPackage(Package))
	return cdc
}

// TestAminoMarshalerSliceStringRepr: slice of AminoMarshaler with string repr.
// Before fix: marshal generated `string(elem)` for [20]byte; unmarshal did
// `ev = string(v)` for [20]byte. After fix: calls MarshalAmino()/UnmarshalAmino().
func TestAminoMarshalerSliceStringRepr(t *testing.T) {
	cdc := newAminoListsCodec()
	orig := ContainerWithAminoLists{
		Addrs: []SimpleAddress{
			{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
	}
	checkAminoListRoundtrip(t, cdc, "SliceStringRepr", orig)
}

// TestAminoMarshalerArrayStringRepr: fixed-size array of AminoMarshaler.
func TestAminoMarshalerArrayStringRepr(t *testing.T) {
	cdc := newAminoListsCodec()
	addr := SimpleAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	orig := ContainerWithAminoLists{
		TopAddrs: [3]SimpleAddress{addr, addr, addr},
	}
	checkAminoListRoundtrip(t, cdc, "ArrayStringRepr", orig)
}

// TestAminoMarshalerSliceNil: nil slice should roundtrip correctly.
// (amino elides empty slices to nil on the wire; only nil is tested.)
func TestAminoMarshalerSliceNil(t *testing.T) {
	cdc := newAminoListsCodec()
	checkAminoListRoundtrip(t, cdc, "NilSlice", ContainerWithAminoLists{Addrs: nil})
}

// TestAminoMarshalerSizeMatchesMarshal: SizeBinary2 must equal len(MarshalBinary2).
// Exercises gen_size.go writeUnpackedListSize AminoMarshaler branch.
func TestAminoMarshalerSizeMatchesMarshal(t *testing.T) {
	cdc := newAminoListsCodec()
	addr1 := SimpleAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	addr2 := SimpleAddress{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	orig := &ContainerWithAminoLists{
		Addrs:    []SimpleAddress{addr1, addr2, addr1},
		TopAddrs: [3]SimpleAddress{addr1, addr2, addr1},
	}

	bz, err := cdc.MarshalBinary2(orig)
	if err != nil {
		t.Fatalf("MarshalBinary2: %v", err)
	}
	size, err := orig.SizeBinary2(cdc)
	if err != nil {
		t.Fatalf("SizeBinary2: %v", err)
	}
	if size != len(bz) {
		t.Errorf("SizeBinary2=%d but MarshalBinary2 produced %d bytes", size, len(bz))
	}
}

// TestAminoMarshalerCrossDecoderParity: bytes from MarshalBinary2 decode
// identically via genproto2 UnmarshalBinary2 and reflect-based Unmarshal.
func TestAminoMarshalerCrossDecoderParity(t *testing.T) {
	cdc := newAminoListsCodec()
	addr1 := SimpleAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	addr2 := SimpleAddress{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	orig := &ContainerWithAminoLists{
		Addrs:    []SimpleAddress{addr1, addr2},
		TopAddrs: [3]SimpleAddress{addr1, addr2, addr1},
	}

	bz, err := cdc.MarshalBinary2(orig)
	if err != nil {
		t.Fatalf("MarshalBinary2: %v", err)
	}

	// Decode via genproto2.
	viaGenproto2 := &ContainerWithAminoLists{}
	if err := viaGenproto2.UnmarshalBinary2(cdc, bz, 0); err != nil {
		t.Fatalf("UnmarshalBinary2: %v", err)
	}

	// Decode via reflect (force by using a fresh ptr and the reflect path).
	viaReflect := &ContainerWithAminoLists{}
	if err := cdc.UnmarshalReflect(bz, viaReflect); err != nil {
		t.Fatalf("UnmarshalReflect: %v", err)
	}

	if !reflect.DeepEqual(viaGenproto2, viaReflect) {
		t.Errorf("genproto2 and reflect decode disagree:\n  genproto2: %#v\n  reflect:   %#v",
			viaGenproto2, viaReflect)
	}
	if !reflect.DeepEqual(*orig, *viaGenproto2) {
		t.Errorf("genproto2 roundtrip mismatch:\n  orig:    %#v\n  decoded: %#v", *orig, *viaGenproto2)
	}
}

// TestAminoMarshalerAminoGenproto2BytesMatch: amino.Marshal (reflect path)
// and MarshalBinary2 (genproto2 path) must produce byte-identical output.
func TestAminoMarshalerAminoGenproto2BytesMatch(t *testing.T) {
	cdc := newAminoListsCodec()
	addr := SimpleAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	orig := ContainerWithAminoLists{
		Addrs:    []SimpleAddress{addr, addr},
		TopAddrs: [3]SimpleAddress{addr, addr, addr},
	}

	bz1, err := cdc.MarshalReflect(orig)
	if err != nil {
		t.Fatalf("MarshalReflect: %v", err)
	}
	msg := &orig
	bz2, err := cdc.MarshalBinary2(msg)
	if err != nil {
		t.Fatalf("MarshalBinary2: %v", err)
	}
	if !bytes.Equal(bz1, bz2) {
		t.Errorf("reflect and genproto2 bytes differ:\n  reflect:   %x\n  genproto2: %x", bz1, bz2)
	}
}

// TestCrossPkgBoxedReprRoundtrip exercises gen_unmarshal.go's writeReprUnmarshal
// with a cross-package struct repr. Before the fix, the generator emitted
// `var repr Inner` (bare name) which fails to compile since Inner lives in
// a different package.
func TestCrossPkgBoxedReprRoundtrip(t *testing.T) {
	cdc := newAminoListsCodec()
	orig := &CrossPkgBoxedRepr{Val: 12345}

	bz, err := cdc.MarshalBinary2(orig)
	if err != nil {
		t.Fatalf("MarshalBinary2: %v", err)
	}
	decoded := &CrossPkgBoxedRepr{}
	if err := decoded.UnmarshalBinary2(cdc, bz, 0); err != nil {
		t.Fatalf("UnmarshalBinary2: %v", err)
	}
	if decoded.Val != orig.Val {
		t.Errorf("roundtrip mismatch: got %d, want %d", decoded.Val, orig.Val)
	}
}

// TestCrossPkgPointerSliceRoundtrip exercises the pointer-slice AminoMarshaler
// path in gen_marshal.go writeListEncode with cross-package element type.
// The fix at line 374 (and the related line 442 for UnpackedList fields)
// requires ctx.goTypeName(ert.Elem()) so that the generated `new(...)`
// uses the qualified type name (e.g. crosspkg.SmallCount).
func TestCrossPkgPointerSliceRoundtrip(t *testing.T) {
	cdc := newAminoListsCodec()
	c1 := crosspkg.SmallCount(1)
	c2 := crosspkg.SmallCount(255)
	orig := &CrossPkgPointerSlice{Counts: []*crosspkg.SmallCount{&c1, &c2}}

	bz, err := cdc.MarshalBinary2(orig)
	if err != nil {
		t.Fatalf("MarshalBinary2: %v", err)
	}
	decoded := &CrossPkgPointerSlice{}
	if err := decoded.UnmarshalBinary2(cdc, bz, 0); err != nil {
		t.Fatalf("UnmarshalBinary2: %v", err)
	}
	if len(decoded.Counts) != len(orig.Counts) {
		t.Fatalf("len mismatch: got %d, want %d", len(decoded.Counts), len(orig.Counts))
	}
	for i := range orig.Counts {
		if *decoded.Counts[i] != *orig.Counts[i] {
			t.Errorf("element %d: got %d, want %d", i, *decoded.Counts[i], *orig.Counts[i])
		}
	}
}

// checkAminoListRoundtrip runs compareEncoding-style checks for list cases.
func checkAminoListRoundtrip(t *testing.T, cdc *amino.Codec, name string, orig ContainerWithAminoLists) {
	t.Helper()

	bz1, err := cdc.MarshalReflect(orig)
	if err != nil {
		t.Fatalf("%s: MarshalReflect: %v", name, err)
	}
	msg := &orig
	bz2, err := cdc.MarshalBinary2(msg)
	if err != nil {
		t.Fatalf("%s: MarshalBinary2: %v", name, err)
	}
	if !bytes.Equal(bz1, bz2) {
		t.Fatalf("%s: bytes mismatch:\n  reflect:   %x\n  genproto2: %x", name, bz1, bz2)
	}
	size, err := msg.SizeBinary2(cdc)
	if err != nil {
		t.Fatalf("%s: SizeBinary2: %v", name, err)
	}
	if size != len(bz2) {
		t.Errorf("%s: SizeBinary2=%d but MarshalBinary2=%d", name, size, len(bz2))
	}

	decoded := &ContainerWithAminoLists{}
	if err := decoded.UnmarshalBinary2(cdc, bz2, 0); err != nil {
		t.Fatalf("%s: UnmarshalBinary2: %v", name, err)
	}
	if !reflect.DeepEqual(orig, *decoded) {
		t.Errorf("%s: roundtrip mismatch:\n  orig:    %#v\n  decoded: %#v", name, orig, *decoded)
	}
}
