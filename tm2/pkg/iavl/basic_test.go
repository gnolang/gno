package iavl

import (
	"bytes"
	mrand "math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	tree := NewMutableTree(memdb.NewMemDB(), 0)
	up := tree.Set([]byte("1"), []byte("one"))
	if up {
		t.Error("Did not expect an update (should have been create)")
	}
	up = tree.Set([]byte("2"), []byte("two"))
	if up {
		t.Error("Did not expect an update (should have been create)")
	}
	up = tree.Set([]byte("2"), []byte("TWO"))
	if !up {
		t.Error("Expected an update")
	}
	up = tree.Set([]byte("5"), []byte("five"))
	if up {
		t.Error("Did not expect an update (should have been create)")
	}

	// Test 0x00
	{
		idx, val := tree.Get([]byte{0x00})
		if val != nil {
			t.Errorf("Expected no value to exist")
		}
		if idx != 0 {
			t.Errorf("Unexpected idx %x", idx)
		}
		if string(val) != "" {
			t.Errorf("Unexpected value %v", string(val))
		}
	}

	// Test "1"
	{
		idx, val := tree.Get([]byte("1"))
		if val == nil {
			t.Errorf("Expected value to exist")
		}
		if idx != 0 {
			t.Errorf("Unexpected idx %x", idx)
		}
		if string(val) != "one" {
			t.Errorf("Unexpected value %v", string(val))
		}
	}

	// Test "2"
	{
		idx, val := tree.Get([]byte("2"))
		if val == nil {
			t.Errorf("Expected value to exist")
		}
		if idx != 1 {
			t.Errorf("Unexpected idx %x", idx)
		}
		if string(val) != "TWO" {
			t.Errorf("Unexpected value %v", string(val))
		}
	}

	// Test "4"
	{
		idx, val := tree.Get([]byte("4"))
		if val != nil {
			t.Errorf("Expected no value to exist")
		}
		if idx != 2 {
			t.Errorf("Unexpected idx %x", idx)
		}
		if string(val) != "" {
			t.Errorf("Unexpected value %v", string(val))
		}
	}

	// Test "6"
	{
		idx, val := tree.Get([]byte("6"))
		if val != nil {
			t.Errorf("Expected no value to exist")
		}
		if idx != 3 {
			t.Errorf("Unexpected idx %x", idx)
		}
		if string(val) != "" {
			t.Errorf("Unexpected value %v", string(val))
		}
	}
}

