package gnolang

import (
	"fmt"
	"io"
	"slices"
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

// published returns every node the recorder saw, in publication order.
func (r *batchRecorder) published() []BlockNode {
	return slices.Concat(r.batches...)
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
// It asserts only on the three caches this path provably leaves cold: pkgID, the
// method's FuncType typeid, and its bound form. DeclaredType's own typeid is
// filled by preprocessing either way, so asserting it would pass with sealing
// removed.
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

// TestSealFillsPackageNodePkgID pins the PackageNode branch in sealBlockNode.
// The field is reached through packageOf(last).GetPkgID() on every preprocess,
// which every vm/qeval performs against the shared package node, so leaving it
// cold is a write two concurrent queries can race on.
//
// A published package usually has it filled by the deploy's own preprocessing,
// so asserting on a deployed node would pass with the branch removed. Sealing a
// fresh node is what makes this test fail when the branch is reverted.
func TestSealFillsPackageNodePkgID(t *testing.T) {
	const path = "gno.vm/t/sealpkgid"

	pn := NewPackageNode("sealpkgid", path, nil)
	require.True(t, pn.pkgID.IsZero(), "a fresh PackageNode must start with no pkgID")

	newSealer().sealBlockNode(pn)

	require.Equal(t, PkgIDFromPkgPath(path), pn.pkgID,
		"sealBlockNode left PackageNode.pkgID for a concurrent reader to fill")
}

// astTypes collects every type a published file node holds outside its static
// block: on a constTypeExpr, or under ATTR_TYPE_VALUE / ATTR_TYPEOF_VALUE.
// These are the types sealExprTypes exists for.
//
// It deliberately does not call sealExprTypes' own reader. Sharing it would
// make the test circular: a source sealExprTypes forgets to read is a source
// this probe would then forget to check.
func astTypes(bns []BlockNode) []Type {
	var out []Type
	add := func(t Type) {
		if t != nil {
			out = append(out, t)
		}
	}
	for _, bn := range bns {
		fn, ok := bn.(*FileNode)
		if !ok {
			continue
		}
		Transcribe(fn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
			if stage != TRANS_ENTER {
				return n, TRANS_CONTINUE
			}
			if cte, ok := n.(*constTypeExpr); ok {
				add(cte.Type)
			}
			if t, ok := n.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
				add(t)
			}
			if t, ok := n.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
				add(t)
			}
			return n, TRANS_CONTINUE
		})
	}
	return out
}

// anonStructsByField indexes every struct type reachable from ts by the names
// of its fields, so one walk serves every field the test asserts on.
func anonStructsByField(ts []Type) map[Name]*StructType {
	out := map[Name]*StructType{}
	seen := map[Type]bool{}
	var walk func(t Type)
	walk = func(t Type) {
		if t == nil || seen[t] {
			return
		}
		seen[t] = true
		switch ct := t.(type) {
		case *StructType:
			for _, f := range ct.Fields {
				out[f.Name] = ct
				walk(f.Type)
			}
		case *PointerType:
			walk(ct.Elt)
		case *SliceType:
			walk(ct.Elt)
		case *ArrayType:
			walk(ct.Elt)
		case *MapType:
			walk(ct.Key)
			walk(ct.Value)
		case *FuncType:
			for i := range ct.Params {
				walk(ct.Params[i].Type)
			}
			for i := range ct.Results {
				walk(ct.Results[i].Type)
			}
		case *tupleType:
			for _, e := range ct.Elts {
				walk(e)
			}
		}
	}
	for _, t := range ts {
		walk(t)
	}
	return out
}

// TestSealFillsExpressionOnlyTypes pins sealExprTypes. A composite type
// written in expression position is defined under no name, so it appears in no
// StaticBlock.Types and no Block.Values — the two places the static-block walk
// looks. It lives on the expression instead, on a block node the defaultStore
// shares with every store forked from it, with its memos cold.
//
// Removing the sealExprTypes call from SaveBlockNodes leaves every case below
// unsealed.
func TestSealFillsExpressionOnlyTypes(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	rec := &batchRecorder{Store: st}

	const path = "gno.vm/t/anonexpr"
	m := NewMachineWithOptions(MachineOptions{PkgPath: path, Store: rec, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "anonexpr",
		Path: path,
		Files: []*std.MemFile{{
			Name: "anonexpr.gno",
			Body: `package anonexpr

func sink(a any) {}

func Anon() {
	sink(struct{ inLiteral int }{1})
	sink(&struct{ inPointer int }{2})
	sink(new(struct{ inNew bool }))
	sink(map[struct{ inMapKey int }]bool{})
	sink([]struct{ inSlice int }{})
	sink(func(p struct{ inParam int }) struct{ inResult int } { return struct{ inResult int }{} })
}

func Assert(v any) {
	if _, ok := v.(struct{ inAssert int }); ok {
		sink(v)
	}
}
`,
		}},
	}, true)

	byField := anonStructsByField(astTypes(rec.published()))
	require.NotEmpty(t, byField, "no AST-held types found; the probe itself is broken")

	for _, field := range []Name{
		"inLiteral", "inPointer", "inNew", "inMapKey",
		"inSlice", "inParam", "inResult", "inAssert",
	} {
		stt := byField[field]
		require.NotNilf(t, stt, "anonymous struct with field %s not found on the published AST", field)
		require.Falsef(t, stt.typeid.IsZero(),
			"struct{%s ...}: typeid left for a concurrent reader to fill", field)
		require.NotEqualf(t, uint8(0), stt.comparable,
			"struct{%s ...}: comparable left for a concurrent reader to fill", field)
	}
}

// TestSealWalksTupleElements pins the *tupleType case in sealType. A tuple is
// only ever reached through an expression's ATTR_TYPEOF_VALUE, so it arrives
// here only once sealExprTypes is walking the AST. Falling through to the
// default branch calls tupleType.TypeID(), which recurses into each element's
// TypeID() and so hides the gap — every other memo on the elements stays cold.
func TestSealWalksTupleElements(t *testing.T) {
	elt := &StructType{
		PkgPath: "gno.vm/t/tuple",
		Fields:  []FieldType{{Name: "a", Type: IntType}},
	}
	tt := &tupleType{Elts: []Type{elt}}

	newSealer().sealType(tt)

	require.False(t, tt.typeid.IsZero(), "tupleType typeid left unsealed")
	require.False(t, elt.typeid.IsZero(), "tuple element typeid left unsealed")
	require.False(t, elt.pkgID.IsZero(), "tuple element pkgID left unsealed")
	require.NotEqual(t, uint8(0), elt.comparable, "tuple element comparable left unsealed")
}
