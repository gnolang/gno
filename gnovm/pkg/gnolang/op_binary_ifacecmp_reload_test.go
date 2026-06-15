package gnolang

import (
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// TestBinaryExprIfaceCmp_SurvivesColdReload guards the persistence assumption
// behind BinaryExpr.ifaceCmp: the verdict is computed at preprocess and is NOT
// amino-persisted, so correctness across a node restart relies on packages
// being re-preprocessed on load (store.go: "Upon restart, all packages will be
// re-preprocessed"). This is the same lifecycle the operands' ATTR_TYPEOF_VALUE
// already depends on.
//
// The realm function compares two interface values whose dynamic type ([]int)
// is uncomparable, which must panic. We persist the realm, then load it into a
// COLD store (fresh node cache over the same backing DB) via the restart
// re-preprocess protocol, and assert the panic still fires — proving ifaceCmp
// was re-established on reload rather than silently defaulting to false.
func TestBinaryExprIfaceCmp_SurvivesColdReload(t *testing.T) {
	t.Parallel()

	const pkgPath = "gno.land/r/recmp"
	mpkg := &std.MemPackage{
		Type: MPUserProd,
		Name: "recmp",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: GenGnoModLatest(pkgPath)},
			{Name: "recmp.gno", Body: `package recmp

func Cmp() bool {
	var a interface{} = []int{1}
	var b interface{} = []int{1}
	return a == b // uncomparable dynamic type via interface: must panic
}
`},
		},
	}

	// Shared backing DBs survive the simulated restart.
	baseDB, iavlDB := memdb.NewMemDB(), memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(baseDB, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(iavlDB, storetypes.StoreOptions{})

	// --- Transaction 1: persist the realm (block nodes + mempackage). ---
	st1 := NewStore(NewAllocator(math.MaxInt64), baseStore, iavlStore)
	tx1 := st1.BeginTransaction(nil, nil, nil, nil)
	m1 := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   tx1,
		Output:  io.Discard,
	})
	m1.RunMemPackage(mpkg, true)
	tx1.Write()

	// --- Cold restart: fresh store (empty node cache) over the same DBs,
	// then run the documented restart protocol that re-preprocesses every
	// package and re-saves block nodes. ---
	st2 := NewStore(NewAllocator(math.MaxInt64), baseStore, iavlStore)
	mRestart := NewMachineWithOptions(MachineOptions{Store: st2, Output: io.Discard})
	mRestart.PreprocessAllFilesAndSaveBlockNodes()

	// --- Call Cmp() on the reloaded realm; expect the uncomparable panic. ---
	pv := st2.GetPackage(pkgPath, false)
	require.NotNil(t, pv, "reloaded package value")
	m2 := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st2,
		Output:  io.Discard,
	})
	m2.SetActivePackage(pv)

	defer func() {
		r := recover()
		require.NotNil(t, r, "comparing uncomparable []int via interface must panic after cold reload")
		require.Contains(t, fmt.Sprintf("%v", r), "comparing uncomparable type",
			"panic should be the uncomparable-type runtime error")
	}()
	m2.Eval(Call("Cmp"))
}
