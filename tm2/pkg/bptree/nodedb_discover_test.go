package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	errIterOpen = errors.New("iterator open failed")
	errIterRead = errors.New("iterator read failed")
)

// failingIterDB fails the iterators of the directions whose flag is set, at
// open when atOpen is set, otherwise through Error() on an exhausted iterator.
type failingIterDB struct {
	dbm.DB
	forward, reverse, atOpen bool
}

func (f failingIterDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	if !f.forward {
		return f.DB.Iterator(start, end)
	}
	return f.fail()
}

func (f failingIterDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if !f.reverse {
		return f.DB.ReverseIterator(start, end)
	}
	return f.fail()
}

func (f failingIterDB) fail() (dbm.Iterator, error) {
	if f.atOpen {
		return nil, errIterOpen
	}
	return erroringIterator{}, nil
}

type erroringIterator struct{}

func (erroringIterator) Domain() (start, end []byte) { return nil, nil }
func (erroringIterator) Valid() bool                 { return false }
func (erroringIterator) Next()                       {}
func (erroringIterator) Key() []byte                 { panic("invalid iterator") }
func (erroringIterator) Value() []byte               { panic("invalid iterator") }
func (erroringIterator) Error() error                { return errIterRead }
func (erroringIterator) Close() error                { return nil }

// scanRootVersions is the full PrefixRoot scan that discoverVersions performed
// before it read the two ends of the range.
func scanRootVersions(t *testing.T, db dbm.DB) (first, latest int64) {
	t.Helper()
	itr, err := db.Iterator([]byte{PrefixRoot}, []byte{PrefixRoot + 1})
	require.NoError(t, err)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		if len(key) != 9 { // prefix(1) + version(8)
			continue
		}
		v := int64(binary.BigEndian.Uint64(key[1:]))
		if first == 0 || v < first {
			first = v
		}
		if v > latest {
			latest = v
		}
	}
	require.NoError(t, itr.Error())
	return first, latest
}

type rootLayout struct {
	versions []int64
	deleted  []int64
	stray    [][]byte
}

func versionRange(from, to int64) []int64 {
	vs := make([]int64, 0, to-from+1)
	for v := from; v <= to; v++ {
		vs = append(vs, v)
	}
	return vs
}

func rootLayouts() map[string]rootLayout {
	r := PrefixRoot
	zeros12 := append([]byte{r}, make([]byte, 12)...)
	ff9 := append([]byte{r}, bytes.Repeat([]byte{0xff}, 9)...)

	layouts := map[string]rootLayout{
		"empty":                  {},
		"one":                    {versions: []int64{7}},
		"contiguous":             {versions: versionRange(1, 5000)},
		"pruned with tombstones": {versions: versionRange(1, 5000), deleted: versionRange(1, 3000)},
		"hardfork start":         {versions: versionRange(1_000_001, 1_000_050)},
		"stray only":             {stray: [][]byte{{r}, {r, 1, 2}}},
		"stray keys at both ends": {
			versions: versionRange(10, 20),
			stray:    [][]byte{{r}, {r, 0, 0}, zeros12, ff9},
		},
		"version zero": {versions: []int64{0, 5, 9}},
		"max version":  {versions: []int64{1, math.MaxInt64}},
	}

	rng := rand.New(rand.NewSource(1))
	for i := range 50 {
		var l rootLayout
		for range rng.Intn(301) {
			l.versions = append(l.versions, 1+rng.Int63n(2_000_000))
		}
		for range rng.Intn(5) {
			n := 1 + rng.Intn(16)
			if n == 9 {
				n = 10
			}
			key := make([]byte, n)
			key[0] = r
			rng.Read(key[1:])
			l.stray = append(l.stray, key)
		}
		layouts[fmt.Sprintf("random %02d", i)] = l
	}
	return layouts
}

// writeLayout also sets two 9-byte keys just outside the root range, so a
// discovery bound that leaks reads them as versions and changes the result.
func writeLayout(t *testing.T, db dbm.DB, l rootLayout) {
	t.Helper()
	for _, v := range l.versions {
		require.NoError(t, db.Set(rootDBKey(v), []byte{1}))
	}
	for _, v := range l.deleted {
		require.NoError(t, db.Set(rootDBKey(v), []byte{1}))
		require.NoError(t, db.Delete(rootDBKey(v)))
	}
	for _, k := range l.stray {
		require.NoError(t, db.Set(k, []byte{1}))
	}

	below := binary.BigEndian.AppendUint64([]byte{PrefixRoot - 1}, 3)
	above := binary.BigEndian.AppendUint64([]byte{PrefixRoot + 1}, 1<<40)
	require.NoError(t, db.Set(below, []byte{1}))
	require.NoError(t, db.Set(above, []byte{1}))
}

