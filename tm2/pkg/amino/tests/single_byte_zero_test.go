package tests

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// SingleByteZero is a synthetic concrete whose hand-written MarshalBinary2
// always emits exactly one byte: 0x00. Used to exercise the Any.Value
// single-0x00 elision rule that generator-emitted MarshalBinary2 methods
// never produce (they always roll back single-0x00 output at the
// writeReprMarshal level).
type SingleByteZero struct{}

func (SingleByteZero) MarshalBinary2(cdc *amino.Codec, buf []byte, offset int) (int, error) {
	return amino.PrependByte(buf, offset, 0x00), nil
}

func (SingleByteZero) SizeBinary2(cdc *amino.Codec) (int, error) {
	return 1, nil
}

func (*SingleByteZero) UnmarshalBinary2(cdc *amino.Codec, bz []byte, anyDepth int) error {
	return nil
}

// singleByteZeroPackage is a minimal pkg containing just SingleByteZero so
// we can register it on a codec without touching the shared tests.Package
// (which would trigger genproto2 code-gen and emit duplicate methods).
var singleByteZeroPackage = pkg.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/amino/tests",
	"sbz",
	pkg.GetCallersDirname(),
).WithTypes(SingleByteZero{})

func init() {
	// Tell amino's dispatch layer that SingleByteZero has native Binary2
	// methods, so MarshalAny takes the fast path through marshalAnyBinary2
	// (and, via the return-then-PrependBytes fallback, MarshalAnyBinary2).
	amino.RegisterGenproto2Type(reflect.TypeOf((*SingleByteZero)(nil)).Elem())
}

// TestMarshalAny_SingleByteZeroElision asserts that when a concrete type's
// MarshalBinary2 produces exactly [0x00], the Any envelope elides field 2
// (Value) entirely, emitting only field 1 (TypeURL) — matching reflect
// (binary_encode.go:302, `len(bz2) == 1 && bz2[0] == 0x00`).
//
// Without the rule:
//   - `MarshalAnyBinary2` (amino.go:663) checks only `innerLen > 0`.
//   - `marshalAnyBinary2` (amino.go:615) checks only `len(valueBz) > 0`.
//
// Both would emit field 2 with value [0x00] onto the wire, diverging from
// reflect. With the rule, both paths elide field 2.
func TestMarshalAny_SingleByteZeroElision(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(singleByteZeroPackage)
	cdc.Seal()

	typeURL := "/sbz.SingleByteZero"
	// Expected wire: tag(1, ByteLength)=0x0A | len | typeURL bytes — no field 2.
	wantField1 := []byte{0x0A, byte(len(typeURL))}
	wantField1 = append(wantField1, []byte(typeURL)...)

	bz, err := cdc.MarshalAny(SingleByteZero{})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(bz, wantField1) {
		t.Errorf("MarshalAny should omit field 2:\nwant: %X\ngot:  %X", wantField1, bz)
	}

	// Direct test of the Prepend-based MarshalAnyBinary2 path.
	// Size returns 3 (1 tag + 1 len + 1 inner-body) per the #21 coupling
	// note — Size is arithmetic and can't inspect bytes. Allocate extra.
	size, err := cdc.SizeAnyBinary2(SingleByteZero{})
	if err != nil {
		t.Fatal(err)
	}
	typeURLFieldSize := 1 + 1 + len(typeURL) // tag + len + bytes
	// The elided field 2 would have been 3 bytes. Size claims it.
	if want := typeURLFieldSize + 3; size != want {
		t.Logf("SizeAnyBinary2 returns %d (over-counts by 3 per #21 coupling; want %d or the elided %d)", size, want, typeURLFieldSize)
	}

	buf := make([]byte, size)
	newOffset, err := cdc.MarshalAnyBinary2(SingleByteZero{}, buf, size)
	if err != nil {
		t.Fatal(err)
	}
	// After the fix, Marshal writes only field-1 bytes; the 3 bytes of
	// would-be field 2 remain as trailing garbage in the pre-sized buffer,
	// trimmed off by buf[newOffset:].
	got := buf[newOffset:]
	if !bytes.Equal(got, wantField1) {
		t.Errorf("MarshalAnyBinary2 should omit field 2:\nwant: %X\ngot:  %X", wantField1, got)
	}
}
