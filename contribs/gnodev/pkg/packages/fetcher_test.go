package packages

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteFetcher_DisabledWithoutRemotes(t *testing.T) {
	f := newRemoteFetcher(nil)
	_, err := f.FetchPackage("gno.land/r/demo/counter")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "-remote")
}

func TestRemoteFetcher_GatesByDomain(t *testing.T) {
	inner := pkgdownload.NewInMemoryFetcher(&std.MemPackage{
		Path:  "gno.land/p/demo/foo",
		Files: []*std.MemFile{{Name: "foo.gno", Body: "package foo\n"}},
	})
	f := &domainFetcher{remotes: map[string]string{"gno.land": "unused"}, inner: inner}

	files, err := f.FetchPackage("gno.land/p/demo/foo")
	require.NoError(t, err)
	require.Len(t, files, 1)

	_, err = f.FetchPackage("other.land/p/demo/foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no -remote configured for domain "other.land"`)
}

func TestLoader_Resolve_RemoteDisabledByDefault(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	l := New(Config{Logger: logger})
	_, err := l.Resolve("gno.land/r/not/onfs")
	require.ErrorIs(t, err, ErrPackageNotFound)
	assert.Contains(t, buf.String(), "remote fetching is disabled")
}