func TestUnit(t *testing.T) {
	t.Parallel()

	expectHash := func(tree *ImmutableTree, hashCount int64) {
		// ensure number of new hash calculations is as expected.
		hash, count := tree.hashWithCount()
		if count != hashCount {
			t.Fatalf("Expected %v new hashes, got %v", hashCount, count)
		}
		// nuke hashes and reconstruct hash, ensure it's the same.
		tree.root.traverse(tree, true, func(node *Node) bool {
			node.hash = nil
			return false
		})
		// ensure that the new hash after nuking is the same as the old.
		newHash, _ := tree.hashWithCount()
		if !bytes.Equal(hash, newHash) {
			t.Fatalf("Expected hash %v but got %v after nuking", hash, newHash)
		}
	}

	expectSet := func(tree *MutableTree, i int, repr string, hashCount int64) {
		origNode := tree.root
		updated := tree.Set(i2b(i), []byte{})
		// ensure node was added & structure is as expected.
		if updated || P(tree.root) != repr {
			t.Fatalf("Adding %v to %v:\nExpected         %v\nUnexpectedly got %v updated:%v",
				i, P(origNode), repr, P(tree.root), updated)
		}
		// ensure hash calculation requirements
		expectHash(tree.ImmutableTree, hashCount)
		tree.root = origNode
	}

	expectRemove := func(tree *MutableTree, i int, repr string, hashCount int64) {
		origNode := tree.root
		value, removed := tree.Remove(i2b(i))
		// ensure node was added & structure is as expected.
		if len(value) != 0 || !removed || P(tree.root) != repr {
			t.Fatalf("Removing %v from %v:\nExpected         %v\nUnexpectedly got %v value:%v removed:%v",
				i, P(origNode), repr, P(tree.root), value, removed)
		}
		// ensure hash calculation requirements
		expectHash(tree.ImmutableTree, hashCount)
		tree.root = origNode
	}

	// ////// Test Set cases:

	// Case 1:
	t1 := T(N(4, 20))

	expectSet(t1, 8, "((4 8) 20)", 3)
	expectSet(t1, 25, "(4 (20 25))", 3)

	t2 := T(N(4, N(20, 25)))

	expectSet(t2, 8, "((4 8) (20 25))", 3)
	expectSet(t2, 30, "((4 20) (25 30))", 4)

	t3 := T(N(N(1, 2), 6))

	expectSet(t3, 4, "((1 2) (4 6))", 4)
	expectSet(t3, 8, "((1 2) (6 8))", 3)

	t4 := T(N(N(1, 2), N(N(5, 6), N(7, 9))))

	expectSet(t4, 8, "(((1 2) (5 6)) ((7 8) 9))", 5)
	expectSet(t4, 10, "(((1 2) (5 6)) (7 (9 10)))", 5)

	// ////// Test Remove cases:

	t10 := T(N(N(1, 2), 3))

	expectRemove(t10, 2, "(1 3)", 1)
	expectRemove(t10, 3, "(1 2)", 0)

	t11 := T(N(N(N(1, 2), 3), N(4, 5)))

	expectRemove(t11, 4, "((1 2) (3 5))", 2)
	expectRemove(t11, 3, "((1 2) (4 5))", 1)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	size := 10000
	keyLen, dataLen := 16, 40

	d, err := db.NewDB("test", "memdb", "")
	require.NoError(t, err)

	defer d.Close()
	t1 := NewMutableTree(d, size)

	// insert a bunch of random nodes
	keys := make([][]byte, size)
	l := int32(len(keys))
	for i := range size {
		key := randBytes(keyLen)
		t1.Set(key, randBytes(dataLen))
		keys[i] = key
	}

	for i := range 10 {
		step := 50 * i
		// remove a bunch of existing keys (may have been deleted twice)
		for range step {
			key := keys[mrand.Int31n(l)]
			t1.Remove(key)
		}
		t1.SaveVersion()
	}
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	type record struct {
		key   string
		value string
	}

	records := make([]*record, 400)
	tree := NewMutableTree(memdb.NewMemDB(), 0)

	randomRecord := func() *record {
		return &record{randstr(20), randstr(20)}
	}

	for i := range records {
		r := randomRecord()
		records[i] = r
		updated := tree.Set([]byte(r.key), []byte{})
		if updated {
			t.Error("should have not been updated")
		}
		updated = tree.Set([]byte(r.key), []byte(r.value))
		if !updated {
			t.Error("should have been updated")
		}
		if tree.Size() != int64(i+1) {
			t.Error("size was wrong", tree.Size(), i+1)
		}
	}

	for _, r := range records {
		if has := tree.Has([]byte(r.key)); !has {
			t.Error("Missing key", r.key)
		}
		if has := tree.Has([]byte(randstr(12))); has {
			t.Error("Table has extra key")
		}
		if _, val := tree.Get([]byte(r.key)); string(val) != r.value {
			t.Error("wrong value")
		}
	}

	for i, x := range records {
		if val, removed := tree.Remove([]byte(x.key)); !removed {
			t.Error("Wasn't removed")
		} else if string(val) != x.value {
			t.Error("Wrong value")
		}
		for _, r := range records[i+1:] {
			if has := tree.Has([]byte(r.key)); !has {
				t.Error("Missing key", r.key)
			}
			if has := tree.Has([]byte(randstr(12))); has {
				t.Error("Table has extra key")
			}
			_, val := tree.Get([]byte(r.key))
			if string(val) != r.value {
				t.Error("wrong value")
			}
		}
		if tree.Size() != int64(len(records)-(i+1)) {
			t.Error("size was wrong", tree.Size(), (len(records) - (i + 1)))
		}
	}
}

