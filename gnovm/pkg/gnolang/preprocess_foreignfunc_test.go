package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

func foreignFuncPkgs() (a, b *std.MemPackage) {
	a = &std.MemPackage{
		Type: MPUserProd, Name: "ffa", Path: "gno.land/r/demo/ffa",
		Files: []*std.MemFile{{Name: "a.gno", Body: "package ffa\n\nvar V string\n\nfunc Set(s string) { V = s }\n"}},
	}
	b = &std.MemPackage{
		Type: MPUserProd, Name: "ffb", Path: "gno.land/r/demo/ffb",
		Files: []*std.MemFile{{Name: "b.gno", Body: "package ffb\n\nimport \"gno.land/r/demo/ffa\"\n\nvar F = ffa.Set\n"}},
	}
	return a, b
}

func foreignFuncStore() Store {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	return NewStore(nil, tm2Store, tm2Store)
}

// Stored source re-preprocessed at node start must not stop the node.
func TestPreprocessForeignFuncAtBoot(t *testing.T) {
	st := foreignFuncStore()
	a, b := foreignFuncPkgs()
	m := NewMachineWithOptions(MachineOptions{PkgPath: a.Path, Store: st, Output: io.Discard})
	m.RunMemPackage(a, true)
	m.Release()
	// b was accepted under earlier rules: stored without preprocessing.
	st.AddMemPackage(b, MPUserProd)

	m2 := NewMachineWithOptions(MachineOptions{PkgPath: "", Store: st, Output: io.Discard})
	defer m2.Release()
	require.NotPanics(t, m2.PreprocessAllFilesAndSaveBlockNodes)
}

// Submitted source takes another realm's function as a value: it runs
// read-only for that realm when called, so preprocessing accepts it.
func TestPreprocessForeignFuncSubmitted(t *testing.T) {
	st := foreignFuncStore()
	a, b := foreignFuncPkgs()
	m := NewMachineWithOptions(MachineOptions{PkgPath: a.Path, Store: st, Output: io.Discard})
	m.RunMemPackage(a, true)
	m.Release()

	m2 := NewMachineWithOptions(MachineOptions{PkgPath: b.Path, Store: st, Output: io.Discard})
	defer m2.Release()
	require.NotPanics(t, func() { m2.RunMemPackage(b, true) })
}