func discover(t *testing.T, db dbm.DB) (first, latest int64) {
	t.Helper()
	ndb := newNodeDB(db, 0, NewNopLogger(), DefaultOptions())
	require.NoError(t, ndb.discoverVersions())
	return ndb.getFirstVersion(), ndb.getLatestVersion()
}

// openPebbleComposition builds the store's read stack: a pebble DB behind a
// prefix, read through a snapshot behind the same prefix.
func openPebbleComposition(t *testing.T, fill func(dbm.DB)) dbm.DB {
	t.Helper()
	pdb, err := pebbledb.NewPebbleDB("discover", t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pdb.Close()) })

	prefix := []byte("s/_/")
	fill(dbm.NewPrefixDB(pdb, prefix))
	require.NoError(t, dbm.NewPrefixDB(pdb, prefix).Set([]byte("o:x"), []byte{1}))

	snap, err := pdb.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, snap.Close()) })

	return dbm.NewPrefixDB(dbm.NewSnapshotDB(snap), prefix)
}

func TestDiscoverVersionsMatchesScan(t *testing.T) {
	for name, l := range rootLayouts() {
		check := func(t *testing.T, db dbm.DB) {
			wantFirst, wantLatest := scanRootVersions(t, db)
			first, latest := discover(t, db)
			require.Equal(t, wantFirst, first, "first")
			require.Equal(t, wantLatest, latest, "latest")
		}
		t.Run("memdb/"+name, func(t *testing.T) {
			db := memdb.NewMemDB()
			writeLayout(t, db, l)
			check(t, db)
		})
		t.Run("pebble/prefix/snapshot/"+name, func(t *testing.T) {
			check(t, openPebbleComposition(t, func(db dbm.DB) { writeLayout(t, db, l) }))
		})
	}
}

func TestDiscoverVersionsKnownLayouts(t *testing.T) {
	cases := map[string]struct{ first, latest int64 }{
		"empty":                   {0, 0},
		"contiguous":              {1, 5000},
		"pruned with tombstones":  {3001, 5000},
		"hardfork start":          {1_000_001, 1_000_050},
		"stray only":              {0, 0},
		"stray keys at both ends": {10, 20},
		"version zero":            {5, 9},
		"max version":             {1, math.MaxInt64},
	}
	layouts := rootLayouts()
	for name, want := range cases {
		l, ok := layouts[name]
		require.True(t, ok, "layout %q", name)
		check := func(t *testing.T, db dbm.DB) {
			first, latest := discover(t, db)
			require.Equal(t, want.first, first, "first")
			require.Equal(t, want.latest, latest, "latest")
		}
		t.Run("memdb/"+name, func(t *testing.T) {
			db := memdb.NewMemDB()
			writeLayout(t, db, l)
			check(t, db)
		})
		t.Run("pebble/prefix/snapshot/"+name, func(t *testing.T) {
			check(t, openPebbleComposition(t, func(db dbm.DB) { writeLayout(t, db, l) }))
		})
	}
}

func TestDiscoverVersionsPropagatesIteratorError(t *testing.T) {
	cases := map[string]struct {
		atOpen bool
		want   error
	}{
		"open": {atOpen: true, want: errIterOpen},
		"read": {atOpen: false, want: errIterRead},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mem := memdb.NewMemDB()
			writeLayout(t, mem, rootLayout{versions: versionRange(1, 10)})
			db := failingIterDB{DB: mem, forward: true, reverse: true, atOpen: tc.atOpen}

			ndb := newNodeDB(db, 0, NewNopLogger(), DefaultOptions())
			ndb.mtx.Lock()
			ndb.firstVersion = 7
			ndb.latestVersion = 9
			ndb.mtx.Unlock()

			require.ErrorIs(t, ndb.discoverVersions(), tc.want)
			require.Equal(t, int64(7), ndb.getFirstVersion())
			require.Equal(t, int64(9), ndb.getLatestVersion())
		})
	}
}
