package iavl

import (
	"bytes"
	"fmt"
	mrand "math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/random"
)

func randstr(length int) string {
	return random.RandStr(length)
}

func i2b(i int) []byte {
	buf := new(bytes.Buffer)
	amino.EncodeInt32(buf, int32(i))
	return buf.Bytes()
}

func b2i(bz []byte) int {
	i, _, _ := amino.DecodeInt32(bz)
	return int(i)
}

// Convenience for a new node
func N(l, r interface{}) *Node {
	var left, right *Node
	if _, ok := l.(*Node); ok {
		left = l.(*Node)
	} else {
		left = NewNode(i2b(l.(int)), nil, 0)
	}
	if _, ok := r.(*Node); ok {
		right = r.(*Node)
	} else {
		right = NewNode(i2b(r.(int)), nil, 0)
	}

	n := &Node{
		key:       right.lmd(nil).key,
		value:     nil,
		leftNode:  left,
		rightNode: right,
	}
	n.calcHeightAndSize(nil)
	return n
}

// Setup a deep node
func T(n *Node) *MutableTree {
	d, err := db.NewDB("test", db.MemDBBackend, "")
	if err != nil {
		panic(err)
	}

	t := NewMutableTree(d, 0)

	n.hashWithCount()
	t.root = n
	return t
}

// Convenience for simple printing of keys & tree structure
func P(n *Node) string {
	if n.height == 0 {
		return fmt.Sprintf("%v", b2i(n.key))
	}
	return fmt.Sprintf("(%v %v)", P(n.leftNode), P(n.rightNode))
}

func randBytes(length int) []byte {
	key := make([]byte, length)
	// math.rand.Read always returns err=nil
	// we do not need cryptographic randomness for this test:
	mrand.Read(key)
	return key
}

type traverser struct {
	first string
	last  string
	count int
}

func (t *traverser) view(key, value []byte) bool {
	if t.first == "" {
		t.first = string(key)
	}
	t.last = string(key)
	t.count++
	return false
}

func expectTraverse(t *testing.T, trav traverser, start, end string, count int) {
	t.Helper()

	if trav.first != start {
		t.Error("Bad start", start, trav.first)
	}
	if trav.last != end {
		t.Error("Bad end", end, trav.last)
	}
	if trav.count != count {
		t.Error("Bad count", count, trav.count)
	}
}

func BenchmarkImmutableAvlTreeMemDB(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	db, err := db.NewDB("test", db.MemDBBackend, "")
	require.NoError(b, err)

	b.ResetTimer()

	benchmarkImmutableAvlTreeWithDB(b, db)
}

func benchmarkImmutableAvlTreeWithDB(b *testing.B, db db.DB) {
	b.Helper()

	defer db.Close()

	b.StopTimer()

	t := NewMutableTree(db, 100000)
	value := []byte{}
	for i := 0; i < 1000000; i++ {
		t.Set(i2b(int(random.RandInt31())), value)
		if i > 990000 && i%1000 == 999 {
			t.SaveVersion()
		}
	}
	b.ReportAllocs()
	t.SaveVersion()

	runtime.GC()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ri := i2b(int(random.RandInt31()))
		t.Set(ri, value)
		t.Remove(ri)
		if i%100 == 99 {
			t.SaveVersion()
		}
	}
}
