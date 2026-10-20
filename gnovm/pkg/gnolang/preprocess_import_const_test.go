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

// TestImportedConstAfterReload checks that an importer folds a constant of
// an imported package from the value the running VM computes, not from the
// value persisted in the package block when the package was deployed.
//
// A package deployed before untyped rune arithmetic was exact persisted
// '\U0010ffff' * '\U0010ffff' wrapped to int32 (-2228223). After a restart
// the package itself is preprocessed again and folds the exact value; an
// importer must see the same value.
func TestImportedConstAfterReload(t *testing.T) {
	const (
		pPath = "gno.land/p/demo/runeconst"
		qPath = "gno.land/p/demo/runeconstuser"
	)
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	mod := func(path string) *std.MemFile {
		return &std.MemFile{Name: "gnomod.toml", Body: "module = \"" + path + "\"\ngno = \"0.9\"\n"}
	}
	run := func(st Store, mpkg *std.MemPackage) *PackageValue {
		m := NewMachineWithOptions(MachineOptions{Store: st, Output: io.Discard})
		defer m.Release()
		_, pv := m.RunMemPackage(mpkg, true)
		return pv
	}

	// Deploy the package, then rewrite its persisted C slot the way the
	// old folding stored it.
	st := NewStore(nil, tm2, tm2)
	pv := run(st, &std.MemPackage{
		Type: MPUserProd, Name: "runeconst", Path: pPath,
		Files: []*std.MemFile{
			mod(pPath),
			{Name: "c.gno", Body: "package runeconst\n\nconst C = '\\U0010ffff' * '\\U0010ffff'\n"},
		},
	})
	pn := pv.GetPackageNode(st)
	idx, ok := pn.GetLocalIndex("C")
	require.True(t, ok)
	blk := pv.GetBlock(st)
	require.Equal(t, int64(1241243320321), blk.Values[idx].GetUntypedRune().Int64())
	wrapped := TypedValue{T: UntypedRuneType}
	wrapped.SetInt32(-2228223)
	blk.Values[idx] = wrapped
	st.SetObject(blk)

	// Restart: a fresh store over the same backend, as the keeper does.
	st2 := NewStore(nil, tm2, tm2)
	m := NewMachineWithOptions(MachineOptions{Store: st2, Output: io.Discard})
	m.PreprocessAllFilesAndSaveBlockNodes()
	m.Release()
	persisted := st2.GetPackage(pPath, false).GetBlock(st2).Values[idx]
	require.Equal(t, int32(-2228223), persisted.GetInt32(), "the persisted slot keeps the old folding")

	qv := run(st2, &std.MemPackage{
		Type: MPUserProd, Name: "runeconstuser", Path: qPath,
		Files: []*std.MemFile{
			mod(qPath),
			{Name: "q.gno", Body: "package runeconstuser\n\nimport \"" + pPath + "\"\n\nvar X int64 = runeconst.C\n"},
		},
	})
	qidx, ok := qv.GetPackageNode(st2).GetLocalIndex("X")
	require.True(t, ok)
	x := qv.GetBlock(st2).Values[qidx]
	if hi, ok := x.V.(*HeapItemValue); ok {
		x = hi.Value
	}
	require.Equal(t, int64(1241243320321), x.GetInt64())
}

// TestUntypedRuneConstLayout pins the persisted form of package constants
// folded from untyped rune operations. A package block is amino-encoded
// with all eight bytes of N, so these bytes are part of the state hash.
//
// Two forms differ from the int32 folding this replaced, although the
// values do not: a comparison of two untyped runes used to leave the
// operand's upper bytes in N under the boolean, and a narrowing conversion
// of an untyped rune used to leave the rune's upper bytes under the
// narrow value. Both now leave N zeroed past the value.
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
	Big  = '\U0010ffff' * '\U0010ffff'
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
		{"Sum", UntypedRuneType, [8]byte{0x62}},
		{"Big", UntypedRuneType, [8]byte{}},
		{"Eq", UntypedBoolType, [8]byte{0x01}}, // was 01 01 00 ...
		{"Neg8", Int8Type, [8]byte{0xff}},      // was ff ff ff ff 00 ...
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
	big := get("Big")
	require.Equal(t, int64(1241243320321), big.GetUntypedRune().Int64())
	var big2 TypedValue
	amino.MustUnmarshal(amino.MustMarshal(big), &big2)
	require.Equal(t, int64(1241243320321), big2.V.(BigintValue).V.Int64())
}
