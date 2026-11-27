package gnolang

import (
	"bytes"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// objectLoadRealm persists the object shapes a read-modify-write transaction
// loads on chain: a binary tree of struct nodes reached through pointers
// (StructValue and HeapItemValue), closures (FuncValue), a package
// block (Block) and the package itself (PackageValue).
const objectLoadRealm = `package bench

type Node struct {
	Key         string
	Value       int
	Height      int
	Left, Right *Node
}

var (
	root    *Node
	size    int
	counter = func() int { size++; return size }
	label   = func(n *Node) string { return n.Key + "=" + itoa(n.Value) }
)

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for ; i > 0; i /= 10 {
		s = string(rune('0'+i%10)) + s
	}
	return s
}

func insert(n *Node, k string, v int) *Node {
	if n == nil {
		return &Node{Key: k, Value: v, Height: 1}
	}
	if k < n.Key {
		n.Left = insert(n.Left, k, v)
	} else {
		n.Right = insert(n.Right, k, v)
	}
	return n
}

func init() {
	for i := 0; i < 200; i++ {
		root = insert(root, itoa((i*7919)%1000), counter())
	}
}

func Get(k string) string {
	for n := root; n != nil; {
		if k == n.Key {
			return label(n)
		}
		if k < n.Key {
			n = n.Left
		} else {
			n = n.Right
		}
	}
	return ""
}
`

type storedObject struct {
	oid  ObjectID
	kind string
	bz   []byte // amino bytes, without the leading hash
}

// persistObjectLoadRealm runs objectLoadRealm against a fresh store and
// returns that store with every object it persisted, sorted by kind.
func persistObjectLoadRealm(tb testing.TB) (Store, storetypes.CommitStore, []storedObject) {
	tb.Helper()
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	wrapped := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
	m := NewMachineWithOptions(MachineOptions{PkgPath: "gno.vm/r/bench", Store: txSt, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "bench",
		Path:  "gno.vm/r/bench",
		Files: []*std.MemFile{{Name: "bench.gno", Body: objectLoadRealm}},
	}, true)
	txSt.Write()
	wrapped.Write()

	var objs []storedObject
	it, err := db.Iterator(nil, nil)
	require.NoError(tb, err)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		if k := string(it.Key()); !strings.HasPrefix(k, "oid:") || strings.HasSuffix(k, "#realm") {
			continue
		}
		bz := bytes.Clone(it.Value()[HashSize:])
		var oo Object
		amino.MustUnmarshal(bz, &oo)
		objs = append(objs, storedObject{oo.GetObjectID(), reflect.TypeOf(oo).Elem().Name(), bz})
	}
	sort.SliceStable(objs, func(i, j int) bool { return objs[i].kind < objs[j].kind })
	require.NotEmpty(tb, objs)
	return st, tm2Store, objs
}

// BenchmarkObjectLoad measures GetObject on an empty transaction cache, the
// path every transaction takes for every object it touches: backend read,
// amino decode, type fill. One op loads every persisted object of one kind.
// It runs unchanged on master, so before and after compare with benchstat:
//
//	go test ./gnovm/pkg/gnolang/ -run x -bench BenchmarkObjectLoad -benchmem -count=6 > old.txt
//	go test ./gnovm/pkg/gnolang/ -run x -bench BenchmarkObjectLoad -benchmem -count=6 > new.txt
//	benchstat old.txt new.txt
func BenchmarkObjectLoad(b *testing.B) {
	st, tm2Store, objs := persistObjectLoadRealm(b)
	byKind := map[string][]ObjectID{}
	var kinds []string
	for _, o := range objs {
		if byKind[o.kind] == nil {
			kinds = append(kinds, o.kind)
		}
		byKind[o.kind] = append(byKind[o.kind], o.oid)
	}
	for _, kind := range kinds {
		oids := byKind[kind]
		b.Run("kind="+kind, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				cw := tm2Store.CacheWrap()
				txSt := st.BeginTransaction(cw, cw, nil, nil)
				for _, oid := range oids {
					if txSt.GetObject(oid) == nil {
						b.Fatalf("object %v not found", oid)
					}
				}
			}
			b.ReportMetric(float64(len(oids)), "objects/op")
		})
	}
}
