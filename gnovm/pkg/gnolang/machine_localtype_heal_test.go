package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// A package deployed before addpkg persisted function-local types carries
// no type record for them, and nothing on the deploy path can heal it
// afterwards: both saveNewPackageValuesAndTypes call sites sit behind
// `if save`. The boot re-preprocess is the only moment every node revisits
// every stored package, so the heal belongs there.
func TestBootHealsFuncLocalTypes(t *testing.T) {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})

	mpkg := &std.MemPackage{
		Type: MPUserAll,
		Name: "localtype",
		Path: "gno.land/p/demo/localtype",
		Files: []*std.MemFile{{Name: "a.gno", Body: `package localtype

func Mk() any {
	type S struct{ N int }
	return S{N: 1}
}
`}},
	}

	store := NewStore(nil, baseStore, iavlStore)
	m := NewMachineWithOptions(MachineOptions{Store: store})
	pn, _ := m.RunMemPackage(mpkg, true)
	var tids []TypeID
	for _, dt := range pn.FuncLocalTypes() {
		tids = append(tids, dt.TypeID())
	}
	m.Release()
	if len(tids) == 0 {
		t.Fatal("no function-local types collected; the fixture no longer exercises the path")
	}
	for _, tid := range tids {
		if baseStore.Get(nil, []byte(backendTypeKey(tid))) == nil {
			t.Fatalf("addpkg did not persist %s", tid)
		}
	}

	// Rewind to what a pre-change deploy left behind.
	for _, tid := range tids {
		baseStore.Delete(nil, []byte(backendTypeKey(tid)))
		if baseStore.Get(nil, []byte(backendTypeKey(tid))) != nil {
			t.Fatalf("could not clear %s", tid)
		}
	}

	// Restart: a fresh store over the same backend, so no cacheTypes entry
	// can stand in for the missing record.
	store2 := NewStore(nil, baseStore, iavlStore)
	m2 := NewMachineWithOptions(MachineOptions{Store: store2})
	m2.PreprocessAllFilesAndSaveBlockNodes()
	m2.Release()

	for _, tid := range tids {
		if baseStore.Get(nil, []byte(backendTypeKey(tid))) == nil {
			t.Fatalf("boot left %s unresolvable", tid)
		}
	}
}
