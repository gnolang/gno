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
	// Zero Hashlet + non-zero NewTime: MarshalAmino returns
	// "0000…00:1337" (non-empty). Checks that only the byte-array part
	// going through hex encoding produces an all-zeros prefix while the
	// ":N" suffix prevents the repr from being empty.
	zeroPkgIDNonzeroTime := &ObjectID{NewTime: 1337}

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

		// ObjectID — AminoMarshaler repr-zeroness surface. Zero PkgID
		// still produces a non-empty repr ("0000...0:N") because the
		// ":N" suffix is always emitted.
		{"ObjectID/zero", zeroOid},
		{"ObjectID/zero-pkgid-nonzero-time", zeroPkgIDNonzeroTime},
		{"ObjectID/populated", populatedOid},

		// Hashlet standalone: typed [20]byte with hex-encoded AminoMarshaler.
		// Zero value produces "0000000000000000000000000000000000000000"
		// (non-empty, 40 chars) — a second exemplar of the repr-zeroness
		// surface distinct from ObjectID's colon-joined form.
		{"Hashlet/zero", func() *Hashlet { h := Hashlet{}; return &h }()},
		{"Hashlet/populated", func() *Hashlet {
			h := NewHashlet([]byte("aaaaaaaaaaaaaaaaaaaa"))
			return &h
		}()},

		// ValuePath variants.
		{"ValuePath/field", vpField},
		{"ValuePath/block-zero", vpBlock},

		// Default TypedValue (nil T, nil V) — empty wire form.
		{"TypedValue/empty", tv},

		// Location/metadata types — the simplest serializable gnolang shapes.
		{"Pos", &Pos{Line: 10, Column: 5}},
		{"Span", &Span{Pos: Pos{Line: 10, Column: 5}, End: Pos{Line: 12, Column: 1}, Num: 0}},
		{"Location", &Location{
			PkgPath: "gno.land/r/foo",
			File:    "foo.gno",
			Span:    Span{Pos: Pos{Line: 3, Column: 1}, End: Pos{Line: 7, Column: 2}},
		}},

		// ValueHash — AminoMarshaler with hex repr, exercises the
		// repr-zeroness guard for a typed byte-array (like Hashlet).
		{"ValueHash/populated", &ValueHash{
			Hashlet: NewHashlet([]byte("01234567890123456789")),
		}},

		// Value types — minimal instances that round-trip cleanly.
		{"RefValue/populated", &RefValue{
			ObjectID: ObjectID{PkgID: PkgID{Hashlet: NewHashlet([]byte("abcdefghabcdefghabcd"))}, NewTime: 5},
			Hash:     ValueHash{Hashlet: NewHashlet([]byte("hhhhhhhhhhhhhhhhhhhh"))},
		}},
		{"RefValue/pkgpath-only", &RefValue{PkgPath: "gno.land/p/demo/foo"}},
		{"ExportRefValue", &ExportRefValue{ObjectID: ":42"}},
		{"HeapItemValue/empty", &HeapItemValue{}},

		// ObjectInfo standalone — carried by every persisted object.
		{"ObjectInfo/populated", &ObjectInfo{
			ID:       ObjectID{PkgID: PkgID{Hashlet: NewHashlet([]byte("pidpidpidpidpidpidpi"))}, NewTime: 7},
			Hash:     ValueHash{Hashlet: NewHashlet([]byte("ihihihihihihihihihih"))},
			ModTime:  100,
			RefCount: 3,
		}},

		// Realm — package-state wrapper.
		{"Realm", &Realm{
			ID:      PkgID{Hashlet: NewHashlet([]byte("rpidrpidrpidrpidrpid"))},
			Path:    "gno.land/r/demo/rlm",
			Time:    42,
			Deposit: 1000,
			Storage: 512,
		}},

		// PointerValue with nil Base/TV — the allocation minimum.
		{"PointerValue/nil", &PointerValue{Index: 0}},

		// SliceValue referring to no backing Value.
		{"SliceValue/empty", &SliceValue{Base: nil, Offset: 0, Length: 0, Maxcap: 0}},
	}
}
