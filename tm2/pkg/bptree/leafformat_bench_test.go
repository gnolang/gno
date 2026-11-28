package bptree

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// Keyspaces used to price the leaf format. The whole point of
// TypeLeafPrefixed is that a leaf holds B ADJACENT keys, so what it saves
// depends entirely on how much structure adjacent keys share. "random" is the
// adversarial case (nothing shared) and is what the benchmarks/ harness uses;
// "oid" is the shape gno.land's main store actually stores.
var keyspaces = map[string]func(i int) []byte{
	// gno.land main store: "oid:<40 hex chars of PkgID>:<NewTime>" — every
	// object of one realm shares the leading 44 bytes. A realm with many
	// objects therefore fills whole leaves with one prefix.
	"oid": func(i int) []byte {
		return fmt.Appendf(nil, "oid:%040x:%d", i/4096, i%4096)
	},
	// A single realm's objects: the extreme case, one prefix for the whole tree.
	"oid-single-realm": func(i int) []byte {
		return fmt.Appendf(nil, "oid:%040x:%d", 7, i)
	},
	// Random 16-byte keys: no shared structure at all.
	"random": func(int) []byte {
		k := make([]byte, 16)
		randSrc.Read(k)
		return k
	},
}

var randSrc = rand.New(rand.NewSource(1))

// nodeRecordBytes totals the payload bytes of every 'B' record in db, both as
// stored and as the legacy layout would have stored it.
func nodeRecordBytes(t *testing.T, db *memdb.MemDB) (stored, legacy, leaves int) {
	t.Helper()
	itr, err := db.Iterator([]byte{PrefixNode}, []byte{PrefixNode + 1})
	if err != nil {
		t.Fatal(err)
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		payload, err := verifyChecksum(itr.Value())
		if err != nil {
			t.Fatalf("checksum: %v", err)
		}
		stored += len(payload)
		if len(payload) > 0 && payload[0] == TypeLeafPrefixed {
			node, err := ReadNode(GetNodeKey(itr.Key()[1:]), payload)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			legacy += len(appendLeafV1(node.(*LeafNode)))
			leaves++
			continue
		}
		legacy += len(payload)
	}
	return stored, legacy, leaves
}

// TestLeafPrefixed_OnDiskSaving reports what the format saves across whole
// trees, per keyspace. It asserts only the two facts worth pinning: the oid
// keyspace must save substantially, and the random keyspace must not cost
// more than the one byte the extra prefix-length field adds per leaf.
func TestLeafPrefixed_OnDiskSaving(t *testing.T) {
	const n = 20_000
	names := []string{"oid", "oid-single-realm", "random"}
	for _, name := range names {
		gen := keyspaces[name]
		t.Run(name, func(t *testing.T) {
			db := memdb.NewMemDB()
			tr := NewMutableTreeWithDB(db, 0, NewNopLogger())
			defer tr.Close()
			val := make([]byte, 40)
			for i := range n {
				mustSet(t, tr, gen(i), val)
			}
			mustSave(t, tr)

			stored, legacy, leaves := nodeRecordBytes(t, db)
			t.Logf("%s: %d leaves, node records %d -> %d bytes (%.1f%% of legacy), %.1f -> %.1f bytes/key",
				name, leaves, legacy, stored, 100*float64(stored)/float64(legacy),
				float64(legacy)/n, float64(stored)/n)

			switch name {
			case "oid", "oid-single-realm":
				if stored*10 > legacy*7 { // expect at least 30% smaller
					t.Errorf("%s: node records only shrank to %.1f%% of legacy", name, 100*float64(stored)/float64(legacy))
				}
			case "random":
				if stored > legacy+2*leaves {
					t.Errorf("random keyspace overhead %d bytes over %d leaves is more than the prefix-length field",
						stored-legacy, leaves)
				}
			}
		})
	}
}

// BenchmarkLeafSerialize prices the encoder and decoder per leaf on each
// keyspace — the CPU side of the format change (an LCP fold on write, a
// two-source key copy on read).
func BenchmarkLeafSerialize(b *testing.B) {
	for _, name := range []string{"oid", "random"} {
		gen := keyspaces[name]
		leaf := &LeafNode{numKeys: B, miniTree: NewMiniMerkle()}
		for i := range B {
			leaf.keys[i] = gen(i)
			leaf.valueHashes[i] = HashLeafSlot(leaf.keys[i], []byte("v"))
			leaf.valueKeys[i] = (&NodeKey{Version: 1, Nonce: uint32(i + 1)}).GetKey()
		}
		rec, err := leaf.appendSerialized(nil)
		if err != nil {
			b.Fatal(err)
		}
		b.Run("encode/"+name, func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(len(rec)), "recbytes")
			buf := make([]byte, 0, leaf.serializedSize())
			for range b.N {
				if _, err := leaf.appendSerialized(buf[:0]); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("decode/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if _, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, rec); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
