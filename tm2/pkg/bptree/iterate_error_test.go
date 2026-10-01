package bptree

import (
	"errors"
	"strings"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// failingValGetDB forces Get() to fail for value rows (PrefixVal), simulating a
// value-resolver failure mid-iteration. Other operations pass through.
type failingValGetDB struct {
	dbm.DB
	err error
}

func (d *failingValGetDB) Get(key []byte) ([]byte, error) {
	if len(key) > 0 && key[0] == PrefixVal {
		return nil, d.err
	}
	return d.DB.Get(key)
}

// TestIterate_PropagatesResolverError (M6) verifies that a value-resolver
// failure during Iterate stops iteration AND returns the error, rather than
// being swallowed as a normal early stop (which would silently truncate reads).
// Covers both MutableTree.Iterate and ImmutableTree.Iterate.
func TestIterate_PropagatesResolverError(t *testing.T) {
	wrapped := &failingValGetDB{DB: memdb.NewMemDB(), err: errors.New("simulated value Get failure")}
	tree := NewMutableTreeWithDB(wrapped, 100, NewNopLogger())
	for _, k := range []string{"a", "b", "c"} {
		if _, err := tree.Set([]byte(k), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	// Commit so values leave the pendingVals buffer and resolve through the
	// (failing) DB Get on the next read.
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	_, err := tree.Iterate(func(_, _ []byte) bool { return false })
	if err == nil || !strings.Contains(err.Error(), "simulated value Get failure") {
		t.Fatalf("MutableTree.Iterate swallowed the resolver error: %v", err)
	}

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = imm.Iterate(func(_, _ []byte) bool { return false })
	if err == nil || !strings.Contains(err.Error(), "simulated value Get failure") {
		t.Fatalf("ImmutableTree.Iterate swallowed the resolver error: %v", err)
	}
}
