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

// stepCountingDB counts the Next calls made on the iterators it opens, and the
// forward iterators it opens.
type stepCountingDB struct {
	dbm.DB
	nextCount, forwardCount int64
}

func (c *stepCountingDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	c.forwardCount++
	itr, err := c.DB.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return stepCountingIterator{Iterator: itr, count: &c.nextCount}, nil
}

func (c *stepCountingDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	itr, err := c.DB.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	return stepCountingIterator{Iterator: itr, count: &c.nextCount}, nil
}

func (c *stepCountingDB) nexts() int64        { return c.nextCount }
func (c *stepCountingDB) forwardOpens() int64 { return c.forwardCount }
func (c *stepCountingDB) reset()              { c.nextCount, c.forwardCount = 0, 0 }

type stepCountingIterator struct {
	dbm.Iterator
	count *int64
}

func (s stepCountingIterator) Next() {
	*s.count++
	s.Iterator.Next()
}

// scanRootVersions is the full scan of the PrefixRoot range, whose results
// discovery must equal.
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
		forward, reverse, atOpen bool
		want                     error
	}{
		"open":         {forward: true, reverse: true, atOpen: true, want: errIterOpen},
		"read":         {forward: true, reverse: true, atOpen: false, want: errIterRead},
		"reverse open": {reverse: true, atOpen: true, want: errIterOpen},
		"reverse read": {reverse: true, atOpen: false, want: errIterRead},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mem := memdb.NewMemDB()
			writeLayout(t, mem, rootLayout{versions: versionRange(1, 10)})
			db := failingIterDB{DB: mem, forward: tc.forward, reverse: tc.reverse, atOpen: tc.atOpen}

			ndb := newNodeDB(db, 0, NewNopLogger(), DefaultOptions())
			presetCounters := func() {
				ndb.mtx.Lock()
				ndb.firstVersion = 7
				ndb.latestVersion = 9
				ndb.mtx.Unlock()
			}

			presetCounters()
			require.ErrorIs(t, ndb.discoverVersions(), tc.want)
			require.Equal(t, int64(7), ndb.getFirstVersion())
			require.Equal(t, int64(9), ndb.getLatestVersion())

			presetCounters()
			require.ErrorIs(t, ndb.discoverLatestVersion(), tc.want)
			require.Equal(t, int64(9), ndb.getLatestVersion())
		})
	}
}

func TestDiscoverVersionsReadsOnlyTheEnds(t *testing.T) {
	mem := memdb.NewMemDB()
	writeLayout(t, mem, rootLayout{versions: versionRange(1, 10_000)})
	cdb := &stepCountingDB{DB: mem}
	ndb := newNodeDB(cdb, 0, NewNopLogger(), DefaultOptions())
	cdb.reset()

	require.NoError(t, ndb.discoverVersions())
	require.Equal(t, int64(1), ndb.getFirstVersion())
	require.Equal(t, int64(10_000), ndb.getLatestVersion())
	require.LessOrEqual(t, cdb.nexts(), int64(2), "discovery must not walk the root range")
}

func TestDiscoverVersionsSkipsNonRootVersions(t *testing.T) {
	writeKeys := func(t *testing.T, db dbm.DB) {
		t.Helper()
		for _, k := range [][]byte{
			rootDBKey(5),
			rootDBKey(9),
			binary.BigEndian.AppendUint64([]byte{PrefixRoot}, 1<<63+1),
			binary.BigEndian.AppendUint64([]byte{PrefixRoot}, 0),
		} {
			require.NoError(t, db.Set(k, []byte{1}))
		}
	}
	check := func(t *testing.T, db dbm.DB) {
		t.Helper()
		first, latest := discover(t, db)
		require.Equal(t, int64(5), first)
		require.Equal(t, int64(9), latest)
	}
	t.Run("memdb", func(t *testing.T) {
		db := memdb.NewMemDB()
		writeKeys(t, db)
		check(t, db)
	})
	t.Run("pebble/prefix/snapshot", func(t *testing.T) {
		check(t, openPebbleComposition(t, func(db dbm.DB) { writeKeys(t, db) }))
	})
}

func TestLoadReadonlyReadsOnlyTheLatestRoot(t *testing.T) {
	mem := memdb.NewMemDB()
	writer := NewMutableTreeWithDB(mem, 0, NewNopLogger())
	for i := range 3 {
		mustSet(t, writer, []byte("k"), []byte{byte(i)})
		mustSave(t, writer)
	}

	cdb := &stepCountingDB{DB: mem}
	tree := NewMutableTreeWithDB(cdb, 0, NewNopLogger())
	cdb.reset()

	latest, err := tree.LoadReadonly()
	require.NoError(t, err)
	require.Equal(t, int64(3), latest)
	require.Zero(t, cdb.forwardOpens(), "a query view must not seek the low end, where pruning leaves tombstones")
}

func TestLoadKeepsPruningOlderVersions(t *testing.T) {
	mem := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mem, 0, NewNopLogger())
	for i := range 5 {
		mustSet(t, tree, []byte("k"), []byte{byte(i)})
		mustSave(t, tree)
	}
	require.NoError(t, tree.PruneVersionsTo(2))

	tree = NewMutableTreeWithDB(mem, 0, NewNopLogger())
	latest, err := tree.Load()
	require.NoError(t, err)
	require.Equal(t, int64(5), latest)

	mustSet(t, tree, []byte("k"), []byte{5})
	require.Equal(t, int64(6), mustSave(t, tree))
	require.NoError(t, tree.PruneVersionsTo(4))
	require.False(t, tree.VersionExists(3), "version 3 survived a prune to 4 after reload")
	require.False(t, tree.VersionExists(4), "version 4 survived a prune to 4 after reload")
}

func TestSaveVersionAfterLoadReadonlyKeepsPruning(t *testing.T) {
	mem := memdb.NewMemDB()
	writer := NewMutableTreeWithDB(mem, 0, NewNopLogger())
	for i := range 5 {
		mustSet(t, writer, []byte("k"), []byte{byte(i)})
		mustSave(t, writer)
	}

	tree := NewMutableTreeWithDB(mem, 0, NewNopLogger())
	latest, err := tree.LoadReadonly()
	require.NoError(t, err)
	require.Equal(t, int64(5), latest)
	mustSet(t, tree, []byte("k"), []byte{5})
	require.Equal(t, int64(6), mustSave(t, tree))

	require.NoError(t, tree.PruneVersionsTo(5))
	for v := int64(1); v <= 5; v++ {
		require.False(t, tree.VersionExists(v), "version %d survived a prune to 5", v)
	}
	require.True(t, tree.VersionExists(6))
	got, err := tree.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte{5}, got)
}

func TestSaveVersionReturnsFirstVersionSeekError(t *testing.T) {
	mem := memdb.NewMemDB()
	writer := NewMutableTreeWithDB(mem, 0, NewNopLogger())
	mustSet(t, writer, []byte("k"), []byte{0})
	mustSave(t, writer)

	tree := NewMutableTreeWithDB(failingIterDB{DB: mem, forward: true, atOpen: true}, 0, NewNopLogger())
	_, err := tree.LoadReadonly()
	require.NoError(t, err)
	mustSet(t, tree, []byte("k"), []byte{1})
	_, _, err = tree.SaveVersion()
	require.ErrorIs(t, err, errIterOpen)
	require.Equal(t, int64(1), tree.ndb.getLatestVersion())
	require.Zero(t, tree.ndb.getFirstVersion())
	require.False(t, tree.VersionExists(2))
}
