package gnolang

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// TestSealSkipsBuiltMethodIndex pins the methodIndex == nil guard in sealType.
// Deleting it makes publication write a fresh empty map into a field concurrent
// readers already hold, which TestConcurrentInitChain only catches in about half
// of its runs without -race.
func TestSealSkipsBuiltMethodIndex(t *testing.T) {
	const methods = methodIndexThreshold + 1

	dt := &DeclaredType{PkgPath: "gno.vm/t/seal", Name: "T", Base: IntType}
	for i := range methods {
		ft := &FuncType{}
		dt.Methods = append(dt.Methods, TypedValue{
			T: ft,
			V: &FuncValue{Name: Name(fmt.Sprintf("M%d", i)), Type: ft},
		})
	}
	dt.buildMethodIndex()
	require.Len(t, dt.methodIndex, methods)

	// A key no rebuild reproduces, so surviving the seal proves the guard held.
	const sentinel = Name("sentinel-not-a-method")
	dt.methodIndex[sentinel] = 0

	newSealer().sealType(dt)

	_, kept := dt.methodIndex[sentinel]
	require.True(t, kept, "sealType rebuilt an already-built methodIndex")
	require.Len(t, dt.methodIndex, methods+1)
}

// batchRecorder counts how one file's block nodes reach the store. The empty
// package node each caller publishes before the file loop is not counted: it
// carries no types yet, so sealing it walks nothing.
type batchRecorder struct {
	Store
	batches [][]BlockNode
	singles int
}

func (r *batchRecorder) SetBlockNodes(bns []BlockNode) {
	r.batches = append(r.batches, bns)
	r.Store.SetBlockNodes(bns)
}

func (r *batchRecorder) SetBlockNode(bn BlockNode) {
	if _, isPkg := bn.(*PackageNode); !isPkg {
		r.singles++
	}
	r.Store.SetBlockNode(bn)
}

// TestSaveBlockNodesPublishesOneBatch pins the call shape that keeps sealing
// proportional to the type graph rather than to the graph times the nodes in it.
// Looping the single-node method instead re-walks the package's whole graph once
// per node, which costs about 40% on genesis and fails nothing else in the tree.
func TestSaveBlockNodesPublishesOneBatch(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	wrapped := tm2Store.CacheWrap()
	rec := &batchRecorder{Store: st.BeginTransaction(wrapped, wrapped, nil, nil)}

	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.vm/t/batch",
		Store:   rec,
		Output:  io.Discard,
	})
	// Several block nodes in one file: two function bodies and a nested block,
	// so per-node publication would be visibly more than one call.
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "batch",
		Path: "gno.vm/t/batch",
		Files: []*std.MemFile{{
			Name: "batch.gno",
			Body: `package batch
type T int
func (t T) M() int { return int(t) }
func main() { x := T(1); { y := x.M(); println(y) } }`,
		}},
	}, true)

	require.Len(t, rec.batches, 1, "SaveBlockNodes should publish one batch per file")
	require.Greater(t, len(rec.batches[0]), 1,
		"the batch should carry the package node plus the file's own block nodes")
	require.Zero(t, rec.singles,
		"a file's block nodes should not be published one at a time")
}

// TestPublicationSeals pins the property the rest of the change rests on, at
// both doors: a package published straight into the store, and one published
// through a transaction's Write, must both come out with the memo caches on
// their shared graph filled.
//
// It asserts only on caches preprocessing provably leaves cold. DeclaredType's
// own typeid and its nameIndex are already filled on this path and would pass
// with sealing removed; pkgID, a method's FuncType typeid and its bound form are
// not.
func TestPublicationSeals(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)

	t.Run("direct", func(t *testing.T) {
		const path = "gno.vm/t/direct"
		deployOneMethodPackage(st, path, "Direct", "M")
		requireSealedType(t, st, path, "Direct")
	})

	t.Run("transaction", func(t *testing.T) {
		const path = "gno.vm/t/tx"
		wrapped := tm2Store.CacheWrap()
		txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
		deployOneMethodPackage(txSt, path, "Tx", "N")
		txSt.Write()
		wrapped.Write()
		requireSealedType(t, st, path, "Tx")
	})
}

// deployOneMethodPackage deploys a one-file package declaring typeName with a
// single method.
func deployOneMethodPackage(store Store, path, typeName, methodName string) {
	name := path[len("gno.vm/t/"):]
	m := NewMachineWithOptions(MachineOptions{PkgPath: path, Store: store, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: name,
		Path: path,
		Files: []*std.MemFile{{
			Name: name + ".gno",
			Body: fmt.Sprintf(`package %s
type %s int
func (v %s) %s(a int) int { return int(v) + a }
`, name, typeName, typeName, methodName),
		}},
	}, true)
}

// requireSealedType reads typeName back off the published package node and
// requires the three caches sealing fills to be populated.
func requireSealedType(t *testing.T, store Store, path, typeName string) {
	t.Helper()

	pn := store.GetBlockNode(PackageNodeLocation(path))
	var dt *DeclaredType
	for _, ty := range pn.GetStaticBlock().Types {
		if d, ok := ty.(*DeclaredType); ok && string(d.Name) == typeName {
			dt = d
			break
		}
	}
	require.NotNil(t, dt, "%s not found on the published package node", typeName)

	require.False(t, dt.pkgID.IsZero(), "DeclaredType.pkgID left unsealed")
	require.Len(t, dt.Methods, 1)
	ft, ok := dt.Methods[0].T.(*FuncType)
	require.True(t, ok)
	require.False(t, ft.typeid.IsZero(), "the method's FuncType typeid left unsealed")
	require.NotNil(t, ft.bound, "the method's bound FuncType left unsealed")
}
