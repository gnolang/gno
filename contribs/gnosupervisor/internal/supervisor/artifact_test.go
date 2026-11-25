package supervisor

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci"
	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci/ocitest"
	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

// imageLeg pushes a gnoland-shaped image and returns a leg pinned to it.
func imageLeg(t *testing.T, reg *ocitest.Registry, version string, files ...ocitest.File) Leg {
	t.Helper()
	if files == nil {
		files = []ocitest.File{
			{Name: "usr/bin/gnoland", Body: "#!/bin/sh\necho " + version + "\n", Mode: 0o755},
			{Name: "gnoroot/gnovm/stdlibs/std/std.gno", Body: "package std"},
			{Name: "etc/os-release", Body: "alpine"},
		}
	}
	digest := reg.PushImage(t, []string{"linux/amd64"}, ocitest.Layer(t, files...))
	return Leg{
		Range:   upgrades.Range{Version: version, From: 1},
		Version: version,
		Entry:   upgrades.Entry{Version: version, Image: upgrades.Image{Ref: "ghcr.io/gnolang/gno/gnoland:" + version, Digest: &digest}},
	}
}

func newStore(t *testing.T, reg *ocitest.Registry) *Store {
	t.Helper()
	c := oci.NewClient()
	c.BaseURL = reg.URL
	c.Backoff = time.Millisecond
	return &Store{Dir: t.TempDir(), Platform: "linux/amd64", Client: c, Log: discard}
}

func TestPrepareCachesAndVerifies(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	leg := imageLeg(t, reg, "v1.0.0")
	s := newStore(t, reg)
	ctx := context.Background()

	art, err := s.Prepare(ctx, leg)
	require.NoError(t, err)
	b, err := os.ReadFile(art.Binary)
	require.NoError(t, err)
	assert.Contains(t, string(b), "v1.0.0")
	fi, err := os.Stat(art.Binary)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), fi.Mode().Perm())
	assert.DirExists(t, filepath.Join(art.GnoRoot, "gnovm", "stdlibs"))
	assert.NoFileExists(t, filepath.Join(filepath.Dir(art.GnoRoot), "etc", "os-release"))

	// A verified cache is used as is.
	served := reg.Total.Load()
	again, err := s.Prepare(ctx, leg)
	require.NoError(t, err)
	assert.Equal(t, art, again)
	assert.Equal(t, served, reg.Total.Load(), "no request for a cached artifact")

	// A damaged cache is pulled again, never run.
	require.NoError(t, os.WriteFile(art.Binary, []byte("tampered"), 0o755))
	again, err = s.Prepare(ctx, leg)
	require.NoError(t, err)
	b, err = os.ReadFile(again.Binary)
	require.NoError(t, err)
	assert.Contains(t, string(b), "v1.0.0")
	assert.Greater(t, reg.Total.Load(), served)
}

func TestPrepareLeavesNothingOnFailure(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	leg := imageLeg(t, reg, "v1.0.0")
	reg.Tamper = func(p string, _ int, b []byte) []byte {
		if strings.Contains(p, "/blobs/") {
			return b[:len(b)/2] // the download never completes
		}
		return b
	}
	s := newStore(t, reg)
	_, err := s.Prepare(context.Background(), leg)
	require.Error(t, err)
	entries, err := os.ReadDir(s.Dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.True(t, strings.HasSuffix(e.Name(), ".lock"),
			"an interrupted pull leaves no artifact and no temporary directory, found %s", e.Name())
	}
}

func TestPrepareRefusesASharedCacheDir(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	leg := imageLeg(t, reg, "v1.0.0")
	s := newStore(t, reg)
	require.NoError(t, os.Chmod(s.Dir, 0o777))
	_, err := s.Prepare(context.Background(), leg)
	assert.ErrorContains(t, err, "writable by group or others")
	assert.Zero(t, reg.Total.Load(), "nothing pulled")
}

// Two supervisors sharing a cache pull an artifact once and both use it.
func TestPrepareConcurrently(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	leg := imageLeg(t, reg, "v1.0.0")
	s := newStore(t, reg)
	other := *s
	errs := make(chan error, 2)
	for _, st := range []*Store{s, &other} {
		go func() {
			_, err := st.Prepare(context.Background(), leg)
			errs <- err
		}()
	}
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	served := reg.Total.Load()
	_, err := s.Prepare(context.Background(), leg)
	require.NoError(t, err)
	assert.Equal(t, served, reg.Total.Load())
}

func TestPrepareRefusesImageWithoutGnoland(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	leg := imageLeg(t, reg, "v1.0.0", ocitest.File{Name: "gnoroot/gnovm/stdlibs/x", Body: "x"})
	s := newStore(t, reg)
	_, err := s.Prepare(context.Background(), leg)
	assert.ErrorContains(t, err, "no regular file /usr/bin/gnoland")

	leg = imageLeg(t, reg, "v1.0.1", ocitest.File{Name: "usr/bin/gnoland", Body: "x"})
	_, err = s.Prepare(context.Background(), leg)
	assert.ErrorContains(t, err, "no /gnoroot/gnovm/stdlibs")
}
