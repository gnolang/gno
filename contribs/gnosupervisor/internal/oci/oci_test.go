package oci_test

import (
	"archive/tar"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci"
	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci/ocitest"
)

var ref = oci.Ref{Host: "ghcr.io", Repository: "gnolang/gno/gnoland"}

func newClient(r *ocitest.Registry) *oci.Client {
	c := oci.NewClient()
	c.BaseURL = r.URL
	c.Backoff = time.Millisecond
	return c
}

func keepAll(string) bool { return true }

// pull resolves digest for linux/amd64 and applies every layer into a new
// directory, as the supervisor does.
func pull(t *testing.T, c *oci.Client, digest string, keep func(string) bool) (string, error) {
	t.Helper()
	ctx := context.Background()
	layers, err := c.ResolveManifest(ctx, ref, digest, "linux/amd64")
	if err != nil {
		return "", err
	}
	root := t.TempDir()
	for _, l := range layers {
		blob, err := c.FetchBlob(ctx, ref, l, t.TempDir())
		if err != nil {
			return "", err
		}
		if err := oci.ApplyLayer(root, blob, l.MediaType, keep); err != nil {
			return "", err
		}
	}
	return root, nil
}

func TestParseRef(t *testing.T) {
	t.Parallel()
	r, err := oci.ParseRef("ghcr.io/gnolang/gno/gnoland:v1.8.0")
	require.NoError(t, err)
	assert.Equal(t, ref, r)
	r, err = oci.ParseRef("localhost:5000/gnoland@sha256:" + strings.Repeat("a", 64))
	require.NoError(t, err)
	assert.Equal(t, oci.Ref{Host: "localhost:5000", Repository: "gnoland"}, r)
	_, err = oci.ParseRef("gnoland:v1.8.0")
	assert.Error(t, err)
}

func TestPullLayersInOrder(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	digest := reg.PushImage(t, []string{"linux/amd64", "linux/arm64"},
		ocitest.Layer(t,
			ocitest.File{Name: "usr/bin/gnoland", Body: "old", Mode: 0o755},
			ocitest.File{Name: "gnoroot/gnovm/stdlibs/std/std.gno", Body: "package std"},
			ocitest.File{Name: "gnoroot/examples/gone.gno", Body: "x"},
			ocitest.File{Name: "gnoroot/tmp/a", Body: "a"},
			ocitest.File{Name: "etc/passwd", Body: "not kept"},
		),
		ocitest.Layer(t,
			ocitest.File{Name: "./usr/bin/gnoland", Body: "new", Mode: 0o755},
			ocitest.File{Name: "gnoroot/examples/.wh.gone.gno"},
			ocitest.File{Name: "gnoroot/tmp/.wh..wh..opq"},
			ocitest.File{Name: "gnoroot/hard.gno", Type: tar.TypeLink, Linkname: "gnoroot/gnovm/stdlibs/std/std.gno"},
		),
	)
	root, err := pull(t, newClient(reg), digest, func(n string) bool {
		return n == "usr/bin/gnoland" || strings.HasPrefix(n, "gnoroot/")
	})
	require.NoError(t, err)

	read := func(p string) string {
		b, err := os.ReadFile(filepath.Join(root, p))
		require.NoError(t, err)
		return string(b)
	}
	assert.Equal(t, "new", read("usr/bin/gnoland"))
	assert.Equal(t, "package std", read("gnoroot/gnovm/stdlibs/std/std.gno"))
	assert.Equal(t, "package std", read("gnoroot/hard.gno"))
	assert.NoFileExists(t, filepath.Join(root, "gnoroot/examples/gone.gno"), "whiteout")
	assert.NoFileExists(t, filepath.Join(root, "gnoroot/tmp/a"), "opaque whiteout")
	assert.NoFileExists(t, filepath.Join(root, "etc/passwd"), "filtered out")
}

func TestPullUnknownPlatform(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	digest := reg.PushImage(t, []string{"linux/arm64"}, ocitest.Layer(t, ocitest.File{Name: "a", Body: "a"}))
	_, err := pull(t, newClient(reg), digest, keepAll)
	assert.ErrorContains(t, err, "no linux/amd64 manifest")
}

func TestPullBearerToken(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	reg.Token = "secret"
	digest := reg.PushImage(t, []string{"linux/amd64"}, ocitest.Layer(t, ocitest.File{Name: "a", Body: "a"}))
	root, err := pull(t, newClient(reg), digest, keepAll)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(root, "a"))
}

