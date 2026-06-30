package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	mock "github.com/gnolang/gno/contribs/gnodev/internal/mock/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	debounceInterval = 50 * time.Millisecond
	os.Exit(m.Run())
}

func setupTestingWatcher(t *testing.T) (*PackageWatcher, string) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.gno"),
		[]byte("package foo\n"), 0o644))

	w, err := NewPackageWatcher(log.NewTestingLogger(t), &mock.ServerEmitter{})
	require.NoError(t, err)
	t.Cleanup(w.Stop)

	w.UpdatePackagesWatch(&packages.Package{
		ImportPath: "gno.land/p/test/foo",
		Dir:        dir,
		Kind:       packages.KindFS,
	})
	return w, dir
}

func waitPackagesUpdate(t *testing.T, w *PackageWatcher) PackageUpdateList {
	t.Helper()

	select {
	case up := <-w.PackagesUpdate:
		return up
	case err := <-w.Errors:
		t.Fatalf("watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("no package update received")
	}
	return nil
}

func TestWatcher_InPlaceWrite(t *testing.T) {
	w, dir := setupTestingWatcher(t)

	f, err := os.OpenFile(filepath.Join(dir, "pkg.gno"), os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	_, err = f.WriteString("// edit\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	up := waitPackagesUpdate(t, w)
	require.NotEmpty(t, up)
}

// TestWatcher_AtomicRenameSave covers the save strategy of sed -i and most
// editors: write a temp file, then rename it over the watched file. The
// rename produces Create/Rename events, not Write.
func TestWatcher_AtomicRenameSave(t *testing.T) {
	w, dir := setupTestingWatcher(t)

	tmp := filepath.Join(dir, ".pkg.gno.tmp")
	require.NoError(t, os.WriteFile(tmp, []byte("package foo\n// edited\n"), 0o644))
	require.NoError(t, os.Rename(tmp, filepath.Join(dir, "pkg.gno")))

	up := waitPackagesUpdate(t, w)
	require.NotEmpty(t, up)
}

// TestWatcher_UpdateAttribution pins file-to-package attribution: an event
// belongs to the package whose directory directly contains the file — not
// to a sibling package whose path merely shares a string prefix.
func TestWatcher_UpdateAttribution(t *testing.T) {
	w, err := NewPackageWatcher(log.NewTestingLogger(t), &mock.ServerEmitter{})
	require.NoError(t, err)
	t.Cleanup(w.Stop)

	root := t.TempDir()
	short := filepath.Join(root, "ab") // package "ab"
	long := filepath.Join(root, "abc") // sibling sharing "ab" as string prefix
	for _, dir := range []string{short, long} {
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg.gno"),
			[]byte("package foo\n"), 0o644))
	}
	w.UpdatePackagesWatch(
		&packages.Package{ImportPath: "gno.land/p/test/ab", Dir: short, Kind: packages.KindFS},
		&packages.Package{ImportPath: "gno.land/p/test/abc", Dir: long, Kind: packages.KindFS},
	)

	changed := filepath.Join(short, "pkg.gno")
	up := w.generatePackagesUpdateList([]string{changed})

	require.Len(t, up, 1, "file in ab/ must not be attributed to sibling abc/")
	require.Equal(t, short, up[0].PackageDir)
	require.Equal(t, []string{changed}, up[0].Files)
}

func TestWatcher_FileRemove(t *testing.T) {
	w, dir := setupTestingWatcher(t)

	extra := filepath.Join(dir, "extra.gno")
	require.NoError(t, os.WriteFile(extra, []byte("package foo\n"), 0o644))
	_ = waitPackagesUpdate(t, w) // consume the create event

	require.NoError(t, os.Remove(extra))

	up := waitPackagesUpdate(t, w)
	require.NotEmpty(t, up)
}
