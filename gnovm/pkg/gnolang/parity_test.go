package gnolang

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/cockroachdb/apd/v3"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
)

// TestCodecParity_Gnolang asserts that every hand-crafted gnovm value
// round-trips byte-identically through both the reflect codec and the
// genproto2 fast path, with SizeBinary2 matching and both decoders
// agreeing.
//
// Add new cases by appending to parityCasesGnolang below.
func TestCodecParity_Gnolang(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	for i, c := range parityCasesGnolang() {
		c := c
		name := fmt.Sprintf("%d/%s", i, c.name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}

func parityCasesGnolang() []struct {
	name string
	v    any
} {
	strVal := StringValue("hello world")

	bigintVal := &BigintValue{V: big.NewInt(0)}
	bigintNeg := &BigintValue{V: big.NewInt(-9223372036854775808)}
	bigintLarge := &BigintValue{V: new(big.Int).Lsh(big.NewInt(1), 200)} // 2^200

	bigdec, _, err := apd.NewFromString("3.14159265358979323846")
	if err != nil {
		panic(err)
	}
	bigdecVal := &BigdecValue{V: bigdec}

	// ObjectID: AminoMarshaler returning "hex:time" string. Zero value
	// becomes a non-empty string ("0000...0:0") — the same
	// repr-zeroness surface the recent AminoMarshaler fix addressed for
	// crypto.Address.
	zeroOid := &ObjectID{}
	pidBytes := []byte("0123456789abcdef0123456789abcdef")
	populatedOid := &ObjectID{
		PkgID:   PkgID{Hashlet: NewHashlet(pidBytes[:HashSize])},
		NewTime: 42,
	}

	// ValuePath covers different VPType cases.
	vpField := &ValuePath{Type: VPField, Depth: 2, Index: 3, Name: "Field1"}
	vpBlock := &ValuePath{Type: VPBlock, Depth: 0, Index: 0, Name: ""}

	// TypedValue — the central gnovm value carrier. Wrapped so it goes
	// through the genproto2 fast path registered on *TypedValue.
	tv := &TypedValue{}
	tv.T = nil
	tv.V = nil

	return []struct {
		name string
		v    any
	}{
		// Edge-value BigintValue: min int64, zero, and a large (>64-bit)
		// positive value that forces multi-word big.Int encoding.
		{"BigintValue/zero", bigintVal},
		{"BigintValue/minint64", bigintNeg},
		{"BigintValue/2^200", bigintLarge},

		// BigdecValue with a non-integer, non-round value.
		{"BigdecValue/pi", bigdecVal},

		// StringValue.
		{"StringValue/nonempty", &strVal},
		{"StringValue/empty", func() *StringValue { v := StringValue(""); return &v }()},

		// ObjectID — AminoMarshaler repr-zeroness surface.
		{"ObjectID/zero", zeroOid},
		{"ObjectID/populated", populatedOid},

		// ValuePath variants.
		{"ValuePath/field", vpField},
		{"ValuePath/block-zero", vpBlock},

		// Default TypedValue (nil T, nil V) — empty wire form.
		{"TypedValue/empty", tv},
	}
}
