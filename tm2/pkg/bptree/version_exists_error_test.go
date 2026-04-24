package bptree

import (
	"errors"
	"strings"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// failingHasDB wraps a DB and forces Has() to return an error for all keys
// whose first byte matches hasFailPrefix. Set Get/Set/Delete pass through.
type failingHasDB struct {
	dbm.DB
	hasFailPrefix byte
	hasFailErr    error
}

func (d *failingHasDB) Has(key []byte) (bool, error) {
	if len(key) > 0 && key[0] == d.hasFailPrefix {
		return false, d.hasFailErr
	}
	return d.DB.Has(key)
}

// TestSaveVersion_PropagatesVersionExistsError verifies that a DB error in
// versionExistsE during SaveVersion is propagated to the caller. Before the
// fix, SaveVersion called VersionExists which swallowed the error and
// returned false — causing SaveVersion to proceed into the "fresh save"
// branch and overwrite an existing version with new data.
func TestSaveVersion_PropagatesVersionExistsError(t *testing.T) {
	inner := memdb.NewMemDB()
	wrapped := &failingHasDB{
		DB:            inner,
		hasFailPrefix: PrefixRoot,
		hasFailErr:    errors.New("simulated root Has failure"),
	}
	tree := NewMutableTreeWithDB(wrapped, 100, NewNopLogger())

	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, _, err := tree.SaveVersion()
	if err == nil {
		t.Fatalf("SaveVersion should have propagated the DB error")
	}
	if !strings.Contains(err.Error(), "simulated root Has failure") {
		t.Fatalf("error does not wrap underlying cause: %v", err)
	}
}
