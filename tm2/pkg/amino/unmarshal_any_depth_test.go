package amino

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// depthTestIface is an interface used only by the depth-propagation
// regression tests below. Having an interface field in a decoded concrete
// type is what exercises the depth check in decodeReflectBinaryInterface.
type depthTestIface interface {
	depthTestMarker()
}

type depthTestHolder struct {
	Inner depthTestIface
}

func (depthTestHolder) depthTestMarker() {}

// TestUnmarshalAny2Depth_PropagatesDepth is a regression test for a bug
// where Codec.UnmarshalAny2 reset the anyDepth counter to 0 when called
// as a fallback from unmarshalAnyBinary2Depth. A deeply-nested Any chain
// that switched paths mid-recursion (genproto2 → reflect) would bypass
// the maxAnyDepth guard.
//
// White-box test: calls the internal depth-aware helper directly. With
// the fix, anyDepth propagates into decodeReflectBinaryAny, so decoding
// a struct with an interface field at depth maxAnyDepth triggers a depth
// error when the interface field's decoder increments depth past the
// limit. Without the fix, depth would reset to 0 and the decode would
// succeed.
func TestUnmarshalAny2Depth_PropagatesDepth(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterPackage(pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino",
		"amino.depthtest",
		"",
	).WithTypes(pkg.Type{
		Type: reflect.TypeOf(depthTestHolder{}),
		Name: "Holder",
	}))

	// Encode a holder (Inner nil is fine — we just need any value to pack
	// into an Any envelope and get a typeURL + value back).
	// Non-nil Inner so the decoder actually enters
	// decodeReflectBinaryInterface and checks the depth guard.
	bz, err := cdc.MarshalAny(&depthTestHolder{Inner: depthTestHolder{}})
	if err != nil {
		t.Fatalf("MarshalAny: %v", err)
	}

	// Call the genproto2 Any-decode entry at a depth where, after the
	// fallback to the reflect path (depthTestHolder has no genproto2
	// methods), decoding the non-nil Inner interface field must exceed
	// maxAnyDepth.
	//
	// With the fix, anyDepth propagates into the reflect path so the
	// Inner field's decodeReflectBinaryInterface call uses anyDepth+1 =
	// maxAnyDepth+1 and errors. Without the fix, the fallback reset
	// anyDepth to 0 and the decode would silently succeed.
	var target depthTestIface
	err = cdc.unmarshalAnyBinary2Depth(bz, &target, maxAnyDepth)
	if err == nil {
		t.Fatalf("expected depth error, got nil")
	}
	if !strings.Contains(err.Error(), "max Any nesting depth") {
		t.Fatalf("expected 'max Any nesting depth' error, got: %v", err)
	}
}

// TestUnmarshalAny2Depth_ZeroDepthSucceeds sanity-checks that normal
// (depth=0) calls still succeed — i.e. the fix didn't break the happy path.
func TestUnmarshalAny2Depth_ZeroDepthSucceeds(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterPackage(pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino",
		"amino.depthtest",
		"",
	).WithTypes(pkg.Type{
		Type: reflect.TypeOf(depthTestHolder{}),
		Name: "Holder",
	}))

	// Non-nil Inner so the decoder actually enters
	// decodeReflectBinaryInterface and checks the depth guard.
	bz, err := cdc.MarshalAny(&depthTestHolder{Inner: depthTestHolder{}})
	if err != nil {
		t.Fatalf("MarshalAny: %v", err)
	}
	typeURL, value := splitAnyForTest(t, bz)

	var target depthTestIface
	if err := cdc.unmarshalAny2Depth(typeURL, value, &target, 0); err != nil {
		t.Fatalf("expected successful decode at depth 0, got: %v", err)
	}
}

// splitAnyForTest parses the Any envelope bytes into (typeURL, value).
func splitAnyForTest(t *testing.T, bz []byte) (string, []byte) {
	t.Helper()

	fnum, typ, n, err := decodeFieldNumberAndTyp3(bz)
	if err != nil {
		t.Fatalf("decode field 1 header: %v", err)
	}
	if fnum != 1 || typ != Typ3ByteLength {
		t.Fatalf("expected Any field 1 TypeURL, got num=%v typ=%v", fnum, typ)
	}
	bz = bz[n:]
	typeURL, n, err := DecodeString(bz)
	if err != nil {
		t.Fatalf("decode TypeURL: %v", err)
	}
	bz = bz[n:]

	var value []byte
	if len(bz) > 0 {
		fnum, typ, n, err = decodeFieldNumberAndTyp3(bz)
		if err != nil {
			t.Fatalf("decode field 2 header: %v", err)
		}
		if fnum != 2 || typ != Typ3ByteLength {
			t.Fatalf("expected Any field 2 Value, got num=%v typ=%v", fnum, typ)
		}
		bz = bz[n:]
		value, _, err = DecodeByteSlice(bz)
		if err != nil {
			t.Fatalf("decode Value: %v", err)
		}
	}
	return typeURL, value
}
