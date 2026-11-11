package amino

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
	"github.com/stretchr/testify/require"
)

type aliasTestIface interface{ aliasTestMarker() }

type aliasTestLeaf struct {
	Bytes []byte
	Str   string
	Arr   [4]byte
}

func (aliasTestLeaf) aliasTestMarker() {}

type aliasTestOuter struct {
	Leaf   aliasTestLeaf
	Any    aliasTestIface
	Nested [][]byte
	Leaves []aliasTestLeaf
}

// TestDecodeDoesNotAliasInput checks that a decoded value shares no memory
// with the input buffer. The reflect decoder reads nested structs, Any values
// and implicit list structs as sub-slices of the input without copying, which
// is only safe while every decoder that keeps bytes copies them.
func TestDecodeDoesNotAliasInput(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterPackage(pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino",
		"amino.aliastest",
		"",
	).WithTypes(
		pkg.Type{Type: reflect.TypeFor[aliasTestLeaf](), Name: "Leaf"},
		pkg.Type{Type: reflect.TypeFor[aliasTestOuter](), Name: "Outer"},
	))

	leaf := aliasTestLeaf{Bytes: []byte("leaf bytes"), Str: "leaf string", Arr: [4]byte{1, 2, 3, 4}}
	want := aliasTestOuter{
		Leaf:   leaf,
		Any:    leaf,
		Nested: [][]byte{[]byte("one"), []byte("two")},
		Leaves: []aliasTestLeaf{leaf, leaf},
	}

	for _, tc := range []struct {
		name   string
		encode func(any) ([]byte, error)
		decode func([]byte, any) error
	}{
		{"Unmarshal", cdc.Marshal, cdc.Unmarshal},
		{"UnmarshalAny", cdc.MarshalAny, cdc.UnmarshalAny},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := tc.encode(want)
			require.NoError(t, err)

			var got aliasTestOuter
			if tc.name == "UnmarshalAny" {
				var iface any
				require.NoError(t, tc.decode(bz, &iface))
				got = iface.(aliasTestOuter)
			} else {
				require.NoError(t, tc.decode(bz, &got))
			}
			require.Equal(t, want, got)

			for i := range bz {
				bz[i] ^= 0xff
			}
			require.Equal(t, want, got, "decoded value changed when the input buffer was overwritten")
		})
	}
}

func TestTypeURLtoFullname(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		wantErr  bool
	}{
		{in: "/tm.PubKeySecp256k1", want: "tm.PubKeySecp256k1"},
		{in: "type.googleapis.com/google.protobuf.Any", want: "google.protobuf.Any"},
		{in: "a/b/c.D", want: "c.D"},
		{in: "a/", want: ""},
		{in: "noslash", wantErr: true},
		{in: "", wantErr: true},
	} {
		got, err := typeURLtoFullname(tc.in)
		if tc.wantErr {
			require.Error(t, err, tc.in)
			continue
		}
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.want, got, tc.in)
	}
}
