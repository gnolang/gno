package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/commands"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBackend = "goleveldb"

// openMainStore opens the "gnolang" DB under <dataDir>/db and returns the main
// store's bptree tree (prefix "s/_/") with the fast index enabled.
func openMainStore(t *testing.T, dataDir string, fastIndex bool) (*bptree.MutableTree, dbm.DB) {
	t.Helper()
	raw, err := dbm.NewDB("gnolang", testBackend, filepath.Join(dataDir, config.DefaultDBDir))
	require.NoError(t, err)
	opts := []bptree.Option{}
	if fastIndex {
		opts = append(opts, bptree.FastIndexOption(true))
	}
	tree := bptree.NewMutableTreeWithDB(
		dbm.NewPrefixDB(raw, []byte(mainStorePrefix)), 1000, bptree.NewNopLogger(), opts...)
	return tree, raw
}

func runVerify(t *testing.T, dataDir string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&out))
	cfg := &fastindexVerifyCfg{dataDir: dataDir, dbBackend: testBackend}
	err := execFastindexVerify(cfg, io)
	return out.String(), err
}

// TestFastindexVerify_Healthy: a normally-maintained store passes with exit 0.
func TestFastindexVerify_Healthy(t *testing.T) {
	dir := t.TempDir()

	tree, raw := openMainStore(t, dir, true)
	if _, err := tree.Set([]byte("a"), []byte("va")); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.Set([]byte("b"), []byte("vb")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	require.NoError(t, raw.Close()) // release the lock before the command opens it

	out, err := runVerify(t, dir)
	require.NoError(t, err)
	assert.Contains(t, out, "OK: fast index is current")
}

// TestFastindexVerify_StampBehind: an index left behind (versions committed with
// the feature off) is a WARN, not a failure — the node rebuilds it on start.
func TestFastindexVerify_StampBehind(t *testing.T) {
	dir := t.TempDir()

	on, raw := openMainStore(t, dir, true)
	if _, err := on.Set([]byte("k"), []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := on.SaveVersion(); err != nil { // stamp=v1, 'F'[k]=v1
		t.Fatal(err)
	}
	require.NoError(t, raw.Close())

	// Reopen with the feature OFF and advance the tree: 'F'[k] and the stamp stay
	// at v1 while the authoritative tree moves to v2 — index behind + stale entry.
	off, raw2 := openMainStore(t, dir, false)
	if _, err := off.Load(); err != nil {
		t.Fatal(err)
	}
	if _, err := off.Set([]byte("k"), []byte("v2")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := off.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	require.NoError(t, raw2.Close())

	out, err := runVerify(t, dir)
	require.NoError(t, err) // behind is benign
	assert.Contains(t, out, "behind the latest version")
}

// TestFastindexVerify_MissingDB: a missing DB directory is a clear error.
func TestFastindexVerify_MissingDB(t *testing.T) {
	_, err := runVerify(t, t.TempDir()) // empty dir: no <dir>/db
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"), "got: %v", err)
}
