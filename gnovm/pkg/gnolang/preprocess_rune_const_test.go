package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// TestUntypedRuneConstLayout pins the persisted form of package constants
// folded from untyped rune operations. A package block is amino-encoded
// with all eight bytes of N, so these bytes are part of the state hash.
//
// The exact folding of untyped rune arithmetic must lay a result out
// exactly as the int32 folding did, including the bytes a comparison or a
// narrowing conversion leaves in N.
func TestUntypedRuneConstLayout(t *testing.T) {
	const path = "gno.land/p/demo/runelayout"
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	m := NewMachineWithOptions(MachineOptions{Store: st, Output: io.Discard})
	defer m.Release()
	_, pv := m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "runelayout", Path: path,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: "module = \"" + path + "\"\ngno = \"0.9\"\n"},
			{Name: "c.gno", Body: `package runelayout

const (
	Lit  = 'a'
	Sum  = 'a' + 1
	Mul  = '\U0010ffff' * 'Ā'
	Neg  = '\x00' - 'a'
	Eq   = 'Ā' == 'Ā'
	Neg8 int8 = 'a' - 'b'
	U16 uint16 = 'Ł'
)
`},
		},
	}, true)

	pn := pv.GetPackageNode(st)
	blk := pv.GetBlock(st)
	get := func(name Name) TypedValue {
		idx, ok := pn.GetLocalIndex(name)
		require.True(t, ok, name)
		return blk.Values[idx]
	}
	tests := []struct {
		name Name
		t    Type
		n    [8]byte
	}{
		{"Lit", UntypedRuneType, [8]byte{0x61}},
		{"Sum", UntypedBigintType, [8]byte{}}, // rune op int is an untyped int, as before
		{"Mul", UntypedRuneType, [8]byte{0x00, 0xff, 0xff, 0x10}},
		{"Neg", UntypedRuneType, [8]byte{0x9f, 0xff, 0xff, 0xff}},
		{"Eq", UntypedBoolType, [8]byte{0x01, 0x01}},
		{"Neg8", Int8Type, [8]byte{0xff, 0xff, 0xff, 0xff}},
		{"U16", Uint16Type, [8]byte{0x41, 0x01}},
	}
	for _, tt := range tests {
		tv := get(tt.name)
		require.Equal(t, tt.t, tv.T, tt.name)
		require.Equal(t, tt.n, tv.N, tt.name)

		// The same bytes survive an amino round trip.
		bz := amino.MustMarshal(tv)
		var tv2 TypedValue
		amino.MustUnmarshal(bz, &tv2)
		require.Equal(t, tt.n, tv2.N, tt.name)
	}
}
