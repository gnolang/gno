package genproto2

// finding4_test.go: pins the codegen output for `[]*Primitive` fields
// without nil_elements. A prior audit hypothesized a bug where
// codegen's unpacked-list pointer-element "else" branch (treating
// 0x00 as length-prefix-of-empty-message) diverged from reflection's
// (treating 0x00 as defaultValue). This test demonstrates the bug
// can't actually fire for `[]*int`: codegen routes it through PACKED
// encoding because int's typ3 is varint (wire-type 0), not byte-
// length (wire-type 2). The unpacked "else" branch is only reached
// by `[]*StructLike` shapes, where the two interpretations
// (`&Zero` from defaultValue vs. `&Zero` from decoding empty
// length-prefix) are semantically equivalent.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

type slicePtrIntNoNilElems struct {
	Xs []*int
}

func TestSlicePtrInt_CodegenUsesPackedEncoding(t *testing.T) {
	cdc := amino.NewCodec()
	p := pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino/genproto2",
		"genproto2",
		pkg.GetCallersDirname(),
	).WithTypes(slicePtrIntNoNilElems{})
	cdc.RegisterPackage(p)
	cdc.Seal()

	ctx := NewP3Context2(cdc)
	src, err := ctx.GenerateProtobuf3ForTypes("finding4_probe",
		reflect.TypeOf((*slicePtrIntNoNilElems)(nil)).Elem(),
	)
	if err != nil {
		t.Fatalf("GenerateProtobuf3ForTypes: %v", err)
	}

	// Marshal: must emit a SINGLE length-prefixed buffer containing
	// all elements as varints. The smoking gun is one
	// PrependFieldNumberAndTyp3 with Typ3ByteLength wrapping all
	// elements, NOT a per-element field-key emit.
	if !strings.Contains(src, "PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)") {
		t.Errorf("expected packed-list field-key emission for []*int (single Typ3ByteLength wrap)")
	}
	// And the per-element emit must be a varint into a contiguous buffer.
	if !strings.Contains(src, "amino.PrependVarint(buf, offset, int64((*e)))") {
		t.Errorf("expected per-element PrependVarint inside the packed buffer")
	}

	// Unmarshal: must read the single length-prefixed buffer (fbz)
	// and loop DecodeVarint until exhausted. NOT the
	// per-key-then-per-element shape.
	if !strings.Contains(src, "fbz, n, err := amino.DecodeByteSlice(bz)") {
		t.Errorf("expected DecodeByteSlice for packed []*int unmarshal")
	}
	if !strings.Contains(src, "for len(fbz) > 0 {") {
		t.Errorf("expected packed-decode loop `for len(fbz) > 0`")
	}

	// Negative: the suspect "else" branch in writeUnpackedListUnmarshal
	// (gen_unmarshal.go:622-633) emits a `var ev <Elem>` line outside
	// any `if len(bz) > 0 && bz[0] == 0x00` guard, then
	// writeByteSliceElementDecode. For []*int this branch is never
	// reached because of the packed-encoding routing above. The
	// outer field-1 case-body has no `0x00 → nil` shortcut for non-
	// struct pointer elements without nil_elements, which is how it
	// SHOULD be (no such code path is reachable).
	if strings.Contains(src, "if len(bz) > 0 && bz[0] == 0x00") {
		t.Errorf("unexpected unpacked-list 0x00-shortcut in []*int unmarshal — packed routing should suppress this")
	}
}
