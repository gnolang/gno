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
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0x0A, 0x01, 0x42})
	f.Fuzz(func(t *testing.T, data []byte) {
		var s tests.PrimitivesStruct
		_ = s.UnmarshalBinary2(nil, data)

		var cs tests.ComplexSt
		cdc := amino.NewCodec()
		cdc.RegisterPackage(tests.Package)
		cdc.Seal()
		_ = cs.UnmarshalBinary2(cdc, data)
	})
}
