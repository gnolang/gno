package genproto2

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// runWithCodecInfo tries to get a TypeInfo for rt, recovering panics. Returns
// (info, errOrPanic) — info is non-nil only when both succeed.
func runWithCodecInfo(rt reflect.Type) (info *amino.TypeInfo, earlyPanic string) {
	defer func() {
		if r := recover(); r != nil {
			earlyPanic = fmt.Sprint(r)
		}
	}()
	cdc := amino.NewCodec()
	var err error
	info, err = cdc.GetTypeInfo(rt)
	if err != nil {
		earlyPanic = err.Error()
	}
	return
}

// assertUnsupportedKind verifies that the generator refuses to emit code for
// the given unsupported type. Accepts either (a) amino's codec rejects it at
// TypeInfo construction, or (b) the generator function panics when called.
// In either case, no silent broken code can be produced.
func assertUnsupportedKind(t *testing.T, rt reflect.Type) {
	t.Helper()

	info, early := runWithCodecInfo(rt)
	if info == nil {
		// Codec already rejected the type — great, no generator call needed.
		if early == "" {
			t.Fatalf("type %v: GetTypeInfo returned nil info and nil error", rt)
		}
		t.Logf("type %v rejected at codec level: %s", rt, early)
		return
	}

	// Codec accepted it; now verify each generator function panics.
	cdc := amino.NewCodec()
	ctx := NewP3Context2(cdc)
	info, _ = cdc.GetTypeInfo(rt)

	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"writePrimitiveEncode", func() {
			var sb strings.Builder
			ctx.writePrimitiveEncode(&sb, "x", info, amino.FieldOptions{}, "\t")
		}},
		{"primitiveValueSizeExpr", func() {
			_ = ctx.primitiveValueSizeExpr("x", info, amino.FieldOptions{})
		}},
		{"writePrimitiveDecodeFrom", func() {
			var sb strings.Builder
			ctx.writePrimitiveDecodeFrom(&sb, "x", info, amino.FieldOptions{}, "\t", "bz")
		}},
	} {
		var got string
		func() {
			defer func() {
				if r := recover(); r != nil {
					got = fmt.Sprint(r)
				}
			}()
			tc.fn()
		}()
		if got == "" {
			t.Errorf("%s(%v): expected panic, got no panic", tc.name, rt)
			continue
		}
		if !strings.Contains(got, "unsupported") {
			t.Errorf("%s(%v): expected panic containing 'unsupported', got %q", tc.name, rt, got)
		}
	}
}

// These tests verify that the generator refuses — either directly (its own
// panic) or transitively (amino's codec rejects the type first) — to emit
// code for unsupported Go kinds. Without this, unsupported kinds would
// silently produce broken code with mismatched marshal/size/unmarshal.

func TestGenproto2_PanicsOnUintptr(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(uintptr(0)))
}

func TestGenproto2_PanicsOnComplex64(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(complex64(0)))
}

func TestGenproto2_PanicsOnComplex128(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(complex128(0)))
}

func TestGenproto2_PanicsOnMap(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(map[string]int{}))
}

func TestGenproto2_PanicsOnChan(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(make(chan int)))
}

func TestGenproto2_PanicsOnFunc(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(func() {}))
}

func TestGenproto2_PanicsOnUnsafePointer(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf(unsafe.Pointer(nil)))
}

// Non-uint8 slice element: []int, []int32, []float64 — these should also
// panic because writePrimitiveEncode's Slice case only handles []byte.
func TestGenproto2_PanicsOnNonByteSliceElement(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf([]int{}))
}

// Non-uint8 array element: [4]int, [4]int32 — same reasoning.
func TestGenproto2_PanicsOnNonByteArrayElement(t *testing.T) {
	assertUnsupportedKind(t, reflect.TypeOf([4]int{}))
}
