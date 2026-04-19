package fuzzbinary

import (
	"testing"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// FuzzUnmarshalBinary2 is the native Go fuzz equivalent of Fuzz (go-fuzz)
// but for genproto2's generated UnmarshalBinary2.
// (Test that deserialize never panics)
func FuzzUnmarshalBinary2(f *testing.F) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	// Hand-written edge-case seeds.
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0x0A, 0x01, 0x42})

	// Generated seeds from representative populated values, so the fuzzer
	// starts exploring from valid wire-format inputs rather than having to
	// discover prefix bytes (e.g., valid Any-encoded interfaces) at random.
	// Regenerating at test init keeps these in sync with any wire-format
	// changes in amino or genproto2.
	addSeed := func(pbm amino.PBMessager2) {
		bz, err := cdc.MarshalBinary2(pbm)
		if err != nil {
			f.Fatalf("seed marshal failed: %v", err)
		}
		f.Add(bz)
	}
	addSeed(&tests.InterfaceHeavy{
		Field1: tests.Concrete1{},
		Field2: tests.Concrete2{},
		Items:  []tests.Interface1{tests.Concrete1{}, tests.Concrete2{}},
		Name:   "fuzz",
	})
	addSeed(&tests.GnoVMTypedValue{
		T: tests.Concrete1{},
		V: tests.Concrete2{},
		N: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	})
	addSeed(&tests.AminoMarshalerStruct1{A: 7, B: -3})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Primitive fields only (no codec needed).
		var s tests.PrimitivesStruct
		_ = s.UnmarshalBinary2(nil, data, 0)

		// Embedded structs, no interface fields.
		var cs tests.ComplexSt
		_ = cs.UnmarshalBinary2(cdc, data, 0)

		// Interface-heavy types that exercise UnmarshalAnyBinary2.
		var tv tests.GnoVMTypedValue
		_ = tv.UnmarshalBinary2(cdc, data, 0)

		var blk tests.GnoVMBlock
		_ = blk.UnmarshalBinary2(cdc, data, 0)

		var node tests.GnoVMNode
		_ = node.UnmarshalBinary2(cdc, data, 0)

		var ih tests.InterfaceHeavy
		_ = ih.UnmarshalBinary2(cdc, data, 0)

		// Custom AminoMarshaler (repr type conversion path).
		var am tests.AminoMarshalerStruct1
		_ = am.UnmarshalBinary2(cdc, data, 0)
	})
}
