package gnolang

import (
	"io"
	"testing"

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
