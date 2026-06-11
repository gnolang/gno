package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestChecksum_RoundTrip pins the helper pair: stamp∘verify is identity, the
// stamped buffer never aliases the payload, and a valid empty payload comes
// back non-nil (TestGetValue_EmptyValueReturnsEmptySlice depends on it).
func TestChecksum_RoundTrip(t *testing.T) {
	for _, payload := range [][]byte{
		{},
		[]byte("x"),
		[]byte("hello world"),
		bytes.Repeat([]byte{0xFF}, 1024),
	} {
		stamped := stampChecksum(payload)
		if len(stamped) != len(payload)+checksumSize {
			t.Fatalf("stamped length %d, want %d", len(stamped), len(payload)+checksumSize)
		}
		if len(payload) > 0 {
			stamped[0] ^= 0 // no-op; aliasing check below mutates payload instead
			payloadCopy := append([]byte(nil), payload...)
			payload[0] ^= 0xFF
			if !bytes.Equal(stamped[:len(payload)], payloadCopy) {
				t.Fatal("stamped buffer aliases the payload")
			}
			payload[0] ^= 0xFF
		}
		got, err := verifyChecksum(stamped)
		if err != nil {
			t.Fatalf("verify(stamp(%q)): %v", payload, err)
		}
		if got == nil {
			t.Fatal("verifyChecksum returned nil payload for a valid record")
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("round-trip mismatch: %q != %q", got, payload)
		}
	}

	// Too-short and corrupt records fail with the sentinel.
	for _, bad := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		if _, err := verifyChecksum(bad); !errors.Is(err, ErrChecksumMismatch) {
			t.Fatalf("verifyChecksum(%v): want ErrChecksumMismatch, got %v", bad, err)
		}
	}
	stamped := stampChecksum([]byte("payload"))
	stamped[2] ^= 0x01
	if _, err := verifyChecksum(stamped); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("flipped byte: want ErrChecksumMismatch, got %v", err)
	}
}

// TestChecksum_CorruptionMatrix flips bytes in persisted records of every
// type and asserts the corruption ALWAYS surfaces loud — an error (direct
// reads) or a panic whose message names the checksum (descent paths, which
// panic via getChild by design; see N37) — and never silent acceptance.
func TestChecksum_CorruptionMatrix(t *testing.T) {
	build := func() (*memdb.MemDB, []string) {
		db := memdb.NewMemDB()
		tree := NewMutableTreeWithDB(db, 0, NewNopLogger()) // no cache: force raw reads
		keys := make([]string, 0, 40)
		for i := range 40 {
			k := fmt.Sprintf("key%02d", i)
			keys = append(keys, k)
			if _, err := tree.Set([]byte(k), []byte("v1-"+k)); err != nil {
				t.Fatal(err)
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
		// Overwrites create orphan records for v2.
		for i := range 10 {
			if _, err := tree.Set([]byte(fmt.Sprintf("key%02d", i)), []byte("v2")); err != nil {
				t.Fatal(err)
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
		return db, keys
	}

	// probe exercises every read path and reports whether corruption
	// surfaced loud (error or checksum-naming panic). Wrong-data detection is
	// not needed: CRC-32C catches every single-byte flip, so the assertion is
	// simply "never silent".
	probe := func(db *memdb.MemDB, keys []string) (loud bool) {
		defer func() {
			if r := recover(); r != nil {
				if !strings.Contains(fmt.Sprint(r), "checksum") {
					t.Fatalf("panic without checksum cause: %v", r)
				}
				loud = true
			}
		}()
		tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
		if _, err := tree.Load(); err != nil {
			return true
		}
		for _, k := range keys {
			if _, err := tree.Get([]byte(k)); err != nil {
				return true
			}
			// Also read through v1, so records that the v2 overwrites
			// displaced (orphaned values, replaced nodes) are exercised too.
			if _, err := tree.GetVersioned([]byte(k), 1); err != nil {
				return true
			}
		}
		itr, _ := tree.Iterator(nil, nil, true)
		for ; itr.Valid(); itr.Next() {
			_ = itr.Key()
			_ = itr.Value()
		}
		if itr.Error() != nil {
			return true
		}
		if err := tree.PruneVersionsTo(1); err != nil {
			return true
		}
		return false
	}

	db, _ := build()
	type rec struct{ k, v string }
	var records []rec
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for ; itr.Valid(); itr.Next() {
		records = append(records, rec{string(itr.Key()), string(itr.Value())})
	}
	itr.Close()
	if len(records) < 10 {
		t.Fatalf("expected a populated DB, got %d records", len(records))
	}

	flips := 0
	for _, r := range records {
		// Sample offsets: first, last, and every 7th byte in between.
		offsets := []int{0, len(r.v) - 1}
		for o := 7; o < len(r.v)-1; o += 7 {
			offsets = append(offsets, o)
		}
		for _, off := range offsets {
			// Corrupt one byte of one record on a FRESH db copy.
			db2, keys2 := build()
			corrupted := []byte(r.v)
			corrupted[off] ^= 0x01
			if err := db2.Set([]byte(r.k), corrupted); err != nil {
				t.Fatal(err)
			}
			flips++
			if !probe(db2, keys2) {
				t.Fatalf("SILENT corruption: record %x type=%c offset %d accepted",
					r.k, r.k[0], off)
			}
		}
	}
	t.Logf("corruption matrix: %d records, %d flips, all loud", len(records), flips)
}

// TestGetRoot_TruncatedRootFailsLoud (N39): a truncated/foreign root record
// must error — never load as an empty tree. Pre-checksum, a 0- or 32-byte
// record was silently accepted as "empty tree", making prune/next-save
// permanently destroy the real state.
func TestGetRoot_TruncatedRootFailsLoud(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	raw, err := db.Get(rootDBKey(1))
	if err != nil || raw == nil {
		t.Fatalf("root record missing: %v", err)
	}
	for _, truncLen := range []int{0, 32, 44, len(raw) - 1} {
		if truncLen >= len(raw) {
			continue
		}
		if err := db.Set(rootDBKey(1), raw[:truncLen]); err != nil {
			t.Fatal(err)
		}
		fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
		_, lerr := fresh.Load()
		if lerr == nil {
			t.Fatalf("truncated root (len=%d) loaded without error (size=%d)", truncLen, fresh.Size())
		}
		if !errors.Is(lerr, ErrChecksumMismatch) && !strings.Contains(lerr.Error(), "corrupt root ref") {
			t.Fatalf("truncated root (len=%d): unexpected error %v", truncLen, lerr)
		}
	}
	// Restore and confirm intact load.
	if err := db.Set(rootDBKey(1), raw); err != nil {
		t.Fatal(err)
	}
	fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatalf("restored root failed to load: %v", err)
	}
	if got, err := fresh.Get([]byte("k")); err != nil || string(got) != "v" {
		t.Fatalf("restored tree: %q, %v", got, err)
	}
}

// TestStagedValue_IndependentOfReadBuffer: the staged batch record must be
// independent of the buffer pendingVals serves to readers, so mutating a
// pre-commit Get result cannot change what is committed (the memdb/boltdb
// retain-by-reference channel of N47).
func TestStagedValue_IndependentOfReadBuffer(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := tree.Set([]byte("k"), []byte("AAAA")); err != nil {
		t.Fatal(err)
	}
	g, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	g[0] = 'Z' // hostile caller mutates the read-your-writes buffer
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	got, err := fresh.Get([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "AAAA" {
		t.Fatalf("committed value corrupted by read-buffer mutation: %q", got)
	}
}