func TestPullRefusesWhatItDidNotPin(t *testing.T) {
	t.Parallel()
	for name, tamper := range map[string]func(path string, attempt int, body []byte) []byte{
		"index altered": func(p string, _ int, b []byte) []byte {
			if strings.Contains(p, "/manifests/") && strings.Contains(string(b), "index") {
				return append(b, ' ')
			}
			return b
		},
		"layer altered": func(p string, _ int, b []byte) []byte {
			if strings.Contains(p, "/blobs/") {
				c := append([]byte(nil), b...)
				c[len(c)/2] ^= 0xff
				return c
			}
			return b
		},
		"layer truncated every time": func(p string, _ int, b []byte) []byte {
			if strings.Contains(p, "/blobs/") {
				return b[:len(b)/2]
			}
			return b
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reg := ocitest.New(t)
			digest := reg.PushImage(t, []string{"linux/amd64"}, ocitest.Layer(t, ocitest.File{Name: "a", Body: strings.Repeat("a", 4096)}))
			reg.Tamper = tamper
			_, err := pull(t, newClient(reg), digest, keepAll)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "after 4 attempts")
		})
	}
}

// An interrupted download is retried, and the retry succeeds.
func TestPullRetriesInterruptedDownload(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	digest := reg.PushImage(t, []string{"linux/amd64"}, ocitest.Layer(t, ocitest.File{Name: "a", Body: strings.Repeat("a", 4096)}))
	reg.Tamper = func(p string, attempt int, b []byte) []byte {
		if strings.Contains(p, "/blobs/") && attempt == 1 {
			return b[:len(b)/3]
		}
		return b
	}
	root, err := pull(t, newClient(reg), digest, keepAll)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(root, "a"))
}

func TestApplyLayerRefusesEscapes(t *testing.T) {
	t.Parallel()
	for name, files := range map[string][]ocitest.File{
		"dot-dot":           {{Name: "../evil", Body: "x"}},
		"any symlink":       {{Name: "l", Type: tar.TypeSymlink, Linkname: "x"}},
		"symlink chain":     {{Name: "s", Type: tar.TypeSymlink, Linkname: "."}, {Name: "e", Type: tar.TypeSymlink, Linkname: "s/s/../.."}},
		"device":            {{Name: "dev", Type: tar.TypeChar}},
		"dangling hardlink": {{Name: "h", Type: tar.TypeLink, Linkname: "missing"}},
		"hardlink escaping": {{Name: "h", Type: tar.TypeLink, Linkname: "../outside"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			layer := filepath.Join(t.TempDir(), "layer")
			require.NoError(t, os.WriteFile(layer, ocitest.Layer(t, files...), 0o644))
			parent := t.TempDir()
			root := filepath.Join(parent, "root")
			require.NoError(t, os.Mkdir(root, 0o755))
			err := oci.ApplyLayer(root, layer, oci.MediaTypeOCILayerGzip, keepAll)
			require.Error(t, err)
			entries, _ := os.ReadDir(parent)
			assert.Len(t, entries, 1, "nothing written next to the root")
		})
	}
}

// A later layer that deletes a whole kept directory, or the directory above
// the binary, hides what lower layers put there.
func TestApplyLayerDirectoryWhiteouts(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool {
		return n == "usr/" || n == "usr/bin/" || n == "usr/bin/gnoland" || strings.HasPrefix(n, "gnoroot/")
	}
	root := t.TempDir()
	apply := func(files ...ocitest.File) {
		layer := filepath.Join(t.TempDir(), "layer")
		require.NoError(t, os.WriteFile(layer, ocitest.Layer(t, files...), 0o644))
		require.NoError(t, oci.ApplyLayer(root, layer, oci.MediaTypeOCILayerGzip, keep))
	}
	apply(ocitest.File{Name: "usr/bin/gnoland", Body: "x"}, ocitest.File{Name: "gnoroot/old.gno", Body: "x"})
	apply(ocitest.File{Name: ".wh.gnoroot"}, ocitest.File{Name: "usr/.wh.bin"}, ocitest.File{Name: "gnoroot/new.gno", Body: "y"})
	assert.NoFileExists(t, filepath.Join(root, "gnoroot/old.gno"))
	assert.FileExists(t, filepath.Join(root, "gnoroot/new.gno"))
	assert.NoFileExists(t, filepath.Join(root, "usr/bin/gnoland"))
}

// The supervisor reads the manifest's type from the body the digest covers,
// never from the Content-Type header.
func TestManifestTypeIgnoresTheHeader(t *testing.T) {
	t.Parallel()
	reg := ocitest.New(t)
	digest := reg.PushImage(t, []string{"linux/amd64"}, ocitest.Layer(t, ocitest.File{Name: "a", Body: "a"}))
	reg.ContentType = "application/vnd.oci.image.manifest.v1+json"
	root, err := pull(t, newClient(reg), digest, keepAll)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(root, "a"))
}

func TestApplyLayerUnsupportedMediaType(t *testing.T) {
	t.Parallel()
	layer := filepath.Join(t.TempDir(), "layer")
	require.NoError(t, os.WriteFile(layer, []byte("x"), 0o644))
	err := oci.ApplyLayer(t.TempDir(), layer, "application/vnd.oci.image.layer.v1.tar+zstd", keepAll)
	assert.ErrorContains(t, err, "unsupported layer media type")
}