func TestIterateRange(t *testing.T) {
	t.Parallel()

	type record struct {
		key   string
		value string
	}

	records := []record{
		{"abc", "123"},
		{"low", "high"},
		{"fan", "456"},
		{"foo", "a"},
		{"foobaz", "c"},
		{"good", "bye"},
		{"foobang", "d"},
		{"foobar", "b"},
		{"food", "e"},
		{"foml", "f"},
	}
	keys := make([]string, len(records))
	for i, r := range records {
		keys[i] = r.key
	}
	sort.Strings(keys)

	tree := NewMutableTree(memdb.NewMemDB(), 0)

	// insert all the data
	for _, r := range records {
		updated := tree.Set([]byte(r.key), []byte(r.value))
		if updated {
			t.Error("should have not been updated")
		}
	}
	// test traversing the whole node works... in order
	viewed := []string{}
	tree.Iterate(func(key []byte, value []byte) bool {
		viewed = append(viewed, string(key))
		return false
	})
	if len(viewed) != len(keys) {
		t.Error("not the same number of keys as expected")
	}
	for i, v := range viewed {
		if v != keys[i] {
			t.Error("Keys out of order", v, keys[i])
		}
	}

	trav := traverser{}
	tree.IterateRange([]byte("foo"), []byte("goo"), true, trav.view)
	expectTraverse(t, trav, "foo", "food", 5)

	trav = traverser{}
	tree.IterateRange([]byte("aaa"), []byte("abb"), true, trav.view)
	expectTraverse(t, trav, "", "", 0)

	trav = traverser{}
	tree.IterateRange(nil, []byte("flap"), true, trav.view)
	expectTraverse(t, trav, "abc", "fan", 2)

	trav = traverser{}
	tree.IterateRange([]byte("foob"), nil, true, trav.view)
	expectTraverse(t, trav, "foobang", "low", 6)

	trav = traverser{}
	tree.IterateRange([]byte("very"), nil, true, trav.view)
	expectTraverse(t, trav, "", "", 0)

	// make sure it doesn't include end
	trav = traverser{}
	tree.IterateRange([]byte("fooba"), []byte("food"), true, trav.view)
	expectTraverse(t, trav, "foobang", "foobaz", 3)

	// make sure backwards also works... (doesn't include end)
	trav = traverser{}
	tree.IterateRange([]byte("fooba"), []byte("food"), false, trav.view)
	expectTraverse(t, trav, "foobaz", "foobang", 3)

	// make sure backwards also works...
	trav = traverser{}
	tree.IterateRange([]byte("g"), nil, false, trav.view)
	expectTraverse(t, trav, "low", "good", 2)
}

func TestPersistence(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()

	// Create some random key value pairs
	records := make(map[string]string)
	for range 10000 {
		records[randstr(20)] = randstr(20)
	}

	// Construct some tree and save it
	t1 := NewMutableTree(db, 0)
	for key, value := range records {
		t1.Set([]byte(key), []byte(value))
	}
	t1.SaveVersion()

	// Load a tree
	t2 := NewMutableTree(db, 0)
	t2.Load()
	for key, value := range records {
		_, t2value := t2.Get([]byte(key))
		if string(t2value) != value {
			t.Fatalf("Invalid value. Expected %v, got %v", value, t2value)
		}
	}
}

func TestProof(t *testing.T) {
	t.Parallel()

	// Construct some random tree
	db := memdb.NewMemDB()
	tree := NewMutableTree(db, 100)
	for range 10 {
		key, value := randstr(20), randstr(20)
		tree.Set([]byte(key), []byte(value))
	}

	// Persist the items so far
	tree.SaveVersion()

	// Add more items so it's not all persisted
	for range 10 {
		key, value := randstr(20), randstr(20)
		tree.Set([]byte(key), []byte(value))
	}

	// Now for each item, construct a proof and verify
	tree.Iterate(func(key []byte, value []byte) bool {
		value2, proof, err := tree.GetWithProof(key)
		assert.NoError(t, err)
		assert.Equal(t, value, value2)
		if assert.NotNil(t, proof) {
			verifyProof(t, proof, tree.WorkingHash())
		}
		return false
	})
}

func TestTreeProof(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	tree := NewMutableTree(db, 100)
	assert.Equal(t, tree.Hash(), []byte(nil))

	// should get false for proof with nil root
	value, proof, err := tree.GetWithProof([]byte("foo"))
	assert.Nil(t, value)
	assert.Nil(t, proof)
	assert.Error(t, proof.Verify([]byte(nil)))
	assert.NoError(t, err)

	// insert lots of info and store the bytes
	keys := make([][]byte, 200)
	for i := range 200 {
		key := randstr(20)
		tree.Set([]byte(key), []byte(key))
		keys[i] = []byte(key)
	}

	tree.SaveVersion()

	// query random key fails
	value, proof, err = tree.GetWithProof([]byte("foo"))
	assert.Nil(t, value)
	assert.NotNil(t, proof)
	assert.NoError(t, err)
	assert.NoError(t, proof.Verify(tree.Hash()))
	assert.NoError(t, proof.VerifyAbsence([]byte("foo")))

	// valid proof for real keys
	root := tree.WorkingHash()
	for _, key := range keys {
		value, proof, err := tree.GetWithProof(key)
		if assert.NoError(t, err) {
			require.Nil(t, err, "Failed to read proof from bytes: %v", err)
			assert.Equal(t, key, value)
			err := proof.Verify(root)
			assert.NoError(t, err, "#### %v", proof.String())
			err = proof.VerifyItem(key, key)
			assert.NoError(t, err, "#### %v", proof.String())
		}
	}
}
